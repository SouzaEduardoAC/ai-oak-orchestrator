package agent

import (
	"context"

	"github.com/ecoza/ai-oak-orchestrator/internal/domain"
	"github.com/ecoza/ai-oak-orchestrator/internal/mcp/llm"
	"go.uber.org/zap"
)

type Orchestrator struct {
	llm    llm.Provider
	logger *zap.Logger
}

func NewOrchestrator(llm llm.Provider, logger *zap.Logger) *Orchestrator {
	return &Orchestrator{
		llm:    llm,
		logger: logger,
	}
}

func (o *Orchestrator) Run(ctx context.Context, session *domain.Session, output chan<- string) error {
	o.logger.Info("Agent starting run", zap.String("session_id", session.ID))

	// Simplified loop: just call LLM once for now
	// In reality this loops: LLM -> [Tool Call -> Result -> LLM] -> Final Answer
	
	prompt := ""
	for _, msg := range session.Messages {
		prompt += msg.Role + ": " + msg.Content + "\n"
	}

	stream, err := o.llm.GenerateStream(ctx, prompt, nil) // No tools yet
	if err != nil {
		return err
	}

	for chunk := range stream {
		output <- chunk
	}

	return nil
}