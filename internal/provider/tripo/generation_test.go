package tripo

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mordor-forge/trident-mcp/internal/provider"
)

func TestTextToModel_Success(t *testing.T) {
	var captured map[string]any
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)

		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"task_id": "txt-task-1"},
		})
	})

	p := newTestProvider(t, handler)
	op, err := p.TextToModel(context.Background(), TextToModelReq("A red apple"))
	if err != nil {
		t.Fatalf("TextToModel: %v", err)
	}
	if op.TaskID != "txt-task-1" {
		t.Errorf("TaskID = %q, want %q", op.TaskID, "txt-task-1")
	}
	if op.Status != "submitted" {
		t.Errorf("Status = %q, want %q", op.Status, "submitted")
	}

	// Verify request body.
	if captured["type"] != "text_to_model" {
		t.Errorf("type = %v, want text_to_model", captured["type"])
	}
	if captured["prompt"] != "A red apple" {
		t.Errorf("prompt = %v, want 'A red apple'", captured["prompt"])
	}
	if captured["model_version"] != defaultModelVersion {
		t.Errorf("model_version = %v, want %v", captured["model_version"], defaultModelVersion)
	}
}

func TestTextToModel_WithOptions(t *testing.T) {
	var captured map[string]any
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)

		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"task_id": "txt-task-2"},
		})
	})

	p := newTestProvider(t, handler)
	_, err := p.TextToModel(context.Background(), TextToModelReqFull("A sword", "blurry", "v2.5", 5000))
	if err != nil {
		t.Fatalf("TextToModel: %v", err)
	}

	if captured["negative_prompt"] != "blurry" {
		t.Errorf("negative_prompt = %v, want 'blurry'", captured["negative_prompt"])
	}
	if captured["model_version"] != "v2.5-20250123" {
		t.Errorf("model_version = %v, want 'v2.5-20250123'", captured["model_version"])
	}
	if captured["face_limit"] != float64(5000) {
		t.Errorf("face_limit = %v, want 5000", captured["face_limit"])
	}
}

func TestTextToModel_EmptyPrompt(t *testing.T) {
	p, _ := New(Config{APIKey: "tsk_test"})
	_, err := p.TextToModel(context.Background(), TextToModelReq(""))
	if err == nil {
		t.Fatal("expected error for empty prompt")
	}
}

func TestImageToModel_WithURL(t *testing.T) {
	var captured map[string]any
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)

		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"task_id": "img-task-1"},
		})
	})

	p := newTestProvider(t, handler)
	op, err := p.ImageToModel(context.Background(), ImageToModelReqURL("https://example.com/photo.jpg"))
	if err != nil {
		t.Fatalf("ImageToModel: %v", err)
	}
	if op.TaskID != "img-task-1" {
		t.Errorf("TaskID = %q, want %q", op.TaskID, "img-task-1")
	}

	if captured["type"] != "image_to_model" {
		t.Errorf("type = %v, want image_to_model", captured["type"])
	}
	file := captured["file"].(map[string]any)
	if file["url"] != "https://example.com/photo.jpg" {
		t.Errorf("file.url = %v", file["url"])
	}
	if file["type"] != "jpg" {
		t.Errorf("file.type = %v, want jpg", file["type"])
	}
}

func TestImageToModel_WithURLPreservesFileType(t *testing.T) {
	var captured map[string]any
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"task_id": "img-task-png"},
		})
	})

	p := newTestProvider(t, handler)
	_, err := p.ImageToModel(context.Background(), ImageToModelReqURL("https://example.com/photo.png?cache=1"))
	if err != nil {
		t.Fatalf("ImageToModel: %v", err)
	}

	file := captured["file"].(map[string]any)
	if file["type"] != "png" {
		t.Errorf("file.type = %v, want png", file["type"])
	}
}

func TestImageToModel_WithUpload(t *testing.T) {
	var taskBody map[string]any
	callCount := 0

	mux := http.NewServeMux()
	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		callCount++
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"image_token": "uploaded-token"},
		})
	})
	mux.HandleFunc("/task", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &taskBody)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"task_id": "img-task-2"},
		})
	})

	p := newTestProvider(t, mux)

	tmpFile := filepath.Join(t.TempDir(), "test.png")
	_ = os.WriteFile(tmpFile, []byte("fake png"), 0o644)

	op, err := p.ImageToModel(context.Background(), ImageToModelReqPath(tmpFile))
	if err != nil {
		t.Fatalf("ImageToModel: %v", err)
	}
	if op.TaskID != "img-task-2" {
		t.Errorf("TaskID = %q", op.TaskID)
	}
	if callCount != 1 {
		t.Errorf("upload called %d times, want 1", callCount)
	}

	file := taskBody["file"].(map[string]any)
	if file["file_token"] != "uploaded-token" {
		t.Errorf("file_token = %v", file["file_token"])
	}
	if file["type"] != "png" {
		t.Errorf("file.type = %v, want png", file["type"])
	}
}

