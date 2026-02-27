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

type ToolProvider interface {
	ListTools(ctx context.Context) map[string]*mcp.Client
}

type Orchestrator struct {
	llm         llm.Provider
	toolManager ToolProvider
	logger      *zap.Logger
	mu          sync.RWMutex
}

func NewOrchestrator(llm llm.Provider, tm ToolProvider, logger *zap.Logger) *Orchestrator {
	return &Orchestrator{
		llm:         llm,
		toolManager: tm,
		logger:      logger,
	}
}

func (o *Orchestrator) Run(ctx context.Context, session *domain.Session, output chan<- domain.AgentEvent, input <-chan domain.AgentCommand) error {
	o.logger.Info("Agent loop started", zap.String("session_id", session.ID))

	for {
		// Signal thinking start
		output <- domain.AgentEvent{Type: domain.EventThinking}

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

		var allTools []domain.Tool
		for _, client := range o.toolManager.ListTools(ctx) {
			tools, err := client.ListTools(ctx)
			if err == nil {
				allTools = append(allTools, tools...)
			}
		}

		prompt := ""
		for _, m := range session.Messages {
			prompt += m.Role + ": " + m.Content + "\n"
		}

		o.mu.RLock()
		stream, err := o.llm.GenerateStream(ctx, prompt, allTools)
		o.mu.RUnlock()
		if err != nil {
			errPayload, _ := json.Marshal(err.Error())
			output <- domain.AgentEvent{Type: domain.EventError, Payload: errPayload}
			return err
		}

		fullResponse := ""
		var activeToolCall *domain.ToolCall
		var streamErr error

		for chunk := range stream {
			if chunk.Error != nil {
				streamErr = chunk.Error
				break
			}
			if chunk.Text != "" {
				fullResponse += chunk.Text
				tokenPayload, _ := json.Marshal(chunk.Text)
				output <- domain.AgentEvent{Type: domain.EventToken, Payload: tokenPayload}
			}
			if chunk.ToolCall != nil {
				activeToolCall = chunk.ToolCall
			}
		}

		if streamErr != nil {
			errPayload, _ := json.Marshal(streamErr.Error())
			output <- domain.AgentEvent{Type: domain.EventError, Payload: errPayload}
			return streamErr
		}

		if activeToolCall == nil {
			session.Messages = append(session.Messages, domain.Message{Role: "assistant", Content: fullResponse})
			break
		}

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
					session.Messages = append(session.Messages, domain.Message{Role: "user", Content: "Tool execution rejected by user." })
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

		o.logger.Info("Executing tool", zap.String("name", activeToolCall.Name))
		result, err := o.executeTool(ctx, activeToolCall, output)
		if err != nil {
			result = &domain.ToolResult{IsError: true, Content: err.Error()}
		}

		// Send tool output to frontend
		resJSON, _ := json.Marshal(result)
		output <- domain.AgentEvent{Type: domain.EventToolOutput, Payload: resJSON}

		session.Messages = append(session.Messages, domain.Message{Role: "assistant", Content: "Call tool " + activeToolCall.Name})
		session.Messages = append(session.Messages, domain.Message{Role: "user", Content: "Tool result: " + string(resJSON)})
	}

	return nil
}

func (o *Orchestrator) executeTool(ctx context.Context, call *domain.ToolCall, output chan<- domain.AgentEvent) (*domain.ToolResult, error) {
	for _, client := range o.toolManager.ListTools(ctx) {
		tools, _ := client.ListTools(ctx)
		for _, t := range tools {
			if t.Name == call.Name {
				client.OnNotification(func(method string, params json.RawMessage) {
					o.logger.Info("Tool notification", zap.String("method", method))
					output <- domain.AgentEvent{
						Type:    domain.EventLog,
						Payload: params,
					}
				})
				return client.CallTool(ctx, call.Name, call.Arguments)
			}
		}
	}
	return nil, nil
}
