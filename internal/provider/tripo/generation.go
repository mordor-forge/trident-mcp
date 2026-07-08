package tripo

import (
	"context"
	"fmt"
	"strings"

	"github.com/mordor-forge/trident-mcp/internal/provider"
)

// TextToModel creates a 3D model from a text prompt.
func (p *TripoProvider) TextToModel(ctx context.Context, req provider.TextToModelRequest) (*provider.ModelOperation, error) {
	if req.Prompt == "" {
		return nil, fmt.Errorf("prompt is required")
	}

	body := map[string]any{
		"prompt": req.Prompt,
	}
	model := resolveModel(req.ModelVersion)
	body["model"] = model
	if req.NegativePrompt != "" {
		body["negative_prompt"] = req.NegativePrompt
	}
	if req.FaceLimit > 0 {
		body["face_limit"] = req.FaceLimit
	}
	setOptionalBool(body, "texture", req.Texture)
	setOptionalBool(body, "pbr", req.PBR)
	setOptionalInt(body, "image_seed", req.ImageSeed)
	setOptionalInt(body, "model_seed", req.ModelSeed)
	setOptionalInt(body, "texture_seed", req.TextureSeed)
	setOptionalString(body, "texture_quality", req.TextureQuality)
	setNonP1Bool(body, "quad", model, req.Quad)
	setNonP1Bool(body, "smart_low_poly", model, req.SmartLowPoly)
	setNonP1Bool(body, "generate_parts", model, req.GenerateParts)
	setOptionalBool(body, "auto_size", req.AutoSize)
	setOptionalString(body, "compress", req.Compress)
	setOptionalBool(body, "export_uv", req.ExportUV)
	setNonP1String(body, "geometry_quality", model, req.GeometryQuality)

	return p.createTask(ctx, "/generation/text-to-model", body)
}

// ImageToModel creates a 3D model from a reference image.
func (p *TripoProvider) ImageToModel(ctx context.Context, req provider.ImageToModelRequest) (*provider.ModelOperation, error) {
	if req.ImagePath == "" && req.ImageURL == "" {
		return nil, fmt.Errorf("exactly one of imagePath or imageUrl is required")
	}
	if req.ImagePath != "" && req.ImageURL != "" {
		return nil, fmt.Errorf("imagePath and imageUrl are mutually exclusive")
	}

	body := map[string]any{}

	// Build the file reference.
	if req.ImagePath != "" {
		token, err := p.uploadFile(ctx, req.ImagePath)
		if err != nil {
			return nil, fmt.Errorf("uploading image: %w", err)
		}
		body["input"] = token
	} else {
		body["input"] = req.ImageURL
	}

	model := resolveModel(req.ModelVersion)
	body["model"] = model
	if req.FaceLimit > 0 {
		body["face_limit"] = req.FaceLimit
	}
	setOptionalBool(body, "texture", req.Texture)
	setOptionalBool(body, "pbr", req.PBR)
	setOptionalInt(body, "model_seed", req.ModelSeed)
	setOptionalInt(body, "texture_seed", req.TextureSeed)
	setOptionalString(body, "texture_quality", req.TextureQuality)
	setNonP1Bool(body, "quad", model, req.Quad)
	setNonP1Bool(body, "smart_low_poly", model, req.SmartLowPoly)
	setNonP1Bool(body, "generate_parts", model, req.GenerateParts)
	setOptionalString(body, "texture_alignment", req.TextureAlignment)
	setOptionalBool(body, "enable_image_autofix", req.EnableImageAutofix)
	setOptionalBool(body, "auto_size", req.AutoSize)
	setOptionalString(body, "orientation", req.Orientation)
	setOptionalString(body, "compress", req.Compress)
	setOptionalBool(body, "export_uv", req.ExportUV)
	setNonP1String(body, "geometry_quality", model, req.GeometryQuality)

	return p.createTask(ctx, "/generation/image-to-model", body)
}

