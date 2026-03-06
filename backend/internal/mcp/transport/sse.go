package transport

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

type SSE struct {
	sseURL     string
	headers    map[string]string
	logger     *zap.Logger
	mu         sync.Mutex
	postURL    string
	ready      chan struct{}
	readyOnce  sync.Once
	httpClient *http.Client
}

func NewSSE(sseURL string, headers map[string]string, logger *zap.Logger) *SSE {
	return &SSE{
		sseURL:     sseURL,
		headers:    headers,
		logger:     logger,
		ready:      make(chan struct{}),
		httpClient: &http.Client{},
	}
}

// WaitReady blocks until the SSE endpoint event has been received or ctx is cancelled.
func (s *SSE) WaitReady(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.ready:
		return nil
	}
}

func (s *SSE) Send(msg interface{}) error {
	s.mu.Lock()
	postURL := s.postURL
	s.mu.Unlock()

	if postURL == "" {
		return fmt.Errorf("SSE transport not ready: post URL not yet received")
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", postURL, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range s.headers {
		req.Header.Set(k, v)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("SSE post failed with status: %s", resp.Status)
	}
	return nil
}

// Listen connects to the SSE stream and dispatches events to handler.
// It automatically reconnects if the stream is dropped.
func (s *SSE) Listen(handler func(json.RawMessage)) {
	for {
		err := s.listenOnce(handler)
		if err != nil {
			s.logger.Warn("SSE: stream error, reconnecting in 3s", zap.String("url", s.sseURL), zap.Error(err))
		} else {
			s.logger.Warn("SSE: stream closed, reconnecting in 3s", zap.String("url", s.sseURL))
		}
		time.Sleep(3 * time.Second)
	}
}

func (s *SSE) listenOnce(handler func(json.RawMessage)) error {
	req, err := http.NewRequest("GET", s.sseURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	for k, v := range s.headers {
		req.Header.Set(k, v)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 10*1024*1024), 10*1024*1024) // 10MB max token size
	var eventType string
	var dataLines []string

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			// Dispatch accumulated event
			if len(dataLines) > 0 {
				data := strings.Join(dataLines, "\n")
				switch eventType {
				case "endpoint":
					postURL := data
					if strings.HasPrefix(data, "/") {
						// Resolve relative URL against the SSE base
						parts := strings.SplitN(s.sseURL, "/", 4)
						if len(parts) >= 3 {
							postURL = parts[0] + "//" + parts[2] + data
						}
					}
					s.mu.Lock()
					s.postURL = postURL
					s.mu.Unlock()
					s.logger.Info("SSE: post URL received", zap.String("url", postURL))
					s.readyOnce.Do(func() { close(s.ready) })

				case "message", "":
					handler(json.RawMessage(data))
				}
			}
			eventType = ""
			dataLines = nil
			continue
		}

		if strings.HasPrefix(line, "event:") {
			eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		} else if strings.HasPrefix(line, "data:") {
			dataLines = append(dataLines, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
	}

	return scanner.Err()
}
