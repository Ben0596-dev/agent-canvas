package controller

import (
	"net/http"

	"agent-canvas/backend/internal/service"
	"github.com/gin-gonic/gin"
)

func NewRouter(
	projectService *service.ProjectService,
	agentService *service.AgentService,
	generationService *service.GenerationService,
	models *service.ModelRegistry,
) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	api := router.Group("/api/v1")
	api.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	api.GET("/readyz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ready"}) })
	NewProjectController(projectService).Register(api)
	if agentService != nil {
		NewAgentController(agentService).Register(api)
	}
	if generationService != nil && models != nil {
		NewGenerationController(generationService, models).Register(api)
	}
	return router
}
