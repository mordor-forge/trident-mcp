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
	Quad            *bool  `json:"quad,omitempty" jsonschema:"Generate quad mesh output when supported"`
	SmartLowPoly    *bool  `json:"smartLowPoly,omitempty" jsonschema:"Enable Tripo smart low-poly mesh optimization when supported"`
	GenerateParts   *bool  `json:"generateParts,omitempty" jsonschema:"Generate semantic mesh parts when supported"`
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
	Quad               *bool  `json:"quad,omitempty" jsonschema:"Generate quad mesh output when supported"`
	SmartLowPoly       *bool  `json:"smartLowPoly,omitempty" jsonschema:"Enable Tripo smart low-poly mesh optimization when supported"`
	GenerateParts      *bool  `json:"generateParts,omitempty" jsonschema:"Generate semantic mesh parts when supported"`
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
	TaskID             string   `json:"taskId,omitempty" jsonschema:"Successful image_to_multiview or edit_multiview task ID to reuse directly. Mutually exclusive with imagePaths and imageUrls"`
	ModelVersion       string   `json:"modelVersion,omitempty" jsonschema:"Model version (e.g. v3.0, v3.1, p1). Defaults to the latest supported H3 model"`
	FaceLimit          int      `json:"faceLimit,omitempty" jsonschema:"Target polygon face count"`
	Texture            *bool    `json:"texture,omitempty" jsonschema:"Enable texturing. Set false to request a base model without textures"`
	PBR                *bool    `json:"pbr,omitempty" jsonschema:"Enable PBR materials. When true, Tripo will also enable texture output"`
	ModelSeed          *int     `json:"modelSeed,omitempty" jsonschema:"Optional seed for geometry generation"`
	TextureSeed        *int     `json:"textureSeed,omitempty" jsonschema:"Optional seed for texture generation"`
	TextureQuality     string   `json:"textureQuality,omitempty" jsonschema:"Texture quality: standard or detailed"`
	Quad               *bool    `json:"quad,omitempty" jsonschema:"Generate quad mesh output when supported"`
	SmartLowPoly       *bool    `json:"smartLowPoly,omitempty" jsonschema:"Enable Tripo smart low-poly mesh optimization when supported"`
	GenerateParts      *bool    `json:"generateParts,omitempty" jsonschema:"Generate semantic mesh parts when supported"`
	TextureAlignment   string   `json:"textureAlignment,omitempty" jsonschema:"Texture alignment priority: original_image or geometry"`
	EnableImageAutofix *bool    `json:"enableImageAutofix,omitempty" jsonschema:"Let Tripo optimize the ordered input views before generation"`
	AutoSize           *bool    `json:"autoSize,omitempty" jsonschema:"Automatically scale the model to real-world dimensions in meters"`
	Orientation        string   `json:"orientation,omitempty" jsonschema:"Model orientation: default or align_image"`
	Compress           string   `json:"compress,omitempty" jsonschema:"Compression mode. Set to geometry to request geometry compression"`
	ExportUV           *bool    `json:"exportUV,omitempty" jsonschema:"Control whether UV unwrapping is performed during generation"`
	GeometryQuality    string   `json:"geometryQuality,omitempty" jsonschema:"Geometry detail mode for H3 models: standard or detailed"`
}

// TextToImageRequest describes a text-to-image generation request.
type TextToImageRequest struct {
	Prompt         string `json:"prompt" jsonschema:"Text prompt for image generation"`
	NegativePrompt string `json:"negativePrompt,omitempty" jsonschema:"What to avoid in the generated image"`
	Model          string `json:"model,omitempty" jsonschema:"Image model, for example seedream_v5, gemini-3-pro, or chat_image_2"`
	Template       string `json:"template,omitempty" jsonschema:"Optional Tripo image-generation template such as asset_extraction or t_pose"`
	Width          int    `json:"width,omitempty" jsonschema:"Optional output width"`
	Height         int    `json:"height,omitempty" jsonschema:"Optional output height"`
	Size           string `json:"size,omitempty" jsonschema:"Optional output size string if supported by the selected image model"`
	Seed           *int   `json:"seed,omitempty" jsonschema:"Optional deterministic seed"`
}

