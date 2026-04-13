package server

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mordor-forge/trident-mcp/internal/provider"
)

// registerPostProcessTools adds retopologize, convert_format, and stylize
// tools to the MCP server.
func (s *Server) registerPostProcessTools() {
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "retopologize",
		Description: "Create a lowpoly version of a generated 3D model. Supports quad mesh or triangle output. This is an async operation — use task_status to poll progress.",
	}, s.handleRetopologize)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "convert_format",
		Description: "Convert a generated 3D model to a different file format (GLTF, FBX, OBJ, STL, USDZ, 3MF). This is an async operation — use task_status to poll progress, then download_model to retrieve the converted file.",
	}, s.handleConvertFormat)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "stylize",
		Description: "Apply a stylization effect to a generated 3D model (lego, voxel, voronoi, minecraft). This is an async operation — use task_status to poll progress, then download_model to retrieve the stylized model.",
	}, s.handleStylize)
}

func (s *Server) handleRetopologize(ctx context.Context, _ *mcp.CallToolRequest, input provider.RetopologyRequest) (*mcp.CallToolResult, provider.ModelOperation, error) {
	op, err := s.postproc.Retopologize(ctx, input)
	if err != nil {
		return nil, provider.ModelOperation{}, fmt.Errorf("retopologize: %w", err)
	}

	meshType := "triangle"
	if input.Quad {
		meshType = "quad"
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(
				"Retopology started (%s mesh)!\n\nTask ID: %s\nStatus: %s\n\nUse task_status to check progress, then download_model to retrieve the result.",
				meshType, op.TaskID, op.Status,
			)},
		},
	}, *op, nil
}

func (s *Server) handleConvertFormat(ctx context.Context, _ *mcp.CallToolRequest, input provider.ConvertRequest) (*mcp.CallToolResult, provider.ModelOperation, error) {
	op, err := s.postproc.ConvertFormat(ctx, input)
	if err != nil {
		return nil, provider.ModelOperation{}, fmt.Errorf("convert format: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(
				"Format conversion to %s started!\n\nTask ID: %s\nStatus: %s\n\nUse task_status to check progress, then download_model to retrieve the file.",
				input.Format, op.TaskID, op.Status,
			)},
		},
	}, *op, nil
}

func (s *Server) handleStylize(ctx context.Context, _ *mcp.CallToolRequest, input provider.StylizeRequest) (*mcp.CallToolResult, provider.ModelOperation, error) {
	op, err := s.postproc.Stylize(ctx, input)
	if err != nil {
		return nil, provider.ModelOperation{}, fmt.Errorf("stylize: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(
				"Stylization (%s) started!\n\nTask ID: %s\nStatus: %s\n\nUse task_status to check progress, then download_model to retrieve the result.",
				input.Style, op.TaskID, op.Status,
			)},
		},
	}, *op, nil
}
