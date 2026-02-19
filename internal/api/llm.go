package api

import (
	"net/http"

	"github.com/ecoza/ai-oak-orchestrator/internal/domain"
	"github.com/ecoza/ai-oak-orchestrator/internal/mcp/llm"
	"github.com/labstack/echo/v4"
)

type LLMHandler struct {
	provider llm.Provider
}

func NewLLMHandler(p llm.Provider) *LLMHandler {
	return &LLMHandler{provider: p}
}

func (h *LLMHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/available", h.ListModels)
}

func (h *LLMHandler) ListModels(c echo.Context) error {
	ids, err := h.provider.ListModels(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	var models []domain.Model
	for _, id := range ids {
		models = append(models, domain.Model{ID: id, Name: id})
	}

	return c.JSON(http.StatusOK, models)
}
