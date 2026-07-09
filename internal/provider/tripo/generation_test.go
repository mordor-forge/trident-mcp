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
		if r.Method != http.MethodPost || r.URL.Path != "/generation/text-to-model" {
			t.Fatalf("request = %s %s, want POST /generation/text-to-model", r.Method, r.URL.Path)
		}

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

	if _, ok := captured["type"]; ok {
		t.Errorf("v3 request must not include legacy type field: %v", captured["type"])
	}
	if captured["prompt"] != "A red apple" {
		t.Errorf("prompt = %v, want 'A red apple'", captured["prompt"])
	}
	if _, ok := captured["model_version"]; ok {
		t.Errorf("v3 request must not include legacy model_version field: %v", captured["model_version"])
	}
	if captured["model"] != "v3.1-20260211" {
		t.Errorf("model = %v, want v3.1-20260211", captured["model"])
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
	_, err := p.TextToModel(context.Background(), provider.TextToModelRequest{
		Prompt:          "A sword",
		NegativePrompt:  "blurry",
		ModelVersion:    "v3.1",
		FaceLimit:       5000,
		Texture:         boolPtr(false),
		PBR:             boolPtr(true),
		ImageSeed:       intPtr(11),
		ModelSeed:       intPtr(22),
		TextureSeed:     intPtr(33),
		TextureQuality:  "detailed",
		Quad:            boolPtr(true),
		SmartLowPoly:    boolPtr(true),
		GenerateParts:   boolPtr(true),
		AutoSize:        boolPtr(true),
		Compress:        "geometry",
		ExportUV:        boolPtr(false),
		GeometryQuality: "detailed",
	})
	if err != nil {
		t.Fatalf("TextToModel: %v", err)
	}

	if captured["negative_prompt"] != "blurry" {
		t.Errorf("negative_prompt = %v, want 'blurry'", captured["negative_prompt"])
	}
	if _, ok := captured["model_version"]; ok {
		t.Errorf("v3 request must not include legacy model_version field: %v", captured["model_version"])
	}
	if captured["model"] != "v3.1-20260211" {
		t.Errorf("model = %v, want v3.1-20260211", captured["model"])
	}
	if captured["face_limit"] != float64(5000) {
		t.Errorf("face_limit = %v, want 5000", captured["face_limit"])
	}
	if captured["texture"] != false {
		t.Errorf("texture = %v, want false", captured["texture"])
	}
	if captured["pbr"] != true {
		t.Errorf("pbr = %v, want true", captured["pbr"])
	}
	if captured["image_seed"] != float64(11) {
		t.Errorf("image_seed = %v, want 11", captured["image_seed"])
	}
	if captured["model_seed"] != float64(22) {
		t.Errorf("model_seed = %v, want 22", captured["model_seed"])
	}
	if captured["texture_seed"] != float64(33) {
		t.Errorf("texture_seed = %v, want 33", captured["texture_seed"])
	}
	if captured["texture_quality"] != "detailed" {
		t.Errorf("texture_quality = %v, want detailed", captured["texture_quality"])
	}
	if captured["quad"] != true {
		t.Errorf("quad = %v, want true", captured["quad"])
	}
	if captured["smart_low_poly"] != true {
		t.Errorf("smart_low_poly = %v, want true", captured["smart_low_poly"])
	}
	if captured["generate_parts"] != true {
		t.Errorf("generate_parts = %v, want true", captured["generate_parts"])
	}
	if captured["auto_size"] != true {
		t.Errorf("auto_size = %v, want true", captured["auto_size"])
	}
	if captured["compress"] != "geometry" {
		t.Errorf("compress = %v, want geometry", captured["compress"])
	}
	if captured["export_uv"] != false {
		t.Errorf("export_uv = %v, want false", captured["export_uv"])
	}
	if captured["geometry_quality"] != "detailed" {
		t.Errorf("geometry_quality = %v, want detailed", captured["geometry_quality"])
	}
}

