package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mordor-forge/trident-mcp/internal/provider"
)

// EmptyInput is used for tools that require no parameters.
type EmptyInput struct{}

// registerConfigTools adds list_models and get_config tools.
func (s *Server) registerConfigTools() {
	if s.models != nil {
		mcp.AddTool(s.mcp, &mcp.Tool{
			Name:        "list_models",
			Description: "List the model versions supported by this server with their capabilities.",
		}, s.handleListModels)
	}

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "get_config",
		Description: "Show current server configuration including active backend and output directory.",
	}, s.handleGetConfig)
}

// modelsResult wraps the model list for structured output.
type modelsResult struct {
	Models []provider.ModelInfo `json:"models"`
}

// configResult is the structured output for get_config.
type configResult struct {
	Backend   string `json:"backend"`
	OutputDir string `json:"outputDir"`
	Version   string `json:"version"`
}

func (s *Server) handleListModels(ctx context.Context, _ *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, modelsResult, error) {
	models, err := s.models.ListModels(ctx)
	if err != nil {
		return nil, modelsResult{}, fmt.Errorf("list models: %w", err)
	}

	var b strings.Builder
	b.WriteString("Available Models\n")
	b.WriteString("================\n\n")
	for _, m := range models {
		fmt.Fprintf(&b, "%-25s  %s\n", m.Name, m.Description)
		fmt.Fprintf(&b, "  ID: %s  Capabilities: [%s]\n\n",
			m.ID, strings.Join(m.Capabilities, ", "))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: b.String()},
		},
	}, modelsResult{Models: models}, nil
}

func (s *Server) handleGetConfig(_ context.Context, _ *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, configResult, error) {
	result := configResult{
		Backend:   s.backend,
		OutputDir: s.outputDir,
		Version:   s.version,
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(
				"Backend: %s\nOutput directory: %s\nVersion: %s",
				result.Backend, result.OutputDir, result.Version,
			)},
		},
	}, result, nil
}
