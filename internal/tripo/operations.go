package tripo

import (
	"errors"
	"fmt"
	"strings"
)

// Every Operation this server exposes, one per task-creating Tripo v3
// endpoint. To cover a new endpoint, add its request type and an Operation
// here, then register it in internal/server.
var (
	TextToModel = Operation[TextToModelRequest]{
		Name: "text_to_model", Path: "/generation/text-to-model", Produces: ProducesMesh,
		Description: "Generate a 3D mesh from a text prompt.",
		check:       func(r *TextToModelRequest, model string) error { return r.check(model) },
	}
	ImageToModel = Operation[ImageToModelRequest]{
		Name: "image_to_model", Path: "/generation/image-to-model", Produces: ProducesMesh,
		Description: "Generate a 3D mesh from one reference image (a local path, URL, file token, or the task ID of a text_to_image / image_to_image task).",
		check:       func(r *ImageToModelRequest, model string) error { return r.check(model) },
	}
	MultiviewToModel = Operation[MultiviewToModelRequest]{
		Name: "multiview_to_model", Path: "/generation/multiview-to-model", Produces: ProducesMesh,
		Description: "Generate a 3D mesh from 2-4 views of one object (front required), or from a successful image_to_multiview / edit_multiview task.",
		check:       checkMultiview,
		wire:        wireMultiview,
	}

	TextToImage = Operation[TextToImageRequest]{
		Name: "text_to_image", Path: "/generation/text-to-image", Produces: ProducesImage,
		Description: "Generate a reference image from a text prompt. Its task ID can be passed directly as image_to_model input.",
		check: func(r *TextToImageRequest, model string) error {
			if r.Watermark != nil && !strings.HasPrefix(model, "seedream") && knownImageModel(model) {
				return fmt.Errorf("watermark is a seedream option; %s does not take it", model)
			}
			return r.check(model)
		},
	}
	ImageToImage = Operation[ImageToImageRequest]{
		Name: "image_to_image", Path: "/generation/image-to-image", Produces: ProducesImage,
		Description: "Edit or derive an image from one or more reference images.",
		check:       checkImageToImage,
	}
	ImageToMultiview = Operation[ImageToMultiviewRequest]{
		Name: "image_to_multiview", Path: "/generation/image-to-multiview", Produces: ProducesViews,
		Description: "Generate front, left, back and right views from one reference image. Pass the task ID to multiview_to_model or edit_multiview.",
	}
	EditMultiview = Operation[EditMultiviewRequest]{
		Name: "edit_multiview", Path: "/generation/edit-multiview", Produces: ProducesViews,
		Description: "Correct individual views of an image_to_multiview task with per-view prompts.",
		check: func(r *EditMultiviewRequest, _ string) error {
			if len(r.Prompts) == 0 || len(r.Prompts) > 4 {
				return fmt.Errorf("prompts takes 1-4 per-view edits, got %d", len(r.Prompts))
			}
			for i, p := range r.Prompts {
				if strings.TrimSpace(p.Prompt) == "" {
					return fmt.Errorf("prompts[%d].prompt is empty", i)
				}
			}
			return nil
		},
	}
	ImageToSplat = Operation[ImageToSplatRequest]{
		Name: "image_to_splat", Path: "/generation/image-to-splat", Produces: ProducesSplat,
		Description: "Generate a Gaussian splat (.splat) from one image. About 30 credits and 4 minutes.",
	}

	ImportModel = Operation[ImportModelRequest]{
		Name: "import_model", Path: "/models/import", Produces: ProducesMesh,
		Description: "Import an existing mesh (GLB, GLTF, FBX, OBJ, STL; up to 150 MB) so other operations can use it by task ID.",
	}
	TextureModel = Operation[TextureModelRequest]{
		Name: "texture_model", Path: "/models/texture", Produces: ProducesMesh,
		Description: "Generate new textures for a mesh from a text, image, or four-view prompt.",
		check:       checkTexture,
	}
	ConvertModel = Operation[ConvertModelRequest]{
		Name: "convert_model", Path: "/models/convert", Produces: ProducesMesh,
		Description: "Convert a mesh to GLTF, FBX, OBJ, STL, USDZ or 3MF, optionally changing topology, scale, pivot, orientation or textures.",
		check:       checkConvert,
	}

	Retopologize = Operation[RetopologizeRequest]{
		Name: "retopologize", Path: "/mesh/decimate", Produces: ProducesMesh,
		Description: "Rebuild a mesh as clean low-poly triangles or quads (smart low-poly retopology).",
		check: func(r *RetopologizeRequest, model string) error {
			if model == "v1.0" && r.FaceLimit == nil {
				return errors.New("face_limit is required with retopology v1.0")
			}
			return nil
		},
	}
	SmartUV = Operation[SmartUVRequest]{
		Name: "smart_uv", Path: "/mesh/uv", Produces: ProducesMesh,
		Description: "Unwrap UVs for a mesh of up to about 80k triangles; also returns a UV layout preview. 20 credits.",
	}
	SegmentMesh = Operation[SegmentMeshRequest]{
		Name: "segment_mesh", Path: "/mesh/segment", Produces: ProducesSegmentation,
		Description: "Split a mesh into parts. Feed the task ID to complete_mesh.",
	}
	CompleteMesh = Operation[CompleteMeshRequest]{
		Name: "complete_mesh", Path: "/mesh/complete", Produces: ProducesMesh,
		Description: "Fill in the parts of a segment_mesh task, optionally only the named ones.",
	}
	SmartSegment = Operation[SmartSegmentRequest]{
		Name: "smart_segment", Path: "/mesh/smartsegment", Produces: ProducesSegmentation,
		Description: "Segment an object guided by a text hint, from an image (85 credits) or a GLB mesh (55 credits).",
		check: func(r *SmartSegmentRequest, _ string) error {
			if r.SegType == "model" && len(r.Transform) != 16 {
				return errors.New("seg_type model requires transform: 16 numbers, a column-major 4x4 matrix")
			}
			if r.SegType == "image" && len(r.Transform) > 0 {
				return errors.New("transform only applies to seg_type model")
			}
			return nil
		},
	}

	RigCheck = Operation[RigCheckRequest]{
		Name: "rig_check", Path: "/animations/rig-check", Produces: ProducesRigCheck,
		Description: "Check whether a mesh can be rigged and detect its skeleton type. Free. The verdict arrives as the task's riggable and rig_type outputs.",
	}
	RigModel = Operation[RigModelRequest]{
		Name: "rig_model", Path: "/animations/rig", Produces: ProducesMesh,
		Description: "Auto-rig a mesh for animation. 25 credits. Run rig_check first.",
		check:       checkRig,
	}
	RetargetAnimation = Operation[RetargetAnimationRequest]{
		Name: "retarget_animation", Path: "/animations/retarget", Produces: ProducesMesh,
		Description: "Apply up to 5 preset animations to a rig_model task. 10 credits per animation.",
		check:       checkRetarget,
	}
)

