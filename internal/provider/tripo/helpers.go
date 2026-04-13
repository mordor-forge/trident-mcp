package tripo

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"mime"
	"net/http"
	neturl "net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	nowUTC               = func() time.Time { return time.Now().UTC() }
	randSource io.Reader = rand.Reader
)

// generateFilename produces a unique filename for a downloaded model.
// Format: model-{UTC timestamp}-{random hex}.{ext}
func generateFilename(ext string) string {
	ts := nowUTC().Format("2006-01-02T15-04-05")
	id := shortID()
	return fmt.Sprintf("model-%s-%s.%s", ts, id, ext)
}

// shortID returns a 16-character random hex string for filename uniqueness.
func shortID() string {
	b := make([]byte, 8)
	if _, err := io.ReadFull(randSource, b); err != nil {
		return fmt.Sprintf("%016x", nowUTC().UnixNano())
	}
	return fmt.Sprintf("%x", b)
}

// formatToExt maps Tripo output format names to file extensions.
func formatToExt(format string) string {
	canonical, ok := normalizeFormat(format)
	if !ok {
		return "glb"
	}

	switch canonical {
	case "GLTF":
		return "glb"
	case "FBX":
		return "fbx"
	case "OBJ":
		return "obj"
	case "STL":
		return "stl"
	case "USDZ":
		return "usdz"
	case "3MF":
		return "3mf"
	default:
		return "glb"
	}
}

// fileTypeFromPath returns the Tripo API file type for image uploads.
func fileTypeFromPath(path string) string {
	if fileType, ok := imageTypeFromExt(filepath.Ext(path)); ok {
		return fileType
	}
	return "jpg"
}

func fileTypeFromURL(rawURL string) string {
	parsed, err := neturl.Parse(rawURL)
	if err != nil {
		return "jpg"
	}
	if fileType, ok := imageTypeFromExt(filepath.Ext(parsed.Path)); ok {
		return fileType
	}
	return "jpg"
}

func imageTypeFromExt(ext string) (string, bool) {
	switch strings.ToLower(ext) {
	case ".png":
		return "png", true
	case ".jpg", ".jpeg":
		return "jpg", true
	case ".webp":
		return "webp", true
	default:
		return "", false
	}
}

func formatFromExt(ext string) (string, bool) {
	switch strings.ToLower(ext) {
	case ".glb", ".gltf":
		return "GLTF", true
	case ".fbx":
		return "FBX", true
	case ".obj":
		return "OBJ", true
	case ".stl":
		return "STL", true
	case ".usdz":
		return "USDZ", true
	case ".3mf":
		return "3MF", true
	default:
		return "", false
	}
}

func formatFromURL(rawURL string) (string, bool) {
	parsed, err := neturl.Parse(rawURL)
	if err != nil {
		return "", false
	}
	if format, ok := formatFromExt(filepath.Ext(parsed.Path)); ok {
		return format, true
	}
	if filename := parsed.Query().Get("filename"); filename != "" {
		return formatFromExt(filepath.Ext(filename))
	}
	return "", false
}

func formatFromHeaders(header http.Header) (string, bool) {
	if filename := filenameFromContentDisposition(header.Get("Content-Disposition")); filename != "" {
		if format, ok := formatFromExt(filepath.Ext(filename)); ok {
			return format, true
		}
	}
	return formatFromContentType(header.Get("Content-Type"))
}

func filenameFromContentDisposition(value string) string {
	if value == "" {
		return ""
	}

	_, params, err := mime.ParseMediaType(value)
	if err != nil {
		return ""
	}
	if filename := params["filename"]; filename != "" {
		return filename
	}
	if filename := params["filename*"]; filename != "" {
		parts := strings.SplitN(filename, "''", 2)
		if len(parts) == 2 {
			decoded, err := neturl.QueryUnescape(parts[1])
			if err == nil {
				return decoded
			}
			return parts[1]
		}
		return filename
	}
	return ""
}

func formatFromContentType(value string) (string, bool) {
	if value == "" {
		return "", false
	}

	mediaType, _, err := mime.ParseMediaType(value)
	if err != nil {
		mediaType = strings.ToLower(strings.TrimSpace(strings.Split(value, ";")[0]))
	}

	switch mediaType {
	case "model/gltf-binary", "model/gltf+json":
		return "GLTF", true
	case "application/x-fbx", "model/fbx":
		return "FBX", true
	case "model/obj":
		return "OBJ", true
	case "model/stl", "application/sla", "application/vnd.ms-pki.stl":
		return "STL", true
	case "model/vnd.usdz+zip":
		return "USDZ", true
	case "application/vnd.ms-package.3dmanufacturing-3dmodel+xml", "model/3mf", "application/vnd.3mfdocument":
		return "3MF", true
	default:
		return "", false
	}
}

// downloadFile downloads a URL to a local file path using the provided client.
// It returns the response headers so callers can infer metadata such as format.
func downloadFile(ctx context.Context, client *http.Client, url, destPath string) (http.Header, error) {
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating download request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("downloading %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download returned status %d", resp.StatusCode)
	}
	headers := resp.Header.Clone()

	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return nil, fmt.Errorf("creating directory: %w", err)
	}

	f, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("creating file: %w", err)
	}

	if _, err := io.Copy(f, resp.Body); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("writing file: %w", err)
	}

	if err := f.Close(); err != nil {
		return nil, fmt.Errorf("closing file: %w", err)
	}

	return headers, nil
}
