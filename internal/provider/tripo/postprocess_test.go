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
		Quad:           true,
		TargetFaces:    4000,
	})
	if err != nil {
		t.Fatalf("Retopologize: %v", err)
	}
	if op.TaskID != "retopo-task-1" {
		t.Errorf("TaskID = %q", op.TaskID)
	}

	if captured["type"] != "highpoly_to_lowpoly" {
		t.Errorf("type = %v, want highpoly_to_lowpoly", captured["type"])
	}
	if captured["original_model_task_id"] != "orig-task" {
		t.Errorf("original_model_task_id = %v", captured["original_model_task_id"])
	}
	if captured["quad"] != true {
		t.Errorf("quad = %v, want true", captured["quad"])
	}
	if captured["face_limit"] != float64(4000) {
		t.Errorf("face_limit = %v, want 4000", captured["face_limit"])
	}
}

func TestRetopologize_TriangleMode(t *testing.T) {
	var captured map[string]any
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"task_id": "conv-task-1"},
		})
	})

	p := newTestProvider(t, handler)
	op, err := p.ConvertFormat(context.Background(), provider.ConvertRequest{
		OriginalTaskID: "orig-task",
		Format:         "FBX",
	})
	if err != nil {
		t.Fatalf("ConvertFormat: %v", err)
	}
	if op.TaskID != "conv-task-1" {
		t.Errorf("TaskID = %q", op.TaskID)
	}

	if captured["type"] != "convert_model" {
		t.Errorf("type = %v, want convert_model", captured["type"])
	}
	if captured["format"] != "FBX" {
		t.Errorf("format = %v, want FBX", captured["format"])
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
		Style:          "lego",
	})
	if err != nil {
		t.Fatalf("Stylize: %v", err)
	}
	if op.TaskID != "style-task-1" {
		t.Errorf("TaskID = %q", op.TaskID)
	}

	if captured["type"] != "stylize_model" {
		t.Errorf("type = %v, want stylize_model", captured["type"])
	}
	if captured["style"] != "lego" {
		t.Errorf("style = %v, want lego", captured["style"])
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
