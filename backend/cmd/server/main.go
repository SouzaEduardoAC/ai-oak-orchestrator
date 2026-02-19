package main

import (
	"log"

	"github.com/ecoza/ai-oak-orchestrator/internal/api/websocket"
	"github.com/ecoza/ai-oak-orchestrator/internal/config"
	"github.com/ecoza/ai-oak-orchestrator/internal/logger"
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

	// 3. Initialize Echo
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// 4. WebSocket Hub
	hub := websocket.NewHub(l)
	go hub.Run()

	e.GET("/ws", hub.HandleWebSocket)

	// 5. Health Check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// 5. Start Server
	if err := e.Start(":" + cfg.Server.Port); err != nil {
		l.Fatal("Server failed to start", zap.Error(err))
	}
}
