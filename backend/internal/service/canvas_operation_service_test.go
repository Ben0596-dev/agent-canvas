package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"agent-canvas/backend/internal/model"
	"agent-canvas/backend/internal/repository/memory"
)

func TestApplyOperationsCreatesTextNodeAndAdvancesRevision(t *testing.T) {
	repository := memory.NewProjectRepository()
	clock := func() time.Time { return time.Date(2026, 11, 20, 8, 0, 0, 0, time.UTC) }
	projectService := NewProjectService(repository, clock, func() string { return "project-1" })
	created, err := projectService.Create(context.Background(), "Plan")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	updated, err := projectService.ApplyOperations(context.Background(), created.ID, "request-1", 1, []model.CanvasOperation{{
		Type: model.OperationNodeCreate,
		Node: &model.CanvasNode{
			ID:       "node-1",
			Type:     model.NodeTypeText,
			Title:    "First step",
			Position: model.CanvasPoint{X: 100, Y: 200},
			Size:     model.CanvasSize{Width: 240, Height: 120},
			Data:     model.CanvasNodeData{Text: "Define the goal"},
		},
	}})
	if err != nil {
		t.Fatalf("apply operations: %v", err)
	}
	if updated.Revision != 2 {
		t.Fatalf("expected revision 2, got %d", updated.Revision)
	}
	var snapshot model.CanvasSnapshot
	if err := json.Unmarshal(updated.SnapshotJSON, &snapshot); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	if len(snapshot.Nodes) != 1 || snapshot.Nodes[0].ID != "node-1" {
		t.Fatalf("unexpected nodes: %+v", snapshot.Nodes)
	}
}

func TestApplyOperationsRejectsStaleRevisionWithoutChangingProject(t *testing.T) {
	repository := memory.NewProjectRepository()
	projectService := NewProjectService(repository, time.Now, func() string { return "project-1" })
	created, err := projectService.Create(context.Background(), "Plan")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	_, err = projectService.ApplyOperations(context.Background(), created.ID, "request-1", 2, []model.CanvasOperation{{
		Type: model.OperationNodeCreate,
		Node: &model.CanvasNode{
			ID:       "node-1",
			Type:     model.NodeTypeText,
			Position: model.CanvasPoint{},
			Size:     model.CanvasSize{Width: 240, Height: 120},
			Data:     model.CanvasNodeData{Text: "Define the goal"},
		},
	}})
	var conflict *RevisionConflictError
	if !errors.As(err, &conflict) || conflict.CurrentRevision != 1 {
		t.Fatalf("expected revision conflict at 1, got %v", err)
	}

	unchanged, getErr := projectService.Get(context.Background(), created.ID)
	if getErr != nil {
		t.Fatalf("get project: %v", getErr)
	}
	if unchanged.Revision != 1 {
		t.Fatalf("expected unchanged revision 1, got %d", unchanged.Revision)
	}
}

func TestApplyOperationsReturnsCurrentProjectForRepeatedRequest(t *testing.T) {
	repository := memory.NewProjectRepository()
	projectService := NewProjectService(repository, time.Now, func() string { return "project-1" })
	created, err := projectService.Create(context.Background(), "Plan")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	operations := []model.CanvasOperation{{
		Type: model.OperationNodeCreate,
		Node: &model.CanvasNode{
			ID:       "node-1",
			Type:     model.NodeTypeText,
			Position: model.CanvasPoint{},
			Size:     model.CanvasSize{Width: 240, Height: 120},
			Data:     model.CanvasNodeData{Text: "Define the goal"},
		},
	}}

	first, err := projectService.ApplyOperations(context.Background(), created.ID, "request-1", 1, operations)
	if err != nil {
		t.Fatalf("first apply: %v", err)
	}
	repeated, err := projectService.ApplyOperations(context.Background(), created.ID, "request-1", 1, operations)
	if err != nil {
		t.Fatalf("repeated apply: %v", err)
	}
	if first.Revision != 2 || repeated.Revision != 2 {
		t.Fatalf("expected idempotent revision 2, got first=%d repeated=%d", first.Revision, repeated.Revision)
	}
	var snapshot model.CanvasSnapshot
	if err := json.Unmarshal(repeated.SnapshotJSON, &snapshot); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	if len(snapshot.Nodes) != 1 {
		t.Fatalf("expected one node after repeated request, got %d", len(snapshot.Nodes))
	}
}

