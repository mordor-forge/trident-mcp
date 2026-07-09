package tripo

import (
	"context"
	"encoding/json"
	"fmt"
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

// taskCreatedHandler returns a handler that expects a v3 task-creation endpoint
// and returns a task ID.
func taskCreatedHandler(wantType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wantPath := map[string]string{
			"convert_model": "/models/convert",
			"stylize_model": "/models/stylize",
		}[wantType]
		if wantPath == "" {
			http.Error(w, fmt.Sprintf("test helper has no endpoint for %q", wantType), http.StatusBadRequest)
			return
		}
		if r.Method != http.MethodPost || r.URL.Path != wantPath {
			http.Error(w, fmt.Sprintf("unexpected request %s %s", r.Method, r.URL.Path), http.StatusBadRequest)
			return
		}

		// Verify auth header.
		auth := r.Header.Get("Authorization")
		if auth != "Bearer tsk_test-key" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
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
	const want = "https://openapi.tripo3d.ai/v3"
	if p.baseURL != want {
		t.Errorf("baseURL = %q, want %q", p.baseURL, want)
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

	// Check that the latest H-series entry is present.
	found := false
	for _, m := range models {
		if m.ID == "v3.1-20260211" {
			found = true
			break
		}
	}
	if !found {
		t.Error("v3.1-20260211 model not found in list")
	}

	// Check that P1 advertises multiview support.
	found = false
	for _, m := range models {
		if m.ID != "P1-20260311" {
			continue
		}
		found = true
		if len(m.Capabilities) != 3 {
			t.Fatalf("P1-20260311 capabilities = %v, want text/image/multiview", m.Capabilities)
		}
		if got := strings.Join(m.Capabilities, ","); !strings.Contains(got, "multiview_to_3d") {
			t.Errorf("P1-20260311 capabilities = %v, want multiview_to_3d", m.Capabilities)
		}
	}
	if !found {
		t.Error("P1-20260311 model not found in list")
	}

	foundImage := false
	foundRig := false
	foundLegacyRig := false
	for _, m := range models {
		switch m.ID {
		case "seedream_v5":
			foundImage = true
			if m.Namespace != "image_generation" {
				t.Errorf("seedream_v5 namespace = %q", m.Namespace)
			}
		case "rig-v2.0":
			foundRig = true
			if m.Namespace != "animation" {
				t.Errorf("rig-v2.0 namespace = %q", m.Namespace)
			}
			if !m.Default {
				t.Error("rig-v2.0 should be the default animation model")
			}
		case "rig-v1.0":
			foundLegacyRig = true
			if m.Default {
				t.Error("rig-v1.0 should not be the default animation model")
			}
		}
	}
	if !foundImage {
		t.Error("seedream_v5 image model not found in list")
	}
	if !foundRig {
		t.Error("rig-v2.0 animation model not found in list")
	}
	if !foundLegacyRig {
		t.Error("rig-v1.0 animation model not found in list")
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
		if r.URL.Path != "/files" {
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
				"file_token": "file_abc-123",
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
	if token != "file_abc-123" {
		t.Errorf("token = %q, want %q", token, "file_abc-123")
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

	_, err := p.createTask(context.Background(), "/generation/text-to-model", map[string]any{"prompt": "test"})
	if err == nil {
		t.Fatal("expected error for missing task_id")
	}
	if !strings.Contains(err.Error(), "missing task_id") {
		t.Errorf("error %q does not mention missing task_id", err)
	}
}
