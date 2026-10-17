package tripo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// Input references something an Operation consumes: an absolute local file
// path, a public http(s) URL, a Tripo file token, or a task ID. Local files
// are uploaded and replaced by their file token before the Operation is sent;
// everything else is passed to Tripo, which tells the remaining forms apart.
type Input string

const (
	maxImageUpload  = 20 << 20
	maxModelUpload  = 150 << 20
	presignAbove    = 50 << 20
	uploadKindImage = "image"
	uploadKindModel = "model"
)

// uploadKinds lists the file types Tripo accepts. Restricting local uploads to
// these keeps an Input from exfiltrating arbitrary files.
var uploadKinds = map[string]string{
	"png": uploadKindImage, "jpg": uploadKindImage, "jpeg": uploadKindImage, "webp": uploadKindImage,
	"glb": uploadKindModel, "gltf": uploadKindModel, "fbx": uploadKindModel, "obj": uploadKindModel,
	"stl": uploadKindModel, "3mf": uploadKindModel, "usdz": uploadKindModel,
}

func isHTTPURL(s string) bool {
	return strings.HasPrefix(s, "https://") || strings.HasPrefix(s, "http://")
}

// localPath reports whether an Input names a local file, and its expanded
// path. Task IDs and file tokens never contain path separators or a file
// extension, so anything that does is treated as a path.
func localPath(in string) (string, bool) {
	if in == "" || isHTTPURL(in) {
		return "", false
	}
	looksLikePath := strings.ContainsAny(in, `/\`) || strings.HasPrefix(in, "~") || uploadKinds[extOf(in)] != ""
	if !looksLikePath {
		return "", false
	}
	if rest, ok := strings.CutPrefix(in, "~"); ok {
		if home, err := os.UserHomeDir(); err == nil {
			in = filepath.Join(home, rest)
		}
	}
	return in, true
}

// resolveInputs uploads every local-file Input reachable from v (a pointer to
// a request) and replaces it with the resulting file token. The same path is
// uploaded once per call.
func (c *Client) resolveInputs(ctx context.Context, v any) error {
	uploaded := map[string]Input{}
	var walk func(reflect.Value) error
	walk = func(v reflect.Value) error {
		switch v.Kind() {
		case reflect.Pointer, reflect.Interface:
			if v.IsNil() {
				return nil
			}
			return walk(v.Elem())
		case reflect.Struct:
			for i := range v.NumField() {
				if v.Type().Field(i).IsExported() {
					if err := walk(v.Field(i)); err != nil {
						return err
					}
				}
			}
		case reflect.Slice, reflect.Array:
			for i := range v.Len() {
				if err := walk(v.Index(i)); err != nil {
					return err
				}
			}
		case reflect.String:
			if v.Type() != reflect.TypeFor[Input]() || !v.CanSet() {
				return nil
			}
			path, ok := localPath(strings.TrimSpace(v.String()))
			if !ok {
				return nil
			}
			token, done := uploaded[path]
			if !done {
				up, err := c.UploadFile(ctx, path)
				if err != nil {
					return err
				}
				token = Input(up.FileToken)
				uploaded[path] = token
			}
			v.SetString(string(token))
		}
		return nil
	}
	return walk(reflect.ValueOf(v))
}

// Upload is a file stored by Tripo.
type Upload struct {
	FileToken string `json:"file_token"`
	// Path is the local file that was uploaded.
	Path  string `json:"path"`
	Bytes int64  `json:"bytes"`
}

// UploadFile uploads a local image (PNG, JPEG, WebP; up to 20 MB) or model
// (GLB, GLTF, FBX, OBJ, STL, 3MF, USDZ; up to 150 MB) and returns its file
// token. Files over 50 MB go through a presigned upload.
func (c *Client) UploadFile(ctx context.Context, path string) (*Upload, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("tripo: resolving %s: %w", path, err)
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("tripo: input file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("tripo: input %s is not a regular file", path)
	}
	ext := extOf(path)
	kind := uploadKinds[ext]
	switch {
	case kind == "":
		return nil, fmt.Errorf("tripo: cannot upload %s: only images (png, jpg, webp) and models (glb, gltf, fbx, obj, stl, 3mf, usdz) are accepted", path)
	case kind == uploadKindImage && info.Size() > maxImageUpload:
		return nil, fmt.Errorf("tripo: image %s is %d MB; Tripo accepts images up to 20 MB", path, info.Size()>>20)
	case kind == uploadKindModel && info.Size() > maxModelUpload:
		return nil, fmt.Errorf("tripo: model %s is %d MB; Tripo accepts models up to 150 MB", path, info.Size()>>20)
	}

	var token string
	if info.Size() > presignAbove {
		token, err = c.presignedUpload(ctx, path, ext)
	} else {
		token, err = c.multipartUpload(ctx, path)
	}
	if err != nil {
		return nil, err
	}
	return &Upload{FileToken: token, Path: path, Bytes: info.Size()}, nil
}

func (c *Client) multipartUpload(ctx context.Context, path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("tripo: reading %s: %w", path, err)
	}
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	part, err := form.CreateFormFile("file", filepath.Base(path))
	if err == nil {
		_, err = part.Write(data)
	}
	if err == nil {
		err = form.Close()
	}
	if err != nil {
		return "", fmt.Errorf("tripo: building upload: %w", err)
	}

	var out struct {
		FileToken string `json:"file_token"`
	}
	// Uploads of up to 50 MB can outlast the API client's timeout on slow
	// links, so they go through the transfer client, bounded by ctx.
	err = c.send(ctx, c.transfer, func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/files", bytes.NewReader(body.Bytes()))
		if err == nil {
			req.Header.Set("Content-Type", form.FormDataContentType())
		}
		return req, err
	}, &out)
	if err != nil {
		return "", err
	}
	if out.FileToken == "" {
		return "", fmt.Errorf("tripo: upload response has no file_token")
	}
	return out.FileToken, nil
}

func (c *Client) presignedUpload(ctx context.Context, path, ext string) (string, error) {
	var presigned struct {
		PresignedURL string `json:"presigned_url"`
		FileToken    string `json:"file_token"`
	}
	if err := c.call(ctx, http.MethodPost, "/files/presign", map[string]string{"format": ext}, &presigned); err != nil {
		return "", err
	}
	if presigned.PresignedURL == "" || presigned.FileToken == "" {
		return "", fmt.Errorf("tripo: presign response is missing presigned_url or file_token")
	}

	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("tripo: opening %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil {
		return "", fmt.Errorf("tripo: reading %s: %w", path, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, presigned.PresignedURL, f)
	if err != nil {
		return "", fmt.Errorf("tripo: building presigned upload: %w", err)
	}
	req.ContentLength = info.Size()
	req.Header.Set("Content-Type", "application/octet-stream")
	resp, err := c.transfer.Do(req)
	if err != nil {
		return "", fmt.Errorf("tripo: presigned upload: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode/100 != 2 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("tripo: presigned upload returned HTTP %d: %s", resp.StatusCode, summarizeBody(msg))
	}
	return presigned.FileToken, nil
}
