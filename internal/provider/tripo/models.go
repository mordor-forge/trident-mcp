package tripo

import (
	"context"
	"strings"

	"github.com/mordor-forge/trident-mcp/internal/provider"
)

type modelCatalogEntry struct {
	Name         string
	ID           string
	Namespace    string
	APIVersion   string
	Description  string
	Capabilities []string
	Aliases      []string
	Default      bool
}

var modelCatalog = []modelCatalogEntry{
	{
		Name:         "Tripo H3.1",
		ID:           "v3.1-20260211",
		Namespace:    "3d_generation",
		APIVersion:   "v3.1-20260211",
		Description:  "Default H-series high-fidelity 3D generation model.",
		Capabilities: []string{"text_to_3d", "image_to_3d", "multiview_to_3d"},
		Aliases:      []string{"tripo-v3.1", "v3.1", "h3.1"},
		Default:      true,
	},
	{
		Name:         "Tripo H3.0",
		ID:           "v3.0-20250812",
		Namespace:    "3d_generation",
		APIVersion:   "v3.0-20250812",
		Description:  "Previous H-series high-fidelity generation model.",
		Capabilities: []string{"text_to_3d", "image_to_3d", "multiview_to_3d"},
		Aliases:      []string{"tripo-v3.0", "v3.0"},
	},
	{
		Name:         "Tripo v2.5",
		ID:           "v2.5-20250123",
		Namespace:    "3d_generation",
		APIVersion:   "v2.5-20250123",
		Description:  "Legacy v2.5 generation model exposed by v3 endpoints.",
		Capabilities: []string{"text_to_3d", "image_to_3d", "multiview_to_3d"},
		Aliases:      []string{"tripo-v2.5", "v2.5"},
	},
	{
		Name:         "Tripo v2.0",
		ID:           "tripo-v2.0",
		Namespace:    "3d_generation",
		APIVersion:   "tripo-v2.0",
		Description:  "Legacy v2.0 generation model exposed by v3 endpoints.",
		Capabilities: []string{"text_to_3d", "image_to_3d", "multiview_to_3d"},
	},
	{
		Name:         "Tripo Turbo",
		ID:           "tripo-turbo",
		Namespace:    "3d_generation",
		APIVersion:   "tripo-turbo",
		Description:  "Fast generation model exposed by v3 endpoint tables.",
		Capabilities: []string{"text_to_3d", "image_to_3d"},
	},
	{
		Name:         "Tripo P1",
		ID:           "P1-20260311",
		Namespace:    "3d_generation",
		APIVersion:   "P1-20260311",
		Description:  "Topology-focused lowpoly generation tuned for clean structured meshes and game-engine workflows.",
		Capabilities: []string{"text_to_3d", "image_to_3d", "multiview_to_3d"},
		Aliases:      []string{"tripo-p1", "p1"},
	},
	{
		Name:         "Seedream v4",
		ID:           "seedream_v4",
		Namespace:    "image_generation",
		APIVersion:   "seedream_v4",
		Description:  "ByteDance image generation model documented as the text-to-image default.",
		Capabilities: []string{"text_to_image"},
		Default:      true,
	},
	{
		Name:         "Seedream v5",
		ID:           "seedream_v5",
		Namespace:    "image_generation",
		APIVersion:   "seedream_v5",
		Description:  "ByteDance image generation and image editing model.",
		Capabilities: []string{"text_to_image", "image_to_image"},
	},
	{
		Name:         "Gemini 2.5 Flash Image",
		ID:           "gemini-2.5-flash",
		Namespace:    "image_generation",
		APIVersion:   "gemini-2.5-flash",
		Description:  "Google fast image model exposed by Tripo image endpoints.",
		Capabilities: []string{"text_to_image", "image_to_image"},
	},
	{
		Name:         "Gemini 3 Pro Image",
		ID:           "gemini-3-pro",
		Namespace:    "image_generation",
		APIVersion:   "gemini-3-pro",
		Description:  "Google high-quality image model exposed by Tripo image endpoints.",
		Capabilities: []string{"text_to_image", "image_to_image"},
	},
	{
		Name:         "Gemini 3.1 Flash Image",
		ID:           "gemini-3.1-flash",
		Namespace:    "image_generation",
		APIVersion:   "gemini-3.1-flash",
		Description:  "Google latest fast image model exposed by Tripo image endpoints.",
		Capabilities: []string{"text_to_image", "image_to_image"},
	},
	{
		Name:         "Chat Image 1",
		ID:           "chat_image_1",
		Namespace:    "image_generation",
		APIVersion:   "chat_image_1",
		Description:  "OpenAI image model exposed by Tripo image endpoints.",
		Capabilities: []string{"text_to_image", "image_to_image"},
	},
	{
		Name:         "Chat Image 1.5",
		ID:           "chat_image_1.5",
		Namespace:    "image_generation",
		APIVersion:   "chat_image_1.5",
		Description:  "OpenAI higher-quality image model exposed by Tripo image endpoints.",
		Capabilities: []string{"text_to_image", "image_to_image"},
	},
	{
		Name:         "Chat Image 2",
		ID:           "chat_image_2",
		Namespace:    "image_generation",
		APIVersion:   "chat_image_2",
		Description:  "OpenAI latest image model exposed by Tripo image endpoints.",
		Capabilities: []string{"text_to_image", "image_to_image"},
	},
	{
		Name:         "Tripo Rig v2.0",
		ID:           "rig-v2.0",
		Namespace:    "animation",
		APIVersion:   "rig-v2.0",
		Description:  "Current rigging model for Tripo animation workflows.",
		Capabilities: []string{"rig_model"},
		Default:      true,
	},
	{
		Name:         "Tripo Rig v1.0",
		ID:           "rig-v1.0",
		Namespace:    "animation",
		APIVersion:   "rig-v1.0",
		Description:  "Legacy rigging model for Tripo animation workflows.",
		Capabilities: []string{"rig_model"},
	},
}

