package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"agent-canvas/backend/internal/model"
	"agent-canvas/backend/internal/repository"
)

var (
	ErrInvalidGeneration      = errors.New("invalid generation request")
	ErrGenerationUnavailable  = errors.New("generation provider is unavailable")
	ErrGenerationTaskNotFound = errors.New("generation task not found")
)

type CreateGenerationInput struct {
	ModelID   string
	Prompt    string
	ImageURL  string
	Size      string
	FPS       int
	Quality   string
	WithAudio bool
}

type GenerationService struct {
	repository     repository.GenerationRepository
	projectService *ProjectService
	models         *ModelRegistry
	imageProviders map[string]repository.ImageGenerationProvider
	videoProviders map[string]repository.VideoGenerationProvider
	now            func() time.Time
	id             func() string
}

func NewGenerationService(
	repository repository.GenerationRepository,
	projectService *ProjectService,
	models *ModelRegistry,
	imageProviders map[string]repository.ImageGenerationProvider,
	videoProviders map[string]repository.VideoGenerationProvider,
	now func() time.Time,
	id func() string,
) *GenerationService {
	return &GenerationService{
		repository: repository, projectService: projectService, models: models,
		imageProviders: imageProviders, videoProviders: videoProviders, now: now, id: id,
	}
}

func (service *GenerationService) Create(ctx context.Context, projectID string, input CreateGenerationInput) (model.GenerationTask, error) {
	if _, err := service.projectService.Get(ctx, projectID); err != nil {
		return model.GenerationTask{}, err
	}
	profile, err := service.models.Resolve(input.ModelID)
	if err != nil {
		return model.GenerationTask{}, fmt.Errorf("%w: %v", ErrInvalidGeneration, err)
	}
	input.Prompt = strings.TrimSpace(input.Prompt)
	if input.Prompt == "" {
		return model.GenerationTask{}, fmt.Errorf("%w: prompt is required", ErrInvalidGeneration)
	}
	if profile.Capability == model.GenerationKindVideo && utf8.RuneCountInString(input.Prompt) > 512 {
		return model.GenerationTask{}, fmt.Errorf("%w: video prompt exceeds 512 characters", ErrInvalidGeneration)
	}
	applyGenerationDefaults(&input, profile)
	if input.FPS != 0 && input.FPS != 30 && input.FPS != 60 {
		return model.GenerationTask{}, fmt.Errorf("%w: fps must be 30 or 60", ErrInvalidGeneration)
	}
	if input.Quality != "" && input.Quality != "speed" && input.Quality != "quality" && input.Quality != "standard" {
		return model.GenerationTask{}, fmt.Errorf("%w: unsupported quality", ErrInvalidGeneration)
	}
	if profile.Capability == model.GenerationKindImage && service.imageProviders[profile.Provider] == nil {
		return model.GenerationTask{}, fmt.Errorf("%w: %s", ErrGenerationUnavailable, profile.Provider)
	}
	if profile.Capability == model.GenerationKindVideo && service.videoProviders[profile.Provider] == nil {
		return model.GenerationTask{}, fmt.Errorf("%w: %s", ErrGenerationUnavailable, profile.Provider)
	}
	parameters, err := json.Marshal(input)
	if err != nil {
		return model.GenerationTask{}, err
	}
	now := service.now()
	task := model.GenerationTask{
		ID: service.id(), ProjectID: projectID, Kind: profile.Capability,
		Provider: profile.Provider, ModelID: profile.ID, UpstreamModel: profile.UpstreamModel,
		Status: model.GenerationTaskQueued, Prompt: input.Prompt, ParametersJSON: parameters,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := service.repository.CreateTask(ctx, task); err != nil {
		return model.GenerationTask{}, err
	}
	go service.execute(context.Background(), task, input)
	return task, nil
}

func (service *GenerationService) Get(ctx context.Context, taskID string) (model.GenerationTask, error) {
	task, err := service.repository.GetTask(ctx, taskID)
	if errors.Is(err, repository.ErrGenerationTaskNotFound) {
		return model.GenerationTask{}, ErrGenerationTaskNotFound
	}
	if err != nil {
		return model.GenerationTask{}, err
	}
	if task.Kind == model.GenerationKindVideo && task.Status == model.GenerationTaskRunning && task.ProviderTaskID != "" {
		provider := service.videoProviders[task.Provider]
		if provider == nil {
			return task, nil
		}
		result, pollErr := provider.GetVideoResult(ctx, task.ProviderTaskID)
		if pollErr != nil {
			return task, nil
		}
		if result.Status == repository.GenerationSucceeded {
			task.Status = model.GenerationTaskSucceeded
			task.ResultURL = result.URL
			task.CoverURL = result.CoverURL
			task.UpdatedAt = service.now()
			_, err = service.repository.UpdateTask(ctx, task, model.GenerationTaskRunning)
		} else if result.Status == repository.GenerationFailed {
			task.Status = model.GenerationTaskFailed
			task.ErrorMessage = result.Error
			task.UpdatedAt = service.now()
			_, err = service.repository.UpdateTask(ctx, task, model.GenerationTaskRunning)
		}
	}
	return task, err
}

func (service *GenerationService) execute(ctx context.Context, task model.GenerationTask, input CreateGenerationInput) {
	task.Status = model.GenerationTaskRunning
	task.UpdatedAt = service.now()
	updated, err := service.repository.UpdateTask(ctx, task, model.GenerationTaskQueued)
	if err != nil || !updated {
		return
	}
	switch task.Kind {
	case model.GenerationKindImage:
		provider := service.imageProviders[task.Provider]
		if provider == nil {
			service.fail(task, fmt.Errorf("image provider %q is unavailable", task.Provider))
			return
		}
		result, err := provider.GenerateImage(ctx, repository.ImageGenerationRequest{
			Model: task.UpstreamModel, Prompt: input.Prompt, Size: input.Size,
		})
		if err != nil {
			service.fail(task, err)
			return
		}
		task.Status = model.GenerationTaskSucceeded
		task.ResultURL = result.URL
		task.UpdatedAt = service.now()
		_, _ = service.repository.UpdateTask(ctx, task, model.GenerationTaskRunning)
	case model.GenerationKindVideo:
		provider := service.videoProviders[task.Provider]
		if provider == nil {
			service.fail(task, fmt.Errorf("video provider %q is unavailable", task.Provider))
			return
		}
		result, err := provider.SubmitVideo(ctx, repository.VideoGenerationRequest{
			RequestID: task.ID, Model: task.UpstreamModel, Prompt: input.Prompt, ImageURL: input.ImageURL,
			Size: input.Size, FPS: input.FPS, Quality: input.Quality, WithAudio: input.WithAudio,
		})
		if err != nil {
			service.fail(task, err)
			return
		}
		task.ProviderTaskID = result.ProviderTaskID
		task.UpdatedAt = service.now()
		_, _ = service.repository.UpdateTask(ctx, task, model.GenerationTaskRunning)
	default:
		service.fail(task, fmt.Errorf("unsupported capability %q", task.Kind))
	}
}

func (service *GenerationService) fail(task model.GenerationTask, taskErr error) {
	task.Status = model.GenerationTaskFailed
	task.ErrorMessage = taskErr.Error()
	task.UpdatedAt = service.now()
	_, _ = service.repository.UpdateTask(context.Background(), task, model.GenerationTaskRunning)
}

func applyGenerationDefaults(input *CreateGenerationInput, profile model.ModelProfile) {
	if input.Size == "" {
		input.Size, _ = profile.Defaults["size"].(string)
	}
	if input.FPS == 0 {
		input.FPS, _ = profile.Defaults["fps"].(int)
	}
	if input.Quality == "" {
		input.Quality, _ = profile.Defaults["quality"].(string)
	}
}