// GenerationOptions are the mesh-generation fields shared by text-, image-
// and multiview-to-model.
type GenerationOptions struct {
	Model           string `json:"model,omitempty" jsonschema:"Model version: v3.1-20260211 (H3.1, default), v3.0-20250812, v2.5-20250123, P1-20260311 (low-poly), P2-20260801 (low-poly preview). Aliases such as v3.1, p1 and p2 work. See list_models."`
	ModelSeed       *int   `json:"model_seed,omitempty" jsonschema:"Seed for geometry generation"`
	TextureSeed     *int   `json:"texture_seed,omitempty" jsonschema:"Seed for texture generation"`
	FaceLimit       *int   `json:"face_limit,omitempty" jsonschema:"Target face count. H3.x up to 1.5M (2M with geometry_quality detailed); P1 50-20,000; P2 48-50,000; quad up to 150k"`
	Texture         *bool  `json:"texture,omitempty" jsonschema:"Generate textures (default true). false is cheaper and required for generate_parts"`
	PBR             *bool  `json:"pbr,omitempty" jsonschema:"Generate PBR materials (default true); implies texture"`
	TextureQuality  string `json:"texture_quality,omitempty" enum:"fast,standard,detailed,extreme" jsonschema:"Texture quality (default standard). detailed is HD (+10 credits), extreme is 8K (+20). fast needs texture_version v3.5-20260815"`
	TextureVersion  string `json:"texture_version,omitempty" catalog:"texture_model" jsonschema:"Texturing model version: v3.5-20260815, v3.0-20250812 or v2.5-20250123. Defaults to the one matching model"`
	Delight         *bool  `json:"delight,omitempty" jsonschema:"Remove baked-in lighting from textures (default true). Only texture_version v3.5-20260815 reads it"`
	GeometryQuality string `json:"geometry_quality,omitempty" enum:"standard,detailed" jsonschema:"detailed (Ultra) adds geometric detail (+20 credits). H3.x only"`
	AutoSize        *bool  `json:"auto_size,omitempty" jsonschema:"Scale the mesh to real-world size in meters"`
	Quad            *bool  `json:"quad,omitempty" jsonschema:"Quad topology; forces FBX output and 10k faces unless face_limit is set (+5 credits). Not on P1"`
	SmartLowPoly    *bool  `json:"smart_low_poly,omitempty" jsonschema:"Hand-crafted-style low-poly topology on H3.x (+10 credits). Use P1/P2 for game-ready low poly instead"`
	GenerateParts   *bool  `json:"generate_parts,omitempty" jsonschema:"Generate separately editable parts (+20 credits). Requires texture false and pbr false. H3.x only"`
	Compress        string `json:"compress,omitempty" enum:"geometry" jsonschema:"Compress the output geometry (meshopt)"`
	ExportUV        *bool  `json:"export_uv,omitempty" jsonschema:"Include UVs in the output (default true)"`
	// Orientation is deliberately not inherited by follow-up Operations, so
	// it is best set on the final convert_model instead.
	ExportOrientation string `json:"export_orientation,omitempty" enum:"+x,-x,+y,-y" jsonschema:"Facing axis of the output (default +x). Later operations do not inherit it; prefer setting it on convert_model"`
}