func TestTextToModel_OmitsSmartLowPolyForP1(t *testing.T) {
	var captured map[string]any
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"task_id": "txt-task-p1"},
		})
	})

	p := newTestProvider(t, handler)
	_, err := p.TextToModel(context.Background(), provider.TextToModelRequest{
		Prompt:          "A low poly key",
		ModelVersion:    "p1",
		Quad:            boolPtr(true),
		SmartLowPoly:    boolPtr(true),
		GenerateParts:   boolPtr(true),
		GeometryQuality: "detailed",
	})
	if err != nil {
		t.Fatalf("TextToModel: %v", err)
	}

	if captured["model"] != "P1-20260311" {
		t.Errorf("model = %v, want P1-20260311", captured["model"])
	}
	if _, ok := captured["quad"]; ok {
		t.Errorf("P1 request must omit unsupported quad field: %v", captured["quad"])
	}
	if _, ok := captured["smart_low_poly"]; ok {
		t.Errorf("P1 request must omit unsupported smart_low_poly field: %v", captured["smart_low_poly"])
	}
	if _, ok := captured["generate_parts"]; ok {
		t.Errorf("P1 request must omit unsupported generate_parts field: %v", captured["generate_parts"])
	}
	if _, ok := captured["geometry_quality"]; ok {
		t.Errorf("P1 request must omit unsupported geometry_quality field: %v", captured["geometry_quality"])
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
		if r.Method != http.MethodPost || r.URL.Path != "/generation/image-to-model" {
			t.Fatalf("request = %s %s, want POST /generation/image-to-model", r.Method, r.URL.Path)
		}
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

	if _, ok := captured["type"]; ok {
		t.Errorf("v3 request must not include legacy type field: %v", captured["type"])
	}
	if captured["input"] != "https://example.com/photo.jpg" {
		t.Errorf("input = %v", captured["input"])
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

	if captured["input"] != "https://example.com/photo.png?cache=1" {
		t.Errorf("input = %v", captured["input"])
	}
}

func TestImageToModel_WithUpload(t *testing.T) {
	var taskBody map[string]any
	callCount := 0

	mux := http.NewServeMux()
	mux.HandleFunc("/files", func(w http.ResponseWriter, r *http.Request) {
		callCount++
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"file_token": "file_uploaded-token"},
		})
	})
	mux.HandleFunc("/generation/image-to-model", func(w http.ResponseWriter, r *http.Request) {
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

	if taskBody["input"] != "file_uploaded-token" {
		t.Errorf("input = %v", taskBody["input"])
	}
}

func TestImageToModel_WithAdvancedOptions(t *testing.T) {
	var captured map[string]any
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"task_id": "img-task-opts"},
		})
	})

	p := newTestProvider(t, handler)
	_, err := p.ImageToModel(context.Background(), provider.ImageToModelRequest{
		ImageURL:           "https://example.com/photo.jpg",
		ModelVersion:       "p1",
		FaceLimit:          4000,
		Texture:            boolPtr(false),
		PBR:                boolPtr(true),
		ModelSeed:          intPtr(42),
		TextureSeed:        intPtr(99),
		TextureQuality:     "detailed",
		Quad:               boolPtr(true),
		SmartLowPoly:       boolPtr(false),
		GenerateParts:      boolPtr(true),
		TextureAlignment:   "geometry",
		EnableImageAutofix: boolPtr(true),
		AutoSize:           boolPtr(true),
		Orientation:        "align_image",
		Compress:           "geometry",
		ExportUV:           boolPtr(false),
	})
	if err != nil {
		t.Fatalf("ImageToModel: %v", err)
	}

	if captured["model"] != "P1-20260311" {
		t.Errorf("model = %v, want P1-20260311", captured["model"])
	}
	if captured["face_limit"] != float64(4000) {
		t.Errorf("face_limit = %v, want 4000", captured["face_limit"])
	}
	if captured["texture"] != false {
		t.Errorf("texture = %v, want false", captured["texture"])
	}
	if captured["pbr"] != true {
		t.Errorf("pbr = %v, want true", captured["pbr"])
	}
	if captured["model_seed"] != float64(42) {
		t.Errorf("model_seed = %v, want 42", captured["model_seed"])
	}
	if captured["texture_seed"] != float64(99) {
		t.Errorf("texture_seed = %v, want 99", captured["texture_seed"])
	}
	if captured["texture_quality"] != "detailed" {
		t.Errorf("texture_quality = %v, want detailed", captured["texture_quality"])
	}
	if _, ok := captured["quad"]; ok {
		t.Errorf("P1 request must omit unsupported quad field: %v", captured["quad"])
	}
	if _, ok := captured["smart_low_poly"]; ok {
		t.Errorf("P1 request must omit unsupported smart_low_poly field: %v", captured["smart_low_poly"])
	}
	if _, ok := captured["generate_parts"]; ok {
		t.Errorf("P1 request must omit unsupported generate_parts field: %v", captured["generate_parts"])
	}
	if captured["texture_alignment"] != "geometry" {
		t.Errorf("texture_alignment = %v, want geometry", captured["texture_alignment"])
	}
	if captured["enable_image_autofix"] != true {
		t.Errorf("enable_image_autofix = %v, want true", captured["enable_image_autofix"])
	}
	if captured["auto_size"] != true {
		t.Errorf("auto_size = %v, want true", captured["auto_size"])
	}
	if captured["orientation"] != "align_image" {
		t.Errorf("orientation = %v, want align_image", captured["orientation"])
	}
	if captured["compress"] != "geometry" {
		t.Errorf("compress = %v, want geometry", captured["compress"])
	}
	if captured["export_uv"] != false {
		t.Errorf("export_uv = %v, want false", captured["export_uv"])
	}
}

