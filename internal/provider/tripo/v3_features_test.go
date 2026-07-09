package tripo

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mordor-forge/trident-mcp/internal/provider"
)

func TestTextToImage_PostsV3Payload(t *testing.T) {
	var captured map[string]any
	p := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/generation/text-to-image" {
			t.Fatalf("request = %s %s, want POST /generation/text-to-image", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"task_id": "text-image-task"},
		})
	}))

	op, err := p.TextToImage(context.Background(), provider.TextToImageRequest{
		Prompt:         "product photo",
		Model:          "seedream_v5",
		NegativePrompt: "blurry",
		Template:       "asset_extraction",
	})
	if err != nil {
		t.Fatalf("TextToImage: %v", err)
	}
	if op.TaskID != "text-image-task" {
		t.Fatalf("TaskID = %q", op.TaskID)
	}
	assertBodyField(t, captured, "prompt", "product photo")
	assertBodyField(t, captured, "model", "seedream_v5")
	assertBodyField(t, captured, "negative_prompt", "blurry")
	assertBodyField(t, captured, "template", "asset_extraction")
}

func TestImageToImage_PostsInput(t *testing.T) {
	var captured map[string]any
	p := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/generation/image-to-image" {
			t.Fatalf("request = %s %s, want POST /generation/image-to-image", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"task_id": "image-image-task"},
		})
	}))

	_, err := p.ImageToImage(context.Background(), provider.ImageToImageRequest{
		ImageURL: "https://example.com/input.png",
		Prompt:   "make it clay",
		Model:    "chat_image_2",
	})
	if err != nil {
		t.Fatalf("ImageToImage: %v", err)
	}
	assertBodyField(t, captured, "input", "https://example.com/input.png")
	assertBodyField(t, captured, "prompt", "make it clay")
	assertBodyField(t, captured, "model", "chat_image_2")
}

func TestImageToMultiviewAndEditMultiview(t *testing.T) {
	paths := []string{}
	p := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		body, _ := io.ReadAll(r.Body)
		var captured map[string]any
		_ = json.Unmarshal(body, &captured)
		assertBodyField(t, captured, "input", "https://example.com/front.png")
		if r.URL.Path == "/generation/edit-multiview" {
			prompts := captured["prompts"].([]any)
			if len(prompts) != 2 {
				t.Fatalf("prompts count = %d, want 2", len(prompts))
			}
			assertPromptView(t, prompts, 0, "turn left view into profile", "left")
			assertPromptView(t, prompts, 1, "remove text from back", "back")
			if _, ok := captured["prompt"]; ok {
				t.Fatalf("edit_multiview body must not include singular prompt: %#v", captured)
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"task_id": strings.Trim(r.URL.Path, "/")},
		})
	}))

	if _, err := p.ImageToMultiview(context.Background(), provider.ImageToMultiviewRequest{
		ImageURL: "https://example.com/front.png",
	}); err != nil {
		t.Fatalf("ImageToMultiview: %v", err)
	}
	if _, err := p.EditMultiview(context.Background(), provider.EditMultiviewRequest{
		Input: "https://example.com/front.png",
		Prompts: []provider.MultiviewEditPrompt{
			{Prompt: "turn left view into profile", View: "left"},
			{Prompt: "remove text from back", View: "back"},
		},
	}); err != nil {
		t.Fatalf("EditMultiview: %v", err)
	}
	want := []string{"/generation/image-to-multiview", "/generation/edit-multiview"}
	for i := range want {
		if paths[i] != want[i] {
			t.Fatalf("paths[%d] = %q, want %q", i, paths[i], want[i])
		}
	}
}

func TestEditMultiview_ExpandsLegacyPromptToAllViews(t *testing.T) {
	var captured map[string]any
	p := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"task_id": "edit-multiview"},
		})
	}))

	if _, err := p.EditMultiview(context.Background(), provider.EditMultiviewRequest{
		Input:  "task_multiview",
		Prompt: "make the bottle glass blue",
	}); err != nil {
		t.Fatalf("EditMultiview: %v", err)
	}

	prompts := captured["prompts"].([]any)
	if len(prompts) != 4 {
		t.Fatalf("prompts count = %d, want 4", len(prompts))
	}
	for i, view := range []string{"front", "left", "back", "right"} {
		assertPromptView(t, prompts, i, "make the bottle glass blue", view)
	}
}

