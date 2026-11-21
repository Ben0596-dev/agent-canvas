package gorm_test

import (
	"context"
	"os"
	"testing"
	"time"

	"agent-canvas/backend/internal/model"
	gormrepo "agent-canvas/backend/internal/repository/gorm"
	"agent-canvas/backend/internal/platform/db"
)

func TestGenerationTaskPersistsAndTransitions(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	database, err := db.Open(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	now := time.Now().UTC()
	project := model.Project{
		ID: "generation-repository-test", UserID: "default", Title: "Test", Revision: 1,
		SnapshotJSON: []byte(`{"nodes":[],"edges":[],"viewport":{"x":0,"y":0,"zoom":1}}`),
		CreatedAt: now, UpdatedAt: now,
	}
	if err := database.Create(&project).Error; err != nil {
		t.Fatalf("create project: %v", err)
	}
	t.Cleanup(func() { _ = database.Delete(&project).Error })

	repository := gormrepo.NewGenerationRepository(database)
	task := model.GenerationTask{
		ID: "generation-task-test", ProjectID: project.ID, Kind: model.GenerationKindImage,
		Provider: "zhipu", ModelID: "canvas-image-fast", UpstreamModel: "cogview-3-flash",
		Status: model.GenerationTaskQueued, Prompt: "future city", ParametersJSON: []byte(`{}`),
		CreatedAt: now, UpdatedAt: now,
	}
	if err := repository.CreateTask(context.Background(), task); err != nil {
		t.Fatalf("create task: %v", err)
	}
	task.Status = model.GenerationTaskSucceeded
	task.ResultURL = "https://cdn.example/image.png"
	updated, err := repository.UpdateTask(context.Background(), task, model.GenerationTaskQueued)
	if err != nil || !updated {
		t.Fatalf("update task: updated=%v err=%v", updated, err)
	}
	stored, err := repository.GetTask(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if stored.Status != model.GenerationTaskSucceeded || stored.ResultURL == "" {
		t.Fatalf("unexpected task: %+v", stored)
	}
}
