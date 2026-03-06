package llm

import (
	"context"
	"fmt"

	"github.com/ecoza/ai-oak-orchestrator/internal/config"
)

func NewProvider(ctx context.Context, cfg config.LLMConfig) (Provider, error) {
	switch cfg.Provider {
	case "gemini":
		return NewGeminiProvider(ctx, cfg.APIKey, cfg.Model)
	case "openai":
		return NewOpenAIProvider(cfg.APIKey, cfg.Model), nil
	case "claude":
		return NewClaudeProvider(cfg.APIKey, cfg.Model, cfg.BaseURL), nil
	case "mock":
		return NewMockProvider(), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}
}