package server

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mordor-forge/trident-mcp/internal/provider"
)

// registerGenerationTools adds text_to_3d, image_to_3d, and multiview_to_3d
// tools to the MCP server.
func (s *Server) registerGenerationTools() {
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "text_to_3d",
		Description: "Generate a 3D model from a text prompt. This is an async operation — use task_status to poll progress and download_model to retrieve the result.",
	}, s.handleTextTo3D)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "image_to_3d",
		Description: "Generate a 3D model from a reference image. Provide a local file path or public URL. This is an async operation — use task_status to poll progress and download_model to retrieve the result.",
	}, s.handleImageTo3D)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "multiview_to_3d",
		Description: "Generate a 3D model from 2-4 reference images taken from different angles. Produces significantly better geometry than single-image input. This is an async operation — use task_status to poll progress and download_model to retrieve the result.",
	}, s.handleMultiviewTo3D)
}

// registerStatusTools adds task_status and download_model tools.
func (s *Server) registerStatusTools() {
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "task_status",
		Description: "Check the status of an async 3D generation or post-processing task. Returns progress info (queued, running, success, or failed).",
	}, s.handleTaskStatus)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "download_model",
		Description: "Download a completed 3D model to a local file using the task's actual output format. If you need a different format, run convert_format first, then download the conversion task.",
	}, s.handleDownloadModel)
}

// --- Input types for status/download tools ---

// TaskIDInput is used for tools that take only a task ID.
type TaskIDInput struct {
	TaskID string `json:"taskId" jsonschema:"Task ID from a previous generation or post-processing call"`
}

// DownloadInput is used for the download_model tool.
type DownloadInput struct {
	TaskID string `json:"taskId" jsonschema:"Task ID of the completed generation"`
	Format string `json:"format,omitempty" jsonschema:"Optional expected task output format to validate: GLTF, FBX, OBJ, STL, USDZ, or 3MF. Use convert_format before download_model to change formats."`
}

// --- Handlers ---

func (s *Server) handleTextTo3D(ctx context.Context, _ *mcp.CallToolRequest, input provider.TextToModelRequest) (*mcp.CallToolResult, provider.ModelOperation, error) {
	op, err := s.generator.TextToModel(ctx, input)
	if err != nil {
		return nil, provider.ModelOperation{}, fmt.Errorf("text to 3D: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(
				"3D generation started!\n\nTask ID: %s\nStatus: %s\n\nUse task_status to check progress, then download_model to retrieve the file.",
				op.TaskID, op.Status,
			)},
		},
	}, *op, nil
}

func (s *Server) handleImageTo3D(ctx context.Context, _ *mcp.CallToolRequest, input provider.ImageToModelRequest) (*mcp.CallToolResult, provider.ModelOperation, error) {
	op, err := s.generator.ImageToModel(ctx, input)
	if err != nil {
		return nil, provider.ModelOperation{}, fmt.Errorf("image to 3D: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(
				"3D generation from image started!\n\nTask ID: %s\nStatus: %s\n\nUse task_status to check progress, then download_model to retrieve the file.",
				op.TaskID, op.Status,
			)},
		},
	}, *op, nil
}

func (s *Server) handleMultiviewTo3D(ctx context.Context, _ *mcp.CallToolRequest, input provider.MultiviewToModelRequest) (*mcp.CallToolResult, provider.ModelOperation, error) {
	op, err := s.generator.MultiviewToModel(ctx, input)
	if err != nil {
		return nil, provider.ModelOperation{}, fmt.Errorf("multiview to 3D: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(
				"Multi-view 3D generation started!\n\nTask ID: %s\nStatus: %s\n\nUse task_status to check progress, then download_model to retrieve the file.",
				op.TaskID, op.Status,
			)},
		},
	}, *op, nil
}

func (s *Server) handleTaskStatus(ctx context.Context, _ *mcp.CallToolRequest, input TaskIDInput) (*mcp.CallToolResult, provider.ModelTaskStatus, error) {
	status, err := s.status.Status(ctx, input.TaskID)
	if err != nil {
		return nil, provider.ModelTaskStatus{}, fmt.Errorf("task status: %w", err)
	}

	text := fmt.Sprintf(
		"Task: %s\nStatus: %s\nProgress: %d%%",
		status.TaskID, status.Status, status.Progress,
	)
	if status.Error != "" {
		text += fmt.Sprintf("\nError: %s", status.Error)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
		},
	}, *status, nil
}

func (s *Server) handleDownloadModel(ctx context.Context, _ *mcp.CallToolRequest, input DownloadInput) (*mcp.CallToolResult, provider.ModelResult, error) {
	result, err := s.status.Download(ctx, input.TaskID, input.Format)
	if err != nil {
		return nil, provider.ModelResult{}, fmt.Errorf("download model: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(
				"Model downloaded!\n\nSaved to: %s\nFormat: %s\nTask: %s",
				result.FilePath, result.Format, result.TaskID,
			)},
		},
	}, *result, nil
}
