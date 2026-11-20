package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"agent-canvas/backend/internal/config"
	"agent-canvas/backend/internal/controller"
	"agent-canvas/backend/internal/platform/db"
	"agent-canvas/backend/internal/repository"
	gormrepo "agent-canvas/backend/internal/repository/gorm"
	"agent-canvas/backend/internal/repository/memory"
	"agent-canvas/backend/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	var projectRepository repository.ProjectRepository
	if cfg.DatabaseURL != "" {
		gormDB, openErr := db.Open(context.Background(), cfg.DatabaseURL)
		if openErr != nil {
			log.Fatal(openErr)
		}
		projectRepository = gormrepo.NewProjectRepository(gormDB)
	} else {
		projectRepository = memory.NewProjectRepository()
	}
	projectService := service.NewProjectService(projectRepository, time.Now, service.NewProjectID)
	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           controller.NewRouter(projectService),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Printf("agent-canvas server listening on %s", cfg.HTTPAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