var (
	modelVersionMap = buildModelVersionMap()
	defaultModel    = "v3.1-20260211"
	// defaultModelVersion is kept for older internal tests and compatibility
	// helpers that still call resolveModelVersion.
	defaultModelVersion = defaultModel
	multiviewModels     = buildMultiviewModels()

	// validFormats maps accepted user input to the canonical Tripo format name.
	validFormats = map[string]string{
		"GLTF": "GLTF",
		"GLB":  "GLTF",
		"FBX":  "FBX",
		"OBJ":  "OBJ",
		"STL":  "STL",
		"USDZ": "USDZ",
		"3MF":  "3MF",
	}

	// validStyles lists supported stylization styles.
	validStyles = map[string]bool{
		"lego":      true,
		"voxel":     true,
		"voronoi":   true,
		"minecraft": true,
	}
)

func buildModelVersionMap() map[string]string {
	versions := make(map[string]string, len(modelCatalog))
	for _, model := range modelCatalog {
		versions[model.ID] = model.APIVersion
	}
	versions["tripo-v3.1"] = "v3.1-20260211"
	versions["v3.1"] = "v3.1-20260211"
	versions["h3.1"] = "v3.1-20260211"
	versions["tripo-v3.0"] = "v3.0-20250812"
	versions["v3.0"] = "v3.0-20250812"
	versions["tripo-v2.5"] = "v2.5-20250123"
	versions["v2.5"] = "v2.5-20250123"
	versions["v2.0"] = "tripo-v2.0"
	versions["turbo"] = "tripo-turbo"
	versions["tripo-p1"] = "P1-20260311"
	versions["p1"] = "P1-20260311"
	return versions
}

func buildMultiviewModels() map[string]bool {
	models := make(map[string]bool)
	for _, model := range modelCatalog {
		if hasCapability(model.Capabilities, "multiview_to_3d") {
			models[model.APIVersion] = true
		}
	}
	return models
}

func hasCapability(capabilities []string, target string) bool {
	for _, capability := range capabilities {
		if capability == target {
			return true
		}
	}
	return false
}

func supportedModels() []provider.ModelInfo {
	models := make([]provider.ModelInfo, 0, len(modelCatalog))
	for _, model := range modelCatalog {
		models = append(models, provider.ModelInfo{
			Name:         model.Name,
			ID:           model.ID,
			Namespace:    model.Namespace,
			Description:  model.Description,
			Capabilities: append([]string(nil), model.Capabilities...),
			Aliases:      append([]string(nil), model.Aliases...),
			Default:      model.Default,
		})
	}
	return models
}

// ListModels returns the built-in model catalog supported by this server.
func (p *TripoProvider) ListModels(_ context.Context) ([]provider.ModelInfo, error) {
	return supportedModels(), nil
}

// resolveModelVersion maps a friendly version name to the full API version string.
// Empty input returns the default version. Unrecognized values are returned as-is.
func resolveModelVersion(version string) string {
	return resolveModel(version)
}

func resolveModel(model string) string {
	model = strings.TrimSpace(model)
	if model == "" {
		return defaultModel
	}
	if full, ok := modelVersionMap[strings.ToLower(model)]; ok {
		return full
	}
	return model
}

// normalizeFormat accepts case-insensitive format names plus the GLB alias.
func normalizeFormat(format string) (string, bool) {
	format = strings.TrimSpace(format)
	if format == "" {
		return "", false
	}
	canonical, ok := validFormats[strings.ToUpper(format)]
	return canonical, ok
}

func normalizeStyle(style string) string {
	return strings.ToLower(strings.TrimSpace(style))
}
