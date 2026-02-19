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

func (m *MockProvider) GenerateStream(ctx context.Context, prompt string, tools []domain.Tool) (<-chan string, error) {
	out := make(chan string)
	go func() {
		defer close(out)
		words := []string{"Hello", " ", "world", " ", "from", " ", "mock", " ", "provider", "."}
		for _, w := range words {
			select {
			case <-ctx.Done():
				return
			case out <- w:
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
	return out, nil
}
