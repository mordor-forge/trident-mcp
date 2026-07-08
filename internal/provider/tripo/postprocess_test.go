package tripo

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/mordor-forge/trident-mcp/internal/provider"
)

func TestRetopologize_Success(t *testing.T) {
	var captured map[string]any
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/mesh/decimate" {
			t.Fatalf("request = %s %s, want POST /mesh/decimate", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"task_id": "retopo-task-1"},
		})
	})

	p := newTestProvider(t, handler)
	op, err := p.Retopologize(context.Background(), provider.RetopologyRequest{
		OriginalTaskID: "orig-task",
		Model:          "P1-20260311",
		Quad:           true,
		TargetFaces:    4000,
		PartNames:      []string{"body", "handle"},
		Bake:           boolPtr(true),
	})
	if err != nil {
		t.Fatalf("Retopologize: %v", err)
	}
	if op.TaskID != "retopo-task-1" {
		t.Errorf("TaskID = %q", op.TaskID)
	}

	if _, ok := captured["type"]; ok {
		t.Errorf("v3 request must not include legacy type field: %v", captured["type"])
	}
	if captured["input"] != "orig-task" {
		t.Errorf("input = %v", captured["input"])
	}
	if captured["quad"] != true {
		t.Errorf("quad = %v, want true", captured["quad"])
	}
	if captured["face_limit"] != float64(4000) {
		t.Errorf("face_limit = %v, want 4000", captured["face_limit"])
	}
	if captured["model"] != "P1-20260311" {
		t.Errorf("model = %v, want P1-20260311", captured["model"])
	}
	if captured["bake"] != true {
		t.Errorf("bake = %v, want true", captured["bake"])
	}
	assertStringSliceBodyField(t, captured, "part_names", []string{"body", "handle"})
}

func TestRetopologize_TriangleMode(t *testing.T) {
	var captured map[string]any
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/mesh/decimate" {
			t.Fatalf("request = %s %s, want POST /mesh/decimate", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"task_id": "retopo-task-2"},
		})
	})

	p := newTestProvider(t, handler)
	_, err := p.Retopologize(context.Background(), provider.RetopologyRequest{
		OriginalTaskID: "orig-task",
		Quad:           false,
	})
	if err != nil {
		t.Fatalf("Retopologize: %v", err)
	}

	if captured["quad"] != false {
		t.Errorf("quad = %v, want false", captured["quad"])
	}
}

func TestRetopologize_MissingTaskID(t *testing.T) {
	p, _ := New(Config{APIKey: "tsk_test"})
	_, err := p.Retopologize(context.Background(), provider.RetopologyRequest{})
	if err == nil {
		t.Fatal("expected error for missing task ID")
	}
}

func TestConvertFormat_Success(t *testing.T) {
	var captured map[string]any
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/models/convert" {
			t.Fatalf("request = %s %s, want POST /models/convert", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"task_id": "conv-task-1"},
		})
	})

	p := newTestProvider(t, handler)
	op, err := p.ConvertFormat(context.Background(), provider.ConvertRequest{
		OriginalTaskID:      "orig-task",
		Format:              "FBX",
		Quad:                boolPtr(true),
		FaceLimit:           3000,
		TextureSize:         2048,
		TextureFormat:       "png",
		Bake:                boolPtr(true),
		PackUV:              boolPtr(false),
		ExportVertexColors:  boolPtr(true),
		PivotToCenterBottom: boolPtr(true),
		ScaleFactor:         0.5,
		PartNames:           []string{"body", "wheel"},
		ExportOrientation:   "y_up",
		FBXPreset:           "unity",
	})
	if err != nil {
		t.Fatalf("ConvertFormat: %v", err)
	}
	if op.TaskID != "conv-task-1" {
		t.Errorf("TaskID = %q", op.TaskID)
	}

	if _, ok := captured["type"]; ok {
		t.Errorf("v3 request must not include legacy type field: %v", captured["type"])
	}
	if captured["input"] != "orig-task" {
		t.Errorf("input = %v", captured["input"])
	}
	if captured["format"] != "FBX" {
		t.Errorf("format = %v, want FBX", captured["format"])
	}
	if captured["quad"] != true {
		t.Errorf("quad = %v, want true", captured["quad"])
	}
	if captured["face_limit"] != float64(3000) {
		t.Errorf("face_limit = %v, want 3000", captured["face_limit"])
	}
	if captured["texture_size"] != float64(2048) {
		t.Errorf("texture_size = %v, want 2048", captured["texture_size"])
	}
	if captured["texture_format"] != "png" {
		t.Errorf("texture_format = %v, want png", captured["texture_format"])
	}
	if captured["bake"] != true {
		t.Errorf("bake = %v, want true", captured["bake"])
	}
	if captured["pack_uv"] != false {
		t.Errorf("pack_uv = %v, want false", captured["pack_uv"])
	}
	if captured["export_vertex_colors"] != true {
		t.Errorf("export_vertex_colors = %v, want true", captured["export_vertex_colors"])
	}
	if captured["pivot_to_center_bottom"] != true {
		t.Errorf("pivot_to_center_bottom = %v, want true", captured["pivot_to_center_bottom"])
	}
	if captured["scale_factor"] != 0.5 {
		t.Errorf("scale_factor = %v, want 0.5", captured["scale_factor"])
	}
	assertStringSliceBodyField(t, captured, "part_names", []string{"body", "wheel"})
	if captured["export_orientation"] != "+y" {
		t.Errorf("export_orientation = %v, want +y", captured["export_orientation"])
	}
	if captured["fbx_preset"] != "unity" {
		t.Errorf("fbx_preset = %v, want unity", captured["fbx_preset"])
	}
}

