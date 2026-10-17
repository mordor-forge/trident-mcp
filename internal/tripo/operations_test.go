package tripo_test

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/mordor-forge/trident-mcp/internal/tripo"
	"github.com/mordor-forge/trident-mcp/internal/tripo/tripotest"
)

func ptr[T any](v T) *T { return &v }

func submit[R any](t *testing.T, fake *tripotest.Server, op tripo.Operation[R], req R) *tripo.Submission {
	t.Helper()
	sub, err := op.Submit(context.Background(), fake.Client(t), req)
	if err != nil {
		t.Fatalf("%s: %v", op.Name, err)
	}
	if sub.TaskID == "" || sub.Operation != op.Name {
		t.Fatalf("%s: unexpected submission %+v", op.Name, sub)
	}
	return sub
}

// assertBody checks that every key in want appears in got with the same
// JSON value; got may hold more keys.
func assertBody(t *testing.T, got map[string]any, want map[string]any) {
	t.Helper()
	for k, v := range want {
		if !reflect.DeepEqual(got[k], v) {
			t.Errorf("body[%q] = %#v, want %#v (body %v)", k, got[k], v, got)
		}
	}
}

func TestOperations_SendTripoPayloads(t *testing.T) {
	fake := tripotest.New(t)
	cases := []struct {
		name string
		run  func() *tripo.Submission
		path string
		want map[string]any
	}{
		{"text_to_model", func() *tripo.Submission {
			return submit(t, fake, tripo.TextToModel, tripo.TextToModelRequest{Prompt: "a brass lantern", GenerationOptions: tripo.GenerationOptions{
				FaceLimit: ptr(8000), TextureQuality: "detailed", Texture: ptr(true), ExportOrientation: "+y",
			}})
		}, "/generation/text-to-model", map[string]any{
			"prompt": "a brass lantern", "model": "v3.1-20260211", "face_limit": 8000.0, "texture_quality": "detailed", "texture": true, "export_orientation": "+y",
		}},
		{"image_to_model", func() *tripo.Submission {
			return submit(t, fake, tripo.ImageToModel, tripo.ImageToModelRequest{Input: "https://example.com/a.png", TextureAlignment: "geometry"})
		}, "/generation/image-to-model", map[string]any{"input": "https://example.com/a.png", "texture_alignment": "geometry", "model": "v3.1-20260211"}},
		{"text_to_image", func() *tripo.Submission {
			return submit(t, fake, tripo.TextToImage, tripo.TextToImageRequest{Prompt: "a sword", ImageOptions: tripo.ImageOptions{AspectRatio: "16:9", Model: "gemini-3-pro"}})
		}, "/generation/text-to-image", map[string]any{"prompt": "a sword", "model": "banana_pro", "aspect_ratio": "16:9"}},
		{"image_to_image", func() *tripo.Submission {
			return submit(t, fake, tripo.ImageToImage, tripo.ImageToImageRequest{Inputs: []tripo.Input{"task-a", "task-b"}, Prompt: "merge image[1] and image[2]"})
		}, "/generation/image-to-image", map[string]any{"inputs": []any{"task-a", "task-b"}, "model": "seedream_v5"}},
		{"image_to_multiview", func() *tripo.Submission {
			return submit(t, fake, tripo.ImageToMultiview, tripo.ImageToMultiviewRequest{Input: "task-img"})
		}, "/generation/image-to-multiview", map[string]any{"input": "task-img"}},
		{"edit_multiview", func() *tripo.Submission {
			return submit(t, fake, tripo.EditMultiview, tripo.EditMultiviewRequest{Input: "task-mv", Prompts: []tripo.MultiviewEdit{{View: "back", Prompt: "add a cape"}}})
		}, "/generation/edit-multiview", map[string]any{"input": "task-mv", "prompts": []any{map[string]any{"view": "back", "prompt": "add a cape"}}}},
		{"image_to_splat", func() *tripo.Submission {
			return submit(t, fake, tripo.ImageToSplat, tripo.ImageToSplatRequest{Input: "https://example.com/a.png", ModelSeed: ptr(7)})
		}, "/generation/image-to-splat", map[string]any{"input": "https://example.com/a.png", "model_seed": 7.0}},
		{"import_model", func() *tripo.Submission {
			return submit(t, fake, tripo.ImportModel, tripo.ImportModelRequest{Input: "https://example.com/a.gltf"})
		}, "/models/import", map[string]any{"input": "https://example.com/a.gltf"}},
		{"texture_model", func() *tripo.Submission {
			return submit(t, fake, tripo.TextureModel, tripo.TextureModelRequest{Input: "task-mesh", Model: "v3.5", TextureQuality: "fast", Compress: "geometry",
				TexturePrompt: &tripo.TexturePrompt{Text: "weathered bronze"}})
		}, "/models/texture", map[string]any{"input": "task-mesh", "model": "v3.5-20260815", "texture_quality": "fast", "compress": "geometry",
			"texture_prompt": map[string]any{"text": "weathered bronze"}}},
		{"convert_model", func() *tripo.Submission {
			return submit(t, fake, tripo.ConvertModel, tripo.ConvertModelRequest{Input: "task-mesh", Format: "FBX", FBXPreset: "mixamo", FlattenBottom: ptr(true), WithAnimation: ptr(false)})
		}, "/models/convert", map[string]any{"input": "task-mesh", "format": "FBX", "fbx_preset": "mixamo", "flatten_bottom": true, "with_animation": false}},
		{"retopologize", func() *tripo.Submission {
			return submit(t, fake, tripo.Retopologize, tripo.RetopologizeRequest{Input: "task-mesh", Model: "v3.0", Quad: ptr(true), SmartUV: ptr(true), FaceLimit: ptr(4000)})
		}, "/mesh/decimate", map[string]any{"input": "task-mesh", "model": "v3.0", "quad": true, "smart_uv": true, "face_limit": 4000.0}},
		{"smart_uv", func() *tripo.Submission {
			return submit(t, fake, tripo.SmartUV, tripo.SmartUVRequest{Input: "task-mesh"})
		}, "/mesh/uv", map[string]any{"input": "task-mesh", "model": "default"}},
		{"segment_mesh", func() *tripo.Submission {
			return submit(t, fake, tripo.SegmentMesh, tripo.SegmentMeshRequest{Input: "task-mesh", Model: "v2.0-20260430", SegmentationGranularity: "detailed"})
		}, "/mesh/segment", map[string]any{"input": "task-mesh", "model": "v2.0-20260430", "segmentation_granularity": "detailed"}},
		{"complete_mesh", func() *tripo.Submission {
			return submit(t, fake, tripo.CompleteMesh, tripo.CompleteMeshRequest{Input: "task-seg", CompletionMode: "quick_cap", PartNames: []string{"lid"}})
		}, "/mesh/complete", map[string]any{"input": "task-seg", "model": "v1.0-20250506", "completion_mode": "quick_cap", "part_names": []any{"lid"}}},
		{"smart_segment", func() *tripo.Submission {
			return submit(t, fake, tripo.SmartSegment, tripo.SmartSegmentRequest{SegType: "image", Input: "https://example.com/a.png", Hint: "wheels and body"})
		}, "/mesh/smartsegment", map[string]any{"seg_type": "image", "input": "https://example.com/a.png", "hint": "wheels and body"}},
		{"rig_check", func() *tripo.Submission {
			return submit(t, fake, tripo.RigCheck, tripo.RigCheckRequest{Input: "task-mesh"})
		}, "/animations/rig-check", map[string]any{"input": "task-mesh"}},
		{"rig_model", func() *tripo.Submission {
			return submit(t, fake, tripo.RigModel, tripo.RigModelRequest{Input: "task-mesh", Model: "rig-v3.0", RigType: "others", Spec: "unreal", OutFormat: "fbx"})
		}, "/animations/rig", map[string]any{"input": "task-mesh", "model": "v3.0-20260909", "rig_type": "others", "spec": "unreal", "out_format": "fbx"}},
		{"retarget_animation", func() *tripo.Submission {
			return submit(t, fake, tripo.RetargetAnimation, tripo.RetargetAnimationRequest{Input: "task-rig", Animations: []string{"preset:walk", "preset:run"}, AnimateInPlace: ptr(true)})
		}, "/animations/retarget", map[string]any{"input": "task-rig", "animations": []any{"preset:walk", "preset:run"}, "animate_in_place": true}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.run()
			assertBody(t, fake.Last(t, tc.path).Body, tc.want)
		})
	}
}

