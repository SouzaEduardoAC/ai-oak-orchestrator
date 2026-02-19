package services

import (
	"context"
	"time"

	"github.com/ecoza/ai-oak-orchestrator/internal/infrastructure/docker"
	"github.com/ecoza/ai-oak-orchestrator/internal/mcp"
	"go.uber.org/zap"
)

type JanitorService struct {
	dockerManager *docker.ContainerManager
	registry      *mcp.Registry
	logger        *zap.Logger
}

func NewJanitorService(dm *docker.ContainerManager, r *mcp.Registry, l *zap.Logger) *JanitorService {
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
	// TODO: List all containers with label 'ai-oak-mcp'
	// TODO: Cross-reference with Registry
	// TODO: Kill orphans
}
