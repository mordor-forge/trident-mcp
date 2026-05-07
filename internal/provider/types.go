package provider

// TextToModelRequest describes a text-to-3D generation request.
type TextToModelRequest struct {
	Prompt          string `json:"prompt" jsonschema:"Text description of the 3D model to generate"`
	NegativePrompt  string `json:"negativePrompt,omitempty" jsonschema:"What to avoid in the generation"`
	ModelVersion    string `json:"modelVersion,omitempty" jsonschema:"Model version (e.g. turbo, v3.0, v3.1, p1). Defaults to the latest supported H3 model"`
	FaceLimit       int    `json:"faceLimit,omitempty" jsonschema:"Target polygon face count"`
	Texture         *bool  `json:"texture,omitempty" jsonschema:"Enable texturing. Set false to request a base model without textures"`
	PBR             *bool  `json:"pbr,omitempty" jsonschema:"Enable PBR materials. When true, Tripo will also enable texture output"`
	ImageSeed       *int   `json:"imageSeed,omitempty" jsonschema:"Optional seed for prompt-to-image generation before 3D reconstruction"`
	ModelSeed       *int   `json:"modelSeed,omitempty" jsonschema:"Optional seed for geometry generation"`
	TextureSeed     *int   `json:"textureSeed,omitempty" jsonschema:"Optional seed for texture generation"`
	TextureQuality  string `json:"textureQuality,omitempty" jsonschema:"Texture quality: standard or detailed"`
	AutoSize        *bool  `json:"autoSize,omitempty" jsonschema:"Automatically scale the model to real-world dimensions in meters"`
	Compress        string `json:"compress,omitempty" jsonschema:"Compression mode. Set to geometry to request geometry compression"`
	ExportUV        *bool  `json:"exportUV,omitempty" jsonschema:"Control whether UV unwrapping is performed during generation"`
	GeometryQuality string `json:"geometryQuality,omitempty" jsonschema:"Geometry detail mode for H3 models: standard or detailed"`
}

// ImageToModelRequest describes an image-to-3D generation request.
type ImageToModelRequest struct {
	ImagePath          string `json:"imagePath,omitempty" jsonschema:"Local file path to the reference image. Mutually exclusive with imageUrl"`
	ImageURL           string `json:"imageUrl,omitempty" jsonschema:"Public URL of the reference image. Mutually exclusive with imagePath"`
	ModelVersion       string `json:"modelVersion,omitempty" jsonschema:"Model version (e.g. turbo, v3.0, v3.1, p1). Defaults to the latest supported H3 model"`
	FaceLimit          int    `json:"faceLimit,omitempty" jsonschema:"Target polygon face count"`
	Texture            *bool  `json:"texture,omitempty" jsonschema:"Enable texturing. Set false to request a base model without textures"`
	PBR                *bool  `json:"pbr,omitempty" jsonschema:"Enable PBR materials. When true, Tripo will also enable texture output"`
	ModelSeed          *int   `json:"modelSeed,omitempty" jsonschema:"Optional seed for geometry generation"`
	TextureSeed        *int   `json:"textureSeed,omitempty" jsonschema:"Optional seed for texture generation"`
	TextureQuality     string `json:"textureQuality,omitempty" jsonschema:"Texture quality: standard or detailed"`
	TextureAlignment   string `json:"textureAlignment,omitempty" jsonschema:"Texture alignment priority: original_image or geometry"`
	EnableImageAutofix *bool  `json:"enableImageAutofix,omitempty" jsonschema:"Let Tripo optimize the reference image before generation"`
	AutoSize           *bool  `json:"autoSize,omitempty" jsonschema:"Automatically scale the model to real-world dimensions in meters"`
	Orientation        string `json:"orientation,omitempty" jsonschema:"Model orientation: default or align_image"`
	Compress           string `json:"compress,omitempty" jsonschema:"Compression mode. Set to geometry to request geometry compression"`
	ExportUV           *bool  `json:"exportUV,omitempty" jsonschema:"Control whether UV unwrapping is performed during generation"`
	GeometryQuality    string `json:"geometryQuality,omitempty" jsonschema:"Geometry detail mode for H3 models: standard or detailed"`
}

