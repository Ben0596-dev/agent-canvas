package repository

import (
	"context"
	"errors"

	"agent-canvas/backend/internal/model"
)

var (
	ErrAgentSessionNotFound = errors.New("agent session not found")
	ErrAgentRunNotFound     = errors.New("agent run not found")
)

type AgentRequest struct {
	Project         model.Project
	Message         string
	SelectedNodeIDs []string
}

type AgentResponse struct {
	Explanation string
	Operations  []model.CanvasOperation
}

type AgentProvider interface {
	Generate(ctx context.Context, request AgentRequest) (AgentResponse, error)
}

type AgentRepository interface {
	CreateSession(ctx context.Context, session model.AgentSession) error
	GetSession(ctx context.Context, projectID string, sessionID string) (model.AgentSession, error)
	ListSessions(ctx context.Context, projectID string) ([]model.AgentSession, error)
	CreateRun(ctx context.Context, run model.AgentRun) error
	GetRun(ctx context.Context, runID string) (model.AgentRun, error)
	ListRuns(ctx context.Context, sessionID string) ([]model.AgentRun, error)
	UpdateRun(ctx context.Context, run model.AgentRun, expectedStatus string) (bool, error)
}
