package llm

import (
	"context"
	"time"

	"github.com/ecoza/ai-oak-orchestrator/internal/domain"
)

type MockProvider struct{
	model string
}

func NewMockProvider() *MockProvider {
	return &MockProvider{model: "mock-default"}
}

func (m *MockProvider) SetModel(name string) {
	m.model = name
}

func (m *MockProvider) ListModels(ctx context.Context) ([]string, error) {
	return []string{"mock-model-1", "mock-model-2"}, nil
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