func TestImageToModel_OmitsSmartLowPolyForP1(t *testing.T) {
	var captured map[string]any
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"task_id": "img-task-p1"},
		})
	})

	p := newTestProvider(t, handler)
	_, err := p.ImageToModel(context.Background(), provider.ImageToModelRequest{
		ImageURL:        "https://example.com/photo.png",
		ModelVersion:    "P1-20260311",
		Quad:            boolPtr(true),
		SmartLowPoly:    boolPtr(true),
		GenerateParts:   boolPtr(true),
		GeometryQuality: "detailed",
	})
	if err != nil {
		t.Fatalf("ImageToModel: %v", err)
	}

	if _, ok := captured["quad"]; ok {
		t.Errorf("P1 request must omit unsupported quad field: %v", captured["quad"])
	}
	if _, ok := captured["smart_low_poly"]; ok {
		t.Errorf("P1 request must omit unsupported smart_low_poly field: %v", captured["smart_low_poly"])
	}
	if _, ok := captured["generate_parts"]; ok {
		t.Errorf("P1 request must omit unsupported generate_parts field: %v", captured["generate_parts"])
	}
	if _, ok := captured["geometry_quality"]; ok {
		t.Errorf("P1 request must omit unsupported geometry_quality field: %v", captured["geometry_quality"])
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
		if r.Method != http.MethodPost || r.URL.Path != "/generation/multiview-to-model" {
			t.Fatalf("request = %s %s, want POST /generation/multiview-to-model", r.Method, r.URL.Path)
		}
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

	if _, ok := captured["type"]; ok {
		t.Errorf("v3 request must not include legacy type field: %v", captured["type"])
	}
	inputs := captured["inputs"].([]any)
	if len(inputs) != 3 {
		t.Errorf("inputs count = %d, want 3", len(inputs))
	}
	assertMultiviewInput(t, inputs, 0, "front", "https://example.com/front.jpg")
	assertMultiviewInput(t, inputs, 1, "left", "https://example.com/side.jpg")
	assertMultiviewInput(t, inputs, 2, "back", "https://example.com/back.jpg")
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

	inputs := captured["inputs"].([]any)
	if len(inputs) != 2 {
		t.Errorf("inputs count = %d, want 2", len(inputs))
	}
	assertMultiviewInput(t, inputs, 0, "front", "https://example.com/front.png")
	assertMultiviewInput(t, inputs, 1, "left", "https://example.com/side.webp")
}

