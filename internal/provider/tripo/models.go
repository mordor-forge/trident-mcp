package tripo

import (
	"context"
	"strings"

	"github.com/mordor-forge/trident-mcp/internal/provider"
)

type modelCatalogEntry struct {
	Name         string
	ID           string
	APIVersion   string
	Description  string
	Capabilities []string
}

var modelCatalog = []modelCatalogEntry{
	{
		Name:         "Tripo Turbo",
		ID:           "turbo",
		APIVersion:   "Turbo-v1.0-20250506",
		Description:  "Fast generation with lower quality. Best for quick iterations.",
		Capabilities: []string{"text_to_3d", "image_to_3d"},
	},
	{
		Name:         "Tripo v1.4",
		ID:           "v1.4",
		APIVersion:   "v1.4-20240625",
		Description:  "Legacy model version.",
		Capabilities: []string{"text_to_3d", "image_to_3d"},
	},
	{
		Name:         "Tripo v2.0",
		ID:           "v2.0",
		APIVersion:   "v2.0-20240919",
		Description:  "First version supporting multi-view input.",
		Capabilities: []string{"text_to_3d", "image_to_3d", "multiview_to_3d"},
	},
	{
		Name:         "Tripo v2.5",
		ID:           "v2.5",
		APIVersion:   "v2.5-20250123",
		Description:  "Improved geometry and texture quality.",
		Capabilities: []string{"text_to_3d", "image_to_3d", "multiview_to_3d"},
	},
	{
		Name:         "Tripo v3.0",
		ID:           "v3.0",
		APIVersion:   "v3.0-20250812",
		Description:  "Major quality upgrade with detailed textures.",
		Capabilities: []string{"text_to_3d", "image_to_3d", "multiview_to_3d"},
	},
	{
		Name:         "Tripo v3.1 (Latest)",
		ID:           "v3.1",
		APIVersion:   "v3.1-20260211",
		Description:  "Latest model with best overall quality.",
		Capabilities: []string{"text_to_3d", "image_to_3d", "multiview_to_3d"},
	},
}

var (
	modelVersionMap     = buildModelVersionMap()
	defaultModelVersion = modelVersionMap["v3.1"]
	multiviewVersions   = buildMultiviewVersions()

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
	return versions
}

func buildMultiviewVersions() map[string]bool {
	versions := make(map[string]bool)
	for _, model := range modelCatalog {
		if hasCapability(model.Capabilities, "multiview_to_3d") {
			versions[model.APIVersion] = true
		}
	}
	return versions
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
			Description:  model.Description,
			Capabilities: append([]string(nil), model.Capabilities...),
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
	version = strings.TrimSpace(version)
	if version == "" {
		return defaultModelVersion
	}
	if full, ok := modelVersionMap[strings.ToLower(version)]; ok {
		return full
	}
	return version
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