// MultiviewToModel creates a 3D model from multiple angle images.
func (p *TripoProvider) MultiviewToModel(ctx context.Context, req provider.MultiviewToModelRequest) (*provider.ModelOperation, error) {
	paths := req.ImagePaths
	urls := req.ImageURLs
	taskID := strings.TrimSpace(req.TaskID)

	sourceCount := 0
	if len(paths) > 0 {
		sourceCount++
	}
	if len(urls) > 0 {
		sourceCount++
	}
	if taskID != "" {
		sourceCount++
	}
	if sourceCount == 0 {
		return nil, fmt.Errorf("exactly one of imagePaths, imageUrls, or taskId is required")
	}
	if sourceCount > 1 {
		return nil, fmt.Errorf("imagePaths, imageUrls, and taskId are mutually exclusive")
	}

	count := len(paths)
	if len(urls) > 0 {
		count = len(urls)
	}
	if taskID == "" && (count < 2 || count > 4) {
		return nil, fmt.Errorf("multiview requires 2-4 images, got %d", count)
	}

	model := resolveModel(req.ModelVersion)
	if !multiviewModels[model] {
		return nil, fmt.Errorf("model %q does not support multiview input", model)
	}

	var inputs []any
	if taskID != "" {
		if err := validateTaskID(taskID); err != nil {
			return nil, fmt.Errorf("taskId: %w", err)
		}
		inputs = []any{map[string]any{"task_id": taskID}}
	} else if len(paths) > 0 {
		values := make([]string, 0, count)
		for i, path := range paths {
			token, err := p.uploadFile(ctx, path)
			if err != nil {
				return nil, fmt.Errorf("uploading image %d: %w", i, err)
			}
			values = append(values, token)
		}
		inputs = orderedMultiviewInputs(values)
	} else {
		inputs = orderedMultiviewInputs(urls)
	}

	body := map[string]any{
		"inputs": inputs,
		"model":  model,
	}
	if req.FaceLimit > 0 {
		body["face_limit"] = req.FaceLimit
	}
	setOptionalBool(body, "texture", req.Texture)
	setOptionalBool(body, "pbr", req.PBR)
	setOptionalInt(body, "model_seed", req.ModelSeed)
	setOptionalInt(body, "texture_seed", req.TextureSeed)
	setOptionalString(body, "texture_quality", req.TextureQuality)
	setNonP1Bool(body, "quad", model, req.Quad)
	setNonP1Bool(body, "smart_low_poly", model, req.SmartLowPoly)
	setNonP1Bool(body, "generate_parts", model, req.GenerateParts)
	setOptionalString(body, "texture_alignment", req.TextureAlignment)
	setOptionalBool(body, "enable_image_autofix", req.EnableImageAutofix)
	setOptionalBool(body, "auto_size", req.AutoSize)
	setOptionalString(body, "orientation", req.Orientation)
	setOptionalString(body, "compress", req.Compress)
	setOptionalBool(body, "export_uv", req.ExportUV)
	setNonP1String(body, "geometry_quality", model, req.GeometryQuality)

	return p.createTask(ctx, "/generation/multiview-to-model", body)
}

func setOptionalBool(body map[string]any, key string, value *bool) {
	if value != nil {
		body[key] = *value
	}
}

func setOptionalInt(body map[string]any, key string, value *int) {
	if value != nil {
		body[key] = *value
	}
}

func setOptionalString(body map[string]any, key, value string) {
	if value = strings.TrimSpace(value); value != "" {
		body[key] = value
	}
}

func setNonP1Bool(body map[string]any, key, model string, value *bool) {
	if value == nil || model == "P1-20260311" {
		return
	}
	body[key] = *value
}

func setNonP1String(body map[string]any, key, model, value string) {
	if model == "P1-20260311" {
		return
	}
	setOptionalString(body, key, value)
}

func orderedMultiviewInputs(values []string) []any {
	views := []string{"front", "left", "back", "right"}
	inputs := make([]any, 0, len(values))
	for i, value := range values {
		inputs = append(inputs, map[string]any{views[i]: value})
	}
	return inputs
}