func TestMultiviewToModel_WithUploadsUsesViewKeyInputs(t *testing.T) {
	var taskBody map[string]any
	uploadCount := 0

	mux := http.NewServeMux()
	mux.HandleFunc("/files", func(w http.ResponseWriter, r *http.Request) {
		uploadCount++
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"file_token": "file_view_" + string(rune('0'+uploadCount))},
		})
	})
	mux.HandleFunc("/generation/multiview-to-model", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &taskBody)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"task_id": "mv-task-upload"},
		})
	})

	dir := t.TempDir()
	front := filepath.Join(dir, "front.png")
	left := filepath.Join(dir, "left.png")
	if err := os.WriteFile(front, []byte("front"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.WriteFile(left, []byte("left"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	p := newTestProvider(t, mux)
	_, err := p.MultiviewToModel(context.Background(), provider.MultiviewToModelRequest{
		ImagePaths: []string{front, left},
	})
	if err != nil {
		t.Fatalf("MultiviewToModel: %v", err)
	}
	if uploadCount != 2 {
		t.Fatalf("upload count = %d, want 2", uploadCount)
	}

	inputs := taskBody["inputs"].([]any)
	if len(inputs) != 2 {
		t.Fatalf("inputs count = %d, want 2", len(inputs))
	}
	assertMultiviewInput(t, inputs, 0, "front", "file_view_1")
	assertMultiviewInput(t, inputs, 1, "left", "file_view_2")
}

func TestMultiviewToModel_WithP1AndAdvancedOptions(t *testing.T) {
	var captured map[string]any
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"task_id": "mv-task-p1"},
		})
	})

	p := newTestProvider(t, handler)
	_, err := p.MultiviewToModel(context.Background(), provider.MultiviewToModelRequest{
		ImageURLs: []string{
			"https://example.com/front.png",
			"https://example.com/left.webp",
			"https://example.com/back.jpg",
		},
		ModelVersion:       "p1",
		FaceLimit:          6000,
		Texture:            boolPtr(false),
		PBR:                boolPtr(true),
		ModelSeed:          intPtr(7),
		TextureSeed:        intPtr(8),
		TextureQuality:     "detailed",
		Quad:               boolPtr(true),
		SmartLowPoly:       boolPtr(false),
		GenerateParts:      boolPtr(true),
		TextureAlignment:   "geometry",
		EnableImageAutofix: boolPtr(true),
		AutoSize:           boolPtr(true),
		Orientation:        "align_image",
		Compress:           "geometry",
		ExportUV:           boolPtr(false),
	})
	if err != nil {
		t.Fatalf("MultiviewToModel: %v", err)
	}

	if captured["model"] != "P1-20260311" {
		t.Errorf("model = %v, want P1-20260311", captured["model"])
	}
	if captured["face_limit"] != float64(6000) {
		t.Errorf("face_limit = %v, want 6000", captured["face_limit"])
	}
	if captured["texture"] != false {
		t.Errorf("texture = %v, want false", captured["texture"])
	}
	if captured["pbr"] != true {
		t.Errorf("pbr = %v, want true", captured["pbr"])
	}
	if captured["model_seed"] != float64(7) {
		t.Errorf("model_seed = %v, want 7", captured["model_seed"])
	}
	if captured["texture_seed"] != float64(8) {
		t.Errorf("texture_seed = %v, want 8", captured["texture_seed"])
	}
	if captured["texture_quality"] != "detailed" {
		t.Errorf("texture_quality = %v, want detailed", captured["texture_quality"])
	}
	if _, ok := captured["quad"]; ok {
		t.Errorf("P1 request must omit unsupported quad field: %v", captured["quad"])
	}
	if _, ok := captured["smart_low_poly"]; ok {
		t.Errorf("P1 request must omit unsupported smart_low_poly field: %v", captured["smart_low_poly"])
	}
	if _, ok := captured["generate_parts"]; ok {
		t.Errorf("P1 request must omit unsupported generate_parts field: %v", captured["generate_parts"])
	}
	if captured["texture_alignment"] != "geometry" {
		t.Errorf("texture_alignment = %v, want geometry", captured["texture_alignment"])
	}
	if captured["enable_image_autofix"] != true {
		t.Errorf("enable_image_autofix = %v, want true", captured["enable_image_autofix"])
	}
	if captured["auto_size"] != true {
		t.Errorf("auto_size = %v, want true", captured["auto_size"])
	}
	if captured["orientation"] != "align_image" {
		t.Errorf("orientation = %v, want align_image", captured["orientation"])
	}
	if captured["compress"] != "geometry" {
		t.Errorf("compress = %v, want geometry", captured["compress"])
	}
	if captured["export_uv"] != false {
		t.Errorf("export_uv = %v, want false", captured["export_uv"])
	}

	inputs := captured["inputs"].([]any)
	if len(inputs) != 3 {
		t.Fatalf("inputs count = %d, want 3", len(inputs))
	}
	assertMultiviewInput(t, inputs, 0, "front", "https://example.com/front.png")
	assertMultiviewInput(t, inputs, 1, "left", "https://example.com/left.webp")
	assertMultiviewInput(t, inputs, 2, "back", "https://example.com/back.jpg")
}