// ImageToImageRequest describes an image-to-image generation request.
type ImageToImageRequest struct {
	Input          string `json:"input,omitempty" jsonschema:"Image URL or file token. Mutually exclusive with imagePath and imageUrl"`
	ImagePath      string `json:"imagePath,omitempty" jsonschema:"Local image path to upload as input"`
	ImageURL       string `json:"imageUrl,omitempty" jsonschema:"Public image URL to use as input"`
	Prompt         string `json:"prompt,omitempty" jsonschema:"Edit prompt"`
	NegativePrompt string `json:"negativePrompt,omitempty" jsonschema:"What to avoid in the generated image"`
	Model          string `json:"model,omitempty" jsonschema:"Image model, for example seedream_v5, gemini-3-pro, or chat_image_2"`
	Template       string `json:"template,omitempty" jsonschema:"Optional image-generation template"`
	Width          int    `json:"width,omitempty" jsonschema:"Optional output width"`
	Height         int    `json:"height,omitempty" jsonschema:"Optional output height"`
	Size           string `json:"size,omitempty" jsonschema:"Optional output size string"`
	Seed           *int   `json:"seed,omitempty" jsonschema:"Optional deterministic seed"`
}

// ImageToMultiviewRequest describes a request to create orthographic views from one image.
type ImageToMultiviewRequest struct {
	Input     string `json:"input,omitempty" jsonschema:"Image URL or file token. Mutually exclusive with imagePath and imageUrl"`
	ImagePath string `json:"imagePath,omitempty" jsonschema:"Local image path to upload as input"`
	ImageURL  string `json:"imageUrl,omitempty" jsonschema:"Public image URL to use as input"`
}

// MultiviewEditPrompt targets an edit instruction at one generated view.
type MultiviewEditPrompt struct {
	Prompt string `json:"prompt" jsonschema:"Edit instruction for this view"`
	View   string `json:"view" jsonschema:"Target view: front, left, back, or right"`
}

// EditMultiviewRequest describes a request to edit generated multiview images.
type EditMultiviewRequest struct {
	Input   string                `json:"input" jsonschema:"Multiview task ID, image URL, or file token"`
	Prompts []MultiviewEditPrompt `json:"prompts,omitempty" jsonschema:"Per-view edit instructions. Each entry requires prompt and view"`
	Prompt  string                `json:"prompt,omitempty" jsonschema:"Convenience edit instruction. If view is omitted, applies to all four views"`
	View    string                `json:"view,omitempty" jsonschema:"Target view for prompt: front, left, back, or right"`
}

// ImageToSplatRequest describes an image-to-Gaussian-splat generation request.
type ImageToSplatRequest struct {
	Input     string `json:"input,omitempty" jsonschema:"Image URL or file token. Mutually exclusive with imagePath and imageUrl"`
	ImagePath string `json:"imagePath,omitempty" jsonschema:"Local image path to upload as input"`
	ImageURL  string `json:"imageUrl,omitempty" jsonschema:"Public image URL to use as input"`
}

// RetopologyRequest describes a retopology/lowpoly conversion request.
type RetopologyRequest struct {
	OriginalTaskID string   `json:"originalTaskId" jsonschema:"Task ID of the model to retopologize"`
	Model          string   `json:"model,omitempty" jsonschema:"Optional Tripo model/version to use for decimation"`
	Quad           bool     `json:"quad,omitempty" jsonschema:"Produce quad mesh instead of triangles"`
	TargetFaces    int      `json:"targetFaces,omitempty" jsonschema:"Target face count for the lowpoly output"`
	PartNames      []string `json:"partNames,omitempty" jsonschema:"Specific segmented part names to decimate"`
	Bake           *bool    `json:"bake,omitempty" jsonschema:"Bake textures into the decimated output when supported"`
}

// ConvertRequest describes a format conversion request.
type ConvertRequest struct {
	OriginalTaskID      string   `json:"originalTaskId" jsonschema:"Task ID of the model to convert"`
	Format              string   `json:"format" jsonschema:"Desired output format: GLTF, FBX, OBJ, STL, USDZ, or 3MF"`
	Quad                *bool    `json:"quad,omitempty" jsonschema:"Convert to quad mesh output when supported"`
	FaceLimit           int      `json:"faceLimit,omitempty" jsonschema:"Target face count for converted output"`
	TextureSize         int      `json:"textureSize,omitempty" jsonschema:"Target texture size for converted output"`
	TextureFormat       string   `json:"textureFormat,omitempty" jsonschema:"Texture image format, such as png, jpg, or webp"`
	Bake                *bool    `json:"bake,omitempty" jsonschema:"Bake material textures during conversion"`
	PackUV              *bool    `json:"packUV,omitempty" jsonschema:"Pack UVs during conversion"`
	ExportVertexColors  *bool    `json:"exportVertexColors,omitempty" jsonschema:"Export vertex colors when supported"`
	PivotToCenterBottom *bool    `json:"pivotToCenterBottom,omitempty" jsonschema:"Move pivot to the model center-bottom"`
	ScaleFactor         float64  `json:"scaleFactor,omitempty" jsonschema:"Scale multiplier for converted output"`
	PartNames           []string `json:"partNames,omitempty" jsonschema:"Specific segmented part names to export"`
	ExportOrientation   string   `json:"exportOrientation,omitempty" jsonschema:"Target export orientation: +x, -x, +y, or -y. x_up and y_up are accepted as aliases"`
	FBXPreset           string   `json:"fbxPreset,omitempty" jsonschema:"FBX preset when exporting FBX, such as unity or unreal"`
}