// MultiviewToModelRequest describes a multi-view image-to-3D generation request.
type MultiviewToModelRequest struct {
	ImagePaths         []string `json:"imagePaths,omitempty" jsonschema:"Local file paths for 2-4 ordered views. Supply them in Tripo's expected order: front, left, back, right. Mutually exclusive with imageUrls"`
	ImageURLs          []string `json:"imageUrls,omitempty" jsonschema:"Public URLs for 2-4 ordered views. Supply them in Tripo's expected order: front, left, back, right. Mutually exclusive with imagePaths"`
	ModelVersion       string   `json:"modelVersion,omitempty" jsonschema:"Model version (e.g. v3.0, v3.1, p1). Defaults to the latest supported H3 model"`
	FaceLimit          int      `json:"faceLimit,omitempty" jsonschema:"Target polygon face count"`
	Texture            *bool    `json:"texture,omitempty" jsonschema:"Enable texturing. Set false to request a base model without textures"`
	PBR                *bool    `json:"pbr,omitempty" jsonschema:"Enable PBR materials. When true, Tripo will also enable texture output"`
	ModelSeed          *int     `json:"modelSeed,omitempty" jsonschema:"Optional seed for geometry generation"`
	TextureSeed        *int     `json:"textureSeed,omitempty" jsonschema:"Optional seed for texture generation"`
	TextureQuality     string   `json:"textureQuality,omitempty" jsonschema:"Texture quality: standard or detailed"`
	TextureAlignment   string   `json:"textureAlignment,omitempty" jsonschema:"Texture alignment priority: original_image or geometry"`
	EnableImageAutofix *bool    `json:"enableImageAutofix,omitempty" jsonschema:"Let Tripo optimize the ordered input views before generation"`
	AutoSize           *bool    `json:"autoSize,omitempty" jsonschema:"Automatically scale the model to real-world dimensions in meters"`
	Orientation        string   `json:"orientation,omitempty" jsonschema:"Model orientation: default or align_image"`
	Compress           string   `json:"compress,omitempty" jsonschema:"Compression mode. Set to geometry to request geometry compression"`
	ExportUV           *bool    `json:"exportUV,omitempty" jsonschema:"Control whether UV unwrapping is performed during generation"`
	GeometryQuality    string   `json:"geometryQuality,omitempty" jsonschema:"Geometry detail mode for H3 models: standard or detailed"`
}

// RetopologyRequest describes a retopology/lowpoly conversion request.
type RetopologyRequest struct {
	OriginalTaskID string `json:"originalTaskId" jsonschema:"Task ID of the model to retopologize"`
	Quad           bool   `json:"quad,omitempty" jsonschema:"Produce quad mesh instead of triangles"`
	TargetFaces    int    `json:"targetFaces,omitempty" jsonschema:"Target face count for the lowpoly output"`
}

// ConvertRequest describes a format conversion request.
type ConvertRequest struct {
	OriginalTaskID string `json:"originalTaskId" jsonschema:"Task ID of the model to convert"`
	Format         string `json:"format" jsonschema:"Desired output format: GLTF, FBX, OBJ, STL, USDZ, or 3MF"`
}

// StylizeRequest describes a stylization request.
type StylizeRequest struct {
	OriginalTaskID string `json:"originalTaskId" jsonschema:"Task ID of the model to stylize"`
	Style          string `json:"style" jsonschema:"Stylization style: lego, voxel, voronoi, or minecraft"`
}

// ModelOperation represents an in-progress async 3D generation task.
type ModelOperation struct {
	TaskID string `json:"taskId"`
	Status string `json:"status"`
}

// ModelTaskStatus represents the current state of a 3D generation task.
type ModelTaskStatus struct {
	TaskID   string `json:"taskId"`
	Status   string `json:"status"`
	Progress int    `json:"progress"`
	Error    string `json:"error,omitempty"`
}

// ModelResult contains the result of a completed model download.
type ModelResult struct {
	FilePath     string `json:"filePath"`
	Format       string `json:"format"`
	TaskID       string `json:"taskId"`
	ModelVersion string `json:"modelVersion,omitempty"`
}

// ModelInfo describes an available model version and its capabilities.
type ModelInfo struct {
	Name         string   `json:"name"`
	ID           string   `json:"id"`
	Description  string   `json:"description"`
	Capabilities []string `json:"capabilities"`
}
