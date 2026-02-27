package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ecoza/ai-oak-orchestrator/internal/domain"
)

type ValkeyStore interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Keys(ctx context.Context, pattern string) ([]string, error)
	Del(ctx context.Context, key string) error
}

type Registry struct {
	valkey ValkeyStore
}

func NewRegistry(v ValkeyStore) *Registry {
	return &Registry{valkey: v}
}

func (r *Registry) SaveConfig(ctx context.Context, cfg domain.ToolConfig) error {
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	return r.valkey.Set(ctx, fmt.Sprintf("mcp:config:%s", cfg.Name), data, 0)
}

func (r *Registry) DeleteConfig(ctx context.Context, name string) error {
	return r.valkey.Del(ctx, fmt.Sprintf("mcp:config:%s", name))
}

func (r *Registry) GetConfigs(ctx context.Context) ([]domain.ToolConfig, error) {
	keys, err := r.valkey.Keys(ctx, "mcp:config:*")
	if err != nil {
		return nil, err
	}

	var configs []domain.ToolConfig
	for _, k := range keys {
		val, err := r.valkey.Get(ctx, k)
		if err != nil {
			continue
		}
		var cfg domain.ToolConfig
		if err := json.Unmarshal([]byte(val), &cfg); err != nil {
			continue
		}
		configs = append(configs, cfg)
	}
	return configs, nil
}
