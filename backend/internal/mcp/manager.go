package mcp

import (
	"context"
	"fmt"
	"sync"

	"github.com/ecoza/ai-oak-orchestrator/internal/infrastructure/docker"
	"github.com/ecoza/ai-oak-orchestrator/internal/mcp/transport"
	"go.uber.org/zap"
)

type ToolManager struct {
	dockerManager *docker.ContainerManager
	registry      *Registry
	activeTools   map[string]*Client
	activeContainers map[string]string // map[toolName]containerID
	mu            sync.RWMutex
	logger        *zap.Logger
}

func NewToolManager(dm *docker.ContainerManager, r *Registry, l *zap.Logger) *ToolManager {
	return &ToolManager{
		dockerManager:    dm,
		registry:         r,
		activeTools:      make(map[string]*Client),
		activeContainers: make(map[string]string),
		logger:           l,
	}
}

func (tm *ToolManager) EnsureTool(ctx context.Context, name string) (*Client, error) {
	tm.mu.RLock()
	client, ok := tm.activeTools[name]
	tm.mu.RUnlock()
	if ok {
		return client, nil
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Double check
	if client, ok := tm.activeTools[name]; ok {
		return client, nil
	}

	// 1. Get config
	configs, err := tm.registry.GetConfigs(ctx)
	if err != nil {
		return nil, err
	}

	var targetConfig *string
	for _, cfg := range configs {
		if cfg.Name == name {
			targetConfig = &cfg.Image
			break
		}
	}

	if targetConfig == nil {
		return nil, fmt.Errorf("tool %s not found in registry", name)
	}

	// 2. Start/Get Container (Singleton Logic)
	containerName := fmt.Sprintf("ai-oak-mcp-%s", name)
	var containerID string

	inspect, err := tm.dockerManager.InspectContainer(ctx, containerName)
	if err == nil {
		// Container exists
		containerID = inspect.ID
		if !inspect.State.Running {
			if err := tm.dockerManager.StartContainer(ctx, containerID); err != nil {
				return nil, fmt.Errorf("failed to start existing container: %w", err)
			}
		}
	} else {
		// Container does not exist, create it
		// First pull if not present (simplified for MVP)
		id, err := tm.dockerManager.CreateContainer(ctx, *targetConfig, nil, containerName)
		if err != nil {
			return nil, fmt.Errorf("failed to create container: %w", err)
		}
		containerID = id
		if err := tm.dockerManager.StartContainer(ctx, containerID); err != nil {
			return nil, fmt.Errorf("failed to start container: %w", err)
		}
	}

	// 3. Attach and Initialize MCP
	hijacked, err := tm.dockerManager.Attach(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to attach to container: %w", err)
	}

	stdioTransport := transport.NewStdio(hijacked.Reader, hijacked.Conn, tm.logger)
	mcpClient := NewClient(stdioTransport)

	if err := mcpClient.Initialize(ctx); err != nil {
		hijacked.Close()
		return nil, fmt.Errorf("failed to initialize MCP: %w", err)
	}

	tm.activeTools[name] = mcpClient
	tm.activeContainers[name] = containerID
	
	return mcpClient, nil
}

func (tm *ToolManager) InitializeAll(ctx context.Context) error {
	configs, err := tm.registry.GetConfigs(ctx)
	if err != nil {
		return err
	}

	for _, cfg := range configs {
		if cfg.Active {
			go func(name string) {
				if _, err := tm.EnsureTool(context.Background(), name); err != nil {
					tm.logger.Error("Failed to auto-start tool", zap.String("name", name), zap.Error(err))
				}
			}(cfg.Name)
		}
	}
	return nil
}

func (tm *ToolManager) ListTools(ctx context.Context) []*Client {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	var clients []*Client
	for _, c := range tm.activeTools {
		clients = append(clients, c)
	}
	return clients
}
