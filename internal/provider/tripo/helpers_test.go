package tripo

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGenerateFilename(t *testing.T) {
	// Pin time and randomness for deterministic output.
	origNow := nowUTC
	origRand := randSource
	t.Cleanup(func() {
		nowUTC = origNow
		randSource = origRand
	})

	nowUTC = func() time.Time {
		return time.Date(2026, 4, 10, 14, 30, 0, 0, time.UTC)
	}
	randSource = bytes.NewReader([]byte{0xde, 0xad, 0xbe, 0xef, 0xca, 0xfe, 0xba, 0xbe})

	name := generateFilename("glb")
	if name != "model-2026-04-10T14-30-00-deadbeefcafebabe.glb" {
		t.Errorf("unexpected filename: %s", name)
	}
}

func TestShortID_FallbackOnReadError(t *testing.T) {
	origNow := nowUTC
	origRand := randSource
	t.Cleanup(func() {
		nowUTC = origNow
		randSource = origRand
	})

	nowUTC = func() time.Time {
		return time.Unix(0, 0x123456789abcdef0)
	}
	// Empty reader causes ReadFull to fail.
	randSource = bytes.NewReader(nil)

	id := shortID()
	if id != "123456789abcdef0" {
		t.Errorf("fallback shortID = %q, want %q", id, "123456789abcdef0")
	}
}

func TestFormatToExt(t *testing.T) {
	tests := []struct {
		format string
		want   string
	}{
		{"GLTF", "glb"},
		{"gltf", "glb"},
		{"GLB", "glb"},
		{"FBX", "fbx"},
		{"OBJ", "obj"},
		{"STL", "stl"},
		{"USDZ", "usdz"},
		{"3MF", "3mf"},
		{"unknown", "glb"},
		{"", "glb"},
	}
	for _, tt := range tests {
		if got := formatToExt(tt.format); got != tt.want {
			t.Errorf("formatToExt(%q) = %q, want %q", tt.format, got, tt.want)
		}
	}
}

func TestFileTypeFromPath(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"photo.png", "png"},
		{"photo.jpg", "jpg"},
		{"photo.jpeg", "jpg"},
		{"photo.webp", "webp"},
		{"photo.tiff", "jpg"}, // fallback
	}
	for _, tt := range tests {
		if got := fileTypeFromPath(tt.path); got != tt.want {
			t.Errorf("fileTypeFromPath(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestFileTypeFromURL(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://example.com/photo.png", "png"},
		{"https://example.com/photo.jpeg?x=1", "jpg"},
		{"https://example.com/photo.webp", "webp"},
		{"https://example.com/photo", "jpg"},
	}
	for _, tt := range tests {
		if got := fileTypeFromURL(tt.url); got != tt.want {
			t.Errorf("fileTypeFromURL(%q) = %q, want %q", tt.url, got, tt.want)
		}
	}
}

func TestFormatFromURL(t *testing.T) {
	tests := []struct {
		url    string
		want   string
		wantOK bool
	}{
		{"https://example.com/model.glb", "GLTF", true},
		{"https://example.com/model.fbx?cache=1", "FBX", true},
		{"https://example.com/download?filename=model.3mf", "3MF", true},
		{"https://example.com/download", "", false},
	}
	for _, tt := range tests {
		got, ok := formatFromURL(tt.url)
		if ok != tt.wantOK || got != tt.want {
			t.Errorf("formatFromURL(%q) = (%q, %v), want (%q, %v)", tt.url, got, ok, tt.want, tt.wantOK)
		}
	}
}

func TestFormatFromHeaders(t *testing.T) {
	tests := []struct {
		name   string
		header http.Header
		want   string
		wantOK bool
	}{
		{
			name: "content disposition filename",
			header: http.Header{
				"Content-Disposition": []string{`attachment; filename="mesh.obj"`},
			},
			want:   "OBJ",
			wantOK: true,
		},
		{
			name: "content type fallback",
			header: http.Header{
				"Content-Type": []string{"model/vnd.usdz+zip"},
			},
			want:   "USDZ",
			wantOK: true,
		},
		{
			name:   "unknown",
			header: http.Header{},
			want:   "",
			wantOK: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := formatFromHeaders(tt.header)
			if ok != tt.wantOK || got != tt.want {
				t.Errorf("formatFromHeaders(%v) = (%q, %v), want (%q, %v)", tt.header, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestDownloadFile(t *testing.T) {
	content := []byte("fake model data")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "model/gltf-binary")
		_, _ = w.Write(content)
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "sub", "model.glb")
	headers, err := downloadFile(context.Background(), srv.Client(), srv.URL, dest)
	if err != nil {
		t.Fatalf("downloadFile: %v", err)
	}
	if headers.Get("Content-Type") != "model/gltf-binary" {
		t.Fatalf("unexpected content type header %q", headers.Get("Content-Type"))
	}

	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("reading downloaded file: %v", err)
	}
	if !bytes.Equal(data, content) {
		t.Errorf("file content = %q, want %q", data, content)
	}
}

func TestDownloadFile_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "model.glb")
	_, err := downloadFile(context.Background(), srv.Client(), srv.URL, dest)
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("error %q does not mention status code", err)
	}
}

func TestDownloadFile_ContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("data"))
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	dest := filepath.Join(t.TempDir(), "model.glb")
	_, err := downloadFile(ctx, srv.Client(), srv.URL, dest)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestResolveModelVersion(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", defaultModelVersion},
		{"v3.1", "v3.1-20260211"},
		{"v2.5", "v2.5-20250123"},
		{"turbo", "Turbo-v1.0-20250506"},
		{"v3.1-20260211", "v3.1-20260211"}, // already full version, returned as-is
		{"custom-version", "custom-version"},
	}
	for _, tt := range tests {
		if got := resolveModelVersion(tt.input); got != tt.want {
			t.Errorf("resolveModelVersion(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestValidateTaskID(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"task-123", false},
		{"abc_def", false},
		{"TASK-ABC-123", false},
		{"", true},
		{"../../../etc/passwd", true},
		{"task id with spaces", true},
		{"task;injection", true},
		{"task\ninjection", true},
	}
	for _, tt := range tests {
		err := validateTaskID(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("validateTaskID(%q) error = %v, wantErr = %v", tt.input, err, tt.wantErr)
		}
	}
}