func TestEditMultiview_RequiresPrompts(t *testing.T) {
	p := newTestProvider(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("unexpected request")
	}))

	_, err := p.EditMultiview(context.Background(), provider.EditMultiviewRequest{Input: "task_multiview"})
	if err == nil {
		t.Fatal("expected missing prompts error")
	}
}

func TestImageToSplat_PostsInput(t *testing.T) {
	var captured map[string]any
	p := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/generation/image-to-splat" {
			t.Fatalf("request = %s %s, want POST /generation/image-to-splat", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"task_id": "splat-task"},
		})
	}))

	_, err := p.ImageToSplat(context.Background(), provider.ImageToSplatRequest{
		ImageURL: "https://example.com/input.webp",
	})
	if err != nil {
		t.Fatalf("ImageToSplat: %v", err)
	}
	assertBodyField(t, captured, "input", "https://example.com/input.webp")
}

func TestModelAndMeshProcessingV3Endpoints(t *testing.T) {
	wantPaths := map[string]bool{
		"/models/import":  false,
		"/models/refine":  false,
		"/models/texture": false,
		"/mesh/segment":   false,
		"/mesh/complete":  false,
	}
	p := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := wantPaths[r.URL.Path]; !ok {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		wantPaths[r.URL.Path] = true
		body, _ := io.ReadAll(r.Body)
		var captured map[string]any
		_ = json.Unmarshal(body, &captured)
		if captured["input"] == nil || captured["input"] == "" {
			t.Fatalf("%s missing input in body %#v", r.URL.Path, captured)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"task_id": strings.Trim(r.URL.Path, "/")},
		})
	}))

	ctx := context.Background()
	calls := []struct {
		name string
		run  func() error
	}{
		{"import", func() error {
			_, err := p.ImportModel(ctx, provider.ImportModelRequest{Input: "file_model"})
			return err
		}},
		{"refine", func() error {
			_, err := p.RefineModel(ctx, provider.RefineModelRequest{Input: "task_model"})
			return err
		}},
		{"texture", func() error {
			_, err := p.TextureModel(ctx, provider.TextureModelRequest{Input: "task_model", TextureQuality: "detailed"})
			return err
		}},
		{"segment", func() error {
			_, err := p.SegmentMesh(ctx, provider.SegmentMeshRequest{Input: "task_model"})
			return err
		}},
		{"complete", func() error {
			_, err := p.CompleteMesh(ctx, provider.CompleteMeshRequest{Input: "task_segment", PartNames: []string{"left_arm"}})
			return err
		}},
	}
	for _, call := range calls {
		if err := call.run(); err != nil {
			t.Fatalf("%s: %v", call.name, err)
		}
	}
	for path, seen := range wantPaths {
		if !seen {
			t.Fatalf("endpoint %s was not called", path)
		}
	}
}

func TestAnimationEndpoints(t *testing.T) {
	p := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var captured map[string]any
		_ = json.Unmarshal(body, &captured)
		if captured["input"] != "task_model" {
			t.Fatalf("%s input = %v", r.URL.Path, captured["input"])
		}
		switch r.URL.Path {
		case "/animations/rig-check":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"data": map[string]any{"riggable": true, "rig_type": "biped"},
			})
		case "/animations/rig", "/animations/retarget":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"data": map[string]any{"task_id": strings.Trim(r.URL.Path, "/")},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))

	rigCheck, err := p.RigCheck(context.Background(), provider.RigCheckRequest{Input: "task_model"})
	if err != nil {
		t.Fatalf("RigCheck: %v", err)
	}
	if !rigCheck.Riggable || rigCheck.RigType != "biped" {
		t.Fatalf("RigCheck = %#v", rigCheck)
	}
	if _, err := p.RigModel(context.Background(), provider.RigModelRequest{
		Input:     "task_model",
		RigType:   "biped",
		Spec:      "mixamo",
		OutFormat: "glb",
	}); err != nil {
		t.Fatalf("RigModel: %v", err)
	}
	if _, err := p.RetargetAnimation(context.Background(), provider.RetargetAnimationRequest{
		Input:     "task_model",
		Animation: "preset:walk",
	}); err != nil {
		t.Fatalf("RetargetAnimation: %v", err)
	}
}

