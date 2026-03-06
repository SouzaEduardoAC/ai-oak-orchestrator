package mcp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ecoza/ai-oak-orchestrator/internal/domain"
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

	var toolCfg *domain.ToolConfig
	for i, cfg := range configs {
		if cfg.Name == name {
			toolCfg = &configs[i]
			break
		}
	}

	if toolCfg == nil {
		return nil, fmt.Errorf("tool %s not found in registry", name)
	}

	// 2. SSE transport path
	if toolCfg.Transport == "sse" || toolCfg.Transport == "http" {
		sseTransport := transport.NewSSE(toolCfg.URL, toolCfg.Headers, tm.logger)
		mcpClient := NewClient(sseTransport)

		readyCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := sseTransport.WaitReady(readyCtx); err != nil {
			return nil, fmt.Errorf("SSE transport did not become ready: %w", err)
		}

		if err := mcpClient.Initialize(ctx); err != nil {
			return nil, fmt.Errorf("failed to initialize SSE MCP: %w", err)
		}

		tm.activeTools[name] = mcpClient
		return mcpClient, nil
	}

	// 3. Docker/stdio path
	containerName := fmt.Sprintf("ai-oak-mcp-%s", name)
	var containerID string

	inspect, err := tm.dockerManager.InspectContainer(ctx, containerName)
	if err == nil {
		containerID = inspect.ID
		if !inspect.State.Running {
			if err := tm.dockerManager.StartContainer(ctx, containerID); err != nil {
				return nil, fmt.Errorf("failed to start existing container: %w", err)
			}
		}
	} else {
		id, err := tm.dockerManager.CreateContainer(ctx, toolCfg.Image, nil, containerName)
		if err != nil {
			return nil, fmt.Errorf("failed to create container: %w", err)
		}
		containerID = id
		if err := tm.dockerManager.StartContainer(ctx, containerID); err != nil {
			return nil, fmt.Errorf("failed to start container: %w", err)
		}
	}

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

func (tm *ToolManager) StopTool(ctx context.Context, name string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	containerID, ok := tm.activeContainers[name]
	if !ok {
		return nil // Already stopped or never started
	}

	delete(tm.activeTools, name)
	delete(tm.activeContainers, name)

	if err := tm.dockerManager.StopContainer(ctx, containerID); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	return tm.dockerManager.RemoveContainer(ctx, containerID)
}

func (tm *ToolManager) ListTools(ctx context.Context) map[string]*Client {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	
	// Return a copy to avoid external modification during map iteration
	clients := make(map[string]*Client)
	for name, c := range tm.activeTools {
		clients[name] = c
	}
	return clients
}
