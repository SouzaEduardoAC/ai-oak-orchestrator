package services

import (
	"context"
	"testing"

	"github.com/ecoza/ai-oak-orchestrator/internal/domain"
	"github.com/docker/docker/api/types"
	"go.uber.org/zap/zaptest"
)

type mockDockerManager struct {
	containers []types.Container
	removed    []string
}

func (m *mockDockerManager) ListManagedContainers(ctx context.Context) ([]types.Container, error) {
	return m.containers, nil
}

func (m *mockDockerManager) RemoveContainer(ctx context.Context, id string) error {
	m.removed = append(m.removed, id)
	return nil
}

type mockToolRegistry struct {
	configs []domain.ToolConfig
}

func (m *mockToolRegistry) GetConfigs(ctx context.Context) ([]domain.ToolConfig, error) {
	return m.configs, nil
}

func TestJanitorCleanup(t *testing.T) {
	logger := zaptest.NewLogger(t)
	
	dm := &mockDockerManager{
		containers: []types.Container{
			{ID: "keep-me", Names: []string{"/keep-me"}},
			{ID: "kill-me", Names: []string{"/kill-me"}},
		},
	}
	
	reg := &mockToolRegistry{
		configs: []domain.ToolConfig{
			{Name: "keep-me", Active: true},
		},
	}

	janitor := NewJanitorService(dm, reg, logger)
	janitor.Cleanup(context.Background())

	if len(dm.removed) != 1 {
		t.Fatalf("Expected 1 container to be removed, got %d", len(dm.removed))
	}

	if dm.removed[0] != "kill-me" {
		t.Errorf("Expected 'kill-me' to be removed, got %s", dm.removed[0])
	}
}