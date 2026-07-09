package tripo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/mordor-forge/trident-mcp/internal/provider"
)

// TextToImage creates an image from a text prompt.
func (p *TripoProvider) TextToImage(ctx context.Context, req provider.TextToImageRequest) (*provider.ModelOperation, error) {
	if strings.TrimSpace(req.Prompt) == "" {
		return nil, fmt.Errorf("prompt is required")
	}
	body := map[string]any{"prompt": req.Prompt}
	setOptionalString(body, "model", req.Model)
	setOptionalString(body, "negative_prompt", req.NegativePrompt)
	setOptionalString(body, "template", req.Template)
	setOptionalString(body, "size", req.Size)
	setOptionalIntValue(body, "width", req.Width)
	setOptionalIntValue(body, "height", req.Height)
	setOptionalInt(body, "seed", req.Seed)
	return p.createTask(ctx, "/generation/text-to-image", body)
}

// ImageToImage creates or edits an image from a source image.
func (p *TripoProvider) ImageToImage(ctx context.Context, req provider.ImageToImageRequest) (*provider.ModelOperation, error) {
	input, err := p.resolveInput(ctx, req.Input, req.ImagePath, req.ImageURL)
	if err != nil {
		return nil, fmt.Errorf("input: %w", err)
	}
	body := map[string]any{"input": input}
	setOptionalString(body, "prompt", req.Prompt)
	setOptionalString(body, "model", req.Model)
	setOptionalString(body, "negative_prompt", req.NegativePrompt)
	setOptionalString(body, "template", req.Template)
	setOptionalString(body, "size", req.Size)
	setOptionalIntValue(body, "width", req.Width)
	setOptionalIntValue(body, "height", req.Height)
	setOptionalInt(body, "seed", req.Seed)
	return p.createTask(ctx, "/generation/image-to-image", body)
}

// ImageToMultiview creates a set of multiview images from one source image.
func (p *TripoProvider) ImageToMultiview(ctx context.Context, req provider.ImageToMultiviewRequest) (*provider.ModelOperation, error) {
	input, err := p.resolveInput(ctx, req.Input, req.ImagePath, req.ImageURL)
	if err != nil {
		return nil, fmt.Errorf("input: %w", err)
	}
	return p.createTask(ctx, "/generation/image-to-multiview", map[string]any{"input": input})
}

// EditMultiview edits an existing multiview image set.
func (p *TripoProvider) EditMultiview(ctx context.Context, req provider.EditMultiviewRequest) (*provider.ModelOperation, error) {
	if strings.TrimSpace(req.Input) == "" {
		return nil, fmt.Errorf("input is required")
	}
	body := map[string]any{"input": req.Input}
	prompts, err := multiviewEditPrompts(req)
	if err != nil {
		return nil, err
	}
	body["prompts"] = prompts
	return p.createTask(ctx, "/generation/edit-multiview", body)
}

// ImageToSplat creates a Gaussian Splat from a source image.
func (p *TripoProvider) ImageToSplat(ctx context.Context, req provider.ImageToSplatRequest) (*provider.ModelOperation, error) {
	input, err := p.resolveInput(ctx, req.Input, req.ImagePath, req.ImageURL)
	if err != nil {
		return nil, fmt.Errorf("input: %w", err)
	}
	return p.createTask(ctx, "/generation/image-to-splat", map[string]any{"input": input})
}

// ImportModel imports an external model into Tripo.
func (p *TripoProvider) ImportModel(ctx context.Context, req provider.ImportModelRequest) (*provider.ModelOperation, error) {
	input, err := p.resolveInput(ctx, req.Input, req.FilePath, req.FileURL)
	if err != nil {
		return nil, fmt.Errorf("input: %w", err)
	}
	return p.createTask(ctx, "/models/import", map[string]any{"input": input})
}

