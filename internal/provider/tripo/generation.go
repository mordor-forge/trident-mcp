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
		"type":   "text_to_model",
		"prompt": req.Prompt,
	}
	if req.NegativePrompt != "" {
		body["negative_prompt"] = req.NegativePrompt
	}
	body["model_version"] = resolveModelVersion(req.ModelVersion)
	if req.FaceLimit > 0 {
		body["face_limit"] = req.FaceLimit
	}
	setOptionalBool(body, "texture", req.Texture)
	setOptionalBool(body, "pbr", req.PBR)
	setOptionalInt(body, "image_seed", req.ImageSeed)
	setOptionalInt(body, "model_seed", req.ModelSeed)
	setOptionalInt(body, "texture_seed", req.TextureSeed)
	setOptionalString(body, "texture_quality", req.TextureQuality)
	setOptionalBool(body, "auto_size", req.AutoSize)
	setOptionalString(body, "compress", req.Compress)
	setOptionalBool(body, "export_uv", req.ExportUV)
	setOptionalString(body, "geometry_quality", req.GeometryQuality)

	return p.createTask(ctx, body)
}

// ImageToModel creates a 3D model from a reference image.
func (p *TripoProvider) ImageToModel(ctx context.Context, req provider.ImageToModelRequest) (*provider.ModelOperation, error) {
	if req.ImagePath == "" && req.ImageURL == "" {
		return nil, fmt.Errorf("exactly one of imagePath or imageUrl is required")
	}
	if req.ImagePath != "" && req.ImageURL != "" {
		return nil, fmt.Errorf("imagePath and imageUrl are mutually exclusive")
	}

	body := map[string]any{
		"type": "image_to_model",
	}

	// Build the file reference.
	fileRef := map[string]any{}
	if req.ImagePath != "" {
		token, err := p.uploadFile(ctx, req.ImagePath)
		if err != nil {
			return nil, fmt.Errorf("uploading image: %w", err)
		}
		fileRef["type"] = fileTypeFromPath(req.ImagePath)
		fileRef["file_token"] = token
	} else {
		fileRef["type"] = fileTypeFromURL(req.ImageURL)
		fileRef["url"] = req.ImageURL
	}
	body["file"] = fileRef

	body["model_version"] = resolveModelVersion(req.ModelVersion)
	if req.FaceLimit > 0 {
		body["face_limit"] = req.FaceLimit
	}
	setOptionalBool(body, "texture", req.Texture)
	setOptionalBool(body, "pbr", req.PBR)
	setOptionalInt(body, "model_seed", req.ModelSeed)
	setOptionalInt(body, "texture_seed", req.TextureSeed)
	setOptionalString(body, "texture_quality", req.TextureQuality)
	setOptionalString(body, "texture_alignment", req.TextureAlignment)
	setOptionalBool(body, "enable_image_autofix", req.EnableImageAutofix)
	setOptionalBool(body, "auto_size", req.AutoSize)
	setOptionalString(body, "orientation", req.Orientation)
	setOptionalString(body, "compress", req.Compress)
	setOptionalBool(body, "export_uv", req.ExportUV)
	setOptionalString(body, "geometry_quality", req.GeometryQuality)

	return p.createTask(ctx, body)
}

// MultiviewToModel creates a 3D model from multiple angle images.
func (p *TripoProvider) MultiviewToModel(ctx context.Context, req provider.MultiviewToModelRequest) (*provider.ModelOperation, error) {
	paths := req.ImagePaths
	urls := req.ImageURLs

	if len(paths) == 0 && len(urls) == 0 {
		return nil, fmt.Errorf("exactly one of imagePaths or imageUrls is required")
	}
	if len(paths) > 0 && len(urls) > 0 {
		return nil, fmt.Errorf("imagePaths and imageUrls are mutually exclusive")
	}

	count := len(paths)
	if count == 0 {
		count = len(urls)
	}
	if count < 2 || count > 4 {
		return nil, fmt.Errorf("multiview requires 2-4 images, got %d", count)
	}

	version := resolveModelVersion(req.ModelVersion)
	if !multiviewVersions[version] {
		return nil, fmt.Errorf("model version %q does not support multiview input", version)
	}

	// Tripo expects a fixed 4-slot array in front/left/back/right order.
	files := make([]map[string]any, 4)
	for i := range files {
		files[i] = map[string]any{}
	}
	if len(paths) > 0 {
		for i, path := range paths {
			token, err := p.uploadFile(ctx, path)
			if err != nil {
				return nil, fmt.Errorf("uploading image %d: %w", i, err)
			}
			files[i] = map[string]any{
				"type":       fileTypeFromPath(path),
				"file_token": token,
			}
		}
	} else {
		for i, url := range urls {
			files[i] = map[string]any{
				"type": fileTypeFromURL(url),
				"url":  url,
			}
		}
	}

	body := map[string]any{
		"type":          "multiview_to_model",
		"files":         files,
		"model_version": version,
	}
	if req.FaceLimit > 0 {
		body["face_limit"] = req.FaceLimit
	}
	setOptionalBool(body, "texture", req.Texture)
	setOptionalBool(body, "pbr", req.PBR)
	setOptionalInt(body, "model_seed", req.ModelSeed)
	setOptionalInt(body, "texture_seed", req.TextureSeed)
	setOptionalString(body, "texture_quality", req.TextureQuality)
	setOptionalString(body, "texture_alignment", req.TextureAlignment)
	setOptionalBool(body, "enable_image_autofix", req.EnableImageAutofix)
	setOptionalBool(body, "auto_size", req.AutoSize)
	setOptionalString(body, "orientation", req.Orientation)
	setOptionalString(body, "compress", req.Compress)
	setOptionalBool(body, "export_uv", req.ExportUV)
	setOptionalString(body, "geometry_quality", req.GeometryQuality)

	return p.createTask(ctx, body)
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