func (g *GenerationOptions) check(string) error {
	textureVersion := modelID("texture_model", g.TextureVersion)
	if g.TextureQuality == "fast" && textureVersion != "v3.5-20260815" {
		return errors.New("texture_quality fast requires texture_version v3.5-20260815")
	}
	if g.Delight != nil && textureVersion != "" && textureVersion != "v3.5-20260815" {
		return errors.New("delight only applies to texture_version v3.5-20260815")
	}
	if isTrue(g.GenerateParts) && (!isFalse(g.Texture) || !isFalse(g.PBR)) {
		return errors.New("generate_parts requires texture false and pbr false")
	}
	if isTrue(g.GenerateParts) && (isTrue(g.Quad) || isTrue(g.SmartLowPoly)) {
		return errors.New("generate_parts cannot be combined with quad or smart_low_poly: Tripo ignores quad and produces no parts with smart_low_poly")
	}
	return nil
}

type TextToModelRequest struct {
	Prompt         string `json:"prompt" jsonschema:"What to generate: one object, its shape, materials and style, up to 1024 characters"`
	NegativePrompt string `json:"negative_prompt,omitempty" jsonschema:"What to avoid, up to 255 characters"`
	ImageSeed      *int   `json:"image_seed,omitempty" jsonschema:"Seed for the intermediate reference image"`
	GenerationOptions
}

type ImageToModelRequest struct {
	Input              Input  `json:"input" jsonschema:"The reference image: absolute local path (uploaded), public URL, file token, or task ID of a text_to_image / image_to_image task. PNG, JPEG or WebP, at least 256x256, up to 20 MB"`
	EnableImageAutofix *bool  `json:"enable_image_autofix,omitempty" jsonschema:"Let Tripo clean up the input image first"`
	TextureAlignment   string `json:"texture_alignment,omitempty" enum:"original_image,geometry" jsonschema:"Align textures to the original image (default) or to the geometry"`
	Orientation        string `json:"orientation,omitempty" enum:"default,align_image" jsonschema:"align_image rotates the mesh to match the image. Needs texture"`
	GenerationOptions
}

// Views are the four camera slots of a multiview set. Each is an Input.
type Views struct {
	Front Input `json:"front,omitempty" jsonschema:"Front view (required)"`
	Left  Input `json:"left,omitempty" jsonschema:"Left view"`
	Back  Input `json:"back,omitempty" jsonschema:"Back view"`
	Right Input `json:"right,omitempty" jsonschema:"Right view"`
}