// RefineModel refines an existing model.
func (p *TripoProvider) RefineModel(ctx context.Context, req provider.RefineModelRequest) (*provider.ModelOperation, error) {
	body, err := inputBody(req.Input)
	if err != nil {
		return nil, err
	}
	setOptionalString(body, "model", req.Model)
	return p.createTask(ctx, "/models/refine", body)
}

// TextureModel generates texture for an existing model.
func (p *TripoProvider) TextureModel(ctx context.Context, req provider.TextureModelRequest) (*provider.ModelOperation, error) {
	body, err := inputBody(req.Input)
	if err != nil {
		return nil, err
	}
	setOptionalString(body, "model", req.Model)
	setOptionalBool(body, "texture", req.Texture)
	setOptionalBool(body, "pbr", req.PBR)
	setOptionalInt(body, "texture_seed", req.TextureSeed)
	setOptionalString(body, "texture_quality", req.TextureQuality)
	setOptionalString(body, "texture_alignment", req.TextureAlignment)
	setOptionalString(body, "text_prompt", req.TextPrompt)
	setOptionalString(body, "image_prompt", req.ImagePrompt)
	setOptionalString(body, "style_image", req.StyleImage)
	setOptionalBool(body, "compress", req.Compress)
	setOptionalBool(body, "bake", req.Bake)
	if len(req.PartNames) > 0 {
		body["part_names"] = req.PartNames
	}
	return p.createTask(ctx, "/models/texture", body)
}

// SegmentMesh segments a model into parts.
func (p *TripoProvider) SegmentMesh(ctx context.Context, req provider.SegmentMeshRequest) (*provider.ModelOperation, error) {
	body, err := inputBody(req.Input)
	if err != nil {
		return nil, err
	}
	setOptionalString(body, "model", req.Model)
	return p.createTask(ctx, "/mesh/segment", body)
}

// CompleteMesh completes a segmented model or selected parts.
func (p *TripoProvider) CompleteMesh(ctx context.Context, req provider.CompleteMeshRequest) (*provider.ModelOperation, error) {
	body, err := inputBody(req.Input)
	if err != nil {
		return nil, err
	}
	setOptionalString(body, "model", req.Model)
	if len(req.PartNames) > 0 {
		body["part_names"] = req.PartNames
	}
	return p.createTask(ctx, "/mesh/complete", body)
}

