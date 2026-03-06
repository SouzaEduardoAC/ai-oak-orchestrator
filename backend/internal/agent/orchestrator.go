package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

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
			listCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			tools, err := client.ListTools(listCtx)
			cancel()
			if err == nil {
				allTools = append(allTools, tools...)
			} else {
				o.logger.Warn("ListTools timed out or failed, skipping client", zap.Error(err))
			}
		}

		// Keep only the last 20 messages to prevent unbounded prompt growth.
		msgs := session.Messages
		if len(msgs) > 20 {
			msgs = msgs[len(msgs)-20:]
		}
		prompt := ""
		for _, m := range msgs {
			prompt += m.Role + ": " + m.Content + "\n"
		}

		// Use the last user message as the relevance query for tool selection.
		query := ""
		for i := len(msgs) - 1; i >= 0; i-- {
			if msgs[i].Role == "user" {
				query = msgs[i].Content
				break
			}
		}
		tools := selectRelevantTools(allTools, query, 15)

		o.logger.Info("Calling GenerateStream", zap.Int("tools_total", len(allTools)), zap.Int("tools_selected", len(tools)), zap.Int("prompt_len", len(prompt)))
		o.mu.RLock()
		stream, err := o.llm.GenerateStream(ctx, prompt, tools)
		o.mu.RUnlock()
		o.logger.Info("GenerateStream returned", zap.Error(err))
		if err != nil {
			errPayload, _ := json.Marshal(err.Error())
			output <- domain.AgentEvent{Type: domain.EventError, Payload: errPayload}
			return err
		}

		// Send periodic progress events while waiting for the LLM to respond,
		// so the frontend does not appear frozen during gateway queue delays.
		llmDone := make(chan struct{})
		go func() {
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()
			elapsed := 0
			for {
				select {
				case <-llmDone:
					return
				case <-ticker.C:
					elapsed += 5
					o.logger.Info("LLM progress tick", zap.Int("elapsed_s", elapsed))
					progressPayload, _ := json.Marshal(fmt.Sprintf("⏳ Waiting for AI response... %ds", elapsed))
					output <- domain.AgentEvent{Type: domain.EventToken, Payload: progressPayload}
				}
			}
		}()

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
		close(llmDone)

		if streamErr != nil {
			errPayload, _ := json.Marshal(streamErr.Error())
			output <- domain.AgentEvent{Type: domain.EventError, Payload: errPayload}
			return streamErr
		}

		if activeToolCall == nil {
			session.Messages = append(session.Messages, domain.Message{Role: "assistant", Content: fullResponse})
			break
		}

		approvalPayload, marshalErr := json.Marshal(activeToolCall)
		o.logger.Info("Tool approval requested", zap.String("name", activeToolCall.Name), zap.String("payload", string(approvalPayload)), zap.Error(marshalErr))
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

		// Send periodic progress events so the frontend knows the tool is still running.
		toolDone := make(chan struct{})
		go func() {
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-toolDone:
					return
				case <-ticker.C:
					progressPayload, _ := json.Marshal("⏳ Tool executing...")
					output <- domain.AgentEvent{Type: domain.EventToken, Payload: progressPayload}
				}
			}
		}()

		result, err := o.executeTool(ctx, activeToolCall, output)
		close(toolDone)
		if err != nil {
			result = &domain.ToolResult{IsError: true, Content: err.Error()}
		}

		// Send tool output to frontend
		resJSON, _ := json.Marshal(result)
		output <- domain.AgentEvent{Type: domain.EventToolOutput, Payload: resJSON}

		// Truncate the tool result stored in the session to prevent huge prompts
		// on subsequent LLM calls.
		resultSummary := string(resJSON)
		if len(resultSummary) > 2000 {
			resultSummary = resultSummary[:2000] + "... [truncated for context]"
		}
		session.Messages = append(session.Messages, domain.Message{Role: "assistant", Content: "Call tool " + activeToolCall.Name})
		session.Messages = append(session.Messages, domain.Message{Role: "user", Content: "Tool result: " + resultSummary})
	}

	return nil
}

// selectRelevantTools returns at most maxTools tools ranked by keyword
// overlap between their name/description and the user's query.
// If allTools fits within maxTools, it is returned as-is.
func selectRelevantTools(allTools []domain.Tool, query string, maxTools int) []domain.Tool {
	if len(allTools) <= maxTools {
		return allTools
	}

	// Tokenise query into a set of lowercase words (3+ chars).
	words := make(map[string]bool)
	for _, w := range strings.Fields(strings.ToLower(query)) {
		w = strings.Trim(w, ".,!?;:'\"()")
		if len(w) >= 3 {
			words[w] = true
		}
	}

	type scored struct {
		tool  domain.Tool
		score int
	}
	scored_tools := make([]scored, len(allTools))
	for i, t := range allTools {
		hay := strings.ToLower(t.Name + " " + t.Description)
		score := 0
		for w := range words {
			if strings.Contains(hay, w) {
				score++
			}
		}
		scored_tools[i] = scored{t, score}
	}

	sort.SliceStable(scored_tools, func(i, j int) bool {
		return scored_tools[i].score > scored_tools[j].score
	})

	result := make([]domain.Tool, maxTools)
	for i := range result {
		result[i] = scored_tools[i].tool
	}
	return result
}

func (o *Orchestrator) executeTool(ctx context.Context, call *domain.ToolCall, output chan<- domain.AgentEvent) (*domain.ToolResult, error) {
	for _, client := range o.toolManager.ListTools(ctx) {
		listCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		tools, _ := client.ListTools(listCtx)
		cancel()
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
