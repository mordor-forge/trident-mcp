// Package server implements the MCP server that exposes 3D model
// generation capabilities as MCP tools over stdio transport.
package server

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mordor-forge/trident-mcp/internal/provider"
)

// Server wraps an MCP server and routes tool calls to provider implementations.
type Server struct {
	mcp       *mcp.Server
	generator provider.ModelGenerator
	imageGen  provider.ImageGenerator
	status    provider.ModelStatus
	postproc  provider.ModelPostProcessor
	modelproc provider.ModelProcessor
	meshproc  provider.MeshProcessor
	animator  provider.Animator
	common    provider.CommonAPI
	models    provider.ModelLister
	backend   string
	outputDir string
	version   string
}

// Options configures server metadata exposed via MCP tools.
type Options struct {
	Backend   string
	OutputDir string
	Version   string
}

// New creates a Server with the given provider implementations.
func New(generator provider.ModelGenerator, status provider.ModelStatus, postproc provider.ModelPostProcessor, models provider.ModelLister, outputDir string) *Server {
	return NewWithOptions(generator, status, postproc, models, Options{
		Backend:   "unknown",
		OutputDir: outputDir,
		Version:   "dev",
	})
}

// NewWithOptions creates a Server with explicit metadata about the configured
// backend and output directory.
func NewWithOptions(generator provider.ModelGenerator, status provider.ModelStatus, postproc provider.ModelPostProcessor, models provider.ModelLister, opts Options) *Server {
	if opts.Version == "" {
		opts.Version = "dev"
	}
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "trident-mcp",
		Version: opts.Version,
	}, nil)

	if opts.Backend == "" {
		opts.Backend = "unknown"
	}

	s := &Server{
		mcp:       mcpServer,
		generator: generator,
		status:    status,
		postproc:  postproc,
		models:    models,
		backend:   opts.Backend,
		outputDir: opts.OutputDir,
		version:   opts.Version,
	}
	s.discoverCapabilities(generator, status, postproc, models)

	if generator != nil && status != nil {
		s.registerGenerationTools()
	}
	if s.imageGen != nil && status != nil {
		s.registerImageTools()
	}
	if status != nil {
		s.registerStatusTools()
	}
	if postproc != nil && status != nil {
		s.registerPostProcessTools()
	}
	if s.modelproc != nil && status != nil {
		s.registerModelProcessTools()
	}
	if s.meshproc != nil && status != nil {
		s.registerMeshTools()
	}
	if s.animator != nil && status != nil {
		s.registerAnimationTools()
	}
	if s.common != nil {
		s.registerCommonTools()
	}
	s.registerConfigTools()

	return s
}

func (s *Server) discoverCapabilities(candidates ...any) {
	for _, candidate := range candidates {
		if candidate == nil {
			continue
		}
		if v, ok := candidate.(provider.ImageGenerator); ok && s.imageGen == nil {
			s.imageGen = v
		}
		if v, ok := candidate.(provider.ModelProcessor); ok && s.modelproc == nil {
			s.modelproc = v
		}
		if v, ok := candidate.(provider.MeshProcessor); ok && s.meshproc == nil {
			s.meshproc = v
		}
		if v, ok := candidate.(provider.Animator); ok && s.animator == nil {
			s.animator = v
		}
		if v, ok := candidate.(provider.CommonAPI); ok && s.common == nil {
			s.common = v
		}
	}
}

// Run starts the MCP server on the stdio transport, blocking until the
// client disconnects or the context is cancelled.
func (s *Server) Run(ctx context.Context) error {
	return s.mcp.Run(ctx, &mcp.StdioTransport{})
}

// MCPServer returns the underlying mcp.Server for testing.
func (s *Server) MCPServer() *mcp.Server {
	return s.mcp
}
