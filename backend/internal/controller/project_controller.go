package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"agent-canvas/backend/internal/model"
	"agent-canvas/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type ProjectController struct {
	service *service.ProjectService
}

type createProjectRequest struct {
	Title string `json:"title" binding:"required"`
}

type updateProjectRequest struct {
	Title    *string         `json:"title"`
	Viewport *model.Viewport `json:"viewport"`
}

type projectResponse struct {
	ID        string               `json:"id"`
	Title     string               `json:"title"`
	Revision  int64                `json:"revision"`
	Snapshot  model.CanvasSnapshot `json:"snapshot"`
	CreatedAt time.Time            `json:"createdAt"`
	UpdatedAt time.Time            `json:"updatedAt"`
}

type applyOperationsRequest struct {
	RequestID        string                  `json:"requestId" binding:"required,min=1,max=64"`
	ExpectedRevision int64                   `json:"expectedRevision" binding:"required,min=1"`
	Operations       []model.CanvasOperation `json:"operations" binding:"required,min=1,max=100"`
}

type applyOperationsResponse struct {
	Revision int64           `json:"revision"`
	Project  projectResponse `json:"project"`
}

type projectSummaryResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Revision  int64     `json:"revision"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type listProjectsResponse struct {
	Items []projectSummaryResponse `json:"items"`
}

func NewProjectController(service *service.ProjectService) *ProjectController {
	return &ProjectController{service: service}
}

func (controller *ProjectController) Register(router *gin.RouterGroup) {
	router.GET("/projects", controller.list)
	router.POST("/projects", controller.create)
	router.GET("/projects/:projectId", controller.get)
	router.PATCH("/projects/:projectId", controller.rename)
	router.DELETE("/projects/:projectId", controller.delete)
	router.POST("/projects/:projectId/operations", controller.applyOperations)
}

func (controller *ProjectController) list(c *gin.Context) {
	projects, err := controller.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "internal_error", "message": "could not list projects"})
		return
	}
	items := make([]projectSummaryResponse, 0, len(projects))
	for _, project := range projects {
		items = append(items, projectSummaryResponse{
			ID: project.ID, Title: project.Title, Revision: project.Revision,
			CreatedAt: project.CreatedAt, UpdatedAt: project.UpdatedAt,
		})
	}
	c.JSON(http.StatusOK, listProjectsResponse{Items: items})
}

func (controller *ProjectController) create(c *gin.Context) {
	var request createProjectRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "invalid_request", "message": "title is required"})
		return
	}
	created, err := controller.service.Create(c.Request.Context(), request.Title)
	if err != nil {
		if errors.Is(err, service.ErrInvalidProjectTitle) {
			c.JSON(http.StatusBadRequest, gin.H{"code": "invalid_title", "message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": "internal_error", "message": "could not create project"})
		return
	}
	response, err := newProjectResponse(created)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "internal_error", "message": "could not encode project"})
		return
	}
	c.JSON(http.StatusCreated, response)
}

func (controller *ProjectController) get(c *gin.Context) {
	found, err := controller.service.Get(c.Request.Context(), c.Param("projectId"))
	if err != nil {
		if errors.Is(err, service.ErrProjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"code": "project_not_found", "message": "project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": "internal_error", "message": "could not load project"})
		return
	}
	response, err := newProjectResponse(found)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "internal_error", "message": "could not encode project"})
		return
	}
	c.JSON(http.StatusOK, response)
}

func (controller *ProjectController) rename(c *gin.Context) {
	var request updateProjectRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "invalid_request", "message": "title is required"})
		return
	}
	if request.Title == nil && request.Viewport == nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "invalid_request", "message": "title or viewport is required"})
		return
	}
	var updated model.Project
	var err error
	if request.Title != nil {
		updated, err = controller.service.Rename(c.Request.Context(), c.Param("projectId"), *request.Title)
	} else {
		updated, err = controller.service.UpdateViewport(c.Request.Context(), c.Param("projectId"), *request.Viewport)
	}
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidProjectTitle):
			c.JSON(http.StatusBadRequest, gin.H{"code": "invalid_title", "message": err.Error()})
		case errors.Is(err, service.ErrInvalidOperation):
			c.JSON(http.StatusBadRequest, gin.H{"code": "invalid_viewport", "message": err.Error()})
		case errors.Is(err, service.ErrProjectNotFound):
			c.JSON(http.StatusNotFound, gin.H{"code": "project_not_found", "message": "project not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "internal_error", "message": "could not update project"})
		}
		return
	}
	response, err := newProjectResponse(updated)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "internal_error", "message": "could not encode project"})
		return
	}
	c.JSON(http.StatusOK, response)
}

func (controller *ProjectController) delete(c *gin.Context) {
	if err := controller.service.Delete(c.Request.Context(), c.Param("projectId")); err != nil {
		if errors.Is(err, service.ErrProjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"code": "project_not_found", "message": "project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": "internal_error", "message": "could not delete project"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (controller *ProjectController) applyOperations(c *gin.Context) {
	var request applyOperationsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "invalid_request", "message": "operation batch is invalid"})
		return
	}
	updated, err := controller.service.ApplyOperations(c.Request.Context(), c.Param("projectId"), request.RequestID, request.ExpectedRevision, request.Operations)
	if err != nil {
		var conflict *service.RevisionConflictError
		switch {
		case errors.As(err, &conflict):
			c.JSON(http.StatusConflict, gin.H{
				"code":            "revision_conflict",
				"message":         "project revision does not match expectedRevision",
				"currentRevision": conflict.CurrentRevision,
			})
		case errors.Is(err, service.ErrInvalidOperation):
			c.JSON(http.StatusBadRequest, gin.H{"code": "invalid_operation", "message": err.Error()})
		case errors.Is(err, service.ErrProjectNotFound):
			c.JSON(http.StatusNotFound, gin.H{"code": "project_not_found", "message": "project not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "internal_error", "message": "could not apply operations"})
		}
		return
	}
	project, err := newProjectResponse(updated)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "internal_error", "message": "could not encode project"})
		return
	}
	c.JSON(http.StatusOK, applyOperationsResponse{Revision: updated.Revision, Project: project})
}

func newProjectResponse(project model.Project) (projectResponse, error) {
	var snapshot model.CanvasSnapshot
	if err := json.Unmarshal(project.SnapshotJSON, &snapshot); err != nil {
		return projectResponse{}, err
	}
	return projectResponse{
		ID:        project.ID,
		Title:     project.Title,
		Revision:  project.Revision,
		Snapshot:  snapshot,
		CreatedAt: project.CreatedAt,
		UpdatedAt: project.UpdatedAt,
	}, nil
}
