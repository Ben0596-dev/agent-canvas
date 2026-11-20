package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"agent-canvas/backend/internal/model"
	"agent-canvas/backend/internal/repository"
)

var (
	ErrInvalidProjectTitle = errors.New("project title must contain 1 to 200 non-whitespace characters")
	ErrProjectNotFound     = errors.New("project not found")
	ErrInvalidOperation    = errors.New("invalid canvas operation")
)

type RevisionConflictError struct {
	CurrentRevision int64
}

func (e *RevisionConflictError) Error() string {
	return fmt.Sprintf("project revision conflict: current revision is %d", e.CurrentRevision)
}

type ProjectService struct {
	repository repository.ProjectRepository
	now        func() time.Time
	id         func() string
}

func NewProjectService(repository repository.ProjectRepository, now func() time.Time, id func() string) *ProjectService {
	return &ProjectService{repository: repository, now: now, id: id}
}

func (s *ProjectService) Create(ctx context.Context, title string) (model.Project, error) {
	title = strings.TrimSpace(title)
	if len(title) == 0 || len([]rune(title)) > 200 {
		return model.Project{}, ErrInvalidProjectTitle
	}
	now := s.now()
	snapshot, err := json.Marshal(model.CanvasSnapshot{
		Nodes:    []model.CanvasNode{},
		Edges:    []model.CanvasEdge{},
		Viewport: model.Viewport{Zoom: 1},
	})
	if err != nil {
		return model.Project{}, err
	}
	project := model.Project{
		ID:           s.id(),
		UserID:       "default",
		Title:        title,
		Revision:     1,
		SnapshotJSON: snapshot,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.repository.Create(ctx, project); err != nil {
		return model.Project{}, err
	}
	return project, nil
}

func (s *ProjectService) Get(ctx context.Context, id string) (model.Project, error) {
	project, err := s.repository.Get(ctx, id)
	if errors.Is(err, repository.ErrProjectNotFound) {
		return model.Project{}, ErrProjectNotFound
	}
	return project, err
}

func (s *ProjectService) List(ctx context.Context) ([]model.Project, error) {
	return s.repository.List(ctx)
}

func (s *ProjectService) Rename(ctx context.Context, id string, title string) (model.Project, error) {
	title = strings.TrimSpace(title)
	if len(title) == 0 || len([]rune(title)) > 200 {
		return model.Project{}, ErrInvalidProjectTitle
	}
	project, err := s.repository.UpdateTitle(ctx, id, title, s.now())
	if errors.Is(err, repository.ErrProjectNotFound) {
		return model.Project{}, ErrProjectNotFound
	}
	return project, err
}

func (s *ProjectService) Delete(ctx context.Context, id string) error {
	err := s.repository.Delete(ctx, id)
	if errors.Is(err, repository.ErrProjectNotFound) {
		return ErrProjectNotFound
	}
	return err
}

func (s *ProjectService) ApplyOperations(ctx context.Context, projectID string, requestID string, expectedRevision int64, operations []model.CanvasOperation) (model.Project, error) {
	if strings.TrimSpace(requestID) == "" || len(requestID) > 64 {
		return model.Project{}, fmt.Errorf("%w: request id must contain 1 to 64 characters", ErrInvalidOperation)
	}
	repeated, found, err := s.repository.GetByOperationRequest(ctx, projectID, requestID)
	if err != nil {
		if errors.Is(err, repository.ErrProjectNotFound) {
			return model.Project{}, ErrProjectNotFound
		}
		return model.Project{}, err
	}
	if found {
		return repeated, nil
	}
	project, err := s.Get(ctx, projectID)
	if err != nil {
		return model.Project{}, err
	}
	if project.Revision != expectedRevision {
		return model.Project{}, &RevisionConflictError{CurrentRevision: project.Revision}
	}
	if len(operations) == 0 || len(operations) > 100 {
		return model.Project{}, fmt.Errorf("%w: operation batch must contain 1 to 100 operations", ErrInvalidOperation)
	}

	var snapshot model.CanvasSnapshot
	if err := json.Unmarshal(project.SnapshotJSON, &snapshot); err != nil {
		return model.Project{}, fmt.Errorf("decode project snapshot: %w", err)
	}
	for index, operation := range operations {
		if err := applyOperation(&snapshot, operation); err != nil {
			return model.Project{}, fmt.Errorf("%w at index %d: %v", ErrInvalidOperation, index, err)
		}
	}
	encoded, err := json.Marshal(snapshot)
	if err != nil {
		return model.Project{}, fmt.Errorf("encode project snapshot: %w", err)
	}
	updated, err := s.repository.UpdateSnapshot(ctx, projectID, requestID, expectedRevision, encoded, s.now())
	if errors.Is(err, repository.ErrRevisionConflict) {
		return model.Project{}, &RevisionConflictError{CurrentRevision: updated.Revision}
	}
	if errors.Is(err, repository.ErrProjectNotFound) {
		return model.Project{}, ErrProjectNotFound
	}
	return updated, err
}

func applyOperation(snapshot *model.CanvasSnapshot, operation model.CanvasOperation) error {
	switch operation.Type {
	case model.OperationNodeCreate:
		return applyNodeCreate(snapshot, operation.Node)
	case model.OperationNodeUpdate:
		return applyNodeUpdate(snapshot, operation)
	case model.OperationNodeMove:
		return applyNodeMove(snapshot, operation.NodeID, operation.Position)
	case model.OperationNodeDelete:
		return applyNodeDelete(snapshot, operation.NodeID)
	case model.OperationEdgeCreate:
		return applyEdgeCreate(snapshot, operation.Edge)
	case model.OperationEdgeDelete:
		return applyEdgeDelete(snapshot, operation.EdgeID)
	default:
		return errors.New("unsupported operation type")
	}
}

func applyNodeCreate(snapshot *model.CanvasSnapshot, input *model.CanvasNode) error {
	if input == nil {
		return errors.New("node is required")
	}
	node := *input
	if strings.TrimSpace(node.ID) == "" || len(node.ID) > 64 {
		return errors.New("node id must contain 1 to 64 characters")
	}
	if node.Type != model.NodeTypeText {
		return errors.New("node type must be text")
	}
	if strings.TrimSpace(node.Data.Text) == "" {
		return errors.New("text node content is required")
	}
	if node.Size.Width <= 0 || node.Size.Height <= 0 {
		return errors.New("node size must be positive")
	}
	for _, existing := range snapshot.Nodes {
		if existing.ID == node.ID {
			return errors.New("node id already exists")
		}
	}
	snapshot.Nodes = append(snapshot.Nodes, node)
	return nil
}

func applyNodeUpdate(snapshot *model.CanvasSnapshot, operation model.CanvasOperation) error {
	index := nodeIndex(snapshot.Nodes, operation.NodeID)
	if index < 0 {
		return errors.New("node does not exist")
	}
	if operation.Title == nil && operation.Data == nil {
		return errors.New("title or data is required")
	}
	if operation.Title != nil {
		if len([]rune(*operation.Title)) > 200 {
			return errors.New("node title exceeds 200 characters")
		}
		snapshot.Nodes[index].Title = *operation.Title
	}
	if operation.Data != nil {
		if strings.TrimSpace(operation.Data.Text) == "" {
			return errors.New("text node content is required")
		}
		snapshot.Nodes[index].Data = *operation.Data
	}
	return nil
}

func applyNodeMove(snapshot *model.CanvasSnapshot, nodeID string, position *model.CanvasPoint) error {
	index := nodeIndex(snapshot.Nodes, nodeID)
	if index < 0 {
		return errors.New("node does not exist")
	}
	if position == nil {
		return errors.New("position is required")
	}
	snapshot.Nodes[index].Position = *position
	return nil
}

func applyNodeDelete(snapshot *model.CanvasSnapshot, nodeID string) error {
	index := nodeIndex(snapshot.Nodes, nodeID)
	if index < 0 {
		return errors.New("node does not exist")
	}
	snapshot.Nodes = append(snapshot.Nodes[:index], snapshot.Nodes[index+1:]...)
	edges := snapshot.Edges[:0]
	for _, edge := range snapshot.Edges {
		if edge.Source != nodeID && edge.Target != nodeID {
			edges = append(edges, edge)
		}
	}
	snapshot.Edges = edges
	return nil
}

func applyEdgeCreate(snapshot *model.CanvasSnapshot, input *model.CanvasEdge) error {
	if input == nil {
		return errors.New("edge is required")
	}
	edge := *input
	if strings.TrimSpace(edge.ID) == "" || len(edge.ID) > 64 {
		return errors.New("edge id must contain 1 to 64 characters")
	}
	if edge.Source == edge.Target {
		return errors.New("self connections are not allowed")
	}
	if nodeIndex(snapshot.Nodes, edge.Source) < 0 || nodeIndex(snapshot.Nodes, edge.Target) < 0 {
		return errors.New("edge endpoints must exist")
	}
	for _, existing := range snapshot.Edges {
		if existing.ID == edge.ID || (existing.Source == edge.Source && existing.Target == edge.Target) {
			return errors.New("edge already exists")
		}
	}
	snapshot.Edges = append(snapshot.Edges, edge)
	return nil
}

func applyEdgeDelete(snapshot *model.CanvasSnapshot, edgeID string) error {
	for index, edge := range snapshot.Edges {
		if edge.ID == edgeID {
			snapshot.Edges = append(snapshot.Edges[:index], snapshot.Edges[index+1:]...)
			return nil
		}
	}
	return errors.New("edge does not exist")
}

func nodeIndex(nodes []model.CanvasNode, nodeID string) int {
	for index, node := range nodes {
		if node.ID == nodeID {
			return index
		}
	}
	return -1
}
