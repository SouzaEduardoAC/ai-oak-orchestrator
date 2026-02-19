package llm

import (
	"context"
	"time"

	"github.com/ecoza/ai-oak-orchestrator/internal/domain"
)

type MockProvider struct{}

func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

func (m *MockProvider) GenerateStream(ctx context.Context, prompt string, tools []domain.Tool) (<-chan domain.Chunk, error) {
	out := make(chan domain.Chunk)
	go func() {
		defer close(out)
		words := []string{"Hello", " ", "world", " ", "from", " ", "mock", " ", "provider", "."}
		for _, w := range words {
			select {
			case <-ctx.Done():
				return
			case out <- domain.Chunk{Text: w}:
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
	return out, nil
}
