package tripo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/mordor-forge/trident-mcp/internal/config"
	"github.com/mordor-forge/trident-mcp/internal/provider"
)

// validTaskID matches Tripo task IDs: alphanumeric, hyphens, underscores.
var validTaskID = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

const defaultBaseURL = "https://api.tripo3d.ai/v2/openapi"

// Compile-time interface satisfaction checks.
var (
	_ provider.ModelGenerator     = (*TripoProvider)(nil)
	_ provider.ModelStatus        = (*TripoProvider)(nil)
	_ provider.ModelPostProcessor = (*TripoProvider)(nil)
	_ provider.ModelLister        = (*TripoProvider)(nil)
)

// TripoProvider implements the provider interfaces by calling the Tripo REST API.
type TripoProvider struct {
	apiKey    string
	baseURL   string
	outputDir string
	client    *http.Client
}

// Config holds the parameters needed to construct a TripoProvider.
type Config struct {
	APIKey    string
	BaseURL   string
	OutputDir string
}

// NewFromConfig creates a TripoProvider from the application-level config.
func NewFromConfig(cfg *config.Config) (*TripoProvider, error) {
	return New(Config{
		APIKey:    cfg.Provider.APIKey,
		OutputDir: cfg.OutputDir,
	})
}

// New creates a TripoProvider with the given configuration.
func New(cfg Config) (*TripoProvider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("tripo API key is required")
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	return &TripoProvider{
		apiKey:    cfg.APIKey,
		baseURL:   baseURL,
		outputDir: cfg.OutputDir,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}, nil
}

// --- API response types ---

type apiResponse struct {
	Code int             `json:"code"`
	Data json.RawMessage `json:"data,omitempty"`
}

type apiError struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}

type taskCreateResponse struct {
	TaskID string `json:"task_id"`
}

type taskStatusResponse struct {
	TaskID          string      `json:"task_id"`
	Type            string      `json:"type"`
	Status          string      `json:"status"`
	Progress        int         `json:"progress"`
	Output          *taskOutput `json:"output,omitempty"`
	ErrorCode       *int        `json:"error_code,omitempty"`
	ErrorMsg        *string     `json:"error_msg,omitempty"`
	RunningLeftTime *int        `json:"running_left_time,omitempty"`
	QueuingNum      *int        `json:"queuing_num,omitempty"`
}

type taskOutput struct {
	Model         string `json:"model,omitempty"`
	BaseModel     string `json:"base_model,omitempty"`
	PBRModel      string `json:"pbr_model,omitempty"`
	RenderedImage string `json:"rendered_image,omitempty"`
}

type uploadResponse struct {
	ImageToken string `json:"image_token"`
}

// --- Validation helpers ---

// validateTaskID ensures a task ID is safe for URL path construction.
func validateTaskID(taskID string) error {
	if taskID == "" {
		return fmt.Errorf("taskID is required")
	}
	if !validTaskID.MatchString(taskID) {
		return fmt.Errorf("invalid taskID: must contain only alphanumeric characters, hyphens, and underscores")
	}
	return nil
}

// --- HTTP helpers ---

// doJSON sends a JSON request and decodes the response.
func (p *TripoProvider) doJSON(ctx context.Context, method, path string, body any) (*apiResponse, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, p.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var apiErr apiError
		if json.Unmarshal(respBody, &apiErr) == nil && apiErr.Message != "" {
			return nil, fmt.Errorf("API error (code %d): %s", apiErr.Code, apiErr.Message)
		}
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var apiResp apiResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	if apiResp.Code != 0 {
		return nil, fmt.Errorf("API error code %d", apiResp.Code)
	}

	return &apiResp, nil
}

// createTask sends a POST /task request and returns the task ID.
func (p *TripoProvider) createTask(ctx context.Context, body map[string]any) (*provider.ModelOperation, error) {
	resp, err := p.doJSON(ctx, http.MethodPost, "/task", body)
	if err != nil {
		return nil, err
	}

	var result taskCreateResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("decoding task response: %w", err)
	}
	if result.TaskID == "" {
		return nil, fmt.Errorf("decoding task response: missing task_id")
	}

	return &provider.ModelOperation{
		TaskID: result.TaskID,
		Status: "submitted",
	}, nil
}

// uploadFile uploads a local file and returns the image token.
func (p *TripoProvider) uploadFile(ctx context.Context, filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("opening file %s: %w", filePath, err)
	}
	defer func() { _ = f.Close() }()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	part, err := w.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return "", fmt.Errorf("creating form file: %w", err)
	}

	if _, err := io.Copy(part, f); err != nil {
		return "", fmt.Errorf("copying file data: %w", err)
	}
	if err := w.Close(); err != nil {
		return "", fmt.Errorf("finalizing upload body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/upload", &buf)
	if err != nil {
		return "", fmt.Errorf("creating upload request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading upload response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("upload returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var apiResp apiResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return "", fmt.Errorf("decoding upload response: %w", err)
	}
	if apiResp.Code != 0 {
		return "", fmt.Errorf("upload API error code %d", apiResp.Code)
	}

	var uploadResult uploadResponse
	if err := json.Unmarshal(apiResp.Data, &uploadResult); err != nil {
		return "", fmt.Errorf("decoding upload data: %w", err)
	}
	if uploadResult.ImageToken == "" {
		return "", fmt.Errorf("decoding upload data: missing image_token")
	}

	return uploadResult.ImageToken, nil
}