// StylizeRequest describes a stylization request.
type StylizeRequest struct {
	OriginalTaskID string `json:"originalTaskId" jsonschema:"Task ID of the model to stylize"`
	Style          string `json:"style" jsonschema:"Stylization style: lego, voxel, voronoi, or minecraft"`
	BlockSize      int    `json:"blockSize,omitempty" jsonschema:"Optional block size for minecraft style"`
}

// ImportModelRequest imports a model file or URL into Tripo.
type ImportModelRequest struct {
	Input    string `json:"input,omitempty" jsonschema:"Model file URL or file token. Mutually exclusive with filePath and fileUrl"`
	FilePath string `json:"filePath,omitempty" jsonschema:"Local model path to upload before import"`
	FileURL  string `json:"fileUrl,omitempty" jsonschema:"Public model URL to import"`
}

// RefineModelRequest refines an existing model task.
type RefineModelRequest struct {
	Input string `json:"input" jsonschema:"Task ID of the model to refine"`
	Model string `json:"model,omitempty" jsonschema:"Optional refinement model version"`
}

// TextureModelRequest generates or regenerates textures for an existing model.
type TextureModelRequest struct {
	Input            string   `json:"input" jsonschema:"Task ID, model URL, or file token to texture"`
	Model            string   `json:"model,omitempty" jsonschema:"Optional texture model version"`
	Texture          *bool    `json:"texture,omitempty" jsonschema:"Enable texture generation"`
	PBR              *bool    `json:"pbr,omitempty" jsonschema:"Enable PBR materials"`
	TextureSeed      *int     `json:"textureSeed,omitempty" jsonschema:"Optional seed for texture generation"`
	TextureQuality   string   `json:"textureQuality,omitempty" jsonschema:"Texture quality: standard, detailed, or extreme"`
	TextureAlignment string   `json:"textureAlignment,omitempty" jsonschema:"Texture alignment priority: original_image or geometry"`
	PartNames        []string `json:"partNames,omitempty" jsonschema:"Specific segmented part names to texture"`
	TextPrompt       string   `json:"textPrompt,omitempty" jsonschema:"Text prompt for texture generation"`
	ImagePrompt      string   `json:"imagePrompt,omitempty" jsonschema:"Image prompt URL or file token for texture generation"`
	StyleImage       string   `json:"styleImage,omitempty" jsonschema:"Style image URL or file token for texture generation"`
	Compress         *bool    `json:"compress,omitempty" jsonschema:"Compress model geometry"`
	Bake             *bool    `json:"bake,omitempty" jsonschema:"Bake generated textures"`
}

// SegmentMeshRequest segments a model into editable parts.
type SegmentMeshRequest struct {
	Input string `json:"input" jsonschema:"Task ID, model URL, or file token to segment"`
	Model string `json:"model,omitempty" jsonschema:"Optional segmentation model version"`
}

// CompleteMeshRequest completes a segmented model or selected parts.
type CompleteMeshRequest struct {
	Input     string   `json:"input" jsonschema:"Segmentation task ID or model source"`
	Model     string   `json:"model,omitempty" jsonschema:"Optional completion model version"`
	PartNames []string `json:"partNames,omitempty" jsonschema:"Part names to complete"`
}

// RigCheckRequest checks whether a model can be rigged.
type RigCheckRequest struct {
	Input string `json:"input" jsonschema:"Task ID, model URL, or file token to check"`
}

