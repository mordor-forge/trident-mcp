package server

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mordor-forge/trident-mcp/internal/tripo"
)

type usageInput struct {
	Limit  int `json:"limit,omitempty" jsonschema:"How many records to return (Tripo's default is 50)"`
	Offset int `json:"offset,omitempty" jsonschema:"How many newest records to skip"`
}

type uploadInput struct {
	Path string `json:"path" jsonschema:"Absolute path of a local image (png, jpg, webp; up to 20 MB) or model (glb, gltf, fbx, obj, stl, 3mf, usdz; up to 150 MB)"`
}

type modelsInput struct {
	Operation string `json:"operation,omitempty" jsonschema:"Only list model versions this tool accepts, e.g. text_to_model or rig_model"`
}

type modelsResult struct {
	CheckedOn string               `json:"checked_on"`
	Models    []tripo.ModelVersion `json:"models"`
}

type configResult struct {
	BaseURL      string `json:"base_url"`
	OutputDir    string `json:"output_dir"`
	Version      string `json:"version"`
	CatalogCheck string `json:"catalog_checked_on"`
}

func (s *Server) registerAccountTools() {
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "get_balance",
		Description: "Get the Tripo account's available and frozen credits (1 credit = US$0.01). Frozen credits are held by running tasks.",
		Annotations: readOnly,
	}, s.handleBalance)
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "get_usage",
		Description: "List recent Tripo tasks with the credits each one consumed.",
		Annotations: readOnly,
	}, s.handleUsage)
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "upload_file",
		Description: "Upload a local image or model to Tripo and get a reusable file token. Rarely needed: every input field already accepts local paths.",
		Annotations: sendsToTripo,
	}, s.handleUpload)
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "list_models",
		Description: "List the model versions this server knows, per tool, with aliases, defaults and the options each version rejects. Static; unknown versions are still sent to Tripo with a warning.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, OpenWorldHint: boolPtr(false)},
	}, s.handleListModels)
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "get_config",
		Description: "Show the Tripo API host, output directory and server version.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, OpenWorldHint: boolPtr(false)},
	}, s.handleConfig)
}

func (s *Server) handleBalance(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, tripo.Balance, error) {
	b, err := s.tripo.GetBalance(ctx)
	if err != nil {
		return nil, tripo.Balance{}, err
	}
	return text(fmt.Sprintf("Available credits: %g\nFrozen by running tasks: %g", b.Balance, b.Frozen)), *b, nil
}

func (s *Server) handleUsage(ctx context.Context, _ *mcp.CallToolRequest, in usageInput) (*mcp.CallToolResult, tripo.Usage, error) {
	u, err := s.tripo.GetUsage(ctx, in.Limit, in.Offset)
	if err != nil {
		return nil, tripo.Usage{}, err
	}
	var b strings.Builder
	var total float64
	for _, r := range u.Records {
		total += r.CreditsConsumed
		fmt.Fprintf(&b, "- %s %s: %g credits\n", r.TaskID, r.Type, r.CreditsConsumed)
	}
	fmt.Fprintf(&b, "%d task(s), %g credits", len(u.Records), total)
	return text(b.String()), *u, nil
}

func (s *Server) handleUpload(ctx context.Context, _ *mcp.CallToolRequest, in uploadInput) (*mcp.CallToolResult, tripo.Upload, error) {
	up, err := s.tripo.UploadFile(ctx, in.Path)
	if err != nil {
		return nil, tripo.Upload{}, err
	}
	return text(fmt.Sprintf("Uploaded %s (%s).\nFile token: %s", up.Path, humanBytes(up.Bytes), up.FileToken)), *up, nil
}

func (s *Server) handleListModels(_ context.Context, _ *mcp.CallToolRequest, in modelsInput) (*mcp.CallToolResult, modelsResult, error) {
	models := tripo.Catalog(in.Operation)
	if len(models) == 0 {
		return nil, modelsResult{}, fmt.Errorf("no model versions for %q; tools with a model field are text_to_model, image_to_model, multiview_to_model, text_to_image, image_to_image, texture_model, retopologize, smart_uv, segment_mesh, complete_mesh and rig_model", in.Operation)
	}
	var b strings.Builder
	family := ""
	for _, m := range models {
		if m.Family != family {
			family = m.Family
			fmt.Fprintf(&b, "\n[%s] for %s\n", family, strings.Join(m.Operations, ", "))
		}
		fmt.Fprintf(&b, "- %s (%s)", m.ID, m.Name)
		if m.Stage != "" {
			fmt.Fprintf(&b, " [%s]", m.Stage)
		}
		switch {
		case in.Operation != "" && slices.Contains(m.DefaultFor, in.Operation):
			b.WriteString(" default")
		case in.Operation == "" && len(m.DefaultFor) > 0:
			fmt.Fprintf(&b, " default for %s", strings.Join(m.DefaultFor, ", "))
		}
		fmt.Fprintf(&b, ": %s", m.Description)
		if len(m.Aliases) > 0 {
			fmt.Fprintf(&b, " Aliases: %s.", strings.Join(m.Aliases, ", "))
		}
		if len(m.Rejects) > 0 {
			fmt.Fprintf(&b, " Rejects: %s.", strings.Join(m.Rejects, ", "))
		}
		b.WriteString("\n")
	}
	return text(fmt.Sprintf("Model catalog (checked against Tripo docs on %s):%s", tripo.CatalogDate, strings.TrimRight(b.String(), "\n"))),
		modelsResult{CheckedOn: tripo.CatalogDate, Models: models}, nil
}

func (s *Server) handleConfig(context.Context, *mcp.CallToolRequest, struct{}) (*mcp.CallToolResult, configResult, error) {
	cfg := configResult{BaseURL: s.tripo.BaseURL(), OutputDir: s.outputDir, Version: s.version, CatalogCheck: tripo.CatalogDate}
	return text(fmt.Sprintf("Tripo API: %s\nOutput directory: %s\nVersion: %s\nModel catalog checked: %s",
		cfg.BaseURL, cfg.OutputDir, cfg.Version, cfg.CatalogCheck)), cfg, nil
}