func TestImageToModel_MissingInput(t *testing.T) {
	p, _ := New(Config{APIKey: "tsk_test"})
	_, err := p.ImageToModel(context.Background(), ImageToModelReqURL(""))
	if err == nil {
		t.Fatal("expected error for missing image input")
	}
}

func TestMultiviewToModel_WithURLs(t *testing.T) {
	var captured map[string]any
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"task_id": "mv-task-1"},
		})
	})

	p := newTestProvider(t, handler)
	urls := []string{
		"https://example.com/front.jpg",
		"https://example.com/side.jpg",
		"https://example.com/back.jpg",
	}
	op, err := p.MultiviewToModel(context.Background(), MultiviewToModelReqURLs(urls))
	if err != nil {
		t.Fatalf("MultiviewToModel: %v", err)
	}
	if op.TaskID != "mv-task-1" {
		t.Errorf("TaskID = %q", op.TaskID)
	}

	if captured["type"] != "multiview_to_model" {
		t.Errorf("type = %v", captured["type"])
	}
	files := captured["files"].([]any)
	if len(files) != 3 {
		t.Errorf("files count = %d, want 3", len(files))
	}
	if files[1].(map[string]any)["type"] != "jpg" {
		t.Errorf("files[1].type = %v, want jpg", files[1].(map[string]any)["type"])
	}
}

func TestMultiviewToModel_WithURLsPreservesFileTypes(t *testing.T) {
	var captured map[string]any
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"task_id": "mv-task-types"},
		})
	})

	p := newTestProvider(t, handler)
	_, err := p.MultiviewToModel(context.Background(), MultiviewToModelReqURLs([]string{
		"https://example.com/front.png",
		"https://example.com/side.webp",
	}))
	if err != nil {
		t.Fatalf("MultiviewToModel: %v", err)
	}

	files := captured["files"].([]any)
	if files[0].(map[string]any)["type"] != "png" {
		t.Errorf("files[0].type = %v, want png", files[0].(map[string]any)["type"])
	}
	if files[1].(map[string]any)["type"] != "webp" {
		t.Errorf("files[1].type = %v, want webp", files[1].(map[string]any)["type"])
	}
}

func TestMultiviewToModel_TooFewImages(t *testing.T) {
	p, _ := New(Config{APIKey: "tsk_test"})
	_, err := p.MultiviewToModel(context.Background(), MultiviewToModelReqURLs([]string{"https://one.jpg"}))
	if err == nil {
		t.Fatal("expected error for too few images")
	}
	if !strings.Contains(err.Error(), "2-4") {
		t.Errorf("error %q doesn't mention required count", err)
	}
}

func TestMultiviewToModel_TooManyImages(t *testing.T) {
	p, _ := New(Config{APIKey: "tsk_test"})
	urls := make([]string, 5)
	for i := range urls {
		urls[i] = "https://img.jpg"
	}
	_, err := p.MultiviewToModel(context.Background(), MultiviewToModelReqURLs(urls))
	if err == nil {
		t.Fatal("expected error for too many images")
	}
}

func TestMultiviewToModel_BothPathsAndURLs(t *testing.T) {
	p, _ := New(Config{APIKey: "tsk_test"})
	_, err := p.MultiviewToModel(context.Background(), provider.MultiviewToModelRequest{
		ImagePaths: []string{"/tmp/a.png", "/tmp/b.png"},
		ImageURLs:  []string{"https://example.com/a.png", "https://example.com/b.png"},
	})
	if err == nil {
		t.Fatal("expected error for mixed path and URL inputs")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("error %q doesn't mention mutual exclusivity", err)
	}
}

func TestMultiviewToModel_UnsupportedVersion(t *testing.T) {
	p, _ := New(Config{APIKey: "tsk_test"})
	req := MultiviewToModelReqURLs([]string{"https://a.jpg", "https://b.jpg"})
	req.ModelVersion = "v1.4"
	_, err := p.MultiviewToModel(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for unsupported multiview version")
	}
	if !strings.Contains(err.Error(), "does not support multiview") {
		t.Errorf("error %q doesn't mention multiview", err)
	}
}

