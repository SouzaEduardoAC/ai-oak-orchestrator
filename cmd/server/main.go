package main

import (
	"context"
	"log"

	"github.com/ecoza/ai-oak-orchestrator/internal/agent"
	"github.com/ecoza/ai-oak-orchestrator/internal/api"
	internalMiddleware "github.com/ecoza/ai-oak-orchestrator/internal/api/middleware"
	"github.com/ecoza/ai-oak-orchestrator/internal/api/websocket"
	"github.com/ecoza/ai-oak-orchestrator/internal/config"
	"github.com/ecoza/ai-oak-orchestrator/internal/infrastructure/docker"
	"github.com/ecoza/ai-oak-orchestrator/internal/infrastructure/valkey"
	"github.com/ecoza/ai-oak-orchestrator/internal/logger"
	"github.com/ecoza/ai-oak-orchestrator/internal/mcp"
	"github.com/ecoza/ai-oak-orchestrator/internal/mcp/llm"
	"github.com/ecoza/ai-oak-orchestrator/internal/services"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Initialize Logger
	l, err := logger.New()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer l.Sync()

	l.Info("Starting AI Oak Orchestrator", zap.String("port", cfg.Server.Port))
	l.Info("Loaded Valkey configuration", zap.String("url", cfg.Valkey.URL))

	// 3. Initialize Infrastructure
	vdb, err := valkey.NewClient(cfg.Valkey.URL)
	if err != nil {
		l.Fatal("Failed to connect to Valkey", zap.Error(err))
	}

	dockerManager, err := docker.NewManager(cfg.Docker.Host)
	if err != nil {
		l.Fatal("Failed to initialize Docker manager", zap.Error(err))
	}

	registry := mcp.NewRegistry(vdb)
	toolManager := mcp.NewToolManager(dockerManager, registry, l)
	go toolManager.InitializeAll(context.Background())

	// Start Janitor
	janitor := services.NewJanitorService(dockerManager, registry, l)
	go janitor.Start(context.Background())

	// 4. Initialize Agent & LLM
	provider, err := llm.NewProvider(context.Background(), cfg.LLM)
	if err != nil {
		l.Fatal("Failed to initialize LLM provider", zap.String("provider", cfg.LLM.Provider), zap.Error(err))
	}
	orch := agent.NewOrchestrator(provider, toolManager, l)

	// 5. Initialize Echo
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// 6. Security Middleware
	auth := internalMiddleware.Auth(cfg.Keycloak.Enabled, cfg.Keycloak.JWKSURL)

	// 7. Register Handlers
	apiGroup := e.Group("/api")
	apiGroup.Use(auth)

	mcpHandler := api.NewMCPHandler(registry, toolManager)
	mcpHandler.RegisterRoutes(apiGroup.Group("/mcp"))

	llmHandler := api.NewLLMHandler(provider)
	llmHandler.RegisterRoutes(apiGroup.Group("/models"))

	hub := websocket.NewHub(l, orch, vdb)
	go hub.Run()

	e.GET("/ws", hub.HandleWebSocket, auth)

	// 8. Health Check (Public)
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// 9. Start Server
	if err := e.Start(":" + cfg.Server.Port); err != nil {
		l.Fatal("Server failed to start", zap.Error(err))
	}
}
