package tripo

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	neturl "net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// Artifact is one Task Output URL saved as a local file.
type Artifact struct {
	// Name is the Output field without its "_url" suffix, e.g. "model",
	// "rendered_image", "generated_image" or "front_view".
	Name string `json:"name"`
	// Path is the absolute local path: <dir>/<task_id>/<name>.<format>.
	Path string `json:"path"`
	// Format is the file extension, e.g. "glb", "fbx", "splat", "png".
	Format string `json:"format"`
	Bytes  int64  `json:"bytes"`
}

// Download saves the Output URLs of a successful Task into dir/<task_id>/,
// one file per Output field, overwriting earlier downloads of the same Task.
// names limits which Outputs are saved; empty means all of them. The Task is
// fetched first because its URLs expire about five minutes after issue.
func (c *Client) Download(ctx context.Context, taskID, dir string, names ...string) (*Task, []Artifact, error) {
	task, err := c.GetTask(ctx, taskID)
	if err != nil {
		return nil, nil, err
	}
	if task.Status != StatusSuccess {
		msg := fmt.Sprintf("tripo: task %s is %s, not success", taskID, task.Status)
		if task.ErrorMessage != "" {
			msg += ": " + task.ErrorMessage
		}
		return task, nil, fmt.Errorf("%s", msg)
	}

	urls := task.OutputURLs()
	if len(names) > 0 {
		wanted := make(map[string]string, len(names))
		for _, name := range names {
			u, ok := urls[name]
			if !ok {
				return task, nil, fmt.Errorf("tripo: task %s has no %q output (available: %s)", taskID, name, strings.Join(sortedKeys(urls), ", "))
			}
			wanted[name] = u
		}
		urls = wanted
	}

	absDir, err := filepath.Abs(filepath.Join(dir, taskID))
	if err != nil {
		return task, nil, fmt.Errorf("tripo: resolving output directory: %w", err)
	}
	if len(urls) > 0 {
		if err := os.MkdirAll(absDir, 0o755); err != nil {
			return task, nil, fmt.Errorf("tripo: creating output directory: %w", err)
		}
	}

	artifacts := make([]Artifact, 0, len(urls))
	for _, name := range sortedKeys(urls) {
		artifact, err := c.downloadArtifact(ctx, urls[name], absDir, name)
		if err != nil {
			return task, artifacts, fmt.Errorf("tripo: downloading %s of task %s: %w", name, taskID, err)
		}
		artifacts = append(artifacts, artifact)
	}
	return task, artifacts, nil
}

func (c *Client) downloadArtifact(ctx context.Context, rawURL, dir, name string) (Artifact, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return Artifact{}, err
	}
	resp, err := c.transfer.Do(req)
	if err != nil {
		return Artifact{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return Artifact{}, fmt.Errorf("HTTP %d (signed URLs expire after about five minutes; download again)", resp.StatusCode)
	}

	tmp, err := os.CreateTemp(dir, "."+name+"-*.part")
	if err != nil {
		return Artifact{}, err
	}
	n, copyErr := io.Copy(tmp, resp.Body)
	if copyErr == nil {
		// CreateTemp makes 0600 files; Artifacts are meant to be opened by
		// other tools and users.
		copyErr = tmp.Chmod(0o644)
	}
	closeErr := tmp.Close()
	if copyErr != nil || closeErr != nil {
		_ = os.Remove(tmp.Name())
		if copyErr != nil {
			return Artifact{}, copyErr
		}
		return Artifact{}, closeErr
	}

	format := artifactFormat(rawURL, resp.Header)
	dest := filepath.Join(dir, name+"."+format)
	if err := os.Rename(tmp.Name(), dest); err != nil {
		_ = os.Remove(tmp.Name())
		return Artifact{}, err
	}
	return Artifact{Name: name, Path: dest, Format: format, Bytes: n}, nil
}

// knownFormats are the extensions Tripo outputs carry.
var knownFormats = map[string]bool{
	"glb": true, "gltf": true, "fbx": true, "obj": true, "stl": true, "usdz": true,
	"3mf": true, "splat": true, "ply": true, "zip": true,
	"png": true, "jpg": true, "jpeg": true, "webp": true, "gif": true, "mp4": true,
}

var contentTypeFormats = map[string]string{
	"model/gltf-binary":   "glb",
	"model/gltf+json":     "gltf",
	"application/x-fbx":   "fbx",
	"model/fbx":           "fbx",
	"model/obj":           "obj",
	"model/stl":           "stl",
	"model/vnd.usdz+zip":  "usdz",
	"model/3mf":           "3mf",
	"application/zip":     "zip",
	"image/png":           "png",
	"image/jpeg":          "jpg",
	"image/webp":          "webp",
	"image/gif":           "gif",
	"video/mp4":           "mp4",
	"application/x-splat": "splat",
}

// artifactFormat infers a file extension from the URL path, then the
// Content-Disposition filename, then the Content-Type.
func artifactFormat(rawURL string, header http.Header) string {
	if u, err := neturl.Parse(rawURL); err == nil {
		if ext := extOf(u.Path); knownFormats[ext] {
			return ext
		}
	}
	if _, params, err := mime.ParseMediaType(header.Get("Content-Disposition")); err == nil {
		if ext := extOf(params["filename"]); knownFormats[ext] {
			return ext
		}
	}
	if mediaType, _, err := mime.ParseMediaType(header.Get("Content-Type")); err == nil {
		if ext, ok := contentTypeFormats[mediaType]; ok {
			return ext
		}
	}
	return "bin"
}

func extOf(p string) string {
	return strings.ToLower(strings.TrimPrefix(path.Ext(p), "."))
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
