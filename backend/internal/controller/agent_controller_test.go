package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	agentprovider "agent-canvas/backend/internal/repository/agent"
	"agent-canvas/backend/internal/repository/memory"
	"agent-canvas/backend/internal/service"
)

func TestCreateAndStreamAgentProposal(t *testing.T) {
	projectRepository := memory.NewProjectRepository()
	projectService := service.NewProjectService(projectRepository, time.Now, func() string { return "project-1" })
	if _, err := projectService.Create(context.Background(), "Launch"); err != nil {
		t.Fatalf("create project: %v", err)
	}
	sequence := 0
	agentService := service.NewAgentService(
		memory.NewAgentRepository(), projectService, agentprovider.LocalProvider{}, time.Now,
		func() string { sequence++; return fmt.Sprintf("agent-%d", sequence) },
	)
	router := NewRouter(projectService, agentService, nil, nil)

	created := httptest.NewRecorder()
	router.ServeHTTP(created, httptest.NewRequest(
		http.MethodPost,
		"/api/v1/projects/project-1/agent/runs",
		bytes.NewBufferString(`{"message":"制定发布目标"}`),
	))
	if created.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", created.Code, created.Body.String())
	}
	var run agentRunResponse
	if err := json.Unmarshal(created.Body.Bytes(), &run); err != nil {
		t.Fatalf("decode run: %v", err)
	}

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		current := httptest.NewRecorder()
		router.ServeHTTP(current, httptest.NewRequest(http.MethodGet, "/api/v1/agent/runs/"+run.ID, nil))
		if current.Code != http.StatusOK {
			t.Fatalf("get run: %d %s", current.Code, current.Body.String())
		}
		if err := json.Unmarshal(current.Body.Bytes(), &run); err != nil {
			t.Fatalf("decode current run: %v", err)
		}
		if run.Status == "succeeded" {
			break
		}
		time.Sleep(time.Millisecond)
	}
	if run.Proposal == nil || len(run.Proposal.Operations) != 1 {
		t.Fatalf("expected proposal, got %+v", run)
	}
	history := httptest.NewRecorder()
	router.ServeHTTP(history, httptest.NewRequest(
		http.MethodGet,
		"/api/v1/projects/project-1/agent/sessions/"+run.SessionID+"/runs",
		nil,
	))
	if history.Code != http.StatusOK || !strings.Contains(history.Body.String(), "制定发布目标") {
		t.Fatalf("unexpected session history %d: %s", history.Code, history.Body.String())
	}

	events := httptest.NewRecorder()
	router.ServeHTTP(events, httptest.NewRequest(http.MethodGet, "/api/v1/agent/runs/"+run.ID+"/events", nil))
	if events.Code != http.StatusOK || !strings.Contains(events.Body.String(), "event: operation.proposed") || !strings.Contains(events.Body.String(), "event: run.completed") {
		t.Fatalf("unexpected SSE response %d: %s", events.Code, events.Body.String())
	}
}
