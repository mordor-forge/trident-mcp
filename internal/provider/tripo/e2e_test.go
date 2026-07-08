//go:build e2e

package tripo

import (
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mordor-forge/trident-mcp/internal/provider"
)

const tinyPNGBase64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+tmZ8AAAAASUVORK5CYII="

const (
	minTextToModelCredits = 100
	minFullMatrixCredits  = 5
)

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

	upload, err := p.UploadFile(context.Background(), provider.UploadFileRequest{FilePath: imagePath})
	if err != nil {
		t.Fatalf("UploadFile: %v", err)
	}
	if upload.FileToken == "" {
		t.Fatal("expected non-empty image token")
	}
}

func TestE2E_CreateFileUpload(t *testing.T) {
	p := newE2EProvider(t)

	presign, err := p.CreateFileUpload(context.Background(), provider.CreateFileUploadRequest{Format: "png"})
	if err != nil {
		t.Fatalf("CreateFileUpload: %v", err)
	}
	if presign.FileToken == "" || presign.PresignedURL == "" {
		t.Fatalf("invalid presign response: %#v", presign)
	}
}

func TestE2E_AccountBalanceAndUsage(t *testing.T) {
	p := newE2EProvider(t)

	balance, err := p.GetBalance(context.Background())
	if err != nil {
		t.Fatalf("GetBalance: %v", err)
	}
	if balance.Balance < 0 || balance.Frozen < 0 {
		t.Fatalf("invalid balance response: %#v", balance)
	}

	usage, err := p.GetUsage(context.Background())
	if err != nil {
		t.Fatalf("GetUsage: %v", err)
	}
	if usage.Records == nil {
		t.Fatal("expected non-nil usage records slice")
	}
}

func TestE2E_TextToModelLifecycle(t *testing.T) {
	if os.Getenv("TRIPO_E2E_GENERATE") != "1" {
		t.Skip("set TRIPO_E2E_GENERATE=1 to run credit-spending generation e2e tests")
	}
	p := newE2EProvider(t)
	requireE2EBalance(t, p, minTextToModelCredits)

	texture := false
	op, err := p.TextToModel(context.Background(), provider.TextToModelRequest{
		Prompt:       "a simple low poly cube game asset",
		ModelVersion: "tripo-p1",
		FaceLimit:    150,
		Texture:      &texture,
	})
	if err != nil {
		t.Fatalf("TextToModel: %v", err)
	}
	if op.TaskID == "" {
		t.Fatal("expected task ID")
	}

	status := waitForE2ETask(t, p, op.TaskID, 10*time.Minute)
	if status.Output == nil || status.Output.ModelURL == "" {
		t.Fatalf("completed task has no model URL: %#v", status.Output)
	}

	batch, err := p.BatchTasks(context.Background(), []string{op.TaskID})
	if err != nil {
		t.Fatalf("BatchTasks: %v", err)
	}
	if _, ok := batch.Tasks[op.TaskID]; !ok {
		t.Fatalf("batch result missing created task %s: %#v", op.TaskID, batch)
	}

	result, err := p.Download(context.Background(), op.TaskID, "")
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	if result.FilePath == "" {
		t.Fatal("download result missing file path")
	}
}

func TestE2E_FullV3CreationMatrix(t *testing.T) {
	if os.Getenv("TRIPO_E2E_FULL") != "1" {
		t.Skip("set TRIPO_E2E_FULL=1 to run the broader credit-spending v3 endpoint matrix")
	}
	p := newE2EProvider(t)
	requireE2EBalance(t, p, minFullMatrixCredits)

	ctx := context.Background()
	imageTask, err := p.TextToImage(ctx, provider.TextToImageRequest{
		Prompt: "single front-facing toy robot reference image on white background",
		Model:  "seedream_v4",
	})
	if err != nil {
		t.Fatalf("TextToImage: %v", err)
	}
	waitForE2ETask(t, p, imageTask.TaskID, 5*time.Minute)

	presign, err := p.CreateFileUpload(ctx, provider.CreateFileUploadRequest{Format: "png"})
	if err != nil {
		t.Fatalf("CreateFileUpload: %v", err)
	}
	if presign.FileToken == "" || presign.PresignedURL == "" {
		t.Fatalf("invalid presign response: %#v", presign)
	}
}

func requireE2EBalance(t *testing.T, p *TripoProvider, minimum float64) {
	t.Helper()

	balance, err := p.GetBalance(context.Background())
	if err != nil {
		t.Fatalf("GetBalance before credit-spending e2e: %v", err)
	}
	if balance.Balance < minimum {
		t.Skipf(
			"Tripo balance %.2f is below %.2f required for this credit-spending e2e test (frozen %.2f)",
			balance.Balance,
			minimum,
			balance.Frozen,
		)
	}
}

func waitForE2ETask(t *testing.T, p *TripoProvider, taskID string, timeout time.Duration) *provider.ModelTaskStatus {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		status, err := p.Status(context.Background(), taskID)
		if err != nil {
			t.Fatalf("Status(%s): %v", taskID, err)
		}
		switch status.Status {
		case "success":
			return status
		case "failed", "cancelled", "banned", "expired":
			t.Fatalf("task %s ended with status %s: %s", taskID, status.Status, status.Error)
		}
		time.Sleep(2 * time.Second)
	}
	t.Fatalf("task %s did not complete within %s", taskID, timeout)
	return nil
}
