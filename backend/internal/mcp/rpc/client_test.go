package rpc

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"
)

type mockTransport struct {
	sent    chan interface{}
	handler func(json.RawMessage)
	mu      sync.Mutex
}

func (m *mockTransport) Send(msg interface{}) error {
	m.sent <- msg
	return nil
}

func (m *mockTransport) Listen(handler func(json.RawMessage)) {
	m.mu.Lock()
	m.handler = handler
	m.mu.Unlock()
}

func (m *mockTransport) receive(raw json.RawMessage) {
	m.mu.Lock()
	h := m.handler
	m.mu.Unlock()
	if h != nil {
		h(raw)
	}
}

func TestClientCall(t *testing.T) {
	transport := &mockTransport{sent: make(chan interface{}, 1)}
	client := NewClient(transport)

	// Simulate server response in a goroutine
	go func() {
		req := (<-transport.sent).(Request)
		resp := Response{
			JSONRPC: JSONRPCVersion,
			Result:  json.RawMessage(`{"success":true}`),
			ID:      req.ID,
		}
		data, _ := json.Marshal(resp)
		transport.receive(data)
	}()

	var result map[string]bool
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := client.Call(ctx, "test", nil, &result)
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}

	if !result["success"] {
		t.Errorf("Expected success: true, got %v", result)
	}
}
