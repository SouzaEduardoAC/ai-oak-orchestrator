package transport

import (
	"bufio"
	"encoding/json"
	"io"
	"sync"

	"go.uber.org/zap"
)

type Stdio struct {
	r      io.Reader
	w      io.Writer
	logger *zap.Logger
	mu     sync.Mutex
}

func NewStdio(r io.Reader, w io.Writer, logger *zap.Logger) *Stdio {
	return &Stdio{
		r:      r,
		w:      w,
		logger: logger,
	}
}

func (s *Stdio) Send(msg interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	s.logger.Debug("Sending MCP message", zap.String("payload", string(data)))
	_, err = s.w.Write(append(data, '\n'))
	return err
}

func (s *Stdio) Listen(handler func(json.RawMessage)) {
	scanner := bufio.NewScanner(s.r)
	for scanner.Scan() {
		line := scanner.Bytes()
		s.logger.Debug("Received MCP message", zap.String("payload", string(line)))
		handler(json.RawMessage(line))
	}
}