func TestMultiviewToModel_OmitsSmartLowPolyForP1(t *testing.T) {
	var captured map[string]any
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"task_id": "mv-task-p1-smart"},
		})
	})

	p := newTestProvider(t, handler)
	_, err := p.MultiviewToModel(context.Background(), provider.MultiviewToModelRequest{
		ImageURLs: []string{
			"https://example.com/front.png",
			"https://example.com/left.png",
		},
		ModelVersion:    "p1",
		Quad:            boolPtr(true),
		SmartLowPoly:    boolPtr(true),
		GenerateParts:   boolPtr(true),
		GeometryQuality: "detailed",
	})
	if err != nil {
		t.Fatalf("MultiviewToModel: %v", err)
	}

	if _, ok := captured["quad"]; ok {
		t.Errorf("P1 request must omit unsupported quad field: %v", captured["quad"])
	}
	if _, ok := captured["smart_low_poly"]; ok {
		t.Errorf("P1 request must omit unsupported smart_low_poly field: %v", captured["smart_low_poly"])
	}
	if _, ok := captured["generate_parts"]; ok {
		t.Errorf("P1 request must omit unsupported generate_parts field: %v", captured["generate_parts"])
	}
	if _, ok := captured["geometry_quality"]; ok {
		t.Errorf("P1 request must omit unsupported geometry_quality field: %v", captured["geometry_quality"])
	}
}

