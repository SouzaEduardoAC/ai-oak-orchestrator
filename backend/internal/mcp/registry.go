package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ecoza/ai-oak-orchestrator/internal/domain"
	"github.com/ecoza/ai-oak-orchestrator/internal/infrastructure/redis"
)

type Registry struct {
	redis *redis.Client
}

func NewRegistry(r *redis.Client) *Registry {
	return &Registry{redis: r}
}

func (r *Registry) SaveConfig(ctx context.Context, cfg domain.ToolConfig) error {
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	return r.redis.Set(ctx, fmt.Sprintf("mcp:config:%s", cfg.Name), data, 0)
}

func (r *Registry) GetConfigs(ctx context.Context) ([]domain.ToolConfig, error) {
	// TODO: Use Redis SCAN to find all keys matching mcp:config:*
	// For now, this is a placeholder
	return nil, nil
}
