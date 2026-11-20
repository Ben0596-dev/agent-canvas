package repository

import (
	"context"
	"errors"
	"time"

	"agent-canvas/backend/internal/model"
)

var (
	ErrProjectNotFound  = errors.New("project not found")
	ErrRevisionConflict = errors.New("project revision conflict")
)

type ProjectRepository interface {
	Create(ctx context.Context, project model.Project) error
	List(ctx context.Context) ([]model.Project, error)
	Get(ctx context.Context, id string) (model.Project, error)
	UpdateTitle(ctx context.Context, id string, title string, updatedAt time.Time) (model.Project, error)
	Delete(ctx context.Context, id string) error
	GetByOperationRequest(ctx context.Context, id string, requestID string) (model.Project, bool, error)
	UpdateSnapshot(ctx context.Context, id string, requestID string, expectedRevision int64, snapshot []byte, updatedAt time.Time) (model.Project, error)
}