// RigCheck checks whether a model can be rigged.
func (p *TripoProvider) RigCheck(ctx context.Context, req provider.RigCheckRequest) (*provider.RigCheckResult, error) {
	body, err := inputBody(req.Input)
	if err != nil {
		return nil, err
	}
	resp, err := p.doJSON(ctx, http.MethodPost, "/animations/rig-check", body)
	if err != nil {
		return nil, err
	}
	var result struct {
		Riggable bool   `json:"riggable"`
		RigType  string `json:"rig_type"`
		Reason   string `json:"reason"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("decoding rig-check response: %w", err)
	}
	return &provider.RigCheckResult{
		Riggable: result.Riggable,
		RigType:  result.RigType,
		Reason:   result.Reason,
	}, nil
}

// RigModel creates a rigged model task.
func (p *TripoProvider) RigModel(ctx context.Context, req provider.RigModelRequest) (*provider.ModelOperation, error) {
	body, err := inputBody(req.Input)
	if err != nil {
		return nil, err
	}
	setOptionalString(body, "model", req.Model)
	setOptionalString(body, "rig_type", req.RigType)
	setOptionalString(body, "spec", req.Spec)
	setOptionalString(body, "out_format", req.OutFormat)
	return p.createTask(ctx, "/animations/rig", body)
}

// RetargetAnimation applies animation presets to a rigged model.
func (p *TripoProvider) RetargetAnimation(ctx context.Context, req provider.RetargetAnimationRequest) (*provider.ModelOperation, error) {
	body, err := inputBody(req.Input)
	if err != nil {
		return nil, err
	}
	if req.Animation != "" && len(req.Animations) > 0 {
		return nil, fmt.Errorf("animation and animations are mutually exclusive")
	}
	setOptionalString(body, "animation", req.Animation)
	if len(req.Animations) > 0 {
		body["animations"] = req.Animations
	}
	setOptionalString(body, "out_format", req.OutFormat)
	setOptionalBool(body, "bake_animation", req.BakeAnimation)
	setOptionalBool(body, "export_with_geometry", req.ExportWithGeometry)
	setOptionalBool(body, "animate_in_place", req.AnimateInPlace)
	return p.createTask(ctx, "/animations/retarget", body)
}

// BatchTasks queries several tasks in one call.
func (p *TripoProvider) BatchTasks(ctx context.Context, taskIDs []string) (*provider.BatchTasksResult, error) {
	if len(taskIDs) == 0 {
		return nil, fmt.Errorf("taskIDs is required")
	}
	if len(taskIDs) > 100 {
		return nil, fmt.Errorf("taskIDs cannot contain more than 100 items")
	}
	for _, taskID := range taskIDs {
		if err := validateTaskID(taskID); err != nil {
			return nil, err
		}
	}
	resp, err := p.doJSON(ctx, http.MethodPost, "/tasks/list", map[string]any{"task_ids": taskIDs})
	if err != nil {
		return nil, err
	}
	var decoded struct {
		Tasks  map[string]taskStatusResponse `json:"tasks"`
		Missed []string                      `json:"missed"`
	}
	if err := json.Unmarshal(resp.Data, &decoded); err != nil {
		return nil, fmt.Errorf("decoding batch task response: %w", err)
	}
	result := &provider.BatchTasksResult{
		Tasks:  make(map[string]provider.ModelTaskStatus, len(decoded.Tasks)),
		Missed: append([]string(nil), decoded.Missed...),
	}
	for id, task := range decoded.Tasks {
		result.Tasks[id] = providerTaskStatus(task)
	}
	return result, nil
}

// UploadFile uploads a local file and returns a reusable Tripo file token.
func (p *TripoProvider) UploadFile(ctx context.Context, req provider.UploadFileRequest) (*provider.FileUpload, error) {
	if strings.TrimSpace(req.FilePath) == "" {
		return nil, fmt.Errorf("filePath is required")
	}
	token, err := p.uploadFile(ctx, req.FilePath)
	if err != nil {
		return nil, err
	}
	return &provider.FileUpload{FileToken: token}, nil
}

// CreateFileUpload returns a presigned upload URL and file token.
func (p *TripoProvider) CreateFileUpload(ctx context.Context, req provider.CreateFileUploadRequest) (*provider.FileUpload, error) {
	if strings.TrimSpace(req.Format) == "" {
		return nil, fmt.Errorf("format is required")
	}
	resp, err := p.doJSON(ctx, http.MethodPost, "/files/presign", map[string]any{"format": req.Format})
	if err != nil {
		return nil, err
	}
	var decoded struct {
		PresignedURL string `json:"presigned_url"`
		FileToken    string `json:"file_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.Unmarshal(resp.Data, &decoded); err != nil {
		return nil, fmt.Errorf("decoding presign response: %w", err)
	}
	if decoded.PresignedURL == "" || decoded.FileToken == "" {
		return nil, fmt.Errorf("decoding presign response: missing presigned_url or file_token")
	}
	return &provider.FileUpload{
		PresignedURL: decoded.PresignedURL,
		FileToken:    decoded.FileToken,
		ExpiresIn:    decoded.ExpiresIn,
	}, nil
}

// GetBalance returns the account's available and frozen credit balance.
func (p *TripoProvider) GetBalance(ctx context.Context) (*provider.AccountBalance, error) {
	resp, err := p.doJSON(ctx, http.MethodGet, "/account/balance", nil)
	if err != nil {
		return nil, err
	}
	var decoded struct {
		Balance float64 `json:"balance"`
		Frozen  float64 `json:"frozen"`
	}
	if err := json.Unmarshal(resp.Data, &decoded); err != nil {
		return nil, fmt.Errorf("decoding account balance response: %w", err)
	}
	return &provider.AccountBalance{
		Balance: decoded.Balance,
		Frozen:  decoded.Frozen,
	}, nil
}

