// Package server exposes the tripo module as MCP tools over stdio. It holds
// no Tripo knowledge of its own: every task-creating tool is generated from a
// tripo.Operation, and the lifecycle tools wrap tripo.Client calls.
package server

import (
	"context"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mordor-forge/trident-mcp/internal/tripo"
)

// Options configures a Server.
type Options struct {
	// OutputDir is where download_task and wait_for_task save Artifacts.
	OutputDir string
	Version   string
	// PollInterval is how often wait_for_task polls Tripo. Defaults to 2 s.
	PollInterval time.Duration
}

// Server is the trident-mcp MCP server.
type Server struct {
	mcp          *mcp.Server
	tripo        *tripo.Client
	outputDir    string
	version      string
	pollInterval time.Duration
}

// New creates a Server backed by a Tripo client.
func New(client *tripo.Client, opts Options) *Server {
	if opts.Version == "" {
		opts.Version = "dev"
	}
	s := &Server{
		mcp: mcp.NewServer(&mcp.Implementation{Name: "trident-mcp", Version: opts.Version}, &mcp.ServerOptions{
			Instructions: instructions,
		}),
		tripo:        client,
		outputDir:    opts.OutputDir,
		version:      opts.Version,
		pollInterval: opts.PollInterval,
	}
	s.registerOperations()
	s.registerTaskTools()
	s.registerAccountTools()
	return s
}

const instructions = `trident-mcp drives the Tripo 3D API. Task-creating tools spend credits and return a task_id at once; call wait_for_task to wait for it and save its outputs (meshes, images, splats) under the output directory. Chain operations by passing a task_id as the next tool's input. Inputs also accept absolute local paths (uploaded automatically), URLs and file tokens. Check get_balance before expensive work and list_models for model versions.`

// registerOperations adds one tool per Tripo Operation.
func (s *Server) registerOperations() {
	addOperation(s, tripo.TextToModel)
	addOperation(s, tripo.ImageToModel)
	addOperation(s, tripo.MultiviewToModel)
	addOperation(s, tripo.TextToImage)
	addOperation(s, tripo.ImageToImage)
	addOperation(s, tripo.ImageToMultiview)
	addOperation(s, tripo.EditMultiview)
	addOperation(s, tripo.ImageToSplat)
	addOperation(s, tripo.ImportModel)
	addOperation(s, tripo.TextureModel)
	addOperation(s, tripo.ConvertModel)
	addOperation(s, tripo.Retopologize)
	addOperation(s, tripo.SmartUV)
	addOperation(s, tripo.SegmentMesh)
	addOperation(s, tripo.CompleteMesh)
	addOperation(s, tripo.SmartSegment)
	addOperation(s, tripo.RigCheck)
	addOperation(s, tripo.RigModel)
	addOperation(s, tripo.RetargetAnimation)
}

// Run serves MCP over stdio until the client disconnects or ctx ends.
func (s *Server) Run(ctx context.Context) error {
	return s.mcp.Run(ctx, &mcp.StdioTransport{})
}

// MCPServer returns the underlying MCP server, for in-memory tests.
func (s *Server) MCPServer() *mcp.Server {
	return s.mcp
}

func text(format string) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: format}}}
}

func boolPtr(b bool) *bool { return &b }

var (
	creditSpending = &mcp.ToolAnnotations{DestructiveHint: boolPtr(false), OpenWorldHint: boolPtr(true)}
	readOnly       = &mcp.ToolAnnotations{ReadOnlyHint: true, OpenWorldHint: boolPtr(true)}
	savesFiles     = &mcp.ToolAnnotations{DestructiveHint: boolPtr(false), IdempotentHint: true, OpenWorldHint: boolPtr(true)}
)
