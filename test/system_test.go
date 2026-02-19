package test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/ecoza/ai-oak-orchestrator/internal/agent"
	"github.com/ecoza/ai-oak-orchestrator/internal/domain"
	"github.com/ecoza/ai-oak-orchestrator/internal/mcp"
	"go.uber.org/zap/zaptest"
)

type mockLLM struct {
	responses []domain.Chunk
}

func (m *mockLLM) GenerateStream(ctx context.Context, prompt string, tools []domain.Tool) (<-chan domain.Chunk, error) {
	ch := make(chan domain.Chunk, len(m.responses))
	for _, r := range m.responses {
		ch <- r
	}
	close(ch)
	return ch, nil
}
func (m *mockLLM) ListModels(ctx context.Context) ([]string, error) { return []string{"mock"}, nil }
func (m *mockLLM) SetModel(name string)                          {}

type mockToolProvider struct {
	clients []*mcp.Client
}
func (m *mockToolProvider) ListTools(ctx context.Context) []*mcp.Client { return m.clients }

func TestSystem_ChatFlow(t *testing.T) {
	logger := zaptest.NewLogger(t)
	
	// 1. Setup mocks
	llm := &mockLLM{
		responses: []domain.Chunk{
			{Text: "I will use a tool."},
			{ToolCall: &domain.ToolCall{ID: "1", Name: "test_tool", Arguments: json.RawMessage(`{}`)}},
			{Text: "Tool executed successfully."},
		},
	}
	tp := &mockToolProvider{}
	
	orch := agent.NewOrchestrator(llm, tp, logger)
	
	session := &domain.Session{ID: "test-session"}
	output := make(chan domain.AgentEvent, 10)
	input := make(chan domain.AgentCommand, 1)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 2. Start Agent Run
	errCh := make(chan error, 1)
	go func() {
		errCh <- orch.Run(ctx, session, output, input)
	}()

	// 3. Process events and provide approval
	var tokens []string
	var approvalRequested bool

	loop:
	for {
		select {
		case <-ctx.Done():
			t.Fatal("Test timed out")
		case err := <-errCh:
			if err != nil {
				t.Fatalf("Run failed: %v", err)
			}
			break loop
		case ev := <-output:
			switch ev.Type {
			case domain.EventToken:
				var text string
				json.Unmarshal(ev.Payload, &text)
				tokens = append(tokens, text)
			case domain.EventApprovalRequest:
				approvalRequested = true
				input <- domain.AgentCommand{Type: domain.CommandApprove}
			}
		}
	}

	// 4. Verify results
	if !approvalRequested {
		t.Error("Expected tool approval request was not emitted")
	}

	foundFinal := false
	for _, tok := range tokens {
		if tok == "Tool executed successfully." {
			foundFinal = true
		}
	}

	if !foundFinal {
		t.Error("Final response token not found in output")
	}
}