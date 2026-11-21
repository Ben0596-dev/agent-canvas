package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"agent-canvas/backend/internal/config"
	"agent-canvas/backend/internal/controller"
	"agent-canvas/backend/internal/platform/db"
	"agent-canvas/backend/internal/repository"
	agentprovider "agent-canvas/backend/internal/repository/agent"
	gormrepo "agent-canvas/backend/internal/repository/gorm"
	"agent-canvas/backend/internal/repository/memory"
	"agent-canvas/backend/internal/repository/provider/zhipu"
	"agent-canvas/backend/internal/service"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to the YAML configuration file")
	flag.Parse()
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	var projectRepository repository.ProjectRepository
	var agentRepository repository.AgentRepository
	var generationRepository repository.GenerationRepository
	if cfg.DatabaseURL != "" {
		gormDB, openErr := db.Open(context.Background(), cfg.DatabaseURL)
		if openErr != nil {
			log.Fatal(openErr)
		}
		projectRepository = gormrepo.NewProjectRepository(gormDB)
		agentRepository = gormrepo.NewAgentRepository(gormDB)
		generationRepository = gormrepo.NewGenerationRepository(gormDB)
	} else {
		projectRepository = memory.NewProjectRepository()
		agentRepository = memory.NewAgentRepository()
		generationRepository = memory.NewGenerationRepository()
	}
	projectService := service.NewProjectService(projectRepository, time.Now, service.NewProjectID)
	models := service.NewZhipuModelRegistry(
		cfg.ZhipuAgentModel,
		cfg.ZhipuImageModel,
		cfg.ZhipuVideoModel,
	)
	var agentProvider repository.AgentProvider = agentprovider.LocalProvider{}
	if cfg.AgentProvider == "openai-compatible" {
		agentProvider = agentprovider.NewOpenAICompatibleProvider(
			cfg.AgentBaseURL,
			cfg.AgentAPIKey,
			cfg.AgentModel,
			&http.Client{Timeout: 2 * time.Minute},
		)
	}
	if cfg.AgentProvider == "zhipu" {
		agentProfile, resolveErr := models.Resolve("canvas-agent")
		if resolveErr != nil {
			log.Fatal(resolveErr)
		}
		agentProvider = agentprovider.NewOpenAICompatibleProvider(
			cfg.ZhipuBaseURL,
			cfg.ZhipuAPIKey,
			agentProfile.UpstreamModel,
			&http.Client{Timeout: 2 * time.Minute},
		)
	}
	agentService := service.NewAgentService(
		agentRepository,
		projectService,
		agentProvider,
		time.Now,
		service.NewProjectID,
	)
	imageProviders := map[string]repository.ImageGenerationProvider{}
	videoProviders := map[string]repository.VideoGenerationProvider{}
	if cfg.ZhipuAPIKey != "" {
		zhipuClient := zhipu.NewClient(
			cfg.ZhipuBaseURL,
			cfg.ZhipuAPIKey,
			&http.Client{Timeout: 2 * time.Minute},
		)
		imageProviders["zhipu"] = zhipuClient
		videoProviders["zhipu"] = zhipuClient
	}
	generationService := service.NewGenerationService(
		generationRepository,
		projectService,
		models,
		imageProviders,
		videoProviders,
		time.Now,
		service.NewProjectID,
	)
	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           controller.NewRouter(projectService, agentService, generationService, models),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      0,
		IdleTimeout:       60 * time.Second,
	}
	log.Printf("agent-canvas server listening on %s", cfg.HTTPAddr)
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()
	shutdownSignal, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	select {
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	case <-shutdownSignal.Done():
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			log.Printf("server shutdown failed: %v", err)
		}
	}
}