type MultiviewToModelRequest struct {
	Inputs           *Views `json:"inputs,omitempty" jsonschema:"Views of one object: front plus at least one of left, back, right. Each is a local path, URL or file token. Mutually exclusive with input"`
	Input            Input  `json:"input,omitempty" jsonschema:"Task ID of a successful image_to_multiview or edit_multiview task. Mutually exclusive with inputs"`
	TextureAlignment string `json:"texture_alignment,omitempty" enum:"original_image,geometry" jsonschema:"Align textures to the original images (default) or to the geometry"`
	Orientation      string `json:"orientation,omitempty" enum:"default,align_image" jsonschema:"align_image rotates the mesh to match the front image. Needs texture"`
	GenerationOptions
}

func checkMultiview(r *MultiviewToModelRequest, model string) error {
	switch {
	case r.Inputs != nil && r.Input != "":
		return errors.New("give either inputs (views) or input (a multiview task ID), not both")
	case r.Inputs == nil && r.Input == "":
		return errors.New("inputs (views) or input (a multiview task ID) is required")
	case r.Inputs != nil:
		if r.Inputs.Front == "" {
			return errors.New("inputs.front is required")
		}
		if r.Inputs.Left == "" && r.Inputs.Back == "" && r.Inputs.Right == "" {
			return errors.New("inputs needs at least one of left, back or right besides front")
		}
	}
	return r.check(model)
}

// wireMultiview converts the views map, or a multiview task ID, into Tripo's
// inputs array: [{"front": ...}, {"back": ...}] or [{"task_id": ...}].
func wireMultiview(body map[string]any) {
	if taskID, ok := body["input"].(string); ok && taskID != "" {
		body["inputs"] = []map[string]any{{"task_id": taskID}}
		delete(body, "input")
		return
	}
	views, _ := body["inputs"].(map[string]any)
	var inputs []map[string]any
	for _, view := range []string{"front", "left", "back", "right"} {
		if v, ok := views[view].(string); ok && v != "" {
			inputs = append(inputs, map[string]any{view: v})
		}
	}
	body["inputs"] = inputs
}

// ImageOptions are the output fields shared by text- and image-to-image.
type ImageOptions struct {
	Model        string `json:"model,omitempty" jsonschema:"Image model: seedream_v4 (text_to_image default), seedream_v5 (image_to_image default), banana, banana2, banana_pro, chat_image_2, chat_image_2.5_flare, chat_image_2.5_sunburst. See list_models"`
	Size         string `json:"size,omitempty" jsonschema:"Output size: 2K, 4K or WxH such as 2048x2048 (default)"`
	Quality      string `json:"quality,omitempty" enum:"low,medium,high,xhigh,max" jsonschema:"chat_image models only (default low); xhigh and max need a 2.5 model"`
	Background   string `json:"background,omitempty" enum:"auto,opaque,transparent" jsonschema:"chat_image 2.5 models only; transparent needs output_format png"`
	AspectRatio  string `json:"aspect_ratio,omitempty" enum:"1:1,2:3,3:2,3:4,4:3,4:5,5:4,9:16,16:9,21:9,1:8,1:4,4:1,8:1" jsonschema:"banana models only (default 1:1); 1:8, 1:4, 4:1 and 8:1 need banana2"`
	OutputFormat string `json:"output_format,omitempty" enum:"png,jpeg" jsonschema:"Image file format (default png)"`
}

type TextToImageRequest struct {
	Prompt    string `json:"prompt" jsonschema:"What to draw. For 3D references: one object, plain background. Append --no <things> to exclude things"`
	Template  string `json:"template,omitempty" enum:"asset_extraction,character_completion,t_pose,variants,figure" jsonschema:"Preset prompt template, e.g. t_pose for rig-ready characters"`
	Watermark *bool  `json:"watermark,omitempty" jsonschema:"Add a watermark (seedream only)"`
	ImageOptions
}

type ImageToImageRequest struct {
	Input    Input   `json:"input,omitempty" jsonschema:"The source image: local path, URL, file token or image task ID"`
	Inputs   []Input `json:"inputs,omitempty" jsonschema:"Several reference images, referenced in the prompt as image[1], image[2], ... (seedream up to 4, banana up to 10, chat_image up to 16)"`
	Prompt   string  `json:"prompt,omitempty" jsonschema:"The edit to make. Required unless template is set"`
	Template string  `json:"template,omitempty" enum:"t_pose,character_completion,3d_enhance,variants,figure" jsonschema:"Preset edit, e.g. 3d_enhance to prepare a photo for image_to_model"`
	ImageOptions
}

