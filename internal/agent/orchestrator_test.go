package agent

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ecoza/ai-oak-orchestrator/internal/domain"
	"github.com/ecoza/ai-oak-orchestrator/internal/mcp"
	"go.uber.org/zap/zaptest"
)

type mockLLM struct {
	responses [][]domain.Chunk
	current   int
}

func (m *mockLLM) GenerateStream(ctx context.Context, prompt string, tools []domain.Tool) (<-chan domain.Chunk, error) {
	ch := make(chan domain.Chunk, 10)
	if m.current < len(m.responses) {
		for _, r := range m.responses[m.current] {
			ch <- r
		}
		m.current++
	}
	close(ch)
	return ch, nil
}

func (m *mockLLM) ListModels(ctx context.Context) ([]string, error) {
	return []string{"mock"}, nil
}

func (m *mockLLM) SetModel(name string) {}

type mockToolProvider struct{}

func (m *mockToolProvider) ListTools(ctx context.Context) map[string]*mcp.Client {
	return nil
}

func TestOrchestratorRun_SimpleText(t *testing.T) {
	logger := zaptest.NewLogger(t)
	llm := &mockLLM{
		responses: [][]domain.Chunk{
			{
				{Text: "Hello"},
				{Text: " world"},
			},
		},
	}
	tp := &mockToolProvider{}
	
	orch := NewOrchestrator(llm, tp, logger)

	session := &domain.Session{ID: "test"}
	output := make(chan domain.AgentEvent, 10)
	input := make(chan domain.AgentCommand)

	err := orch.Run(context.Background(), session, output, input)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Verify events
	var received []string
	close(output)
	for ev := range output {
		if ev.Type == domain.EventToken {
			var text string
			json.Unmarshal(ev.Payload, &text)
			received = append(received, text)
		}
	}

	if len(received) != 2 || received[0] != "Hello" || received[1] != " world" {
		t.Errorf("Unexpected tokens: %v", received)
	}
}

func TestOrchestratorRun_ToolApproval(t *testing.T) {
	logger := zaptest.NewLogger(t)
	
	// 1. LLM emits a ToolCall
	// 2. LLM then emits a final answer (in the second iteration)
	llm := &mockLLM{
		responses: [][]domain.Chunk{
			{{ToolCall: &domain.ToolCall{ID: "call_1", Name: "test_tool"}}},
			{{Text: "Final result"}},
		},
	}
	tp := &mockToolProvider{}
	
	orch := NewOrchestrator(llm, tp, logger)

	session := &domain.Session{ID: "test"}
	output := make(chan domain.AgentEvent, 10)
	input := make(chan domain.AgentCommand, 1)

	// In a goroutine, wait for the approval request and send an approve command
	go func() {
		for ev := range output {
			if ev.Type == domain.EventApprovalRequest {
				input <- domain.AgentCommand{Type: domain.CommandApprove}
				return
			}
		}
	}()

	err := orch.Run(context.Background(), session, output, input)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Verify we got the token after approval
	close(output)
	foundFinal := false
	for ev := range output {
		if ev.Type == domain.EventToken {
			var text string
			json.Unmarshal(ev.Payload, &text)
			if text == "Final result" {
				foundFinal = true
			}
		}
	}

	if !foundFinal {
		t.Error("Did not receive final answer after tool approval")
	}
}