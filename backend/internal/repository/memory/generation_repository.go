package memory

import (
	"context"
	"sync"

	"agent-canvas/backend/internal/model"
	"agent-canvas/backend/internal/repository"
)

type GenerationRepository struct {
	mu    sync.RWMutex
	tasks map[string]model.GenerationTask
}

func NewGenerationRepository() *GenerationRepository {
	return &GenerationRepository{tasks: make(map[string]model.GenerationTask)}
}

func (r *GenerationRepository) CreateTask(_ context.Context, task model.GenerationTask) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks[task.ID] = task
	return nil
}

func (r *GenerationRepository) GetTask(_ context.Context, taskID string) (model.GenerationTask, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	task, ok := r.tasks[taskID]
	if !ok {
		return model.GenerationTask{}, repository.ErrGenerationTaskNotFound
	}
	return task, nil
}

func (r *GenerationRepository) UpdateTask(_ context.Context, task model.GenerationTask, expectedStatus string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	current, ok := r.tasks[task.ID]
	if !ok {
		return false, repository.ErrGenerationTaskNotFound
	}
	if current.Status != expectedStatus {
		return false, nil
	}
	r.tasks[task.ID] = task
	return true, nil
}