func TestApplyOperationsRejectsEntireBatchWhenOneOperationIsInvalid(t *testing.T) {
	repository := memory.NewProjectRepository()
	projectService := NewProjectService(repository, time.Now, func() string { return "project-1" })
	created, err := projectService.Create(context.Background(), "Plan")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	validNode := &model.CanvasNode{
		ID:   "node-1",
		Type: model.NodeTypeText,
		Size: model.CanvasSize{Width: 240, Height: 120},
		Data: model.CanvasNodeData{Text: "Define the goal"},
	}

	_, err = projectService.ApplyOperations(context.Background(), created.ID, "request-1", 1, []model.CanvasOperation{
		{Type: model.OperationNodeCreate, Node: validNode},
		{Type: model.OperationNodeCreate, Node: validNode},
	})
	if !errors.Is(err, ErrInvalidOperation) {
		t.Fatalf("expected invalid operation, got %v", err)
	}
	unchanged, getErr := projectService.Get(context.Background(), created.ID)
	if getErr != nil {
		t.Fatalf("get project: %v", getErr)
	}
	if unchanged.Revision != 1 {
		t.Fatalf("expected unchanged revision 1, got %d", unchanged.Revision)
	}
}

func TestApplyOperationsSupportsEditMoveConnectAndCascadingDelete(t *testing.T) {
	repository := memory.NewProjectRepository()
	projectService := NewProjectService(repository, time.Now, func() string { return "project-1" })
	created, err := projectService.Create(context.Background(), "Plan")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	node := func(id, text string) *model.CanvasNode {
		return &model.CanvasNode{
			ID: id, Type: model.NodeTypeText,
			Size: model.CanvasSize{Width: 240, Height: 120},
			Data: model.CanvasNodeData{Text: text},
		}
	}
	project, err := projectService.ApplyOperations(context.Background(), created.ID, "request-1", 1, []model.CanvasOperation{
		{Type: model.OperationNodeCreate, Node: node("node-1", "Draft")},
		{Type: model.OperationNodeCreate, Node: node("node-2", "Review")},
	})
	if err != nil {
		t.Fatalf("create nodes: %v", err)
	}
	title := "First step"
	data := model.CanvasNodeData{Text: "Define the goal"}
	position := model.CanvasPoint{X: 320, Y: 160}
	project, err = projectService.ApplyOperations(context.Background(), created.ID, "request-2", project.Revision, []model.CanvasOperation{
		{Type: model.OperationNodeUpdate, NodeID: "node-1", Title: &title, Data: &data},
		{Type: model.OperationNodeMove, NodeID: "node-1", Position: &position},
		{Type: model.OperationEdgeCreate, Edge: &model.CanvasEdge{ID: "edge-1", Source: "node-1", Target: "node-2"}},
	})
	if err != nil {
		t.Fatalf("update canvas: %v", err)
	}
	var snapshot model.CanvasSnapshot
	if err := json.Unmarshal(project.SnapshotJSON, &snapshot); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	if snapshot.Nodes[0].Title != title || snapshot.Nodes[0].Position != position || len(snapshot.Edges) != 1 {
		t.Fatalf("unexpected updated snapshot: %+v", snapshot)
	}

	project, err = projectService.ApplyOperations(context.Background(), created.ID, "request-3", project.Revision, []model.CanvasOperation{{
		Type: model.OperationNodeDelete, NodeID: "node-1",
	}})
	if err != nil {
		t.Fatalf("delete node: %v", err)
	}
	if err := json.Unmarshal(project.SnapshotJSON, &snapshot); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	if len(snapshot.Nodes) != 1 || snapshot.Nodes[0].ID != "node-2" || len(snapshot.Edges) != 0 {
		t.Fatalf("expected node and related edge removed, got %+v", snapshot)
	}
}