func TestStatus_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/task/task-abc" {
			http.Error(w, "wrong path", http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{
				"task_id":  "task-abc",
				"status":   "running",
				"progress": 50,
			},
		})
	})

	p := newTestProvider(t, handler)
	status, err := p.Status(context.Background(), "task-abc")
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if status.TaskID != "task-abc" {
		t.Errorf("TaskID = %q", status.TaskID)
	}
	if status.Status != "running" {
		t.Errorf("Status = %q, want running", status.Status)
	}
	if status.Progress != 50 {
		t.Errorf("Progress = %d, want 50", status.Progress)
	}
}

func TestStatus_Failed(t *testing.T) {
	errMsg := "content policy violation"
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{
				"task_id":   "task-fail",
				"status":    "failed",
				"progress":  0,
				"error_msg": errMsg,
			},
		})
	})

	p := newTestProvider(t, handler)
	status, err := p.Status(context.Background(), "task-fail")
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if status.Error != errMsg {
		t.Errorf("Error = %q, want %q", status.Error, errMsg)
	}
}

func TestStatus_EmptyTaskID(t *testing.T) {
	p, _ := New(Config{APIKey: "tsk_test"})
	_, err := p.Status(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty task ID")
	}
}

func TestDownload_Success(t *testing.T) {
	modelContent := []byte("fake glb model data")

	// Model file server.
	fileSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(modelContent)
	}))
	defer fileSrv.Close()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{
				"task_id":  "dl-task-1",
				"status":   "success",
				"progress": 100,
				"output": map[string]any{
					"pbr_model": fileSrv.URL + "/model.glb",
				},
			},
		})
	})

	p := newTestProvider(t, handler)
	result, err := p.Download(context.Background(), "dl-task-1", "GLTF")
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	if result.TaskID != "dl-task-1" {
		t.Errorf("TaskID = %q", result.TaskID)
	}
	if result.Format != "GLTF" {
		t.Errorf("Format = %q, want GLTF", result.Format)
	}
	if !strings.HasSuffix(result.FilePath, ".glb") {
		t.Errorf("FilePath %q doesn't end with .glb", result.FilePath)
	}

	// Verify file was downloaded.
	data, err := os.ReadFile(result.FilePath)
	if err != nil {
		t.Fatalf("reading downloaded file: %v", err)
	}
	if string(data) != string(modelContent) {
		t.Errorf("downloaded content mismatch")
	}
}

func TestDownload_TaskNotComplete(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{
				"task_id":  "dl-task-2",
				"status":   "running",
				"progress": 50,
			},
		})
	})

	p := newTestProvider(t, handler)
	_, err := p.Download(context.Background(), "dl-task-2", "GLTF")
	if err == nil {
		t.Fatal("expected error for incomplete task")
	}
	if !strings.Contains(err.Error(), "not complete") {
		t.Errorf("error %q doesn't mention completion", err)
	}
}

func TestDownload_DefaultFormat(t *testing.T) {
	fileSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("data"))
	}))
	defer fileSrv.Close()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{
				"task_id":  "dl-task-3",
				"status":   "success",
				"progress": 100,
				"output":   map[string]any{"model": fileSrv.URL + "/m.glb"},
			},
		})
	})

	p := newTestProvider(t, handler)
	result, err := p.Download(context.Background(), "dl-task-3", "")
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	if result.Format != "GLTF" {
		t.Errorf("default format = %q, want GLTF", result.Format)
	}
}

func TestDownload_DetectsActualFormatFromTaskOutput(t *testing.T) {
	fileSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("fbx-data"))
	}))
	defer fileSrv.Close()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{
				"task_id":  "dl-task-fbx",
				"status":   "success",
				"progress": 100,
				"output":   map[string]any{"model": fileSrv.URL + "/m.fbx"},
			},
		})
	})

	p := newTestProvider(t, handler)
	result, err := p.Download(context.Background(), "dl-task-fbx", "")
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	if result.Format != "FBX" {
		t.Errorf("Format = %q, want FBX", result.Format)
	}
	if !strings.HasSuffix(result.FilePath, ".fbx") {
		t.Errorf("FilePath %q doesn't end with .fbx", result.FilePath)
	}
}