func checkImageToImage(r *ImageToImageRequest, model string) error {
	if r.Input == "" && len(r.Inputs) == 0 {
		return errors.New("input or inputs is required")
	}
	if r.Prompt == "" && r.Template == "" {
		return errors.New("prompt is required unless template is set")
	}
	if limit := referenceLimit(model); limit > 0 && len(r.Inputs) > limit {
		return fmt.Errorf("%s takes up to %d reference images, got %d", model, limit, len(r.Inputs))
	}
	return r.check(model)
}

// referenceLimit is how many inputs image_to_image accepts per model family;
// zero means unknown.
func referenceLimit(model string) int {
	switch {
	case strings.HasPrefix(model, "seedream"):
		return 4
	case strings.HasPrefix(model, "banana"):
		return 10
	case strings.HasPrefix(model, "chat_image"):
		return 16
	default:
		return 0
	}
}

func knownImageModel(model string) bool {
	_, ok := lookupModel("text_to_image", model)
	if !ok {
		_, ok = lookupModel("image_to_image", model)
	}
	return ok
}

// check validates the options each image model family supports. Models the
// Catalog does not know are left to Tripo.
func (o *ImageOptions) check(model string) error {
	if o.Background == "transparent" && o.OutputFormat != "" && o.OutputFormat != "png" {
		return errors.New("background transparent requires output_format png")
	}
	if !knownImageModel(model) {
		return nil
	}
	chat25 := strings.HasPrefix(model, "chat_image_2.5")
	switch {
	case o.Quality != "" && !strings.HasPrefix(model, "chat_image"):
		return fmt.Errorf("quality only applies to chat_image models, not %s", model)
	case (o.Quality == "xhigh" || o.Quality == "max") && !chat25:
		return fmt.Errorf("quality %s needs a chat_image_2.5 model", o.Quality)
	case o.Background != "" && !chat25:
		return fmt.Errorf("background only applies to chat_image_2.5 models, not %s", model)
	case o.AspectRatio != "" && !strings.HasPrefix(model, "banana"):
		return fmt.Errorf("aspect_ratio only applies to banana models, not %s; use size", model)
	}
	switch o.AspectRatio {
	case "1:8", "1:4", "4:1", "8:1":
		if model != "banana2" {
			return fmt.Errorf("aspect_ratio %s needs banana2", o.AspectRatio)
		}
	}
	return nil
}

type ImageToMultiviewRequest struct {
	Input Input `json:"input" jsonschema:"The reference image: local path, URL, file token or image task ID"`
}

// MultiviewEdit is one per-view correction.
type MultiviewEdit struct {
	View   string `json:"view" enum:"front,left,back,right" jsonschema:"Which view to edit"`
	Prompt string `json:"prompt" jsonschema:"What to change in that view"`
}

type EditMultiviewRequest struct {
	Input   Input           `json:"input" jsonschema:"Task ID of an image_to_multiview or edit_multiview task"`
	Prompts []MultiviewEdit `json:"prompts" jsonschema:"1-4 per-view edits (5 credits each)"`
}

type ImageToSplatRequest struct {
	Input     Input `json:"input" jsonschema:"The source image: local path, URL or file token. PNG, JPEG or WebP, at least 256 px"`
	ModelSeed *int  `json:"model_seed,omitempty" jsonschema:"Seed for the splat"`
}

type ImportModelRequest struct {
	Input Input `json:"input" jsonschema:"The mesh: local path (GLB, FBX, OBJ, STL), URL (GLTF only as a URL) or file token. Up to 150 MB"`
}

// TexturePrompt guides texture_model. Give exactly one of text, image or images.
type TexturePrompt struct {
	Text       string  `json:"text,omitempty" jsonschema:"Describe the materials and colors"`
	Image      Input   `json:"image,omitempty" jsonschema:"One reference image"`
	Images     []Input `json:"images,omitempty" jsonschema:"Four reference views in front, left, back, right order"`
	StyleImage Input   `json:"style_image,omitempty" jsonschema:"Style reference, used together with text"`
}

