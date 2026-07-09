package provider

import "context"

// ModelGenerator handles 3D model creation from various inputs.
type ModelGenerator interface {
	TextToModel(ctx context.Context, req TextToModelRequest) (*ModelOperation, error)
	ImageToModel(ctx context.Context, req ImageToModelRequest) (*ModelOperation, error)
	MultiviewToModel(ctx context.Context, req MultiviewToModelRequest) (*ModelOperation, error)
}

// ImageGenerator handles v3 image-generation tasks.
type ImageGenerator interface {
	TextToImage(ctx context.Context, req TextToImageRequest) (*ModelOperation, error)
	ImageToImage(ctx context.Context, req ImageToImageRequest) (*ModelOperation, error)
	ImageToMultiview(ctx context.Context, req ImageToMultiviewRequest) (*ModelOperation, error)
	EditMultiview(ctx context.Context, req EditMultiviewRequest) (*ModelOperation, error)
	ImageToSplat(ctx context.Context, req ImageToSplatRequest) (*ModelOperation, error)
}

// ModelStatus handles async task polling and model download.
type ModelStatus interface {
	Status(ctx context.Context, taskID string) (*ModelTaskStatus, error)
	Download(ctx context.Context, taskID string, format string) (*ModelResult, error)
}

// ModelPostProcessor handles retopology, format conversion, and stylization.
type ModelPostProcessor interface {
	Retopologize(ctx context.Context, req RetopologyRequest) (*ModelOperation, error)
	ConvertFormat(ctx context.Context, req ConvertRequest) (*ModelOperation, error)
	Stylize(ctx context.Context, req StylizeRequest) (*ModelOperation, error)
}

// ModelProcessor handles v3 model processing tasks.
type ModelProcessor interface {
	ImportModel(ctx context.Context, req ImportModelRequest) (*ModelOperation, error)
	RefineModel(ctx context.Context, req RefineModelRequest) (*ModelOperation, error)
	TextureModel(ctx context.Context, req TextureModelRequest) (*ModelOperation, error)
}

// MeshProcessor handles v3 mesh processing tasks.
type MeshProcessor interface {
	SegmentMesh(ctx context.Context, req SegmentMeshRequest) (*ModelOperation, error)
	CompleteMesh(ctx context.Context, req CompleteMeshRequest) (*ModelOperation, error)
}

// Animator handles rigging and animation tasks.
type Animator interface {
	RigCheck(ctx context.Context, req RigCheckRequest) (*RigCheckResult, error)
	RigModel(ctx context.Context, req RigModelRequest) (*ModelOperation, error)
	RetargetAnimation(ctx context.Context, req RetargetAnimationRequest) (*ModelOperation, error)
}

// CommonAPI handles common Tripo v3 APIs.
type CommonAPI interface {
	BatchTasks(ctx context.Context, taskIDs []string) (*BatchTasksResult, error)
	UploadFile(ctx context.Context, req UploadFileRequest) (*FileUpload, error)
	CreateFileUpload(ctx context.Context, req CreateFileUploadRequest) (*FileUpload, error)
	GetBalance(ctx context.Context) (*AccountBalance, error)
	GetUsage(ctx context.Context) (*AccountUsageResult, error)
}

// ModelLister provides model/version discovery.
type ModelLister interface {
	ListModels(ctx context.Context) ([]ModelInfo, error)
}
