package memory

import (
	"context"
	"sort"
	"sync"

	"agent-canvas/backend/internal/model"
	"agent-canvas/backend/internal/repository"
)

type AgentRepository struct {
	mu       sync.RWMutex
	sessions map[string]model.AgentSession
	runs     map[string]model.AgentRun
}

func NewAgentRepository() *AgentRepository {
	return &AgentRepository{
		sessions: make(map[string]model.AgentSession),
		runs:     make(map[string]model.AgentRun),
	}
}

func (r *AgentRepository) CreateSession(_ context.Context, session model.AgentSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[session.ID] = session
	return nil
}

func (r *AgentRepository) GetSession(_ context.Context, projectID string, sessionID string) (model.AgentSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	session, ok := r.sessions[sessionID]
	if !ok || session.ProjectID != projectID {
		return model.AgentSession{}, repository.ErrAgentSessionNotFound
	}
	return session, nil
}

func (r *AgentRepository) ListSessions(_ context.Context, projectID string) ([]model.AgentSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	sessions := make([]model.AgentSession, 0)
	for _, session := range r.sessions {
		if session.ProjectID == projectID {
			sessions = append(sessions, session)
		}
	}
	sort.Slice(sessions, func(i, j int) bool { return sessions[i].UpdatedAt.After(sessions[j].UpdatedAt) })
	return sessions, nil
}

func (r *AgentRepository) CreateRun(_ context.Context, run model.AgentRun) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.runs[run.ID] = run
	return nil
}

func (r *AgentRepository) GetRun(_ context.Context, runID string) (model.AgentRun, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	run, ok := r.runs[runID]
	if !ok {
		return model.AgentRun{}, repository.ErrAgentRunNotFound
	}
	return run, nil
}

func (r *AgentRepository) ListRuns(_ context.Context, sessionID string) ([]model.AgentRun, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	runs := make([]model.AgentRun, 0)
	for _, run := range r.runs {
		if run.SessionID == sessionID {
			runs = append(runs, run)
		}
	}
	sort.Slice(runs, func(i, j int) bool { return runs[i].CreatedAt.Before(runs[j].CreatedAt) })
	return runs, nil
}

func (r *AgentRepository) UpdateRun(_ context.Context, run model.AgentRun, expectedStatus string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	current, ok := r.runs[run.ID]
	if !ok {
		return false, repository.ErrAgentRunNotFound
	}
	if current.Status != expectedStatus {
		return false, nil
	}
	r.runs[run.ID] = run
	return true, nil
}