func TestMultiviewToModel_WiresViewsAndTaskReuse(t *testing.T) {
	fake := tripotest.New(t)

	submit(t, fake, tripo.MultiviewToModel, tripo.MultiviewToModelRequest{Inputs: &tripo.Views{
		Front: "https://example.com/f.png", Back: "file-token-b", Right: "https://example.com/r.png",
	}, GenerationOptions: tripo.GenerationOptions{Model: "p1"}})
	body := fake.Last(t, "/generation/multiview-to-model").Body
	assertBody(t, body, map[string]any{
		"model": "P1-20260311",
		"inputs": []any{
			map[string]any{"front": "https://example.com/f.png"},
			map[string]any{"back": "file-token-b"},
			map[string]any{"right": "https://example.com/r.png"},
		},
	})

	submit(t, fake, tripo.MultiviewToModel, tripo.MultiviewToModelRequest{Input: "task-mv"})
	body = fake.Last(t, "/generation/multiview-to-model").Body
	assertBody(t, body, map[string]any{"inputs": []any{map[string]any{"task_id": "task-mv"}}})
	if _, ok := body["input"]; ok {
		t.Errorf("task reuse should not also send input: %v", body)
	}
}

func TestSubmit_ModelVersions(t *testing.T) {
	fake := tripotest.New(t)

	t.Run("aliases and case resolve to IDs", func(t *testing.T) {
		sub := submit(t, fake, tripo.TextToModel, tripo.TextToModelRequest{Prompt: "x", GenerationOptions: tripo.GenerationOptions{Model: "p1-20260311"}})
		if sub.Model != "P1-20260311" {
			t.Fatalf("model = %q", sub.Model)
		}
	})
	t.Run("unset fields a version rejects are dropped", func(t *testing.T) {
		submit(t, fake, tripo.TextToModel, tripo.TextToModelRequest{Prompt: "x", GenerationOptions: tripo.GenerationOptions{Model: "P1", Quad: ptr(false)}})
		if _, sent := fake.Last(t, "/generation/text-to-model").Body["quad"]; sent {
			t.Fatal("quad:false should be dropped for P1")
		}
	})
	t.Run("set fields a version rejects fail before any request", func(t *testing.T) {
		before := len(fake.Requests())
		_, err := tripo.TextToModel.Submit(context.Background(), fake.Client(t), tripo.TextToModelRequest{Prompt: "x", GenerationOptions: tripo.GenerationOptions{Model: "p1", Quad: ptr(true)}})
		if err == nil || !strings.Contains(err.Error(), "does not support quad") {
			t.Fatalf("err = %v", err)
		}
		if len(fake.Requests()) != before {
			t.Fatal("rejected request reached Tripo")
		}
	})
	t.Run("P2 accepts quad", func(t *testing.T) {
		submit(t, fake, tripo.TextToModel, tripo.TextToModelRequest{Prompt: "x", GenerationOptions: tripo.GenerationOptions{Model: "p2", Quad: ptr(true)}})
		assertBody(t, fake.Last(t, "/generation/text-to-model").Body, map[string]any{"model": "P2-20260801", "quad": true})
	})
	t.Run("unknown versions pass through with a warning", func(t *testing.T) {
		sub := submit(t, fake, tripo.TextToModel, tripo.TextToModelRequest{Prompt: "x", GenerationOptions: tripo.GenerationOptions{Model: "v4.0-20270101"}})
		if sub.Model != "v4.0-20270101" || len(sub.Warnings) != 1 {
			t.Fatalf("submission = %+v", sub)
		}
	})
	t.Run("texture_version aliases resolve", func(t *testing.T) {
		submit(t, fake, tripo.TextToModel, tripo.TextToModelRequest{Prompt: "x", GenerationOptions: tripo.GenerationOptions{TextureVersion: "v3.5", TextureQuality: "fast"}})
		assertBody(t, fake.Last(t, "/generation/text-to-model").Body, map[string]any{"texture_version": "v3.5-20260815"})
	})
}

