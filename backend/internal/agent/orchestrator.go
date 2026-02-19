package agent

import (
	"context"
	"encoding/json"

	"github.com/ecoza/ai-oak-orchestrator/internal/domain"
	"github.com/ecoza/ai-oak-orchestrator/internal/mcp"
	"github.com/ecoza/ai-oak-orchestrator/internal/mcp/llm"
	"go.uber.org/zap"
)

type Orchestrator struct {
	llm        llm.Provider
	mcpClients map[string]*mcp.Client
	logger     *zap.Logger
}

func NewOrchestrator(llm llm.Provider, logger *zap.Logger) *Orchestrator {
	return &Orchestrator{
		llm:        llm,
		mcpClients: make(map[string]*mcp.Client),
		logger:     logger,
	}
}

func (o *Orchestrator) RegisterMCPClient(name string, client *mcp.Client) {
	o.mcpClients[name] = client
}

func (o *Orchestrator) Run(ctx context.Context, session *domain.Session, output chan<- string) error {
	o.logger.Info("Agent loop started", zap.String("session_id", session.ID))

	for {
		// 1. Get current tools from all clients
		var allTools []domain.Tool
		// Simplified: we'd ideally cache these or pass them in
		for _, client := range o.mcpClients {
			tools, err := client.ListTools(ctx)
			if err == nil {
				allTools = append(allTools, tools...)
			}
		}

		// 2. Format prompt (simplified history)
		prompt := ""
		for _, m := range session.Messages {
			prompt += m.Role + ": " + m.Content + "\n"
		}

		// 3. Generate stream
		stream, err := o.llm.GenerateStream(ctx, prompt, allTools)
		if err != nil {
			return err
		}

		fullResponse := ""
		for chunk := range stream {
			fullResponse += chunk
			output <- chunk
		}

		// 4. Check for tool calls in fullResponse (Simplification for MVP)
		// Real implementation would use GenAI ToolCall parts instead of parsing text
		toolCall := o.findToolCall(fullResponse)
		if toolCall == nil {
			// No more tool calls, we are done
			session.Messages = append(session.Messages, domain.Message{Role: "assistant", Content: fullResponse})
			break
		}

		// 5. Execute Tool
		o.logger.Info("Executing tool", zap.String("name", toolCall.Name))
		result, err := o.executeTool(ctx, toolCall)
		if err != nil {
			result = &domain.ToolResult{IsError: true, Content: err.Error()}
		}

		// 6. Append Result and Loop
		resJSON, _ := json.Marshal(result)
		session.Messages = append(session.Messages, domain.Message{Role: "assistant", Content: "Call tool " + toolCall.Name})
		session.Messages = append(session.Messages, domain.Message{Role: "user", Content: "Tool result: " + string(resJSON)})
	}

	return nil
}

func (o *Orchestrator) findToolCall(text string) *domain.ToolCall {
	// TODO: Implement regex/JSON extraction logic
	return nil
}

func (o *Orchestrator) executeTool(ctx context.Context, call *domain.ToolCall) (*domain.ToolResult, error) {
	// For now, iterate all clients until we find the tool
	// Real implementation would have a registry mapping tool names to clients
	for _, client := range o.mcpClients {
		tools, _ := client.ListTools(ctx)
		for _, t := range tools {
			if t.Name == call.Name {
				return client.CallTool(ctx, call.Name, call.Arguments)
			}
		}
	}
	return nil, nil
}
