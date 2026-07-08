package tripo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/mordor-forge/trident-mcp/internal/provider"
)

// Status checks the current state of a task.
func (p *TripoProvider) Status(ctx context.Context, taskID string) (*provider.ModelTaskStatus, error) {
	if err := validateTaskID(taskID); err != nil {
		return nil, err
	}

	resp, err := p.doJSON(ctx, http.MethodGet, "/tasks/"+taskID, nil)
	if err != nil {
		return nil, err
	}

	var status taskStatusResponse
	if err := json.Unmarshal(resp.Data, &status); err != nil {
		return nil, fmt.Errorf("decoding task status: %w", err)
	}

	result := providerTaskStatus(status)

	return &result, nil
}

// Download retrieves a completed model and saves it to disk.
func (p *TripoProvider) Download(ctx context.Context, taskID string, format string) (*provider.ModelResult, error) {
	if err := validateTaskID(taskID); err != nil {
		return nil, err
	}
	var requestedFormat string
	if format != "" {
		var ok bool
		requestedFormat, ok = normalizeFormat(format)
		if !ok {
			return nil, fmt.Errorf("unsupported format %q (valid: GLTF, FBX, OBJ, STL, USDZ, 3MF)", format)
		}
	}

	// Get task status to find download URL.
	resp, err := p.doJSON(ctx, http.MethodGet, "/tasks/"+taskID, nil)
	if err != nil {
		return nil, err
	}

	var status taskStatusResponse
	if err := json.Unmarshal(resp.Data, &status); err != nil {
		return nil, fmt.Errorf("decoding task status: %w", err)
	}

	if status.Status != "success" {
		return nil, fmt.Errorf("task %s is not complete (status: %s)", taskID, status.Status)
	}
	if status.Output == nil {
		return nil, fmt.Errorf("task %s has no output", taskID)
	}

	modelURL := modelDownloadURL(status.Output, requestedFormat)
	if modelURL == "" {
		return nil, fmt.Errorf("task %s has no model download URL", taskID)
	}

	actualFormat, haveActualFormat := formatFromURL(modelURL)
	if requestedFormat != "" && haveActualFormat && requestedFormat != actualFormat {
		return nil, fmt.Errorf(
			"task %s provides %s output; run convert_format before download_model to get %s",
			taskID, actualFormat, requestedFormat,
		)
	}

	tempPath := filepath.Join(p.outputDir, generateFilename("tmp"))
	headers, err := downloadFile(ctx, p.client, modelURL, tempPath)
	if err != nil {
		return nil, fmt.Errorf("downloading model: %w", err)
	}

	if !haveActualFormat {
		actualFormat, haveActualFormat = formatFromHeaders(headers)
	}
	if !haveActualFormat {
		_ = os.Remove(tempPath)
		return nil, fmt.Errorf("could not determine downloaded model format from URL or response headers")
	}
	if requestedFormat != "" && requestedFormat != actualFormat {
		_ = os.Remove(tempPath)
		return nil, fmt.Errorf(
			"task %s provides %s output; run convert_format before download_model to get %s",
			taskID, actualFormat, requestedFormat,
		)
	}

	filename := generateFilename(formatToExt(actualFormat))
	destPath := filepath.Join(p.outputDir, filename)
	if err := os.Rename(tempPath, destPath); err != nil {
		_ = os.Remove(tempPath)
		return nil, fmt.Errorf("renaming downloaded model: %w", err)
	}

	return &provider.ModelResult{
		FilePath: destPath,
		Format:   actualFormat,
		TaskID:   taskID,
	}, nil
}

func modelDownloadURL(output *taskOutput, requestedFormat string) string {
	candidates := []string{
		firstNonEmpty(output.PBRModelURL, output.PBRModel),
		firstNonEmpty(output.ModelURL, output.Model),
		firstNonEmpty(output.BaseModelURL, output.BaseModel),
	}
	candidates = append(candidates, output.ModelURLs...)

	var fallback string
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if fallback == "" {
			fallback = candidate
		}
		if requestedFormat == "" {
			return candidate
		}
		if actualFormat, ok := formatFromURL(candidate); ok && actualFormat == requestedFormat {
			return candidate
		}
	}
	return fallback
}
