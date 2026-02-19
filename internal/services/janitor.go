package services

import (
	"context"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/ecoza/ai-oak-orchestrator/internal/domain"
	"go.uber.org/zap"
)

type DockerManager interface {
	ListManagedContainers(ctx context.Context) ([]types.Container, error)
	RemoveContainer(ctx context.Context, id string) error
}

type ToolRegistry interface {
	GetConfigs(ctx context.Context) ([]domain.ToolConfig, error)
}

type JanitorService struct {
	dockerManager DockerManager
	registry      ToolRegistry
	logger        *zap.Logger
}

func NewJanitorService(dm DockerManager, r ToolRegistry, l *zap.Logger) *JanitorService {
	return &JanitorService{
		dockerManager: dm,
		registry:      r,
		logger:        l,
	}
}

func (s *JanitorService) Start(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.Cleanup(ctx)
		}
	}
}

func (s *JanitorService) Cleanup(ctx context.Context) {
	s.logger.Info("Janitor starting cleanup cycle")

	// 1. Get all managed containers from Docker
	containers, err := s.dockerManager.ListManagedContainers(ctx)
	if err != nil {
		s.logger.Error("Failed to list managed containers", zap.Error(err))
		return
	}

	// 2. Get all configured tools from Registry
	configs, err := s.registry.GetConfigs(ctx)
	if err != nil {
		s.logger.Error("Failed to get tool configs from registry", zap.Error(err))
		return
	}

	configMap := make(map[string]bool)
	for _, cfg := range configs {
		configMap[cfg.Name] = true
	}

	// 3. Identify and kill orphans
	for _, c := range containers {
		// Docker container names usually start with a slash
		name := ""
		if len(c.Names) > 0 {
			name = c.Names[0]
			if name[0] == '/' {
				name = name[1:]
			}
		}

		if name == "" {
			continue
		}

		// If container is not in registry, it's an orphan
		if _, ok := configMap[name]; !ok {
			s.logger.Warn("Killing orphaned MCP container", zap.String("id", c.ID), zap.String("name", name))
			if err := s.dockerManager.RemoveContainer(ctx, c.ID); err != nil {
				s.logger.Error("Failed to remove orphan container", zap.String("id", c.ID), zap.Error(err))
			}
		}
	}
}
