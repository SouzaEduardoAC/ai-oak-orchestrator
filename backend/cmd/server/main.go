package main

import (
	"context"
	"log"

	"github.com/ecoza/ai-oak-orchestrator/internal/agent"
	"github.com/ecoza/ai-oak-orchestrator/internal/api"
	internalMiddleware "github.com/ecoza/ai-oak-orchestrator/internal/api/middleware"
	"github.com/ecoza/ai-oak-orchestrator/internal/api/websocket"
	"github.com/ecoza/ai-oak-orchestrator/internal/config"
	"github.com/ecoza/ai-oak-orchestrator/internal/infrastructure/redis"
	"github.com/ecoza/ai-oak-orchestrator/internal/logger"
	"github.com/ecoza/ai-oak-orchestrator/internal/mcp"
	"github.com/ecoza/ai-oak-orchestrator/internal/mcp/llm"
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

	// 3. Initialize Infrastructure
	rdb, err := redis.NewClient(cfg.Redis.URL)
	if err != nil {
		l.Fatal("Failed to connect to Redis", zap.Error(err))
	}

	registry := mcp.NewRegistry(rdb)

	// 4. Initialize Agent & LLM
	provider, err := llm.NewProvider(context.Background(), cfg.LLM)
	if err != nil {
		l.Fatal("Failed to initialize LLM provider", zap.String("provider", cfg.LLM.Provider), zap.Error(err))
	}
	orch := agent.NewOrchestrator(provider, l)

	// 5. Initialize Echo
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// 6. Security Middleware
	auth := internalMiddleware.Auth(cfg.Keycloak.JWKSURL)

	// 7. Register Handlers
	mcpHandler := api.NewMCPHandler(registry)
	// Protected API group
	apiGroup := e.Group("/api")
	apiGroup.Use(auth)
	// TODO: refactor mcpHandler to use apiGroup
	mcpHandler.RegisterRoutes(e)

	llmHandler := api.NewLLMHandler(provider)
	llmHandler.RegisterRoutes(e)

	hub := websocket.NewHub(l, orch)
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