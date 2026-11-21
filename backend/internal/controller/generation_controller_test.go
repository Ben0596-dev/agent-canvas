package controller

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"agent-canvas/backend/internal/model"
	"agent-canvas/backend/internal/repository"
	"agent-canvas/backend/internal/repository/memory"
	"agent-canvas/backend/internal/service"
)

type controllerImageProvider struct{}

func (controllerImageProvider) GenerateImage(context.Context, repository.ImageGenerationRequest) (repository.ImageGenerationResult, error) {
	return repository.ImageGenerationResult{URL: "https://cdn.example/image.png"}, nil
}

func TestCreateAndGetGenerationTask(t *testing.T) {
	projectRepository := memory.NewProjectRepository()
	projectService := service.NewProjectService(projectRepository, time.Now, func() string { return "project-1" })
	if _, err := projectService.Create(context.Background(), "Media"); err != nil {
		t.Fatalf("create project: %v", err)
	}
	sequence := 0
	models := service.NewModelRegistry([]model.ModelProfile{{
		ID: "canvas-image-fast", Provider: "zhipu", UpstreamModel: "cogview-3-flash",
		Capability: model.GenerationKindImage, Defaults: map[string]any{"size": "1024x1024"},
	}})
	generationService := service.NewGenerationService(
		memory.NewGenerationRepository(), projectService, models,
		map[string]repository.ImageGenerationProvider{"zhipu": controllerImageProvider{}}, nil,
		time.Now, func() string { sequence++; return fmt.Sprintf("task-%d", sequence) },
	)
	router := NewRouter(projectService, nil, generationService, models)

	created := httptest.NewRecorder()
	router.ServeHTTP(created, httptest.NewRequest(
		http.MethodPost, "/api/v1/projects/project-1/generations",
		bytes.NewBufferString(`{"modelId":"canvas-image-fast","prompt":"未来城市"}`),
	))
	if created.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", created.Code, created.Body.String())
	}

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		current := httptest.NewRecorder()
		router.ServeHTTP(current, httptest.NewRequest(http.MethodGet, "/api/v1/generation-tasks/task-1", nil))
		if current.Code == http.StatusOK && strings.Contains(current.Body.String(), `"status":"succeeded"`) {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("generation task did not succeed")
}
