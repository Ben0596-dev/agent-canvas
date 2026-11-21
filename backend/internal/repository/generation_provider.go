package repository

import "context"

const (
	GenerationRunning   = "running"
	GenerationSucceeded = "succeeded"
	GenerationFailed    = "failed"
)

type ImageGenerationRequest struct {
	Model  string
	Prompt string
	Size   string
}

type ImageGenerationResult struct {
	URL string
}

type VideoGenerationRequest struct {
	RequestID string
	Model     string
	Prompt    string
	ImageURL  string
	Size      string
	FPS       int
	Quality   string
	WithAudio bool
}

type VideoGenerationResult struct {
	ProviderTaskID string
	Status         string
	URL            string
	CoverURL       string
	Error          string
}

type ImageGenerationProvider interface {
	GenerateImage(ctx context.Context, request ImageGenerationRequest) (ImageGenerationResult, error)
}

type VideoGenerationProvider interface {
	SubmitVideo(ctx context.Context, request VideoGenerationRequest) (VideoGenerationResult, error)
	GetVideoResult(ctx context.Context, providerTaskID string) (VideoGenerationResult, error)
}