type TextureModelRequest struct {
	Input            Input          `json:"input" jsonschema:"The mesh: task ID, local path, URL or file token (GLB, GLTF, FBX, OBJ, STL)"`
	Model            string         `json:"model,omitempty" jsonschema:"Texturing model: v3.0-20250812 (default), v3.5-20260815 (newest; fast quality and delight), v2.5-20250123"`
	TexturePrompt    *TexturePrompt `json:"texture_prompt,omitempty" jsonschema:"What the textures should look like"`
	PBR              *bool          `json:"pbr,omitempty" jsonschema:"Generate PBR materials (default true)"`
	TextureSeed      *int           `json:"texture_seed,omitempty" jsonschema:"Seed for texture generation"`
	TextureAlignment string         `json:"texture_alignment,omitempty" enum:"original_image,geometry" jsonschema:"Align to the reference image or to the geometry"`
	TextureQuality   string         `json:"texture_quality,omitempty" enum:"fast,standard,detailed,extreme" jsonschema:"fast (v3.5 only) and standard cost 10 credits, detailed 20, extreme (8K) 30"`
	PartNames        []string       `json:"part_names,omitempty" jsonschema:"Only texture these parts"`
	Compress         string         `json:"compress,omitempty" enum:"geometry" jsonschema:"Compress the output geometry"`
	Bake             *bool          `json:"bake,omitempty" jsonschema:"Bake textures (default true)"`
	Delight          *bool          `json:"delight,omitempty" jsonschema:"Remove baked-in lighting (default true; v3.5 only)"`
}

func checkTexture(r *TextureModelRequest, model string) error {
	if p := r.TexturePrompt; p != nil {
		given := 0
		for _, set := range []bool{p.Text != "", p.Image != "", len(p.Images) > 0} {
			if set {
				given++
			}
		}
		switch {
		case given != 1:
			return errors.New("texture_prompt takes exactly one of text, image or images")
		case p.StyleImage != "" && p.Text == "":
			return errors.New("texture_prompt.style_image goes together with text")
		case len(p.Images) > 0 && len(p.Images) != 4:
			return fmt.Errorf("texture_prompt.images takes exactly 4 views (front, left, back, right), got %d", len(p.Images))
		}
	}
	if r.TextureQuality == "fast" && model != "v3.5-20260815" {
		return errors.New("texture_quality fast requires model v3.5-20260815")
	}
	return nil
}

type ConvertModelRequest struct {
	Input                  Input    `json:"input" jsonschema:"Task ID of the mesh to convert (or a file token / URL)"`
	Format                 string   `json:"format" enum:"GLTF,FBX,OBJ,STL,USDZ,3MF" jsonschema:"Target format. 3MF is single-color"`
	Quad                   *bool    `json:"quad,omitempty" jsonschema:"Remesh to quads (forces FBX)"`
	ForceSymmetry          *bool    `json:"force_symmetry,omitempty" jsonschema:"Symmetric quad remesh; needs quad"`
	FaceLimit              *int     `json:"face_limit,omitempty" jsonschema:"Target face count (default keeps the original)"`
	FlattenBottom          *bool    `json:"flatten_bottom,omitempty" jsonschema:"Flatten the base, e.g. for 3D printing"`
	FlattenBottomThreshold *float64 `json:"flatten_bottom_threshold,omitempty" jsonschema:"How much of the base to flatten (default 0.01)"`
	TextureSize            *int     `json:"texture_size,omitempty" jsonschema:"Texture resolution in pixels (default 4096)"`
	TextureFormat          string   `json:"texture_format,omitempty" enum:"JPEG,PNG,WEBP,BMP,DPX,HDR,OPEN_EXR,TARGA,TIFF" jsonschema:"Texture image format (default JPEG)"`
	Bake                   *bool    `json:"bake,omitempty" jsonschema:"Bake textures (default true)"`
	PackUV                 *bool    `json:"pack_uv,omitempty" jsonschema:"Pack UV islands"`
	ExportVertexColors     *bool    `json:"export_vertex_colors,omitempty" jsonschema:"Export vertex colors (OBJ and GLTF only)"`
	PivotToCenterBottom    *bool    `json:"pivot_to_center_bottom,omitempty" jsonschema:"Move the pivot to the bottom center"`
	ScaleFactor            *float64 `json:"scale_factor,omitempty" jsonschema:"Uniform scale (default 1)"`
	WithAnimation          *bool    `json:"with_animation,omitempty" jsonschema:"Keep animations (default true)"`
	AnimateInPlace         *bool    `json:"animate_in_place,omitempty" jsonschema:"Remove root motion from animations"`
	PartNames              []string `json:"part_names,omitempty" jsonschema:"Only export these parts"`
	ExportOrientation      string   `json:"export_orientation,omitempty" enum:"+x,-x,+y,-y" jsonschema:"Facing axis (default +x)"`
	FBXPreset              string   `json:"fbx_preset,omitempty" enum:"blender,3dsmax,mixamo,bake_scale" jsonschema:"FBX target application (default blender)"`
}