func TestConvertFormat_NormalizesExportOrientationAliases(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"+x", "+x"},
		{"-x", "-x"},
		{"+y", "+y"},
		{"-y", "-y"},
		{"+X", "+x"},
		{" +y ", "+y"},
		{"x_up", "+x"},
		{"y_up", "+y"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var captured map[string]any
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(body, &captured)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"code": 0,
					"data": map[string]any{"task_id": "conv-task-orientation"},
				})
			})

			p := newTestProvider(t, handler)
			_, err := p.ConvertFormat(context.Background(), provider.ConvertRequest{
				OriginalTaskID:    "orig-task",
				Format:            "FBX",
				ExportOrientation: tt.input,
			})
			if err != nil {
				t.Fatalf("ConvertFormat: %v", err)
			}
			if captured["export_orientation"] != tt.want {
				t.Errorf("export_orientation = %v, want %s", captured["export_orientation"], tt.want)
			}
		})
	}
}

func TestConvertFormat_InvalidExportOrientation(t *testing.T) {
	p, _ := New(Config{APIKey: "tsk_test"})
	_, err := p.ConvertFormat(context.Background(), provider.ConvertRequest{
		OriginalTaskID:    "orig-task",
		Format:            "FBX",
		ExportOrientation: "z_up",
	})
	if err == nil {
		t.Fatal("expected error for invalid export orientation")
	}
	if !strings.Contains(err.Error(), "unsupported exportOrientation") {
		t.Errorf("error %q doesn't mention exportOrientation", err)
	}
}

func TestConvertFormat_InvalidFormat(t *testing.T) {
	p, _ := New(Config{APIKey: "tsk_test"})
	_, err := p.ConvertFormat(context.Background(), provider.ConvertRequest{
		OriginalTaskID: "orig-task",
		Format:         "INVALID",
	})
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("error %q doesn't mention format", err)
	}
}

func TestConvertFormat_AllValidFormats(t *testing.T) {
	handler := taskCreatedHandler("convert_model")

	for _, format := range []string{"GLTF", "FBX", "OBJ", "STL", "USDZ", "3MF"} {
		t.Run(format, func(t *testing.T) {
			p := newTestProvider(t, handler)
			_, err := p.ConvertFormat(context.Background(), provider.ConvertRequest{
				OriginalTaskID: "task-1",
				Format:         format,
			})
			if err != nil {
				t.Fatalf("ConvertFormat(%s): %v", format, err)
			}
		})
	}
}

func TestConvertFormat_NormalizesAliases(t *testing.T) {
	var captured map[string]any
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/models/convert" {
			t.Fatalf("request = %s %s, want POST /models/convert", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"task_id": "conv-task-alias"},
		})
	})

	p := newTestProvider(t, handler)
	_, err := p.ConvertFormat(context.Background(), provider.ConvertRequest{
		OriginalTaskID: "orig-task",
		Format:         "glb",
	})
	if err != nil {
		t.Fatalf("ConvertFormat: %v", err)
	}
	if captured["format"] != "GLTF" {
		t.Errorf("format = %v, want GLTF", captured["format"])
	}
}

func TestConvertFormat_MissingFields(t *testing.T) {
	p, _ := New(Config{APIKey: "tsk_test"})

	// Missing task ID.
	_, err := p.ConvertFormat(context.Background(), provider.ConvertRequest{Format: "FBX"})
	if err == nil {
		t.Fatal("expected error for missing task ID")
	}

	// Missing format.
	_, err = p.ConvertFormat(context.Background(), provider.ConvertRequest{OriginalTaskID: "task-1"})
	if err == nil {
		t.Fatal("expected error for missing format")
	}
}