// GetUsage returns per-task account credit usage history.
func (p *TripoProvider) GetUsage(ctx context.Context) (*provider.AccountUsageResult, error) {
	resp, err := p.doJSON(ctx, http.MethodGet, "/account/usage", nil)
	if err != nil {
		return nil, err
	}
	var decoded []struct {
		TaskID          string  `json:"task_id"`
		Type            string  `json:"type"`
		CreditsConsumed float64 `json:"credits_consumed"`
		CreatedAt       string  `json:"created_at"`
	}
	if err := json.Unmarshal(resp.Data, &decoded); err != nil {
		return nil, fmt.Errorf("decoding account usage response: %w", err)
	}
	result := &provider.AccountUsageResult{
		Records: make([]provider.AccountUsageRecord, 0, len(decoded)),
	}
	for _, record := range decoded {
		result.Records = append(result.Records, provider.AccountUsageRecord{
			TaskID:          record.TaskID,
			Type:            record.Type,
			CreditsConsumed: record.CreditsConsumed,
			CreatedAt:       record.CreatedAt,
		})
	}
	return result, nil
}

func (p *TripoProvider) resolveInput(ctx context.Context, input, filePath, fileURL string) (string, error) {
	input = strings.TrimSpace(input)
	filePath = strings.TrimSpace(filePath)
	fileURL = strings.TrimSpace(fileURL)
	count := 0
	for _, value := range []string{input, filePath, fileURL} {
		if value != "" {
			count++
		}
	}
	if count != 1 {
		return "", fmt.Errorf("exactly one input source is required")
	}
	if input != "" {
		return input, nil
	}
	if fileURL != "" {
		return fileURL, nil
	}
	token, err := p.uploadFile(ctx, filePath)
	if err != nil {
		return "", err
	}
	return token, nil
}

func inputBody(input string) (map[string]any, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("input is required")
	}
	return map[string]any{"input": input}, nil
}

func multiviewEditPrompts(req provider.EditMultiviewRequest) ([]map[string]string, error) {
	if len(req.Prompts) > 0 {
		prompts := make([]map[string]string, 0, len(req.Prompts))
		for i, item := range req.Prompts {
			prompt := strings.TrimSpace(item.Prompt)
			if prompt == "" {
				return nil, fmt.Errorf("prompts[%d].prompt is required", i)
			}
			view, ok := normalizeMultiviewEditView(item.View)
			if !ok {
				return nil, fmt.Errorf("prompts[%d].view must be one of front, left, back, or right", i)
			}
			prompts = append(prompts, map[string]string{
				"prompt": prompt,
				"view":   view,
			})
		}
		return prompts, nil
	}

	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		return nil, fmt.Errorf("prompts is required")
	}
	if strings.TrimSpace(req.View) != "" {
		view, ok := normalizeMultiviewEditView(req.View)
		if !ok {
			return nil, fmt.Errorf("view must be one of front, left, back, or right")
		}
		return []map[string]string{{"prompt": prompt, "view": view}}, nil
	}

	return []map[string]string{
		{"prompt": prompt, "view": "front"},
		{"prompt": prompt, "view": "left"},
		{"prompt": prompt, "view": "back"},
		{"prompt": prompt, "view": "right"},
	}, nil
}

func normalizeMultiviewEditView(value string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "front", "left", "back", "right":
		return strings.ToLower(strings.TrimSpace(value)), true
	default:
		return "", false
	}
}

func setOptionalIntValue(body map[string]any, key string, value int) {
	if value > 0 {
		body[key] = value
	}
}

func setOptionalFloatValue(body map[string]any, key string, value float64) {
	if value != 0 {
		body[key] = value
	}
}
