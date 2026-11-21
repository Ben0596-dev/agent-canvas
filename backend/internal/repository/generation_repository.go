package repository

import (
	"context"
	"errors"

	"agent-canvas/backend/internal/model"
)

var ErrGenerationTaskNotFound = errors.New("generation task not found")

type GenerationRepository interface {
	CreateTask(ctx context.Context, task model.GenerationTask) error
	GetTask(ctx context.Context, taskID string) (model.GenerationTask, error)
	UpdateTask(ctx context.Context, task model.GenerationTask, expectedStatus string) (bool, error)
}