func TestBatchTasksAndPresign(t *testing.T) {
	p := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/tasks/list":
			body, _ := io.ReadAll(r.Body)
			var captured map[string]any
			_ = json.Unmarshal(body, &captured)
			ids := captured["task_ids"].([]any)
			if len(ids) != 2 {
				t.Fatalf("task_ids = %#v", captured["task_ids"])
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"data": map[string]any{
					"tasks": map[string]any{
						"task_a": map[string]any{"task_id": "task_a", "status": "success"},
					},
					"missed": []string{"task_b"},
				},
			})
		case "/files/presign":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"data": map[string]any{
					"presigned_url": "https://storage.example/upload",
					"file_token":    "file_presigned",
					"expires_in":    1800,
				},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))

	batch, err := p.BatchTasks(context.Background(), []string{"task_a", "task_b"})
	if err != nil {
		t.Fatalf("BatchTasks: %v", err)
	}
	if len(batch.Tasks) != 1 || len(batch.Missed) != 1 {
		t.Fatalf("BatchTasks = %#v", batch)
	}

	presign, err := p.CreateFileUpload(context.Background(), provider.CreateFileUploadRequest{Format: "glb"})
	if err != nil {
		t.Fatalf("CreateFileUpload: %v", err)
	}
	if presign.FileToken != "file_presigned" || presign.ExpiresIn != 1800 {
		t.Fatalf("CreateFileUpload = %#v", presign)
	}
}

func TestUploadFilePostsMultipartToFiles(t *testing.T) {
	var sawFile bool
	p := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/files" {
			t.Fatalf("request = %s %s, want POST /files", r.Method, r.URL.Path)
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("FormFile(file): %v", err)
		}
		_ = file.Close()
		sawFile = true
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"file_token": "file_direct"},
		})
	}))

	tmpFile := filepath.Join(t.TempDir(), "mesh.glb")
	if err := os.WriteFile(tmpFile, []byte("glb"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	result, err := p.UploadFile(context.Background(), provider.UploadFileRequest{FilePath: tmpFile})
	if err != nil {
		t.Fatalf("UploadFile: %v", err)
	}
	if !sawFile {
		t.Fatal("server did not receive multipart file")
	}
	if result.FileToken != "file_direct" {
		t.Fatalf("FileToken = %q, want file_direct", result.FileToken)
	}
}

func TestAccountBalanceAndUsage(t *testing.T) {
	p := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		switch r.URL.Path {
		case "/account/balance":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"data": map[string]any{
					"balance": 10000.5,
					"frozen":  200.25,
				},
			})
		case "/account/usage":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"data": []map[string]any{
					{
						"task_id":          "task_abc123",
						"type":             "text_to_model",
						"credits_consumed": 5.25,
						"created_at":       "2026-07-08T10:00:00Z",
					},
				},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))

	balance, err := p.GetBalance(context.Background())
	if err != nil {
		t.Fatalf("GetBalance: %v", err)
	}
	if balance.Balance != 10000.5 || balance.Frozen != 200.25 {
		t.Fatalf("GetBalance = %#v", balance)
	}

	usage, err := p.GetUsage(context.Background())
	if err != nil {
		t.Fatalf("GetUsage: %v", err)
	}
	if len(usage.Records) != 1 {
		t.Fatalf("usage records = %#v", usage.Records)
	}
	record := usage.Records[0]
	if record.TaskID != "task_abc123" || record.Type != "text_to_model" || record.CreditsConsumed != 5.25 {
		t.Fatalf("usage record = %#v", record)
	}
}

func assertBodyField(t *testing.T, body map[string]any, key string, want any) {
	t.Helper()
	if got := body[key]; got != want {
		t.Fatalf("body[%q] = %#v, want %#v in %#v", key, got, want, body)
	}
}

func assertPromptView(t *testing.T, prompts []any, index int, wantPrompt string, wantView string) {
	t.Helper()
	item, ok := prompts[index].(map[string]any)
	if !ok {
		t.Fatalf("prompts[%d] = %#v, want object", index, prompts[index])
	}
	if item["prompt"] != wantPrompt {
		t.Errorf("prompts[%d].prompt = %v, want %s", index, item["prompt"], wantPrompt)
	}
	if item["view"] != wantView {
		t.Errorf("prompts[%d].view = %v, want %s", index, item["view"], wantView)
	}
}
