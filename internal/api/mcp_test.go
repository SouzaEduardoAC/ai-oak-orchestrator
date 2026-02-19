package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ecoza/ai-oak-orchestrator/internal/domain"
	"github.com/ecoza/ai-oak-orchestrator/internal/mcp"
	"github.com/labstack/echo/v4"
)

type mockRedis struct {
	data map[string]string
}

func (m *mockRedis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	switch v := value.(type) {
	case []byte:
		m.data[key] = string(v)
	case string:
		m.data[key] = v
	default:
		b, _ := json.Marshal(value)
		m.data[key] = string(b)
	}
	return nil
}

func (m *mockRedis) Get(ctx context.Context, key string) (string, error) {
	val, ok := m.data[key]
	if !ok {
		return "", nil
	}
	return val, nil
}

func (m *mockRedis) Keys(ctx context.Context, pattern string) ([]string, error) {
	var keys []string
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys, nil
}

func (m *mockRedis) Del(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

func TestAddTool(t *testing.T) {
	e := echo.New()
	mr := &mockRedis{data: make(map[string]string)}
	reg := mcp.NewRegistry(mr)
	h := NewMCPHandler(reg, nil)

	tool := domain.ToolConfig{
		Name:  "weather",
		Image: "mcp/weather",
	}
	body, _ := json.Marshal(tool)
	req := httptest.NewRequest(http.MethodPost, "/api/mcp/tools", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.AddTool(c); err != nil {
		t.Fatalf("AddTool failed: %v", err)
	}

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", rec.Code)
	}

	var resp domain.ToolConfig
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Name != tool.Name {
		t.Errorf("Expected name %s, got %s", tool.Name, resp.Name)
	}
}

func TestListTools(t *testing.T) {
	e := echo.New()
	mr := &mockRedis{data: make(map[string]string)}
	reg := mcp.NewRegistry(mr)
	h := NewMCPHandler(reg, nil)

	// Pre-seed a tool
	reg.SaveConfig(context.Background(), domain.ToolConfig{Name: "test"})

	req := httptest.NewRequest(http.MethodGet, "/api/mcp/tools", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.ListTools(c); err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var tools []domain.ToolConfig
	json.Unmarshal(rec.Body.Bytes(), &tools)
	if len(tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(tools))
	}
}