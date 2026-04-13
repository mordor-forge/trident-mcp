package tripo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mordor-forge/trident-mcp/internal/config"
)

// newTestProvider creates a TripoProvider pointing at a test server.
func newTestProvider(t *testing.T, handler http.Handler) *TripoProvider {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	p, err := New(Config{
		APIKey:    "tsk_test-key",
		BaseURL:   srv.URL,
		OutputDir: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("creating test provider: %v", err)
	}
	return p
}

// taskCreatedHandler returns a handler that expects a POST /task and returns a task ID.
func taskCreatedHandler(wantType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/task" {
			http.Error(w, "unexpected request", http.StatusBadRequest)
			return
		}

		// Verify auth header.
		auth := r.Header.Get("Authorization")
		if auth != "Bearer tsk_test-key" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		_ = json.Unmarshal(body, &req)

		if reqType, ok := req["type"].(string); !ok || reqType != wantType {
			http.Error(w, fmt.Sprintf("expected type %q, got %v", wantType, req["type"]), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{
				"task_id": "test-task-123",
			},
		})
	}
}

func TestNewFromConfig(t *testing.T) {
	cfg := &config.Config{
		Provider:  config.ProviderConfig{APIKey: "tsk_from-config"},
		OutputDir: t.TempDir(),
	}
	p, err := NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("NewFromConfig: %v", err)
	}
	if p.apiKey != "tsk_from-config" {
		t.Errorf("apiKey = %q, want %q", p.apiKey, "tsk_from-config")
	}
	if p.outputDir != cfg.OutputDir {
		t.Errorf("outputDir = %q, want %q", p.outputDir, cfg.OutputDir)
	}
}

func TestNew_RequiresAPIKey(t *testing.T) {
	_, err := New(Config{APIKey: ""})
	if err == nil {
		t.Fatal("expected error for missing API key")
	}
}

func TestNew_DefaultBaseURL(t *testing.T) {
	p, err := New(Config{APIKey: "tsk_test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.baseURL != defaultBaseURL {
		t.Errorf("baseURL = %q, want %q", p.baseURL, defaultBaseURL)
	}
}

func TestNew_CustomBaseURL(t *testing.T) {
	p, err := New(Config{APIKey: "tsk_test", BaseURL: "https://custom.api"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.baseURL != "https://custom.api" {
		t.Errorf("baseURL = %q, want %q", p.baseURL, "https://custom.api")
	}
}

func TestListModels(t *testing.T) {
	p, _ := New(Config{APIKey: "tsk_test"})
	models, err := p.ListModels(context.Background())
	if err != nil {
		t.Fatalf("ListModels: %v", err)
	}
	if len(models) == 0 {
		t.Fatal("expected at least one model")
	}

	// Check that v3.1 is present.
	found := false
	for _, m := range models {
		if m.ID == "v3.1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("v3.1 model not found in list")
	}
}

// --- API error handling ---

func TestDoJSON_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code":    2000,
			"message": "invalid task type",
		})
	}))
	defer srv.Close()

	p, _ := New(Config{APIKey: "tsk_test", BaseURL: srv.URL})
	_, err := p.doJSON(context.Background(), http.MethodGet, "/test", nil)
	if err == nil {
		t.Fatal("expected error for API error response")
	}
	if !strings.Contains(err.Error(), "invalid task type") {
		t.Errorf("error %q does not contain API message", err)
	}
}

func TestDoJSON_NonZeroCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 1001,
		})
	}))
	defer srv.Close()

	p, _ := New(Config{APIKey: "tsk_test", BaseURL: srv.URL})
	_, err := p.doJSON(context.Background(), http.MethodGet, "/test", nil)
	if err == nil {
		t.Fatal("expected error for non-zero code")
	}
	if !strings.Contains(err.Error(), "1001") {
		t.Errorf("error %q does not contain code", err)
	}
}

// --- Upload tests ---

func TestUploadFile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/upload" {
			http.Error(w, "wrong path", http.StatusBadRequest)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "wrong method", http.StatusMethodNotAllowed)
			return
		}
		if r.Header.Get("Authorization") != "Bearer tsk_test-key" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Verify multipart form.
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "no file field", http.StatusBadRequest)
			return
		}
		defer func() { _ = file.Close() }()

		if header.Filename != "test.png" {
			t.Errorf("filename = %q, want %q", header.Filename, "test.png")
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{
				"image_token": "token-abc-123",
			},
		})
	}))
	defer srv.Close()

	p, _ := New(Config{APIKey: "tsk_test-key", BaseURL: srv.URL})

	// Create a temp file.
	tmpFile := filepath.Join(t.TempDir(), "test.png")
	_ = os.WriteFile(tmpFile, []byte("fake png data"), 0o644)

	token, err := p.uploadFile(context.Background(), tmpFile)
	if err != nil {
		t.Fatalf("uploadFile: %v", err)
	}
	if token != "token-abc-123" {
		t.Errorf("token = %q, want %q", token, "token-abc-123")
	}
}

func TestUploadFile_FileNotFound(t *testing.T) {
	p, _ := New(Config{APIKey: "tsk_test", BaseURL: "http://localhost"})
	_, err := p.uploadFile(context.Background(), "/nonexistent/file.png")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestUploadFile_NonZeroCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 1002,
			"data": map[string]any{},
		})
	}))
	defer srv.Close()

	p, _ := New(Config{APIKey: "tsk_test", BaseURL: srv.URL})

	tmpFile := filepath.Join(t.TempDir(), "test.png")
	_ = os.WriteFile(tmpFile, []byte("fake png data"), 0o644)

	_, err := p.uploadFile(context.Background(), tmpFile)
	if err == nil {
		t.Fatal("expected error for upload API code")
	}
	if !strings.Contains(err.Error(), "upload API error code") {
		t.Errorf("error %q does not contain upload API code", err)
	}
}

func TestCreateTask_MissingTaskID(t *testing.T) {
	p := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{},
		})
	}))

	_, err := p.createTask(context.Background(), map[string]any{"type": "text_to_model"})
	if err == nil {
		t.Fatal("expected error for missing task_id")
	}
	if !strings.Contains(err.Error(), "missing task_id") {
		t.Errorf("error %q does not mention missing task_id", err)
	}
}
