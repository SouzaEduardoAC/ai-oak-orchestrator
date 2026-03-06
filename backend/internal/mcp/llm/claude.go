package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/ecoza/ai-oak-orchestrator/internal/domain"
)

type ClaudeProvider struct {
	apiKey  string
	model   string
	baseURL string
}

func NewClaudeProvider(apiKey, model, baseURL string) *ClaudeProvider {
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}
	return &ClaudeProvider{
		apiKey:  apiKey,
		model:   model,
		baseURL: baseURL,
	}
}

func (p *ClaudeProvider) SetModel(name string) {
	p.model = name
}

func (p *ClaudeProvider) ListModels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/v1/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	models := make([]string, 0, len(result.Data))
	for _, m := range result.Data {
		models = append(models, m.ID)
	}
	return models, nil
}

type claudeRequest struct {
	Model     string          `json:"model"`
	Messages  []claudeMessage `json:"messages"`
	MaxTokens int             `json:"max_tokens"`
	Tools     []claudeTool    `json:"tools,omitempty"`
	Stream    bool            `json:"stream"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

func (p *ClaudeProvider) GenerateStream(ctx context.Context, prompt string, tools []domain.Tool) (<-chan domain.Chunk, error) {
	out := make(chan domain.Chunk, 4)

	reqBody := claudeRequest{
		Model: p.model,
		Messages: []claudeMessage{
			{Role: "user", Content: prompt},
		},
		MaxTokens: 4096,
		Stream:    true,
	}

	if len(tools) > 0 {
		for _, t := range tools {
			reqBody.Tools = append(reqBody.Tools, claudeTool{
				Name:        t.Name,
				Description: t.Description,
				InputSchema: t.Schema,
			})
		}
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(out)

		// Use a context with timeout so the HTTP call (including waiting for
		// response headers from a slow/queued gateway) is bounded. Running
		// client.Do inside the goroutine means the caller never blocks here.
		reqCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
		defer cancel()

		req, err := http.NewRequestWithContext(reqCtx, "POST", p.baseURL+"/v1/messages", bytes.NewBuffer(data))
		if err != nil {
			select {
			case out <- domain.Chunk{Error: err}:
			case <-ctx.Done():
			}
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-api-key", p.apiKey)
		req.Header.Set("anthropic-version", "2023-06-01")

		resp, err := (&http.Client{}).Do(req)
		if err != nil {
			select {
			case out <- domain.Chunk{Error: err}:
			case <-ctx.Done():
			}
			return
		}
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
		var toolCallID string
		var toolName string
		var toolArgs strings.Builder

		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			line = strings.TrimPrefix(line, "data: ")

			var event struct {
				Type string `json:"type"`
				// content_block_start
				ContentBlock struct {
					Type string `json:"type"`
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"content_block"`
				// content_block_delta
				Delta struct {
					Type         string `json:"type"`
					Text         string `json:"text"`
					PartialJSON  string `json:"partial_json"`
				} `json:"delta"`
			}

			if err := json.Unmarshal([]byte(line), &event); err != nil {
				continue
			}

			switch event.Type {
			case "content_block_start":
				if event.ContentBlock.Type == "tool_use" {
					toolCallID = event.ContentBlock.ID
					toolName = event.ContentBlock.Name
					toolArgs.Reset() // reset args for each new tool block
				}
			case "content_block_delta":
				if event.Delta.Type == "text_delta" {
					select {
					case <-ctx.Done():
						return
					case out <- domain.Chunk{Text: event.Delta.Text}:
					}
				} else if event.Delta.Type == "input_json_delta" {
					toolArgs.WriteString(event.Delta.PartialJSON)
				}
			}
		}

		if err := scanner.Err(); err != nil {
			out <- domain.Chunk{Error: err}
			return
		}

		if toolName != "" {
			args := toolArgs.String()
			if args == "" {
				args = "{}"
			}
			select {
			case <-ctx.Done():
				return
			case out <- domain.Chunk{
				ToolCall: &domain.ToolCall{
					ID:        toolCallID,
					Name:      toolName,
					Arguments: json.RawMessage(args),
				},
			}:
			}
		}
	}()

	return out, nil
}
