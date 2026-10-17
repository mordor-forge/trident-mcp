package tripo

import (
	"fmt"
	"slices"
	"strings"
)

// CatalogDate is when the Catalog was last checked against Tripo's docs.
const CatalogDate = "2026-10-17"

// ModelVersion is a Tripo AI model an Operation can run on. The Catalog is
// static on purpose (ADR-0002): it is what this server knows, not what a
// given account is entitled to. Unknown versions are still sent to Tripo,
// with a warning, so new releases work before the Catalog catches up.
type ModelVersion struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Family      string   `json:"family"`
	Description string   `json:"description"`
	Aliases     []string `json:"aliases,omitempty"`
	// Operations that accept this version. Empty means every Operation of the family.
	Operations []string `json:"operations"`
	// DefaultFor lists the Operations that use this version when none is given.
	DefaultFor []string `json:"default_for,omitempty"`
	// Rejects lists request fields Tripo refuses for this version. Unset or
	// false values of them are dropped; anything else is an error.
	Rejects []string `json:"rejects,omitempty"`
	Stage   string   `json:"stage,omitempty"`
}

var familyOperations = map[string][]string{
	"3d":         {"text_to_model", "image_to_model", "multiview_to_model"},
	"image":      {"text_to_image", "image_to_image"},
	"texture":    {"texture_model"},
	"retopology": {"retopologize"},
	"uv":         {"smart_uv"},
	"segment":    {"segment_mesh"},
	"complete":   {"complete_mesh"},
	"rig":        {"rig_model"},
}

// Fields only H3.x supports.
var h3OnlyFields = []string{"texture_quality", "geometry_quality", "auto_size", "quad", "smart_low_poly", "generate_parts", "compress"}

var catalog = []ModelVersion{
	{ID: "v3.1-20260211", Name: "Tripo H3.1", Family: "3d", Aliases: []string{"v3.1", "h3.1", "tripo-v3.1"},
		DefaultFor:  familyOperations["3d"],
		Description: "Default high-fidelity model. Best overall geometry and texture quality."},
	{ID: "v3.0-20250812", Name: "Tripo H3.0", Family: "3d", Aliases: []string{"v3.0", "h3.0", "tripo-v3.0"},
		Description: "Previous high-fidelity model."},
	{ID: "v2.5-20250123", Name: "Tripo v2.5", Family: "3d", Aliases: []string{"v2.5", "tripo-v2.5"},
		Rejects:     h3OnlyFields,
		Description: "Legacy model; lacks the H3 quality, quad, low-poly and parts options."},
	{ID: "P1-20260311", Name: "Tripo P1", Family: "3d", Aliases: []string{"p1", "tripo-p1"},
		Rejects:     []string{"quad", "smart_low_poly", "generate_parts", "geometry_quality"},
		Description: "Low-poly, game-ready topology (50-20,000 faces). Control the budget with face_limit."},
	{ID: "P2-20260801", Name: "Tripo P2", Family: "3d", Aliases: []string{"p2", "tripo-p2"}, Stage: "preview",
		Rejects:     []string{"smart_low_poly", "generate_parts", "geometry_quality"},
		Description: "Next-generation low-poly topology (48-50,000 faces; quad up to 25,000). About 3x P1's cost."},

	{ID: "seedream_v4", Name: "Seedream v4", Family: "image", Operations: []string{"text_to_image"},
		DefaultFor:  []string{"text_to_image"},
		Description: "Text-to-image default. Supports watermark."},
	{ID: "seedream_v5", Name: "Seedream v5", Family: "image", DefaultFor: []string{"image_to_image"},
		Description: "Image-to-image default; up to 4 reference images."},
	{ID: "banana", Name: "Nano Banana", Family: "image", Aliases: []string{"gemini-2.5-flash"},
		Description: "Google Gemini 2.5 Flash Image; up to 1K. Supports aspect_ratio."},
	{ID: "banana2", Name: "Nano Banana 2", Family: "image", Aliases: []string{"gemini-3.1-flash"},
		Description: "Google Gemini 3.1 Flash Image; up to 4K, extra-wide aspect ratios."},
	{ID: "banana_pro", Name: "Nano Banana Pro", Family: "image", Aliases: []string{"gemini-3-pro"},
		Description: "Google Gemini 3 Pro Image; up to 4K, up to 10 reference images."},
	{ID: "chat_image_1.5", Name: "Chat Image 1.5", Family: "image", Stage: "retiring 2026-12-01",
		Description: "OpenAI image model; up to 1K."},
	{ID: "chat_image_2", Name: "Chat Image 2", Family: "image",
		Description: "OpenAI image model; supports quality and custom sizes, up to 16 reference images."},
	{ID: "chat_image_2.5_flare", Name: "Chat Image 2.5 Flare", Family: "image",
		Description: "OpenAI image model; adds quality xhigh/max and transparent backgrounds."},
	{ID: "chat_image_2.5_sunburst", Name: "Chat Image 2.5 Sunburst", Family: "image",
		Description: "OpenAI image model; adds quality xhigh/max and transparent backgrounds."},

	{ID: "v3.5-20260815", Name: "Texture v3.5", Family: "texture", Aliases: []string{"v3.5"},
		Description: "Newest texturing model; enables texture_quality fast and delight."},
	{ID: "v3.0-20250812", Name: "Texture v3.0", Family: "texture", Aliases: []string{"v3.0"},
		DefaultFor: []string{"texture_model"}, Rejects: []string{"delight"},
		Description: "Default texturing model."},
	{ID: "v2.5-20250123", Name: "Texture v2.5", Family: "texture", Aliases: []string{"v2.5"},
		Rejects:     []string{"delight"},
		Description: "Legacy texturing model."},

	{ID: "v2.0", Name: "Retopology v2.0", Family: "retopology", DefaultFor: []string{"retopologize"},
		Rejects:     []string{"smart_uv"},
		Description: "Default smart low-poly retopology (500-20,000 triangles, 500-10,000 quads)."},
	{ID: "v3.0", Name: "Retopology v3.0", Family: "retopology",
		Description: "Newest retopology; the only version that supports smart_uv."},
	{ID: "v1.0", Name: "Retopology v1.0", Family: "retopology",
		Rejects:     []string{"smart_uv", "bake", "part_names"},
		Description: "Plain decimation; face_limit is required (up to 2M triangles / 150k quads)."},

	{ID: "default", Name: "Smart UV", Family: "uv", DefaultFor: []string{"smart_uv"},
		Description: "UV unwrapping for meshes up to about 80k triangles."},

	{ID: "v1.0-20250506", Name: "Segmentation v1.0", Family: "segment", DefaultFor: []string{"segment_mesh"},
		Rejects:     []string{"segmentation_granularity", "split_by_connectivity", "ref_image"},
		Description: "Geometry-based part segmentation."},
	{ID: "v2.0-20260430", Name: "Segmentation v2.0", Family: "segment", Stage: "beta",
		Description: "Semantic, labelled segmentation with granularity and reference-image control."},

	{ID: "v1.0-20250506", Name: "Completion v1.0", Family: "complete", DefaultFor: []string{"complete_mesh"},
		Description: "Fills in segmented parts."},

	{ID: "v1.0-20240301", Name: "Rig v1.0", Family: "rig", Aliases: []string{"v1.0", "rig-v1.0"},
		DefaultFor:  []string{"rig_model"},
		Description: "Biped rigs only, with 90+ preset:biped:* animations."},
	{ID: "v2.5-20260210", Name: "Rig v2.5", Family: "rig", Aliases: []string{"v2.5", "rig-v2.5", "rig-v2.0"},
		Description: "All creature types (quadruped, hexapod, avian, ...); no humanoid presets."},
	{ID: "v3.0-20260909", Name: "Rig v3.0", Family: "rig", Aliases: []string{"v3.0", "rig-v3.0"},
		Description: "Best humanoid rigs; adds rig_type others and the unity/unreal/vrm/actorcore specs."},
}

