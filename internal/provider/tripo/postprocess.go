package tripo

import (
	"context"
	"fmt"

	"github.com/mordor-forge/trident-mcp/internal/provider"
)

// Retopologize creates a lowpoly version of a model.
// The Tripo API uses the "highpoly_to_lowpoly" task type, with the quad
// parameter controlling whether to produce quad or triangle mesh output.
func (p *TripoProvider) Retopologize(ctx context.Context, req provider.RetopologyRequest) (*provider.ModelOperation, error) {
	if err := validateTaskID(req.OriginalTaskID); err != nil {
		return nil, fmt.Errorf("originalTaskId: %w", err)
	}

	body := map[string]any{
		"type":                   "highpoly_to_lowpoly",
		"original_model_task_id": req.OriginalTaskID,
		"quad":                   req.Quad,
	}
	if req.TargetFaces > 0 {
		body["face_limit"] = req.TargetFaces
	}

	return p.createTask(ctx, body)
}

// ConvertFormat converts a model to a different file format.
func (p *TripoProvider) ConvertFormat(ctx context.Context, req provider.ConvertRequest) (*provider.ModelOperation, error) {
	if err := validateTaskID(req.OriginalTaskID); err != nil {
		return nil, fmt.Errorf("originalTaskId: %w", err)
	}
	if req.Format == "" {
		return nil, fmt.Errorf("format is required")
	}
	format, ok := normalizeFormat(req.Format)
	if !ok {
		return nil, fmt.Errorf("unsupported format %q (valid: GLTF, FBX, OBJ, STL, USDZ, 3MF)", req.Format)
	}

	body := map[string]any{
		"type":                   "convert_model",
		"original_model_task_id": req.OriginalTaskID,
		"format":                 format,
	}

	return p.createTask(ctx, body)
}

// Stylize applies a stylization effect to a model.
func (p *TripoProvider) Stylize(ctx context.Context, req provider.StylizeRequest) (*provider.ModelOperation, error) {
	if err := validateTaskID(req.OriginalTaskID); err != nil {
		return nil, fmt.Errorf("originalTaskId: %w", err)
	}
	if req.Style == "" {
		return nil, fmt.Errorf("style is required")
	}
	style := normalizeStyle(req.Style)
	if !validStyles[style] {
		return nil, fmt.Errorf("unsupported style %q (valid: lego, voxel, voronoi, minecraft)", req.Style)
	}

	body := map[string]any{
		"type":                   "stylize_model",
		"original_model_task_id": req.OriginalTaskID,
		"style":                  style,
	}

	return p.createTask(ctx, body)
}
