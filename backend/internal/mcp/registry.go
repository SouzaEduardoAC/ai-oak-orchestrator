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

	keys, err := r.redis.Keys(ctx, "mcp:config:*")

	if err != nil {

		return nil, err

	}



	var configs []domain.ToolConfig

	for _, k := range keys {

		val, err := r.redis.Get(ctx, k)

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
