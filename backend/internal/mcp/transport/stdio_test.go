package transport

import (
	"bytes"
	"encoding/json"
	"testing"

	"go.uber.org/zap/zaptest"
)

func TestStdioSend(t *testing.T) {
	var buf bytes.Buffer
	logger := zaptest.NewLogger(t)
	stdio := NewStdio(nil, &buf, logger)

	msg := map[string]string{"test": "data"}
	err := stdio.Send(msg)
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	expected := `{"test":"data"}` + "\n"
	if buf.String() != expected {
		t.Errorf("Expected %q, got %q", expected, buf.String())
	}
}

func TestStdioListen(t *testing.T) {
	input := `{"test":"data"}` + "\n"
	buf := bytes.NewBufferString(input)
	logger := zaptest.NewLogger(t)
	stdio := NewStdio(buf, nil, logger)

	called := false
	stdio.Listen(func(raw json.RawMessage) {
		called = true
		if string(raw) != `{"test":"data"}` {
			t.Errorf("Unexpected message: %s", string(raw))
		}
	})

	if !called {
		t.Error("Handler not called")
	}
}
