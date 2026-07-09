package server

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mordor-forge/trident-mcp/internal/provider"
)

func (s *Server) registerImageTools() {
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "text_to_image",
		Description: "Generate an image from a text prompt using Tripo v3 image-generation models.",
	}, s.handleTextToImage)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "image_to_image",
		Description: "Generate or edit an image from a source image using Tripo v3 image models.",
	}, s.handleImageToImage)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "image_to_multiview",
		Description: "Generate front/left/back/right multiview reference images from a single source image.",
	}, s.handleImageToMultiview)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "edit_multiview",
		Description: "Edit an existing multiview image set with per-view prompt instructions.",
	}, s.handleEditMultiview)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "image_to_splat",
		Description: "Generate a Gaussian Splat task from a source image.",
	}, s.handleImageToSplat)
}

func (s *Server) handleTextToImage(ctx context.Context, _ *mcp.CallToolRequest, input provider.TextToImageRequest) (*mcp.CallToolResult, provider.ModelOperation, error) {
	op, err := s.imageGen.TextToImage(ctx, input)
	if err != nil {
		return nil, provider.ModelOperation{}, fmt.Errorf("text to image: %w", err)
	}
	return operationResult("Image generation started", op), *op, nil
}

func (s *Server) handleImageToImage(ctx context.Context, _ *mcp.CallToolRequest, input provider.ImageToImageRequest) (*mcp.CallToolResult, provider.ModelOperation, error) {
	op, err := s.imageGen.ImageToImage(ctx, input)
	if err != nil {
		return nil, provider.ModelOperation{}, fmt.Errorf("image to image: %w", err)
	}
	return operationResult("Image-to-image generation started", op), *op, nil
}

func (s *Server) handleImageToMultiview(ctx context.Context, _ *mcp.CallToolRequest, input provider.ImageToMultiviewRequest) (*mcp.CallToolResult, provider.ModelOperation, error) {
	op, err := s.imageGen.ImageToMultiview(ctx, input)
	if err != nil {
		return nil, provider.ModelOperation{}, fmt.Errorf("image to multiview: %w", err)
	}
	return operationResult("Image-to-multiview generation started", op), *op, nil
}

func (s *Server) handleEditMultiview(ctx context.Context, _ *mcp.CallToolRequest, input provider.EditMultiviewRequest) (*mcp.CallToolResult, provider.ModelOperation, error) {
	op, err := s.imageGen.EditMultiview(ctx, input)
	if err != nil {
		return nil, provider.ModelOperation{}, fmt.Errorf("edit multiview: %w", err)
	}
	return operationResult("Multiview edit started", op), *op, nil
}

func (s *Server) handleImageToSplat(ctx context.Context, _ *mcp.CallToolRequest, input provider.ImageToSplatRequest) (*mcp.CallToolResult, provider.ModelOperation, error) {
	op, err := s.imageGen.ImageToSplat(ctx, input)
	if err != nil {
		return nil, provider.ModelOperation{}, fmt.Errorf("image to splat: %w", err)
	}
	return operationResult("Image-to-splat generation started", op), *op, nil
}

func operationResult(message string, op *provider.ModelOperation) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(
				"%s.\n\nTask ID: %s\nStatus: %s\n\nUse get_task or task_status to check progress.",
				message, op.TaskID, op.Status,
			)},
		},
	}
}