// RigCheckResult reports rigging compatibility.
type RigCheckResult struct {
	Riggable bool   `json:"riggable"`
	RigType  string `json:"rigType,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

// RigModelRequest rigs a model for animation.
type RigModelRequest struct {
	Input     string `json:"input" jsonschema:"Task ID, model URL, or file token to rig"`
	Model     string `json:"model,omitempty" jsonschema:"Rigging model: rig-v2.0 or rig-v1.0"`
	RigType   string `json:"rigType,omitempty" jsonschema:"Rig type such as biped, quadruped, hexapod, octopod, avian, serpentine, or aquatic"`
	Spec      string `json:"spec,omitempty" jsonschema:"Skeleton specification: tripo or mixamo"`
	OutFormat string `json:"outFormat,omitempty" jsonschema:"Output format: glb or fbx"`
}

// RetargetAnimationRequest applies one or more animations to a rigged model.
type RetargetAnimationRequest struct {
	Input              string   `json:"input" jsonschema:"Rigged model task ID"`
	Animation          string   `json:"animation,omitempty" jsonschema:"Single preset animation identifier"`
	Animations         []string `json:"animations,omitempty" jsonschema:"Multiple preset animation identifiers"`
	OutFormat          string   `json:"outFormat,omitempty" jsonschema:"Output format: glb or fbx"`
	BakeAnimation      *bool    `json:"bakeAnimation,omitempty" jsonschema:"Bake animation into the model"`
	ExportWithGeometry *bool    `json:"exportWithGeometry,omitempty" jsonschema:"Export with geometry included"`
	AnimateInPlace     *bool    `json:"animateInPlace,omitempty" jsonschema:"Play animation in place without displacement"`
}

// BatchTasksResult contains a batch task query response.
type BatchTasksResult struct {
	Tasks  map[string]ModelTaskStatus `json:"tasks"`
	Missed []string                   `json:"missed"`
}

// AccountBalance contains account credit balance information.
type AccountBalance struct {
	Balance float64 `json:"balance"`
	Frozen  float64 `json:"frozen"`
}

// AccountUsageRecord contains one account credit usage entry.
type AccountUsageRecord struct {
	TaskID          string  `json:"taskId"`
	Type            string  `json:"type,omitempty"`
	CreditsConsumed float64 `json:"creditsConsumed"`
	CreatedAt       string  `json:"createdAt,omitempty"`
}

// AccountUsageResult contains account credit usage history.
type AccountUsageResult struct {
	Records []AccountUsageRecord `json:"records"`
}

// CreateFileUploadRequest requests a presigned upload URL.
type CreateFileUploadRequest struct {
	Format string `json:"format" jsonschema:"File extension without leading dot, such as glb, fbx, png, or webp"`
}

// UploadFileRequest uploads a local file to Tripo's file-token API.
type UploadFileRequest struct {
	FilePath string `json:"filePath" jsonschema:"Local file path to upload to Tripo"`
}

// FileUpload contains presigned upload metadata.
type FileUpload struct {
	PresignedURL string `json:"presignedUrl"`
	FileToken    string `json:"fileToken"`
	ExpiresIn    int    `json:"expiresIn"`
}

// ModelOperation represents an in-progress async 3D generation task.
type ModelOperation struct {
	TaskID string `json:"taskId"`
	Status string `json:"status"`
}

// ModelTaskStatus represents the current state of a 3D generation task.
type ModelTaskStatus struct {
	TaskID          string         `json:"taskId"`
	Type            string         `json:"type,omitempty"`
	Status          string         `json:"status"`
	Progress        int            `json:"progress"`
	Output          *TaskOutput    `json:"output,omitempty"`
	Error           string         `json:"error,omitempty"`
	CreditsConsumed float64        `json:"creditsConsumed,omitempty"`
	CreatedAt       string         `json:"createdAt,omitempty"`
	CompletedAt     string         `json:"completedAt,omitempty"`
	RawOutput       map[string]any `json:"rawOutput,omitempty"`
}

// TaskOutput contains known Tripo v3 task output fields while preserving
// unknown output keys that can vary by endpoint.
type TaskOutput struct {
	ModelURL         string         `json:"modelUrl,omitempty"`
	BaseModelURL     string         `json:"baseModelUrl,omitempty"`
	PBRModelURL      string         `json:"pbrModelUrl,omitempty"`
	RenderedImageURL string         `json:"renderedImageUrl,omitempty"`
	ImageURL         string         `json:"imageUrl,omitempty"`
	ImageURLs        []string       `json:"imageUrls,omitempty"`
	ModelURLs        []string       `json:"modelUrls,omitempty"`
	Extra            map[string]any `json:"extra,omitempty"`
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
	Namespace    string   `json:"namespace,omitempty"`
	Description  string   `json:"description"`
	Capabilities []string `json:"capabilities"`
	Aliases      []string `json:"aliases,omitempty"`
	Default      bool     `json:"default,omitempty"`
}
