package gorm

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"agent-canvas/backend/internal/model"
	"agent-canvas/backend/internal/repository"
	"gorm.io/gorm"
)

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(ctx context.Context, project model.Project) error {
	return r.db.WithContext(ctx).Create(&project).Error
}

func (r *ProjectRepository) List(ctx context.Context) ([]model.Project, error) {
	var projects []model.Project
	err := r.db.WithContext(ctx).Where("user_id = ?", "default").Order("updated_at DESC").Find(&projects).Error
	return projects, err
}

func (r *ProjectRepository) Get(ctx context.Context, id string) (model.Project, error) {
	var project model.Project
	if err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, "default").First(&project).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Project{}, repository.ErrProjectNotFound
		}
		return model.Project{}, err
	}
	return project, nil
}

func (r *ProjectRepository) UpdateTitle(ctx context.Context, id string, title string, updatedAt time.Time) (model.Project, error) {
	result := r.db.WithContext(ctx).Model(&model.Project{}).
		Where("id = ? AND user_id = ?", id, "default").
		Updates(map[string]any{"title": title, "updated_at": updatedAt})
	if result.Error != nil {
		return model.Project{}, result.Error
	}
	if result.RowsAffected == 0 {
		return model.Project{}, repository.ErrProjectNotFound
	}
	return r.Get(ctx, id)
}

func (r *ProjectRepository) UpdateViewport(ctx context.Context, id string, viewport model.Viewport, updatedAt time.Time) (model.Project, error) {
	encoded, err := json.Marshal(viewport)
	if err != nil {
		return model.Project{}, err
	}
	result := r.db.WithContext(ctx).Model(&model.Project{}).
		Where("id = ? AND user_id = ?", id, "default").
		Updates(map[string]any{
			"snapshot_json": gorm.Expr("jsonb_set(snapshot_json, '{viewport}', ?::jsonb, true)", string(encoded)),
			"updated_at":    updatedAt,
		})
	if result.Error != nil {
		return model.Project{}, result.Error
	}
	if result.RowsAffected == 0 {
		return model.Project{}, repository.ErrProjectNotFound
	}
	return r.Get(ctx, id)
}

func (r *ProjectRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, "default").Delete(&model.Project{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repository.ErrProjectNotFound
	}
	return nil
}

func (r *ProjectRepository) GetByOperationRequest(ctx context.Context, id string, requestID string) (model.Project, bool, error) {
	var request model.OperationRequest
	err := r.db.WithContext(ctx).Where("project_id = ? AND request_id = ?", id, requestID).First(&request).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Project{}, false, nil
	}
	if err != nil {
		return model.Project{}, false, err
	}
	project, err := r.Get(ctx, id)
	return project, true, err
}

func (r *ProjectRepository) UpdateSnapshot(ctx context.Context, id string, requestID string, expectedRevision int64, snapshot []byte, updatedAt time.Time) (model.Project, error) {
	var updated model.Project
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var request model.OperationRequest
		err := tx.Where("project_id = ? AND request_id = ?", id, requestID).First(&request).Error
		if err == nil {
			return tx.Where("id = ? AND user_id = ?", id, "default").First(&updated).Error
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		result := tx.Model(&model.Project{}).
			Where("id = ? AND user_id = ? AND revision = ?", id, "default", expectedRevision).
			Updates(map[string]any{
				"snapshot_json": snapshot,
				"revision":      expectedRevision + 1,
				"updated_at":    updatedAt,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			if err := tx.Where("id = ? AND user_id = ?", id, "default").First(&updated).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return repository.ErrProjectNotFound
				}
				return err
			}
			return repository.ErrRevisionConflict
		}
		if err := tx.Create(&model.OperationRequest{
			ProjectID:      id,
			RequestID:      requestID,
			ResultRevision: expectedRevision + 1,
			CreatedAt:      updatedAt,
		}).Error; err != nil {
			return err
		}
		return tx.Where("id = ? AND user_id = ?", id, "default").First(&updated).Error
	})
	if err != nil {
		repeated, found, lookupErr := r.GetByOperationRequest(ctx, id, requestID)
		if lookupErr == nil && found {
			return repeated, nil
		}
	}
	return updated, err
}