func TestDownload_DetectsFormatFromResponseHeaders(t *testing.T) {
	fileSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Disposition", `attachment; filename="mesh.obj"`)
		_, _ = w.Write([]byte("obj-data"))
	}))
	defer fileSrv.Close()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{
				"task_id":  "dl-task-header",
				"status":   "success",
				"progress": 100,
				"output":   map[string]any{"model": fileSrv.URL + "/download"},
			},
		})
	})

	p := newTestProvider(t, handler)
	result, err := p.Download(context.Background(), "dl-task-header", "")
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	if result.Format != "OBJ" {
		t.Errorf("Format = %q, want OBJ", result.Format)
	}
	if !strings.HasSuffix(result.FilePath, ".obj") {
		t.Errorf("FilePath %q doesn't end with .obj", result.FilePath)
	}
}

func TestDownload_RejectsMismatchedRequestedFormat(t *testing.T) {
	fileSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("glb-data"))
	}))
	defer fileSrv.Close()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{
				"task_id":  "dl-task-mismatch",
				"status":   "success",
				"progress": 100,
				"output":   map[string]any{"pbr_model": fileSrv.URL + "/m.glb"},
			},
		})
	})

	p := newTestProvider(t, handler)
	_, err := p.Download(context.Background(), "dl-task-mismatch", "FBX")
	if err == nil {
		t.Fatal("expected mismatch error")
	}
	if !strings.Contains(err.Error(), "convert_format") {
		t.Errorf("error %q doesn't mention convert_format", err)
	}
}

func TestDownload_InvalidRequestedFormat(t *testing.T) {
	p, _ := New(Config{APIKey: "tsk_test"})
	_, err := p.Download(context.Background(), "task-1", "INVALID")
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("error %q doesn't mention unsupported format", err)
	}
}

func TestDownload_EmptyTaskID(t *testing.T) {
	p, _ := New(Config{APIKey: "tsk_test"})
	_, err := p.Download(context.Background(), "", "GLTF")
	if err == nil {
		t.Fatal("expected error for empty task ID")
	}
}

func TestDownload_InvalidTaskID(t *testing.T) {
	p, _ := New(Config{APIKey: "tsk_test"})
	_, err := p.Download(context.Background(), "../../../etc/passwd", "GLTF")
	if err == nil {
		t.Fatal("expected error for path traversal task ID")
	}
	if !strings.Contains(err.Error(), "invalid taskID") {
		t.Errorf("error %q doesn't mention invalid taskID", err)
	}
}

func TestStatus_InvalidTaskID(t *testing.T) {
	p, _ := New(Config{APIKey: "tsk_test"})
	_, err := p.Status(context.Background(), "task;injection")
	if err == nil {
		t.Fatal("expected error for task ID with special characters")
	}
}

func TestImageToModel_BothPathAndURL(t *testing.T) {
	p, _ := New(Config{APIKey: "tsk_test", BaseURL: "http://localhost"})
	_, err := p.ImageToModel(context.Background(), provider.ImageToModelRequest{
		ImagePath: "/nonexistent/file.png",
		ImageURL:  "https://example.com/photo.jpg",
	})
	if err == nil {
		t.Fatal("expected validation error for mixed path and URL inputs")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("error %q doesn't mention mutual exclusivity", err)
	}
}

func TestDownload_NoOutputURLs(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{
				"task_id":  "dl-task-empty",
				"status":   "success",
				"progress": 100,
				"output":   map[string]any{},
			},
		})
	})

	p := newTestProvider(t, handler)
	_, err := p.Download(context.Background(), "dl-task-empty", "GLTF")
	if err == nil {
		t.Fatal("expected error for missing download URLs")
	}
	if !strings.Contains(err.Error(), "no model download URL") {
		t.Errorf("error %q doesn't mention missing URL", err)
	}
}

// --- Request builders (reduce test boilerplate) ---

func TextToModelReq(prompt string) provider.TextToModelRequest {
	return provider.TextToModelRequest{Prompt: prompt}
}

func TextToModelReqFull(prompt, negative, version string, faces int) provider.TextToModelRequest {
	return provider.TextToModelRequest{
		Prompt:         prompt,
		NegativePrompt: negative,
		ModelVersion:   version,
		FaceLimit:      faces,
	}
}

func ImageToModelReqURL(url string) provider.ImageToModelRequest {
	return provider.ImageToModelRequest{ImageURL: url}
}

func ImageToModelReqPath(path string) provider.ImageToModelRequest {
	return provider.ImageToModelRequest{ImagePath: path}
}

func MultiviewToModelReqURLs(urls []string) provider.MultiviewToModelRequest {
	return provider.MultiviewToModelRequest{ImageURLs: urls}
}
