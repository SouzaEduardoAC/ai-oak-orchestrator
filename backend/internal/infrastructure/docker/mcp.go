package docker

import (
	"context"

	"github.com/ecoza/ai-oak-orchestrator/internal/mcp"
	"github.com/ecoza/ai-oak-orchestrator/internal/mcp/transport"
	"go.uber.org/zap"
)

type ManagedMCPContainer struct {
	Manager     *ContainerManager
	ContainerID string
	MCPClient   *mcp.Client
	logger      *zap.Logger
}

func (m *ContainerManager) NewMCPClient(ctx context.Context, containerID string, logger *zap.Logger) (*ManagedMCPContainer, error) {
	hijacked, err := m.Attach(ctx, containerID)
	if err != nil {
		return nil, err
	}

	// The hijacked response Conn is an io.ReadWriter
	stdioTransport := transport.NewStdio(hijacked.Reader, hijacked.Conn, logger)
	mcpClient := mcp.NewClient(stdioTransport)

	// Initialize the protocol
	if err := mcpClient.Initialize(ctx); err != nil {
		hijacked.Close()
		return nil, err
	}

	return &ManagedMCPContainer{
		Manager:     m,
		ContainerID: containerID,
		MCPClient:   mcpClient,
		logger:      logger,
	}, nil
}

func (mc *ManagedMCPContainer) Close() error {
	return mc.Manager.StopContainer(context.Background(), mc.ContainerID)
}
