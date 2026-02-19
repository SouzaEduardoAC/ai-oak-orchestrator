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

func (o *Orchestrator) Run(ctx context.Context, session *domain.Session, output chan<- domain.AgentEvent, input <-chan domain.AgentCommand) error {
	o.logger.Info("Agent loop started", zap.String("session_id", session.ID))

	for {
		// 1. Get current tools from all clients
		var allTools []domain.Tool
		for _, client := range o.mcpClients {
			tools, err := client.ListTools(ctx)
			if err == nil {
				allTools = append(allTools, tools...)
			}
		}

		// 2. Format prompt
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
		var activeToolCall *domain.ToolCall

		for chunk := range stream {
			if chunk.Text != "" {
				fullResponse += chunk.Text
				output <- domain.AgentEvent{Type: domain.EventToken, Payload: json.RawMessage(`"` + chunk.Text + `"`)}
			}
			if chunk.ToolCall != nil {
				activeToolCall = chunk.ToolCall
			}
		}

		// 4. Check for tool calls
		if activeToolCall == nil {
			// No more tool calls, we are done
			session.Messages = append(session.Messages, domain.Message{Role: "assistant", Content: fullResponse})
			break
		}

		// 5. Tool Approval Wait
		o.logger.Info("Tool approval requested", zap.String("name", activeToolCall.Name))
		approvalPayload, _ := json.Marshal(activeToolCall)
		output <- domain.AgentEvent{Type: domain.EventApprovalRequest, Payload: approvalPayload}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case cmd := <-input:
			if cmd.Type == domain.CommandReject {
				o.logger.Info("Tool rejected by user", zap.String("name", activeToolCall.Name))
				session.Messages = append(session.Messages, domain.Message{Role: "user", Content: "Tool execution rejected by user."})
				continue // Loop back to LLM with rejection message
			}
			// Fallthrough if approved
		}

		// 6. Execute Tool
		o.logger.Info("Executing tool", zap.String("name", activeToolCall.Name))
		result, err := o.executeTool(ctx, activeToolCall)
		if err != nil {
			result = &domain.ToolResult{IsError: true, Content: err.Error()}
		}

		// 7. Append Result and Loop
		resJSON, _ := json.Marshal(result)
		session.Messages = append(session.Messages, domain.Message{Role: "assistant", Content: "Call tool " + activeToolCall.Name})
		session.Messages = append(session.Messages, domain.Message{Role: "user", Content: "Tool result: " + string(resJSON)})
	}

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
