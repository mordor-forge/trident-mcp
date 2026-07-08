package server

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mordor-forge/trident-mcp/internal/provider"
)

func (s *Server) registerModelProcessTools() {
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "import_model",
		Description: "Import an external model URL or file token into Tripo for downstream processing.",
	}, s.handleImportModel)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "refine_model",
		Description: "Refine an existing Tripo model task.",
	}, s.handleRefineModel)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "texture_model",
		Description: "Generate or regenerate textures for an existing model.",
	}, s.handleTextureModel)
}

func (s *Server) registerMeshTools() {
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "segment_mesh",
		Description: "Segment a model into editable parts.",
	}, s.handleSegmentMesh)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "complete_mesh",
		Description: "Complete a segmented model or selected parts.",
	}, s.handleCompleteMesh)
}

func (s *Server) registerAnimationTools() {
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "rig_check",
		Description: "Check whether a model can be rigged and get the recommended rig type.",
	}, s.handleRigCheck)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "rig_model",
		Description: "Automatically rig a model for animation.",
	}, s.handleRigModel)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "retarget_animation",
		Description: "Apply one or more preset animations to a rigged model.",
	}, s.handleRetargetAnimation)
}

func (s *Server) handleImportModel(ctx context.Context, _ *mcp.CallToolRequest, input provider.ImportModelRequest) (*mcp.CallToolResult, provider.ModelOperation, error) {
	op, err := s.modelproc.ImportModel(ctx, input)
	if err != nil {
		return nil, provider.ModelOperation{}, fmt.Errorf("import model: %w", err)
	}
	return operationResult("Model import started", op), *op, nil
}

func (s *Server) handleRefineModel(ctx context.Context, _ *mcp.CallToolRequest, input provider.RefineModelRequest) (*mcp.CallToolResult, provider.ModelOperation, error) {
	op, err := s.modelproc.RefineModel(ctx, input)
	if err != nil {
		return nil, provider.ModelOperation{}, fmt.Errorf("refine model: %w", err)
	}
	return operationResult("Model refinement started", op), *op, nil
}

func (s *Server) handleTextureModel(ctx context.Context, _ *mcp.CallToolRequest, input provider.TextureModelRequest) (*mcp.CallToolResult, provider.ModelOperation, error) {
	op, err := s.modelproc.TextureModel(ctx, input)
	if err != nil {
		return nil, provider.ModelOperation{}, fmt.Errorf("texture model: %w", err)
	}
	return operationResult("Model texturing started", op), *op, nil
}

func (s *Server) handleSegmentMesh(ctx context.Context, _ *mcp.CallToolRequest, input provider.SegmentMeshRequest) (*mcp.CallToolResult, provider.ModelOperation, error) {
	op, err := s.meshproc.SegmentMesh(ctx, input)
	if err != nil {
		return nil, provider.ModelOperation{}, fmt.Errorf("segment mesh: %w", err)
	}
	return operationResult("Mesh segmentation started", op), *op, nil
}

func (s *Server) handleCompleteMesh(ctx context.Context, _ *mcp.CallToolRequest, input provider.CompleteMeshRequest) (*mcp.CallToolResult, provider.ModelOperation, error) {
	op, err := s.meshproc.CompleteMesh(ctx, input)
	if err != nil {
		return nil, provider.ModelOperation{}, fmt.Errorf("complete mesh: %w", err)
	}
	return operationResult("Mesh completion started", op), *op, nil
}

func (s *Server) handleRigCheck(ctx context.Context, _ *mcp.CallToolRequest, input provider.RigCheckRequest) (*mcp.CallToolResult, provider.RigCheckResult, error) {
	result, err := s.animator.RigCheck(ctx, input)
	if err != nil {
		return nil, provider.RigCheckResult{}, fmt.Errorf("rig check: %w", err)
	}
	text := fmt.Sprintf("Riggable: %v", result.Riggable)
	if result.RigType != "" {
		text += fmt.Sprintf("\nRig type: %s", result.RigType)
	}
	if result.Reason != "" {
		text += fmt.Sprintf("\nReason: %s", result.Reason)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
		},
	}, *result, nil
}

func (s *Server) handleRigModel(ctx context.Context, _ *mcp.CallToolRequest, input provider.RigModelRequest) (*mcp.CallToolResult, provider.ModelOperation, error) {
	op, err := s.animator.RigModel(ctx, input)
	if err != nil {
		return nil, provider.ModelOperation{}, fmt.Errorf("rig model: %w", err)
	}
	return operationResult("Model rigging started", op), *op, nil
}

func (s *Server) handleRetargetAnimation(ctx context.Context, _ *mcp.CallToolRequest, input provider.RetargetAnimationRequest) (*mcp.CallToolResult, provider.ModelOperation, error) {
	op, err := s.animator.RetargetAnimation(ctx, input)
	if err != nil {
		return nil, provider.ModelOperation{}, fmt.Errorf("retarget animation: %w", err)
	}
	return operationResult("Animation retargeting started", op), *op, nil
}
