package mcp

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/ecoza/ai-oak-orchestrator/internal/domain"
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

func TestRegistry_SaveAndGet(t *testing.T) {
	mr := &mockRedis{data: make(map[string]string)}
	reg := NewRegistry(mr)

	cfg := domain.ToolConfig{
		Name:   "test-tool",
		Image:  "test-image",
		Active: true,
	}

	ctx := context.Background()
	err := reg.SaveConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	configs, err := reg.GetConfigs(ctx)
	if err != nil {
		t.Fatalf("GetConfigs failed: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 config, got %d", len(configs))
	}

	if configs[0].Name != cfg.Name || configs[0].Image != cfg.Image {
		t.Errorf("Config mismatch. Got %+v, expected %+v", configs[0], cfg)
	}
}