func checkConvert(r *ConvertModelRequest, _ string) error {
	if isTrue(r.Quad) && r.Format != "FBX" {
		return fmt.Errorf("quad output is always FBX; set format FBX (got %s)", r.Format)
	}
	if isTrue(r.ForceSymmetry) && !isTrue(r.Quad) {
		return errors.New("force_symmetry requires quad")
	}
	return nil
}

type RetopologizeRequest struct {
	Input     Input    `json:"input" jsonschema:"Task ID of the mesh (or a file token / URL)"`
	Model     string   `json:"model,omitempty" enum:"v1.0,v2.0,v3.0" jsonschema:"Retopology version: v2.0 (default), v3.0 (newest, supports smart_uv), v1.0 (plain decimation; needs face_limit)"`
	FaceLimit *int     `json:"face_limit,omitempty" jsonschema:"Target faces: 500-20,000 triangles or 500-10,000 quads on v2.0/v3.0"`
	Quad      *bool    `json:"quad,omitempty" jsonschema:"Quad topology instead of triangles"`
	SmartUV   *bool    `json:"smart_uv,omitempty" jsonschema:"Also unwrap UVs (v3.0 only)"`
	Bake      *bool    `json:"bake,omitempty" jsonschema:"Bake the original textures onto the new mesh (default true; not on v1.0)"`
	PartNames []string `json:"part_names,omitempty" jsonschema:"Only process these parts (not on v1.0)"`
}

type SmartUVRequest struct {
	Input             Input  `json:"input" jsonschema:"Task ID of the mesh (or a file token / URL); up to about 80k triangles"`
	Model             string `json:"model,omitempty" enum:"default" jsonschema:"UV model version (only default)"`
	Bake              *bool  `json:"bake,omitempty" jsonschema:"Bake textures onto the new UVs (default true)"`
	ExportOrientation string `json:"export_orientation,omitempty" enum:"+x,-x,+y,-y" jsonschema:"Facing axis of the output"`
	Compress          string `json:"compress,omitempty" enum:"geometry" jsonschema:"Compress the output geometry"`
}

type SegmentMeshRequest struct {
	Input                   Input  `json:"input" jsonschema:"Task ID of the mesh (or a file token / URL)"`
	Model                   string `json:"model,omitempty" jsonschema:"Segmentation version: v1.0-20250506 (default, geometric) or v2.0-20260430 (beta, semantic)"`
	SegmentationGranularity string `json:"segmentation_granularity,omitempty" enum:"simple,balanced,detailed" jsonschema:"How finely to split (v2.0 only; default balanced)"`
	SplitByConnectivity     *bool  `json:"split_by_connectivity,omitempty" jsonschema:"Also split disconnected pieces (v2.0 only; default true)"`
	RefImage                Input  `json:"ref_image,omitempty" jsonschema:"Reference image guiding the parts (v2.0 only; overrides granularity)"`
}

type CompleteMeshRequest struct {
	Input          Input    `json:"input" jsonschema:"Task ID of a segment_mesh task"`
	Model          string   `json:"model,omitempty" jsonschema:"Completion version (only v1.0-20250506)"`
	PartNames      []string `json:"part_names,omitempty" jsonschema:"Only complete these parts"`
	CompletionMode string   `json:"completion_mode,omitempty" enum:"ai_completion,quick_cap" jsonschema:"ai_completion (default, 50 credits) regenerates hidden geometry; quick_cap (30) just closes holes"`
}

