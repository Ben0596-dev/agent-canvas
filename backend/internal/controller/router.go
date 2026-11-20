package controller

import (
	"net/http"

	"agent-canvas/backend/internal/service"
	"github.com/gin-gonic/gin"
)

func NewRouter(projectService *service.ProjectService) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	api := router.Group("/api/v1")
	api.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	api.GET("/readyz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ready"}) })
	NewProjectController(projectService).Register(api)
	return router
}
