package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"agent-canvas/backend/internal/model"
	"agent-canvas/backend/internal/repository"
	"agent-canvas/backend/internal/repository/memory"
)

type fakeImageProvider struct{}

func (fakeImageProvider) GenerateImage(context.Context, repository.ImageGenerationRequest) (repository.ImageGenerationResult, error) {
	return repository.ImageGenerationResult{URL: "https://cdn.example/image.png"}, nil
}

type fakeVideoProvider struct{}

func (fakeVideoProvider) SubmitVideo(_ context.Context, request repository.VideoGenerationRequest) (repository.VideoGenerationResult, error) {
	return repository.VideoGenerationResult{ProviderTaskID: "provider-" + request.RequestID, Status: repository.GenerationRunning}, nil
}

func (fakeVideoProvider) GetVideoResult(context.Context, string) (repository.VideoGenerationResult, error) {
	return repository.VideoGenerationResult{
		ProviderTaskID: "provider-task-2", Status: repository.GenerationSucceeded,
		URL: "https://cdn.example/video.mp4", CoverURL: "https://cdn.example/cover.jpg",
	}, nil
}

func TestGenerationServiceUsesModelRegistryForImageAndVideo(t *testing.T) {
	projectRepository := memory.NewProjectRepository()
	projectService := NewProjectService(projectRepository, time.Now, func() string { return "project-1" })
	project, err := projectService.Create(context.Background(), "Media")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	sequence := 0
	service := NewGenerationService(
		memory.NewGenerationRepository(), projectService,
		NewModelRegistry([]model.ModelProfile{
			{ID: "canvas-image-fast", Provider: "zhipu", UpstreamModel: "cogview-3-flash", Capability: model.GenerationKindImage},
			{ID: "canvas-video-fast", Provider: "zhipu", UpstreamModel: "cogvideox-flash", Capability: model.GenerationKindVideo},
		}),
		map[string]repository.ImageGenerationProvider{"zhipu": fakeImageProvider{}},
		map[string]repository.VideoGenerationProvider{"zhipu": fakeVideoProvider{}},
		time.Now,
		func() string { sequence++; return fmt.Sprintf("task-%d", sequence) },
	)

	image, err := service.Create(context.Background(), project.ID, CreateGenerationInput{
		ModelID: "canvas-image-fast", Prompt: "未来城市", Size: "1024x1024",
	})
	if err != nil {
		t.Fatalf("create image task: %v", err)
	}
	image = waitForGenerationTask(t, service, image.ID)
	if image.Status != model.GenerationTaskSucceeded || image.ResultURL == "" || image.UpstreamModel != "cogview-3-flash" {
		t.Fatalf("unexpected image task: %+v", image)
	}

	video, err := service.Create(context.Background(), project.ID, CreateGenerationInput{
		ModelID: "canvas-video-fast", Prompt: "猫在奔跑", Size: "1920x1080", FPS: 30,
	})
	if err != nil {
		t.Fatalf("create video task: %v", err)
	}
	video = waitForGenerationTask(t, service, video.ID)
	if video.Status != model.GenerationTaskSucceeded || video.ResultURL == "" || video.UpstreamModel != "cogvideox-flash" {
		t.Fatalf("unexpected video task: %+v", video)
	}
}

func waitForGenerationTask(t *testing.T, service *GenerationService, taskID string) model.GenerationTask {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		task, err := service.Get(context.Background(), taskID)
		if err != nil {
			t.Fatalf("get generation task: %v", err)
		}
		if task.Status == model.GenerationTaskSucceeded || task.Status == model.GenerationTaskFailed {
			return task
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("generation task did not finish")
	return model.GenerationTask{}
}
