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

type mockValkey struct {
	data map[string]string
}

func (m *mockValkey) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
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

func (m *mockValkey) Get(ctx context.Context, key string) (string, error) {
	val, ok := m.data[key]
	if !ok {
		return "", nil
	}
	return val, nil
}

func (m *mockValkey) Keys(ctx context.Context, pattern string) ([]string, error) {
	var keys []string
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys, nil
}

func (m *mockValkey) Del(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

func TestAddTool(t *testing.T) {
	e := echo.New()
	mv := &mockValkey{data: make(map[string]string)}
	reg := mcp.NewRegistry(mv)
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

func TestAddTool_Validation(t *testing.T) {
	e := echo.New()
	mv := &mockValkey{data: make(map[string]string)}
	reg := mcp.NewRegistry(mv)
	h := NewMCPHandler(reg, nil)

	tests := []struct {
		name         string
		toolName     string
		expectedCode int
	}{
		{"Valid name", "weather-api-1", http.StatusCreated},
		{"Invalid spaces", "weather api", http.StatusBadRequest},
		{"Invalid caps", "WeatherAPI", http.StatusBadRequest},
		{"Invalid special chars", "weather@api", http.StatusBadRequest},
		{"Invalid path traversal", "../weather", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := domain.ToolConfig{
				Name:  tt.toolName,
				Image: "mcp/weather",
			}
			body, _ := json.Marshal(tool)
			req := httptest.NewRequest(http.MethodPost, "/api/mcp/tools", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := h.AddTool(c)
			
			// Echo HTTP errors are returned directly, not wrapped in rec.Code yet
			if he, ok := err.(*echo.HTTPError); ok {
				if he.Code != tt.expectedCode {
					t.Errorf("Expected status %d, got %d", tt.expectedCode, he.Code)
				}
			} else if tt.expectedCode != http.StatusCreated {
				t.Errorf("Expected error but got success for %s", tt.toolName)
			} else if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

func TestListTools(t *testing.T) {
	e := echo.New()
	mv := &mockValkey{data: make(map[string]string)}
	reg := mcp.NewRegistry(mv)
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
