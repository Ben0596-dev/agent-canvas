package controller

import (
	"errors"
	"net/http"
	"time"

	"agent-canvas/backend/internal/model"
	"agent-canvas/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type GenerationController struct {
	service *service.GenerationService
	models  *service.ModelRegistry
}

type createGenerationRequest struct {
	ModelID   string `json:"modelId" binding:"required"`
	Prompt    string `json:"prompt" binding:"required,max=5000"`
	ImageURL  string `json:"imageUrl"`
	Size      string `json:"size"`
	FPS       int    `json:"fps"`
	Quality   string `json:"quality"`
	WithAudio bool   `json:"withAudio"`
}

type generationTaskResponse struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"projectId"`
	Kind      string    `json:"kind"`
	ModelID   string    `json:"modelId"`
	Status    string    `json:"status"`
	Prompt    string    `json:"prompt"`
	ResultURL string    `json:"resultUrl,omitempty"`
	CoverURL  string    `json:"coverUrl,omitempty"`
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func NewGenerationController(generationService *service.GenerationService, models *service.ModelRegistry) *GenerationController {
	return &GenerationController{service: generationService, models: models}
}

func (controller *GenerationController) Register(router *gin.RouterGroup) {
	router.GET("/models", controller.listModels)
	router.POST("/projects/:projectId/generations", controller.create)
	router.GET("/generation-tasks/:taskId", controller.get)
}

func (controller *GenerationController) listModels(c *gin.Context) {
	profiles := controller.models.List()
	items := make([]gin.H, 0, len(profiles))
	for _, profile := range profiles {
		items = append(items, gin.H{
			"id": profile.ID, "provider": profile.Provider, "upstreamModel": profile.UpstreamModel,
			"capability": profile.Capability, "defaults": profile.Defaults,
		})
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (controller *GenerationController) create(c *gin.Context) {
	var request createGenerationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "invalid_request", "message": "generation request is invalid"})
		return
	}
	task, err := controller.service.Create(c.Request.Context(), c.Param("projectId"), service.CreateGenerationInput{
		ModelID: request.ModelID, Prompt: request.Prompt, ImageURL: request.ImageURL,
		Size: request.Size, FPS: request.FPS, Quality: request.Quality, WithAudio: request.WithAudio,
	})
	if err != nil {
		writeGenerationError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, newGenerationTaskResponse(task))
}

func (controller *GenerationController) get(c *gin.Context) {
	task, err := controller.service.Get(c.Request.Context(), c.Param("taskId"))
	if err != nil {
		writeGenerationError(c, err)
		return
	}
	c.JSON(http.StatusOK, newGenerationTaskResponse(task))
}

func newGenerationTaskResponse(task model.GenerationTask) generationTaskResponse {
	return generationTaskResponse{
		ID: task.ID, ProjectID: task.ProjectID, Kind: task.Kind, ModelID: task.ModelID,
		Status: task.Status, Prompt: task.Prompt, ResultURL: task.ResultURL,
		CoverURL: task.CoverURL, Error: task.ErrorMessage, CreatedAt: task.CreatedAt, UpdatedAt: task.UpdatedAt,
	}
}

func writeGenerationError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidGeneration):
		c.JSON(http.StatusBadRequest, gin.H{"code": "invalid_generation", "message": err.Error()})
	case errors.Is(err, service.ErrGenerationUnavailable):
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": "generation_unavailable", "message": err.Error()})
	case errors.Is(err, service.ErrProjectNotFound):
		c.JSON(http.StatusNotFound, gin.H{"code": "project_not_found", "message": "project not found"})
	case errors.Is(err, service.ErrGenerationTaskNotFound):
		c.JSON(http.StatusNotFound, gin.H{"code": "generation_task_not_found", "message": "generation task not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"code": "internal_error", "message": "generation request failed"})
	}
}
