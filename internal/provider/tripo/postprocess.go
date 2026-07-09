package tripo

import (
	"context"
	"fmt"
	"strings"

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
		"input": req.OriginalTaskID,
		"quad":  req.Quad,
	}
	if req.TargetFaces > 0 {
		body["face_limit"] = req.TargetFaces
	}
	setOptionalString(body, "model", req.Model)
	setOptionalBool(body, "bake", req.Bake)
	if len(req.PartNames) > 0 {
		body["part_names"] = req.PartNames
	}

	return p.createTask(ctx, "/mesh/decimate", body)
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
		"input":  req.OriginalTaskID,
		"format": format,
	}
	setOptionalBool(body, "quad", req.Quad)
	setOptionalIntValue(body, "face_limit", req.FaceLimit)
	setOptionalIntValue(body, "texture_size", req.TextureSize)
	setOptionalString(body, "texture_format", req.TextureFormat)
	setOptionalBool(body, "bake", req.Bake)
	setOptionalBool(body, "pack_uv", req.PackUV)
	setOptionalBool(body, "export_vertex_colors", req.ExportVertexColors)
	setOptionalBool(body, "pivot_to_center_bottom", req.PivotToCenterBottom)
	setOptionalFloatValue(body, "scale_factor", req.ScaleFactor)
	exportOrientation, ok := normalizeExportOrientation(req.ExportOrientation)
	if !ok {
		return nil, fmt.Errorf("unsupported exportOrientation %q (valid: +x, -x, +y, -y)", req.ExportOrientation)
	}
	setOptionalString(body, "export_orientation", exportOrientation)
	setOptionalString(body, "fbx_preset", req.FBXPreset)
	if len(req.PartNames) > 0 {
		body["part_names"] = req.PartNames
	}

	return p.createTask(ctx, "/models/convert", body)
}

func normalizeExportOrientation(value string) (string, bool) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "":
		return "", true
	case "+x", "-x", "+y", "-y":
		return normalized, true
	case "x_up":
		return "+x", true
	case "y_up":
		return "+y", true
	default:
		return "", false
	}
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
		"input": req.OriginalTaskID,
		"style": style,
	}
	setOptionalIntValue(body, "block_size", req.BlockSize)

	return p.createTask(ctx, "/models/stylize", body)
}
