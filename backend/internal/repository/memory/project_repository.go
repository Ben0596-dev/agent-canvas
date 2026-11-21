package memory

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"sync"
	"time"

	"agent-canvas/backend/internal/model"
	"agent-canvas/backend/internal/repository"
)

type ProjectRepository struct {
	mu       sync.RWMutex
	projects map[string]model.Project
	requests map[string]struct{}
}

func NewProjectRepository() *ProjectRepository {
	return &ProjectRepository{
		projects: make(map[string]model.Project),
		requests: make(map[string]struct{}),
	}
}

func (r *ProjectRepository) Create(_ context.Context, project model.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.projects[project.ID] = project
	return nil
}

func (r *ProjectRepository) List(_ context.Context) ([]model.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	projects := make([]model.Project, 0, len(r.projects))
	for _, project := range r.projects {
		projects = append(projects, project)
	}
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].UpdatedAt.After(projects[j].UpdatedAt)
	})
	return projects, nil
}

func (r *ProjectRepository) Get(_ context.Context, id string) (model.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	project, ok := r.projects[id]
	if !ok {
		return model.Project{}, repository.ErrProjectNotFound
	}
	return project, nil
}

func (r *ProjectRepository) UpdateTitle(_ context.Context, id string, title string, updatedAt time.Time) (model.Project, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	project, ok := r.projects[id]
	if !ok {
		return model.Project{}, repository.ErrProjectNotFound
	}
	project.Title = title
	project.UpdatedAt = updatedAt
	r.projects[id] = project
	return project, nil
}

func (r *ProjectRepository) UpdateViewport(_ context.Context, id string, viewport model.Viewport, updatedAt time.Time) (model.Project, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	project, ok := r.projects[id]
	if !ok {
		return model.Project{}, repository.ErrProjectNotFound
	}
	var snapshot model.CanvasSnapshot
	if err := json.Unmarshal(project.SnapshotJSON, &snapshot); err != nil {
		return model.Project{}, err
	}
	snapshot.Viewport = viewport
	encoded, err := json.Marshal(snapshot)
	if err != nil {
		return model.Project{}, err
	}
	project.SnapshotJSON = encoded
	project.UpdatedAt = updatedAt
	r.projects[id] = project
	return project, nil
}

func (r *ProjectRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.projects[id]; !ok {
		return repository.ErrProjectNotFound
	}
	delete(r.projects, id)
	for key := range r.requests {
		if strings.HasPrefix(key, id+"\x00") {
			delete(r.requests, key)
		}
	}
	return nil
}

func (r *ProjectRepository) GetByOperationRequest(_ context.Context, id string, requestID string) (model.Project, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	project, ok := r.projects[id]
	if !ok {
		return model.Project{}, false, repository.ErrProjectNotFound
	}
	_, repeated := r.requests[id+"\x00"+requestID]
	return project, repeated, nil
}

func (r *ProjectRepository) UpdateSnapshot(_ context.Context, id string, requestID string, expectedRevision int64, snapshot []byte, updatedAt time.Time) (model.Project, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	project, ok := r.projects[id]
	if !ok {
		return model.Project{}, repository.ErrProjectNotFound
	}
	requestKey := id + "\x00" + requestID
	if _, repeated := r.requests[requestKey]; repeated {
		return project, nil
	}
	if project.Revision != expectedRevision {
		return project, repository.ErrRevisionConflict
	}
	project.SnapshotJSON = append([]byte(nil), snapshot...)
	project.Revision++
	project.UpdatedAt = updatedAt
	r.projects[id] = project
	r.requests[requestKey] = struct{}{}
	return project, nil
}