func init() {
	for i := range catalog {
		if len(catalog[i].Operations) == 0 {
			catalog[i].Operations = familyOperations[catalog[i].Family]
		}
	}
}

// Catalog returns the built-in model versions, optionally only those an
// Operation accepts.
func Catalog(operation string) []ModelVersion {
	out := make([]ModelVersion, 0, len(catalog))
	for _, m := range catalog {
		if operation == "" || slices.Contains(m.Operations, operation) {
			out = append(out, m)
		}
	}
	return out
}

// lookupModel finds the version an Operation accepts under id or alias,
// ignoring case.
func lookupModel(operation, value string) (ModelVersion, bool) {
	value = strings.TrimSpace(value)
	for _, m := range catalog {
		if !slices.Contains(m.Operations, operation) {
			continue
		}
		if strings.EqualFold(m.ID, value) || slices.ContainsFunc(m.Aliases, func(a string) bool { return strings.EqualFold(a, value) }) {
			return m, true
		}
	}
	return ModelVersion{}, false
}

// modelID resolves an id or alias for an Operation, returning value
// unchanged when the Catalog does not know it.
func modelID(operation, value string) string {
	if m, ok := lookupModel(operation, value); ok {
		return m.ID
	}
	return value
}

func defaultModel(operation string) (ModelVersion, bool) {
	for _, m := range catalog {
		if slices.Contains(m.DefaultFor, operation) {
			return m, true
		}
	}
	return ModelVersion{}, false
}

// applyModel resolves body["model"] for an Operation: aliases become IDs, an
// absent model becomes the Operation's default, and fields the version
// rejects are dropped when unset or refused when set.
func applyModel(operation string, body map[string]any) (model string, warnings []string, err error) {
	raw, _ := body["model"].(string)
	var version ModelVersion
	var known bool
	if strings.TrimSpace(raw) == "" {
		version, known = defaultModel(operation)
	} else {
		version, known = lookupModel(operation, raw)
	}
	switch {
	case known:
		body["model"] = version.ID
	case strings.TrimSpace(raw) != "":
		body["model"] = strings.TrimSpace(raw)
		warnings = append(warnings, fmt.Sprintf("model %q is not in the built-in catalog for %s (checked %s); sent to Tripo as-is", raw, operation, CatalogDate))
		return body["model"].(string), warnings, nil
	default:
		return "", nil, nil
	}
	if err := applyRejects(version, body); err != nil {
		return "", nil, err
	}
	return version.ID, warnings, nil
}

func applyRejects(version ModelVersion, body map[string]any) error {
	for _, field := range version.Rejects {
		value, present := body[field]
		if !present {
			continue
		}
		if isZeroJSON(value) {
			delete(body, field)
			continue
		}
		return fmt.Errorf("tripo: %s (%s) does not support %s", version.ID, version.Name, field)
	}
	return nil
}

func isZeroJSON(v any) bool {
	switch v := v.(type) {
	case nil:
		return true
	case bool:
		return !v
	case float64:
		return v == 0
	case string:
		return v == ""
	case []any:
		return len(v) == 0
	case map[string]any:
		return len(v) == 0
	default:
		return false
	}
}
