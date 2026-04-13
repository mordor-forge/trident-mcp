//go:build e2e

package tripo

import (
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

const tinyPNGBase64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+tmZ8AAAAASUVORK5CYII="

func newE2EProvider(t *testing.T) *TripoProvider {
	t.Helper()

	apiKey := os.Getenv("TRIPO_API_KEY")
	if apiKey == "" {
		t.Skip("TRIPO_API_KEY is not set")
	}

	p, err := New(Config{
		APIKey:    apiKey,
		OutputDir: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("creating E2E provider: %v", err)
	}
	return p
}

func writeTinyPNG(t *testing.T) string {
	t.Helper()

	data, err := base64.StdEncoding.DecodeString(tinyPNGBase64)
	if err != nil {
		t.Fatalf("decoding embedded PNG: %v", err)
	}

	path := filepath.Join(t.TempDir(), "tiny.png")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("writing temp PNG: %v", err)
	}
	return path
}

func TestE2E_UploadFile(t *testing.T) {
	p := newE2EProvider(t)
	imagePath := writeTinyPNG(t)

	token, err := p.uploadFile(context.Background(), imagePath)
	if err != nil {
		t.Fatalf("uploadFile: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty image token")
	}
}
