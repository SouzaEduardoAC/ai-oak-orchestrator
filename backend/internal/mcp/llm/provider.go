package llm

import (
	"context"

	"github.com/ecoza/ai-oak-orchestrator/internal/domain"
)

type Provider interface {
	GenerateStream(ctx context.Context, prompt string, tools []domain.Tool) (<-chan domain.Chunk, error)
}
