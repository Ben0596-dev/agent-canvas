package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"agent-canvas/backend/internal/repository/memory"
	"agent-canvas/backend/internal/service"
)

func TestCreateAndGetProject(t *testing.T) {
	repository := memory.NewProjectRepository()
	projectService := service.NewProjectService(repository, func() time.Time {
		return time.Date(2026, 11, 19, 0, 0, 0, 0, time.UTC)
	}, func() string { return "project-1" })
	router := NewRouter(projectService)

	create := httptest.NewRecorder()
	router.ServeHTTP(create, httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewBufferString(`{"title":"  Sprint plan  "}`)))
	if create.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", create.Code, create.Body.String())
	}
	var created projectResponse
	if err := json.Unmarshal(create.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if created.Title != "Sprint plan" || created.Revision != 1 {
		t.Fatalf("unexpected project: %+v", created)
	}

	get := httptest.NewRecorder()
	router.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/api/v1/projects/project-1", nil))
	if get.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", get.Code, get.Body.String())
	}
}

func TestCreateProjectRejectsBlankTitle(t *testing.T) {
	projectService := service.NewProjectService(memory.NewProjectRepository(), time.Now, func() string { return "project-1" })
	router := NewRouter(projectService)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewBufferString(`{"title":"   "}`)))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", response.Code, response.Body.String())
	}
}

func TestGetProjectReturnsNotFound(t *testing.T) {
	projectService := service.NewProjectService(memory.NewProjectRepository(), time.Now, func() string { return "project-1" })
	router := NewRouter(projectService)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/projects/missing", nil))
	if response.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", response.Code, response.Body.String())
	}
}

func TestApplyOperationsCreatesNode(t *testing.T) {
	repository := memory.NewProjectRepository()
	projectService := service.NewProjectService(repository, time.Now, func() string { return "project-1" })
	if _, err := projectService.Create(context.Background(), "Plan"); err != nil {
		t.Fatalf("create project: %v", err)
	}
	router := NewRouter(projectService)

	response := httptest.NewRecorder()
	body := `{
		"requestId":"request-1",
		"expectedRevision":1,
		"operations":[{
			"type":"node.create",
			"node":{
				"id":"node-1",
				"type":"text",
				"title":"First step",
				"position":{"x":100,"y":200},
				"size":{"width":240,"height":120},
				"data":{"text":"Define the goal"}
			}
		}]
	}`
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/v1/projects/project-1/operations", bytes.NewBufferString(body)))
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
	}
	var result applyOperationsResponse
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if result.Revision != 2 || len(result.Project.Snapshot.Nodes) != 1 {
		t.Fatalf("unexpected response: %+v", result)
	}
}

func TestApplyOperationsReturnsRevisionConflict(t *testing.T) {
	repository := memory.NewProjectRepository()
	projectService := service.NewProjectService(repository, time.Now, func() string { return "project-1" })
	if _, err := projectService.Create(context.Background(), "Plan"); err != nil {
		t.Fatalf("create project: %v", err)
	}
	router := NewRouter(projectService)

	response := httptest.NewRecorder()
	body := `{
		"requestId":"request-1",
		"expectedRevision":2,
		"operations":[{
			"type":"node.create",
			"node":{
				"id":"node-1",
				"type":"text",
				"title":"First step",
				"position":{"x":100,"y":200},
				"size":{"width":240,"height":120},
				"data":{"text":"Define the goal"}
			}
		}]
	}`
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/v1/projects/project-1/operations", bytes.NewBufferString(body)))
	if response.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", response.Code, response.Body.String())
	}
	var conflict struct {
		Code            string `json:"code"`
		CurrentRevision int64  `json:"currentRevision"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &conflict); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if conflict.Code != "revision_conflict" || conflict.CurrentRevision != 1 {
		t.Fatalf("unexpected conflict: %+v", conflict)
	}
}

func TestListRenameAndDeleteProject(t *testing.T) {
	repository := memory.NewProjectRepository()
	projectService := service.NewProjectService(repository, time.Now, func() string { return "project-1" })
	if _, err := projectService.Create(context.Background(), "Draft"); err != nil {
		t.Fatalf("create project: %v", err)
	}
	router := NewRouter(projectService)

	list := httptest.NewRecorder()
	router.ServeHTTP(list, httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil))
	if list.Code != http.StatusOK || !bytes.Contains(list.Body.Bytes(), []byte(`"title":"Draft"`)) {
		t.Fatalf("unexpected list response %d: %s", list.Code, list.Body.String())
	}

	rename := httptest.NewRecorder()
	router.ServeHTTP(rename, httptest.NewRequest(http.MethodPatch, "/api/v1/projects/project-1", bytes.NewBufferString(`{"title":"Launch plan"}`)))
	if rename.Code != http.StatusOK || !bytes.Contains(rename.Body.Bytes(), []byte(`"title":"Launch plan"`)) {
		t.Fatalf("unexpected rename response %d: %s", rename.Code, rename.Body.String())
	}

	deleted := httptest.NewRecorder()
	router.ServeHTTP(deleted, httptest.NewRequest(http.MethodDelete, "/api/v1/projects/project-1", nil))
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", deleted.Code, deleted.Body.String())
	}

	missing := httptest.NewRecorder()
	router.ServeHTTP(missing, httptest.NewRequest(http.MethodGet, "/api/v1/projects/project-1", nil))
	if missing.Code != http.StatusNotFound {
		t.Fatalf("expected deleted project to return 404, got %d", missing.Code)
	}
}
