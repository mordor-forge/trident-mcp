package server

import (
	"context"
	"fmt"
	"sort"

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
		Description: "Generate a 3D model from 2-4 ordered reference images. Supply views in Tripo's expected order: front, left, back, right. This is an async operation — use task_status to poll progress and download_model to retrieve the result.",
	}, s.handleMultiviewTo3D)
}

// registerStatusTools adds task_status and download_model tools.
func (s *Server) registerStatusTools() {
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "task_status",
		Description: "Check the status of an async 3D generation or post-processing task. Returns progress info and the current state, such as queued, running, success, failed, cancelled, or expired.",
	}, s.handleTaskStatus)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "get_task",
		Description: "Get the current status and output metadata for a Tripo task by task ID.",
	}, s.handleTaskStatus)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "download_model",
		Description: "Download a completed 3D model to a local file using the task's actual output format. If you need a different format, run convert_format first, then download the conversion task.",
	}, s.handleDownloadModel)
}

// registerCommonTools adds account, task batch, and upload utility tools.
func (s *Server) registerCommonTools() {
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "get_tasks",
		Description: "Batch query up to 100 Tripo tasks by task ID.",
	}, s.handleGetTasks)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "get_balance",
		Description: "Query the Tripo account's available and frozen credit balance.",
	}, s.handleGetBalance)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "get_usage",
		Description: "Query Tripo account credit usage history by task.",
	}, s.handleGetUsage)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "upload_file",
		Description: "Upload a local file to Tripo and return a reusable file token.",
	}, s.handleUploadFile)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "create_file_upload",
		Description: "Create a presigned upload URL and file token for large Tripo file uploads.",
	}, s.handleCreateFileUpload)
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

// TaskIDsInput is used for batch task queries.
type TaskIDsInput struct {
	TaskIDs []string `json:"taskIds" jsonschema:"Task IDs to query, maximum 100"`
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

func (s *Server) handleGetTasks(ctx context.Context, _ *mcp.CallToolRequest, input TaskIDsInput) (*mcp.CallToolResult, provider.BatchTasksResult, error) {
	result, err := s.common.BatchTasks(ctx, input.TaskIDs)
	if err != nil {
		return nil, provider.BatchTasksResult{}, fmt.Errorf("get tasks: %w", err)
	}
	taskIDs := make([]string, 0, len(result.Tasks))
	for taskID := range result.Tasks {
		taskIDs = append(taskIDs, taskID)
	}
	sort.Strings(taskIDs)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(
				"Batch task query returned %d task(s), %d missed.\nTasks: %v\nMissed: %v",
				len(result.Tasks), len(result.Missed), taskIDs, result.Missed,
			)},
		},
	}, *result, nil
}

func (s *Server) handleGetBalance(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, provider.AccountBalance, error) {
	result, err := s.common.GetBalance(ctx)
	if err != nil {
		return nil, provider.AccountBalance{}, fmt.Errorf("get balance: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(
				"Available credits: %g\nFrozen credits: %g",
				result.Balance, result.Frozen,
			)},
		},
	}, *result, nil
}

func (s *Server) handleGetUsage(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, provider.AccountUsageResult, error) {
	result, err := s.common.GetUsage(ctx)
	if err != nil {
		return nil, provider.AccountUsageResult{}, fmt.Errorf("get usage: %w", err)
	}
	taskIDs := make([]string, 0, len(result.Records))
	for _, record := range result.Records {
		if record.TaskID != "" {
			taskIDs = append(taskIDs, record.TaskID)
		}
	}
	sort.Strings(taskIDs)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(
				"Account usage returned %d record(s).\nTasks: %v",
				len(result.Records), taskIDs,
			)},
		},
	}, *result, nil
}

func (s *Server) handleUploadFile(ctx context.Context, _ *mcp.CallToolRequest, input provider.UploadFileRequest) (*mcp.CallToolResult, provider.FileUpload, error) {
	result, err := s.common.UploadFile(ctx, input)
	if err != nil {
		return nil, provider.FileUpload{}, fmt.Errorf("upload file: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(
				"File uploaded.\n\nFile token: %s",
				result.FileToken,
			)},
		},
	}, *result, nil
}

func (s *Server) handleCreateFileUpload(ctx context.Context, _ *mcp.CallToolRequest, input provider.CreateFileUploadRequest) (*mcp.CallToolResult, provider.FileUpload, error) {
	result, err := s.common.CreateFileUpload(ctx, input)
	if err != nil {
		return nil, provider.FileUpload{}, fmt.Errorf("create file upload: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(
				"File upload prepared.\n\nFile token: %s\nExpires in: %d seconds",
				result.FileToken, result.ExpiresIn,
			)},
		},
	}, *result, nil
}