func TestSubmit_RejectsInvalidRequestsLocally(t *testing.T) {
	fake := tripotest.New(t)
	c := fake.Client(t)
	ctx := context.Background()
	cases := map[string]func() error{
		"missing required prompt": func() error {
			_, err := tripo.TextToModel.Submit(ctx, c, tripo.TextToModelRequest{})
			return err
		},
		"enum value outside the schema": func() error {
			_, err := tripo.TextToModel.Submit(ctx, c, tripo.TextToModelRequest{Prompt: "x", GenerationOptions: tripo.GenerationOptions{TextureQuality: "ultra"}})
			return err
		},
		"generate_parts with textures on": func() error {
			_, err := tripo.TextToModel.Submit(ctx, c, tripo.TextToModelRequest{Prompt: "x", GenerationOptions: tripo.GenerationOptions{GenerateParts: ptr(true)}})
			return err
		},
		"fast texture without v3.5": func() error {
			_, err := tripo.TextureModel.Submit(ctx, c, tripo.TextureModelRequest{Input: "task-a", TextureQuality: "fast"})
			return err
		},
		"multiview without front": func() error {
			_, err := tripo.MultiviewToModel.Submit(ctx, c, tripo.MultiviewToModelRequest{Inputs: &tripo.Views{Back: "a", Left: "b"}})
			return err
		},
		"multiview with views and task": func() error {
			_, err := tripo.MultiviewToModel.Submit(ctx, c, tripo.MultiviewToModelRequest{Inputs: &tripo.Views{Front: "a", Back: "b"}, Input: "task-mv"})
			return err
		},
		"six animations": func() error {
			_, err := tripo.RetargetAnimation.Submit(ctx, c, tripo.RetargetAnimationRequest{Input: "task-rig", Animations: []string{"preset:a", "preset:b", "preset:c", "preset:d", "preset:e", "preset:f"}})
			return err
		},
		"quadruped on rig v1.0": func() error {
			_, err := tripo.RigModel.Submit(ctx, c, tripo.RigModelRequest{Input: "task-mesh", RigType: "quadruped"})
			return err
		},
		"smart_uv on retopology v2.0": func() error {
			_, err := tripo.Retopologize.Submit(ctx, c, tripo.RetopologizeRequest{Input: "task-mesh", SmartUV: ptr(true)})
			return err
		},
		"quad GLTF conversion": func() error {
			_, err := tripo.ConvertModel.Submit(ctx, c, tripo.ConvertModelRequest{Input: "task-mesh", Format: "GLTF", Quad: ptr(true)})
			return err
		},
		"edit_multiview without prompts": func() error {
			_, err := tripo.EditMultiview.Submit(ctx, c, tripo.EditMultiviewRequest{Input: "task-mv", Prompts: []tripo.MultiviewEdit{}})
			return err
		},
	}
	for name, run := range cases {
		t.Run(name, func(t *testing.T) {
			if err := run(); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
	if reqs := fake.Requests(); len(reqs) != 0 {
		t.Fatalf("invalid requests reached Tripo: %v", reqs)
	}
}

func writeFile(t *testing.T, name string, size int) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, make([]byte, size), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestSubmit_UploadsLocalInputs(t *testing.T) {
	fake := tripotest.New(t)
	img := writeFile(t, "ref.png", 1024)

	submit(t, fake, tripo.ImageToModel, tripo.ImageToModelRequest{Input: tripo.Input(img)})
	upload := fake.Last(t, "/files")
	if upload.FileName != "ref.png" || upload.FileBytes != 1024 {
		t.Fatalf("upload = %+v", upload)
	}
	token := fake.Last(t, "/generation/image-to-model").Body["input"].(string)
	if !strings.HasPrefix(token, "file-") {
		t.Fatalf("input = %q, want the uploaded file token", token)
	}

	// Inputs nested in objects and lists are uploaded too, once per path.
	submit(t, fake, tripo.MultiviewToModel, tripo.MultiviewToModelRequest{Inputs: &tripo.Views{Front: tripo.Input(img), Back: tripo.Input(img)}})
	views := fake.Last(t, "/generation/multiview-to-model").Body["inputs"].([]any)
	front := views[0].(map[string]any)["front"]
	back := views[1].(map[string]any)["back"]
	if front != back || !strings.HasPrefix(front.(string), "file-") {
		t.Fatalf("views = %v", views)
	}
	submit(t, fake, tripo.TextureModel, tripo.TextureModelRequest{Input: "task-mesh", TexturePrompt: &tripo.TexturePrompt{Image: tripo.Input(img)}})
	prompt := fake.Last(t, "/models/texture").Body["texture_prompt"].(map[string]any)
	if !strings.HasPrefix(prompt["image"].(string), "file-") {
		t.Fatalf("texture_prompt = %v", prompt)
	}
}

func TestSubmit_RefusesBadLocalInputs(t *testing.T) {
	fake := tripotest.New(t)
	c := fake.Client(t)
	secret := writeFile(t, "id_rsa.txt", 10)
	for name, input := range map[string]string{
		"missing file":      filepath.Join(t.TempDir(), "nope.png"),
		"non-media file":    secret,
		"directory":         t.TempDir() + "/",
		"oversized image":   writeFile(t, "big.png", 21<<20),
		"relative not here": "missing-ref.jpg",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := tripo.ImageToModel.Submit(context.Background(), c, tripo.ImageToModelRequest{Input: tripo.Input(input)}); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
	if reqs := fake.Requests(); len(reqs) != 0 {
		t.Fatalf("requests reached Tripo: %v", reqs)
	}
}

func TestUploadFile_LargeFilesUsePresign(t *testing.T) {
	fake := tripotest.New(t)
	path := filepath.Join(t.TempDir(), "big.glb")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Truncate(51 << 20); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	up, err := fake.Client(t).UploadFile(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(up.FileToken, "file-") || up.Bytes != 51<<20 {
		t.Fatalf("upload = %+v", up)
	}
	assertBody(t, fake.Last(t, "/files/presign").Body, map[string]any{"format": "glb"})
}

func TestInputSchema_CarriesEnumsAndDescriptions(t *testing.T) {
	s, err := tripo.TextToModel.InputSchema()
	if err != nil {
		t.Fatal(err)
	}
	quality := s.Properties["texture_quality"]
	if quality == nil || len(quality.Enum) != 4 || quality.Description == "" {
		t.Fatalf("texture_quality schema = %+v", quality)
	}
	if !reflect.DeepEqual(s.Required, []string{"prompt"}) {
		t.Fatalf("required = %v", s.Required)
	}

	edit, err := tripo.EditMultiview.InputSchema()
	if err != nil {
		t.Fatal(err)
	}
	view := edit.Properties["prompts"].Items.Properties["view"]
	if len(view.Enum) != 4 {
		t.Fatalf("nested view enum = %v", view.Enum)
	}
}
