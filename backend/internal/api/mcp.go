package api

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/ecoza/ai-oak-orchestrator/internal/domain"
	"github.com/ecoza/ai-oak-orchestrator/internal/mcp"
	"github.com/labstack/echo/v4"
)

var validToolName = regexp.MustCompile(`^[a-z0-9-]+$`)

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
	g.POST("/add", h.AddTool)
	g.DELETE("/tools/:name", h.DeleteTool)
	g.DELETE("/:name", h.DeleteTool)
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
	var input struct {
		Name   string            `json:"name"`
		Config domain.ToolConfig `json:"config"`
	}
	
	if err := c.Bind(&input); err == nil && input.Name != "" {
		if !validToolName.MatchString(input.Name) {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid tool name: must contain only lowercase letters, numbers, and hyphens")
		}
		if input.Config.Name == "" {
			input.Config.Name = input.Name
		}
		if err := h.registry.SaveConfig(c.Request().Context(), input.Config); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusCreated, input.Config)
	}

	var cfg domain.ToolConfig
	if err := c.Bind(&cfg); err != nil {
		return err
	}

	if !validToolName.MatchString(cfg.Name) {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid tool name: must contain only lowercase letters, numbers, and hyphens")
	}

	if err := h.registry.SaveConfig(c.Request().Context(), cfg); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, cfg)
}

func (h *MCPHandler) DeleteTool(c echo.Context) error {
	name := c.Param("name")
	ctx := c.Request().Context()

	if err := h.toolManager.StopTool(ctx, name); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to cleanup resources: %v", err))
	}

	if err := h.registry.DeleteConfig(ctx, name); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *MCPHandler) GetHealth(c echo.Context) error {
	tools := h.toolManager.ListTools(c.Request().Context())
	
	type McpHealth struct {
		Name                string `json:"name"`
		Status              string `json:"status"`
		LastCheck           int64  `json:"lastCheck"`
		LastSuccess         int64  `json:"lastSuccess"`
		ConsecutiveFailures int    `json:"consecutiveFailures"`
	}

	var mcps []McpHealth
	healthyCount := 0
	
	now := time.Now().Unix() * 1000
	for name := range tools {
		status := "healthy"
		healthyCount++
		mcps = append(mcps, McpHealth{
			Name:        name,
			Status:      status,
			LastCheck:   now,
			LastSuccess: now,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"mcps": mcps,
		"summary": map[string]interface{}{
			"total":       len(tools),
			"healthy":     healthyCount,
			"unhealthy":   len(tools) - healthyCount,
			"reconnecting": 0,
		},
	})
}