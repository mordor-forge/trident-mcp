package provider

// TextToModelRequest describes a text-to-3D generation request.
type TextToModelRequest struct {
	Prompt         string `json:"prompt" jsonschema:"Text description of the 3D model to generate"`
	NegativePrompt string `json:"negativePrompt,omitempty" jsonschema:"What to avoid in the generation"`
	ModelVersion   string `json:"modelVersion,omitempty" jsonschema:"Model version (e.g. v2.5, v3.0, v3.1). Defaults to latest"`
	FaceLimit      int    `json:"faceLimit,omitempty" jsonschema:"Target polygon face count"`
}

// ImageToModelRequest describes an image-to-3D generation request.
type ImageToModelRequest struct {
	ImagePath      string `json:"imagePath,omitempty" jsonschema:"Local file path to the reference image. Mutually exclusive with imageUrl"`
	ImageURL       string `json:"imageUrl,omitempty" jsonschema:"Public URL of the reference image. Mutually exclusive with imagePath"`
	ModelVersion   string `json:"modelVersion,omitempty" jsonschema:"Model version. Defaults to latest"`
	FaceLimit      int    `json:"faceLimit,omitempty" jsonschema:"Target polygon face count"`
	TextureQuality string `json:"textureQuality,omitempty" jsonschema:"Texture quality: standard or detailed"`
	Orientation    string `json:"orientation,omitempty" jsonschema:"Model orientation: default or align_image"`
}

// MultiviewToModelRequest describes a multi-view image-to-3D generation request.
type MultiviewToModelRequest struct {
	ImagePaths   []string `json:"imagePaths,omitempty" jsonschema:"Local file paths to 2-4 reference images from different angles. Mutually exclusive with imageUrls"`
	ImageURLs    []string `json:"imageUrls,omitempty" jsonschema:"Public URLs for 2-4 reference images from different angles. Mutually exclusive with imagePaths"`
	ModelVersion string   `json:"modelVersion,omitempty" jsonschema:"Model version. Defaults to latest"`
	FaceLimit    int      `json:"faceLimit,omitempty" jsonschema:"Target polygon face count"`
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
