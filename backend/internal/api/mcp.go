package api

import (
	"net/http"

	"github.com/ecoza/ai-oak-orchestrator/internal/domain"
	"github.com/ecoza/ai-oak-orchestrator/internal/mcp"
	"github.com/labstack/echo/v4"
)

type MCPHandler struct {
	registry *mcp.Registry
}

func NewMCPHandler(r *mcp.Registry) *MCPHandler {
	return &MCPHandler{registry: r}
}

func (h *MCPHandler) RegisterRoutes(e *echo.Echo) {
	g := e.Group("/api/mcp")
	g.GET("/tools", h.ListTools)
	g.POST("/tools", h.AddTool)
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
