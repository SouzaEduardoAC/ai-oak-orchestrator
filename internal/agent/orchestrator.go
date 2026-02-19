package agent

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/ecoza/ai-oak-orchestrator/internal/domain"
	"github.com/ecoza/ai-oak-orchestrator/internal/mcp"
	"github.com/ecoza/ai-oak-orchestrator/internal/mcp/llm"
	"go.uber.org/zap"
)

type Orchestrator struct {
	llm         llm.Provider
	toolManager *mcp.ToolManager
	logger      *zap.Logger
	mu          sync.RWMutex
}

func NewOrchestrator(llm llm.Provider, tm *mcp.ToolManager, logger *zap.Logger) *Orchestrator {
	return &Orchestrator{
		llm:         llm,
		toolManager: tm,
		logger:      logger,
	}
}

func (o *Orchestrator) Run(ctx context.Context, session *domain.Session, output chan<- domain.AgentEvent, input <-chan domain.AgentCommand) error {
	o.logger.Info("Agent loop started", zap.String("session_id", session.ID))

	for {
		// 0. Check for pending commands (e.g. SetModel)
		select {
		case cmd := <-input:
			if cmd.Type == domain.CommandSetModel {
				var modelName string
				if err := json.Unmarshal(cmd.Payload, &modelName); err == nil {
					o.mu.Lock()
					o.llm.SetModel(modelName)
					o.mu.Unlock()
					o.logger.Info("Model changed", zap.String("model", modelName))
				}
			}
		default:
		}

		// 1. Get current tools from all clients
		var allTools []domain.Tool
		for _, client := range o.toolManager.ListTools(ctx) {
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
		o.mu.RLock()
		stream, err := o.llm.GenerateStream(ctx, prompt, allTools)
		o.mu.RUnlock()
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

		var approved, rejected bool
		for !approved && !rejected {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case cmd := <-input:
				switch cmd.Type {
				case domain.CommandApprove:
					approved = true
				case domain.CommandReject:
					rejected = true
					o.logger.Info("Tool rejected by user", zap.String("name", activeToolCall.Name))
					session.Messages = append(session.Messages, domain.Message{Role: "user", Content: "Tool execution rejected by user."})
				case domain.CommandSetModel:
					var modelName string
					if err := json.Unmarshal(cmd.Payload, &modelName); err == nil {
						o.mu.Lock()
						o.llm.SetModel(modelName)
						o.mu.Unlock()
						o.logger.Info("Model changed", zap.String("model", modelName))
					}
				}
			}
		}

		if rejected {
			continue
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
	for _, client := range o.toolManager.ListTools(ctx) {
		tools, _ := client.ListTools(ctx)
		for _, t := range tools {
			if t.Name == call.Name {
				return client.CallTool(ctx, call.Name, call.Arguments)
			}
		}
	}
	return nil, nil
}
