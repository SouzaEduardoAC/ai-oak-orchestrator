package api

import (
	"fmt"
	"net/http"

	"github.com/ecoza/ai-oak-orchestrator/internal/domain"
	"github.com/ecoza/ai-oak-orchestrator/internal/mcp"
	"github.com/labstack/echo/v4"
)

type MCPHandler struct {
	registry    *mcp.Registry
	toolManager *mcp.ToolManager
}

func NewMCPHandler(r *mcp.Registry, tm *mcp.ToolManager) *MCPHandler {
	return &MCPHandler{
		registry:    r,
		toolManager: tm,
	}
}

func (h *MCPHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/tools", h.ListTools)
	g.POST("/tools", h.AddTool)
	g.DELETE("/tools/:name", h.DeleteTool)
	g.GET("/health", h.GetHealth)
}

func (h *MCPHandler) ListTools(c echo.Context) error {
	configs, err := h.registry.GetConfigs(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, configs)
}

func (h *MCPHandler) AddTool(c echo.Context) error {
	var cfg domain.ToolConfig
	if err := c.Bind(&cfg); err != nil {
		return err
	}

	if err := h.registry.SaveConfig(c.Request().Context(), cfg); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, cfg)
}

func (h *MCPHandler) DeleteTool(c echo.Context) error {
	name := c.Param("name")
	ctx := c.Request().Context()

	// 1. Stop and remove container if active
	if err := h.toolManager.StopTool(ctx, name); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to cleanup resources: %v", err))
	}

	// 2. Remove from registry
	if err := h.registry.DeleteConfig(ctx, name); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *MCPHandler) GetHealth(c echo.Context) error {
	// For now, return a simple map. In a real app, we'd query Docker/RPC health.
	tools := h.toolManager.ListTools(c.Request().Context())
	health := map[string]interface{}{
		"active_tool_count": len(tools),
		"status":            "healthy",
	}
	return c.JSON(http.StatusOK, health)
}
