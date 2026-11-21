package gorm

import (
	"context"
	"errors"

	"agent-canvas/backend/internal/model"
	"agent-canvas/backend/internal/repository"
	"gorm.io/gorm"
)

type GenerationRepository struct {
	db *gorm.DB
}

func NewGenerationRepository(db *gorm.DB) *GenerationRepository {
	return &GenerationRepository{db: db}
}

func (r *GenerationRepository) CreateTask(ctx context.Context, task model.GenerationTask) error {
	return r.db.WithContext(ctx).Create(&task).Error
}

func (r *GenerationRepository) GetTask(ctx context.Context, taskID string) (model.GenerationTask, error) {
	var task model.GenerationTask
	err := r.db.WithContext(ctx).Where("id = ?", taskID).First(&task).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.GenerationTask{}, repository.ErrGenerationTaskNotFound
	}
	return task, err
}

func (r *GenerationRepository) UpdateTask(ctx context.Context, task model.GenerationTask, expectedStatus string) (bool, error) {
	result := r.db.WithContext(ctx).Model(&model.GenerationTask{}).
		Where("id = ? AND status = ?", task.ID, expectedStatus).
		Updates(map[string]any{
			"status":           task.Status,
			"provider_task_id": task.ProviderTaskID,
			"result_url":       task.ResultURL,
			"cover_url":        task.CoverURL,
			"error_message":    task.ErrorMessage,
			"updated_at":       task.UpdatedAt,
		})
	return result.RowsAffected == 1, result.Error
}
