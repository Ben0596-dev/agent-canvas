package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"agent-canvas/backend/internal/model"
	"agent-canvas/backend/internal/repository"
)

var (
	ErrInvalidAgentMessage  = errors.New("agent message must contain 1 to 10000 characters")
	ErrAgentSessionNotFound = errors.New("agent session not found")
	ErrAgentRunNotFound     = errors.New("agent run not found")
)

type AgentService struct {
	repository     repository.AgentRepository
	projectService *ProjectService
	provider       repository.AgentProvider
	now            func() time.Time
	id             func() string
	mu             sync.Mutex
	cancels        map[string]context.CancelFunc
}

func NewAgentService(
	repository repository.AgentRepository,
	projectService *ProjectService,
	provider repository.AgentProvider,
	now func() time.Time,
	id func() string,
) *AgentService {
	return &AgentService{
		repository: repository, projectService: projectService, provider: provider,
		now: now, id: id, cancels: make(map[string]context.CancelFunc),
	}
}

func (s *AgentService) CreateRun(ctx context.Context, projectID string, sessionID string, message string, selectedNodeIDs []string) (model.AgentRun, error) {
	message = strings.TrimSpace(message)
	if message == "" || utf8.RuneCountInString(message) > 10000 {
		return model.AgentRun{}, ErrInvalidAgentMessage
	}
	project, err := s.projectService.Get(ctx, projectID)
	if err != nil {
		return model.AgentRun{}, err
	}
	now := s.now()
	if sessionID == "" {
		sessionID = s.id()
		title := []rune(message)
		if len(title) > 50 {
			title = title[:50]
		}
		if err := s.repository.CreateSession(ctx, model.AgentSession{
			ID: sessionID, ProjectID: projectID, Title: string(title), CreatedAt: now, UpdatedAt: now,
		}); err != nil {
			return model.AgentRun{}, err
		}
	} else if _, err := s.repository.GetSession(ctx, projectID, sessionID); err != nil {
		if errors.Is(err, repository.ErrAgentSessionNotFound) {
			return model.AgentRun{}, ErrAgentSessionNotFound
		}
		return model.AgentRun{}, err
	}
	selectedJSON, err := json.Marshal(selectedNodeIDs)
	if err != nil {
		return model.AgentRun{}, err
	}
	run := model.AgentRun{
		ID: s.id(), ProjectID: projectID, SessionID: sessionID, Status: model.AgentRunQueued,
		Message: message, SelectedNodesJSON: selectedJSON, ExpectedRevision: project.Revision,
		OperationsJSON: []byte(`[]`), CreatedAt: now, UpdatedAt: now,
	}
	if err := s.repository.CreateRun(ctx, run); err != nil {
		return model.AgentRun{}, err
	}
	runContext, cancel := context.WithCancel(context.Background())
	s.mu.Lock()
	s.cancels[run.ID] = cancel
	s.mu.Unlock()
	go s.execute(runContext, run, project, selectedNodeIDs)
	return run, nil
}

func (s *AgentService) execute(ctx context.Context, run model.AgentRun, project model.Project, selectedNodeIDs []string) {
	defer func() {
		s.mu.Lock()
		delete(s.cancels, run.ID)
		s.mu.Unlock()
	}()
	run.Status = model.AgentRunRunning
	run.UpdatedAt = s.now()
	updated, err := s.repository.UpdateRun(ctx, run, model.AgentRunQueued)
	if err != nil || !updated {
		return
	}
	response, err := s.provider.Generate(ctx, repository.AgentRequest{
		Project: project, Message: run.Message, SelectedNodeIDs: selectedNodeIDs,
	})
	if err != nil {
		s.failRun(run, err)
		return
	}
	if err := s.projectService.PreviewOperations(ctx, project.ID, project.Revision, response.Operations); err != nil {
		s.failRun(run, err)
		return
	}
	operationsJSON, err := json.Marshal(response.Operations)
	if err != nil {
		s.failRun(run, err)
		return
	}
	run.Status = model.AgentRunSucceeded
	run.Explanation = response.Explanation
	run.OperationsJSON = operationsJSON
	run.Proposal = model.AgentProposal{
		Explanation: response.Explanation, ExpectedRevision: project.Revision, Operations: response.Operations,
	}
	run.UpdatedAt = s.now()
	_, _ = s.repository.UpdateRun(context.Background(), run, model.AgentRunRunning)
}

func (s *AgentService) failRun(run model.AgentRun, runErr error) {
	run.Status = model.AgentRunFailed
	run.ErrorMessage = runErr.Error()
	run.UpdatedAt = s.now()
	_, _ = s.repository.UpdateRun(context.Background(), run, model.AgentRunRunning)
}

func (s *AgentService) GetRun(ctx context.Context, runID string) (model.AgentRun, error) {
	run, err := s.repository.GetRun(ctx, runID)
	if errors.Is(err, repository.ErrAgentRunNotFound) {
		return model.AgentRun{}, ErrAgentRunNotFound
	}
	if err != nil {
		return model.AgentRun{}, err
	}
	if run.Status == model.AgentRunSucceeded {
		var operations []model.CanvasOperation
		if err := json.Unmarshal(run.OperationsJSON, &operations); err != nil {
			return model.AgentRun{}, err
		}
		run.Proposal = model.AgentProposal{
			Explanation: run.Explanation, ExpectedRevision: run.ExpectedRevision, Operations: operations,
		}
	}
	return run, nil
}

func (s *AgentService) ListSessions(ctx context.Context, projectID string) ([]model.AgentSession, error) {
	if _, err := s.projectService.Get(ctx, projectID); err != nil {
		return nil, err
	}
	return s.repository.ListSessions(ctx, projectID)
}

func (s *AgentService) ListRuns(ctx context.Context, projectID string, sessionID string) ([]model.AgentRun, error) {
	if _, err := s.repository.GetSession(ctx, projectID, sessionID); err != nil {
		if errors.Is(err, repository.ErrAgentSessionNotFound) {
			return nil, ErrAgentSessionNotFound
		}
		return nil, err
	}
	runs, err := s.repository.ListRuns(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	for index := range runs {
		if runs[index].Status == model.AgentRunSucceeded {
			var operations []model.CanvasOperation
			if err := json.Unmarshal(runs[index].OperationsJSON, &operations); err != nil {
				return nil, err
			}
			runs[index].Proposal = model.AgentProposal{
				Explanation:      runs[index].Explanation,
				ExpectedRevision: runs[index].ExpectedRevision,
				Operations:       operations,
			}
		}
	}
	return runs, nil
}

func (s *AgentService) CancelRun(ctx context.Context, runID string) (model.AgentRun, error) {
	run, err := s.GetRun(ctx, runID)
	if err != nil {
		return model.AgentRun{}, err
	}
	if isTerminalRun(run.Status) {
		return run, nil
	}
	previousStatus := run.Status
	run.Status = model.AgentRunCancelled
	run.UpdatedAt = s.now()
	updated, err := s.repository.UpdateRun(ctx, run, previousStatus)
	if err != nil {
		return model.AgentRun{}, err
	}
	if !updated {
		return s.GetRun(ctx, runID)
	}
	s.mu.Lock()
	cancel := s.cancels[runID]
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	return run, nil
}

func isTerminalRun(status string) bool {
	return status == model.AgentRunSucceeded || status == model.AgentRunFailed || status == model.AgentRunCancelled
}