type SmartSegmentRequest struct {
	SegType     string    `json:"seg_type" enum:"image,model" jsonschema:"Segment an image (85 credits) or a GLB mesh (55 credits)"`
	Input       Input     `json:"input" jsonschema:"The image or GLB mesh: local path, URL or file token (task IDs are not accepted)"`
	Granularity string    `json:"granularity,omitempty" enum:"coarse,medium,fine" jsonschema:"How finely to split (default medium)"`
	Hint        string    `json:"hint,omitempty" jsonschema:"Describe the parts you want"`
	Transform   []float64 `json:"transform,omitempty" jsonschema:"seg_type model only (required): 16 numbers, the mesh's column-major 4x4 transform"`
}

type RigCheckRequest struct {
	Input Input `json:"input" jsonschema:"Task ID of the mesh (or a file token / URL; GLB, FBX, OBJ, STL)"`
}

type RigModelRequest struct {
	Input     Input  `json:"input" jsonschema:"Task ID of the mesh to rig"`
	Model     string `json:"model,omitempty" jsonschema:"Rig version: v1.0-20240301 (default, biped only), v2.5-20260210 (all creature types), v3.0-20260909 (best humanoids, adds others)"`
	RigType   string `json:"rig_type,omitempty" enum:"biped,quadruped,hexapod,octopod,avian,serpentine,aquatic,others" jsonschema:"Skeleton type (default biped); use rig_check's rig_type"`
	Spec      string `json:"spec,omitempty" enum:"tripo,mixamo,actorcore,unity,unreal,vrm" jsonschema:"Bone naming for bipeds (default tripo); unity, unreal, vrm and actorcore need v2.5 or v3.0"`
	OutFormat string `json:"out_format,omitempty" enum:"glb,fbx" jsonschema:"Output format (default glb)"`
}

func checkRig(r *RigModelRequest, model string) error {
	if r.Spec != "" && r.Spec != "tripo" && r.RigType != "" && r.RigType != "biped" {
		return fmt.Errorf("spec %s only applies to biped rigs, not rig_type %s", r.Spec, r.RigType)
	}
	if model == "v1.0-20240301" {
		if r.RigType != "" && r.RigType != "biped" {
			return fmt.Errorf("rig v1.0-20240301 only rigs bipeds; use model v2.5-20260210 or v3.0-20260909 for rig_type %s", r.RigType)
		}
		switch r.Spec {
		case "unity", "unreal", "vrm", "actorcore":
			return fmt.Errorf("spec %s needs model v2.5-20260210 or v3.0-20260909", r.Spec)
		}
	}
	if r.RigType == "others" && model != "v3.0-20260909" {
		return errors.New("rig_type others needs model v3.0-20260909")
	}
	return nil
}

type RetargetAnimationRequest struct {
	Input              Input    `json:"input" jsonschema:"Task ID of a rig_model task"`
	Animation          string   `json:"animation,omitempty" jsonschema:"One preset, e.g. preset:idle, preset:walk, preset:run, preset:jump, preset:quadruped:walk, or preset:biped:* for v1.0 rigs"`
	Animations         []string `json:"animations,omitempty" jsonschema:"Up to 5 presets in one task, instead of animation"`
	OutFormat          string   `json:"out_format,omitempty" enum:"glb,fbx" jsonschema:"Output format (default glb)"`
	ExportWithGeometry *bool    `json:"export_with_geometry,omitempty" jsonschema:"Include the mesh, not just the skeleton (default true)"`
	AnimateInPlace     *bool    `json:"animate_in_place,omitempty" jsonschema:"Remove root motion"`
}

func checkRetarget(r *RetargetAnimationRequest, _ string) error {
	switch {
	case r.Animation != "" && len(r.Animations) > 0:
		return errors.New("give either animation or animations, not both")
	case r.Animation == "" && len(r.Animations) == 0:
		return errors.New("animation or animations is required")
	case len(r.Animations) > 5:
		return fmt.Errorf("at most 5 animations per task, got %d", len(r.Animations))
	}
	for _, a := range append([]string{r.Animation}, r.Animations...) {
		if a != "" && !strings.HasPrefix(a, "preset:") {
			return fmt.Errorf("animation %q should be a preset such as preset:walk", a)
		}
	}
	return nil
}

func isTrue(b *bool) bool  { return b != nil && *b }
func isFalse(b *bool) bool { return b != nil && !*b }