func TestMultiviewToModel_WithTaskID(t *testing.T) {
	var captured map[string]any
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/generation/multiview-to-model" {
			t.Fatalf("request = %s %s, want POST /generation/multiview-to-model", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"task_id": "mv-task-from-task"},
		})
	})

	p := newTestProvider(t, handler)
	_, err := p.MultiviewToModel(context.Background(), provider.MultiviewToModelRequest{
		TaskID:       "multiview-task-123",
		ModelVersion: "v3.1",
		Texture:      boolPtr(true),
	})
	if err != nil {
		t.Fatalf("MultiviewToModel: %v", err)
	}

	inputs := captured["inputs"].([]any)
	if len(inputs) != 1 {
		t.Fatalf("inputs count = %d, want 1", len(inputs))
	}
	taskInput := inputs[0].(map[string]any)
	if taskInput["task_id"] != "multiview-task-123" {
		t.Errorf("task_id = %v, want multiview-task-123", taskInput["task_id"])
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

func TestMultiviewToModel_TaskIDMutuallyExclusive(t *testing.T) {
	p, _ := New(Config{APIKey: "tsk_test"})
	_, err := p.MultiviewToModel(context.Background(), provider.MultiviewToModelRequest{
		TaskID:    "multiview-task-123",
		ImageURLs: []string{"https://example.com/front.png", "https://example.com/left.png"},
	})
	if err == nil {
		t.Fatal("expected error for mixed task ID and URL inputs")
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
		if r.URL.Path != "/tasks/task-abc" {
			http.Error(w, "wrong path", http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{
				"task_id":  "task-abc",
				"status":   "running",
				"progress": 50,
				"output": map[string]any{
					"model_url":          "https://cdn.example/model.glb",
					"rendered_image_url": "https://cdn.example/render.png",
					"extra_url":          "https://cdn.example/extra.bin",
				},
				"credits_consumed": 3,
				"created_at":       "2026-07-08T10:00:00Z",
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
	if status.Output == nil {
		t.Fatal("Output is nil, want v3 task output")
	}
	if status.Output.ModelURL != "https://cdn.example/model.glb" {
		t.Errorf("Output.ModelURL = %q", status.Output.ModelURL)
	}
	if status.CreditsConsumed != 3 {
		t.Errorf("CreditsConsumed = %v, want 3", status.CreditsConsumed)
	}
	if status.CreatedAt != "2026-07-08T10:00:00Z" {
		t.Errorf("CreatedAt = %q", status.CreatedAt)
	}
	if status.Output.Extra["extra_url"] == nil {
		t.Errorf("Output.Extra missing extra_url: %#v", status.Output.Extra)
	}
}

func TestStatus_DecimalCreditsConsumed(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{
				"task_id":          "task-credits",
				"status":           "success",
				"progress":         100,
				"credits_consumed": 5.25,
			},
		})
	})

	p := newTestProvider(t, handler)
	status, err := p.Status(context.Background(), "task-credits")
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if status.CreditsConsumed != 5.25 {
		t.Errorf("CreditsConsumed = %v, want 5.25", status.CreditsConsumed)
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

func TestDownload_UsesModelURLsOutput(t *testing.T) {
	fileSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("animated-glb"))
	}))
	defer fileSrv.Close()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{
				"task_id":  "dl-task-model-urls",
				"status":   "success",
				"progress": 100,
				"output": map[string]any{
					"model_urls": []string{fileSrv.URL + "/animated.glb"},
				},
			},
		})
	})

	p := newTestProvider(t, handler)
	result, err := p.Download(context.Background(), "dl-task-model-urls", "")
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	if result.Format != "GLTF" {
		t.Errorf("Format = %q, want GLTF", result.Format)
	}
}

func TestDownload_SelectsRequestedFormatFromModelURLs(t *testing.T) {
	fileSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/animated.fbx" {
			t.Fatalf("download path = %s, want /animated.fbx", r.URL.Path)
		}
		_, _ = w.Write([]byte("animated-fbx"))
	}))
	defer fileSrv.Close()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{
				"task_id":  "dl-task-model-urls-format",
				"status":   "success",
				"progress": 100,
				"output": map[string]any{
					"model_urls": []string{
						fileSrv.URL + "/animated.glb",
						fileSrv.URL + "/animated.fbx",
					},
				},
			},
		})
	})

	p := newTestProvider(t, handler)
	result, err := p.Download(context.Background(), "dl-task-model-urls-format", "FBX")
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	if result.Format != "FBX" {
		t.Errorf("Format = %q, want FBX", result.Format)
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

func assertMultiviewInput(t *testing.T, inputs []any, index int, view string, want string) {
	t.Helper()
	if index >= len(inputs) {
		t.Fatalf("inputs[%d] missing from %#v", index, inputs)
	}
	item, ok := inputs[index].(map[string]any)
	if !ok {
		t.Fatalf("inputs[%d] = %#v, want view-key object", index, inputs[index])
	}
	got, ok := item[view]
	if !ok {
		t.Fatalf("inputs[%d] missing view %q: %#v", index, view, item)
	}
	if got != want {
		t.Errorf("inputs[%d][%q] = %v, want %s", index, view, got, want)
	}
}

func boolPtr(v bool) *bool {
	return &v
}

func intPtr(v int) *int {
	return &v
}
