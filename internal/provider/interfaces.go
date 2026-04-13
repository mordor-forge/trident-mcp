package provider

import "context"

// ModelGenerator handles 3D model creation from various inputs.
type ModelGenerator interface {
	TextToModel(ctx context.Context, req TextToModelRequest) (*ModelOperation, error)
	ImageToModel(ctx context.Context, req ImageToModelRequest) (*ModelOperation, error)
	MultiviewToModel(ctx context.Context, req MultiviewToModelRequest) (*ModelOperation, error)
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

// ModelLister provides model/version discovery.
type ModelLister interface {
	ListModels(ctx context.Context) ([]ModelInfo, error)
}
