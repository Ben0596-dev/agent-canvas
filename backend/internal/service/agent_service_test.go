package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"agent-canvas/backend/internal/model"
	"agent-canvas/backend/internal/repository"
	"agent-canvas/backend/internal/repository/memory"
)

type proposalProvider struct{}

func (proposalProvider) Generate(_ context.Context, request repository.AgentRequest) (repository.AgentResponse, error) {
	return repository.AgentResponse{
		Explanation: "已根据目标创建第一步。",
		Operations: []model.CanvasOperation{{
			Type: model.OperationNodeCreate,
			Node: &model.CanvasNode{
				ID:       "agent-node-1",
				Type:     model.NodeTypeText,
				Title:    "目标",
				Position: model.CanvasPoint{X: 120, Y: 120},
				Size:     model.CanvasSize{Width: 240, Height: 140},
				Data:     model.CanvasNodeData{Text: request.Message},
			},
		}},
	}, nil
}

type blockingProvider struct {
	started chan struct{}
}

func (provider blockingProvider) Generate(ctx context.Context, _ repository.AgentRequest) (repository.AgentResponse, error) {
	close(provider.started)
	<-ctx.Done()
	return repository.AgentResponse{}, ctx.Err()
}

func TestAgentRunProducesProposalWithoutMutatingProject(t *testing.T) {
	projectRepository := memory.NewProjectRepository()
	projectService := NewProjectService(projectRepository, time.Now, func() string { return "project-1" })
	project, err := projectService.Create(context.Background(), "Launch")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	sequence := 0
	agentService := NewAgentService(
		memory.NewAgentRepository(),
		projectService,
		proposalProvider{},
		time.Now,
		func() string {
			sequence++
			return fmt.Sprintf("id-%d", sequence)
		},
	)

	run, err := agentService.CreateRun(context.Background(), project.ID, "", "制定发布目标", nil)
	if err != nil {
		t.Fatalf("create run: %v", err)
	}
	if run.Status != model.AgentRunQueued {
		t.Fatalf("expected queued run, got %s", run.Status)
	}

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		run, err = agentService.GetRun(context.Background(), run.ID)
		if err != nil {
			t.Fatalf("get run: %v", err)
		}
		if run.Status == model.AgentRunSucceeded {
			break
		}
		time.Sleep(time.Millisecond)
	}
	if run.Status != model.AgentRunSucceeded || len(run.Proposal.Operations) != 1 {
		t.Fatalf("unexpected completed run: %+v", run)
	}
	unchanged, err := projectService.Get(context.Background(), project.ID)
	if err != nil {
		t.Fatalf("get project: %v", err)
	}
	if unchanged.Revision != 1 {
		t.Fatalf("Agent proposal changed project revision to %d", unchanged.Revision)
	}
}

func TestCancelAgentRunKeepsCancelledTerminalState(t *testing.T) {
	projectRepository := memory.NewProjectRepository()
	projectService := NewProjectService(projectRepository, time.Now, func() string { return "project-1" })
	project, err := projectService.Create(context.Background(), "Launch")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	sequence := 0
	started := make(chan struct{})
	agentService := NewAgentService(
		memory.NewAgentRepository(), projectService, blockingProvider{started: started}, time.Now,
		func() string { sequence++; return fmt.Sprintf("id-%d", sequence) },
	)
	run, err := agentService.CreateRun(context.Background(), project.ID, "", "制定发布目标", nil)
	if err != nil {
		t.Fatalf("create run: %v", err)
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("provider did not start")
	}
	cancelled, err := agentService.CancelRun(context.Background(), run.ID)
	if err != nil {
		t.Fatalf("cancel run: %v", err)
	}
	if cancelled.Status != model.AgentRunCancelled {
		t.Fatalf("expected cancelled, got %s", cancelled.Status)
	}
	time.Sleep(5 * time.Millisecond)
	stored, err := agentService.GetRun(context.Background(), run.ID)
	if err != nil {
		t.Fatalf("get run: %v", err)
	}
	if stored.Status != model.AgentRunCancelled {
		t.Fatalf("provider completion overwrote cancelled state with %s", stored.Status)
	}
}
