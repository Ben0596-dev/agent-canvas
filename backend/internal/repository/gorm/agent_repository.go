package gorm

import (
	"context"
	"errors"

	"agent-canvas/backend/internal/model"
	"agent-canvas/backend/internal/repository"
	"gorm.io/gorm"
)

type AgentRepository struct {
	db *gorm.DB
}

func NewAgentRepository(db *gorm.DB) *AgentRepository {
	return &AgentRepository{db: db}
}

func (r *AgentRepository) CreateSession(ctx context.Context, session model.AgentSession) error {
	return r.db.WithContext(ctx).Create(&session).Error
}

func (r *AgentRepository) GetSession(ctx context.Context, projectID string, sessionID string) (model.AgentSession, error) {
	var session model.AgentSession
	err := r.db.WithContext(ctx).Where("id = ? AND project_id = ?", sessionID, projectID).First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.AgentSession{}, repository.ErrAgentSessionNotFound
	}
	return session, err
}

func (r *AgentRepository) ListSessions(ctx context.Context, projectID string) ([]model.AgentSession, error) {
	var sessions []model.AgentSession
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Order("updated_at DESC").Find(&sessions).Error
	return sessions, err
}

func (r *AgentRepository) CreateRun(ctx context.Context, run model.AgentRun) error {
	return r.db.WithContext(ctx).Omit("Proposal").Create(&run).Error
}

func (r *AgentRepository) GetRun(ctx context.Context, runID string) (model.AgentRun, error) {
	var run model.AgentRun
	err := r.db.WithContext(ctx).Where("id = ?", runID).First(&run).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.AgentRun{}, repository.ErrAgentRunNotFound
	}
	return run, err
}

func (r *AgentRepository) ListRuns(ctx context.Context, sessionID string) ([]model.AgentRun, error) {
	var runs []model.AgentRun
	err := r.db.WithContext(ctx).Where("session_id = ?", sessionID).Order("created_at ASC").Find(&runs).Error
	return runs, err
}

func (r *AgentRepository) UpdateRun(ctx context.Context, run model.AgentRun, expectedStatus string) (bool, error) {
	result := r.db.WithContext(ctx).Model(&model.AgentRun{}).
		Where("id = ? AND status = ?", run.ID, expectedStatus).
		Updates(map[string]any{
			"status":          run.Status,
			"explanation":     run.Explanation,
			"operations_json": run.OperationsJSON,
			"error_message":   run.ErrorMessage,
			"updated_at":      run.UpdatedAt,
		})
	return result.RowsAffected == 1, result.Error
}