func TestStylize_Success(t *testing.T) {
	var captured map[string]any
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"task_id": "style-task-1"},
		})
	})

	p := newTestProvider(t, handler)
	op, err := p.Stylize(context.Background(), provider.StylizeRequest{
		OriginalTaskID: "orig-task",
		Style:          "minecraft",
		BlockSize:      80,
	})
	if err != nil {
		t.Fatalf("Stylize: %v", err)
	}
	if op.TaskID != "style-task-1" {
		t.Errorf("TaskID = %q", op.TaskID)
	}

	if _, ok := captured["type"]; ok {
		t.Errorf("v3 request must not include legacy type field: %v", captured["type"])
	}
	if captured["input"] != "orig-task" {
		t.Errorf("input = %v", captured["input"])
	}
	if captured["style"] != "minecraft" {
		t.Errorf("style = %v, want minecraft", captured["style"])
	}
	if captured["block_size"] != float64(80) {
		t.Errorf("block_size = %v, want 80", captured["block_size"])
	}
}

func assertStringSliceBodyField(t *testing.T, body map[string]any, key string, want []string) {
	t.Helper()

	gotRaw, ok := body[key].([]any)
	if !ok {
		t.Fatalf("%s = %#v, want string slice", key, body[key])
	}
	if len(gotRaw) != len(want) {
		t.Fatalf("%s length = %d, want %d", key, len(gotRaw), len(want))
	}
	for i, value := range want {
		if gotRaw[i] != value {
			t.Fatalf("%s[%d] = %#v, want %#v", key, i, gotRaw[i], value)
		}
	}
}

func TestStylize_AllValidStyles(t *testing.T) {
	handler := taskCreatedHandler("stylize_model")

	for _, style := range []string{"lego", "voxel", "voronoi", "minecraft"} {
		t.Run(style, func(t *testing.T) {
			p := newTestProvider(t, handler)
			_, err := p.Stylize(context.Background(), provider.StylizeRequest{
				OriginalTaskID: "task-1",
				Style:          style,
			})
			if err != nil {
				t.Fatalf("Stylize(%s): %v", style, err)
			}
		})
	}
}

func TestStylize_NormalizesCase(t *testing.T) {
	var captured map[string]any
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"task_id": "style-task-upper"},
		})
	})

	p := newTestProvider(t, handler)
	_, err := p.Stylize(context.Background(), provider.StylizeRequest{
		OriginalTaskID: "orig-task",
		Style:          "LEGO",
	})
	if err != nil {
		t.Fatalf("Stylize: %v", err)
	}
	if captured["style"] != "lego" {
		t.Errorf("style = %v, want lego", captured["style"])
	}
}

func TestStylize_InvalidStyle(t *testing.T) {
	p, _ := New(Config{APIKey: "tsk_test"})
	_, err := p.Stylize(context.Background(), provider.StylizeRequest{
		OriginalTaskID: "task-1",
		Style:          "cartoon",
	})
	if err == nil {
		t.Fatal("expected error for invalid style")
	}
	if !strings.Contains(err.Error(), "unsupported style") {
		t.Errorf("error %q doesn't mention style", err)
	}
}

func TestStylize_MissingFields(t *testing.T) {
	p, _ := New(Config{APIKey: "tsk_test"})

	_, err := p.Stylize(context.Background(), provider.StylizeRequest{Style: "lego"})
	if err == nil {
		t.Fatal("expected error for missing task ID")
	}

	_, err = p.Stylize(context.Background(), provider.StylizeRequest{OriginalTaskID: "task-1"})
	if err == nil {
		t.Fatal("expected error for missing style")
	}
}

func TestPostProcess_InvalidOriginalTaskID(t *testing.T) {
	p, _ := New(Config{APIKey: "tsk_test"})

	tests := []struct {
		name string
		run  func() error
	}{
		{
			name: "retopologize",
			run: func() error {
				_, err := p.Retopologize(context.Background(), provider.RetopologyRequest{
					OriginalTaskID: "task with spaces",
				})
				return err
			},
		},
		{
			name: "convert",
			run: func() error {
				_, err := p.ConvertFormat(context.Background(), provider.ConvertRequest{
					OriginalTaskID: "task;bad",
					Format:         "FBX",
				})
				return err
			},
		},
		{
			name: "stylize",
			run: func() error {
				_, err := p.Stylize(context.Background(), provider.StylizeRequest{
					OriginalTaskID: "../task",
					Style:          "lego",
				})
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.run()
			if err == nil {
				t.Fatal("expected invalid task ID error")
			}
			if !strings.Contains(err.Error(), "invalid taskID") {
				t.Errorf("error %q does not mention invalid taskID", err)
			}
		})
	}
}
