package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ecoza/ai-oak-orchestrator/internal/domain"
)

type ClaudeProvider struct {
	apiKey string
	model  string
}

func NewClaudeProvider(apiKey, model string) *ClaudeProvider {
	return &ClaudeProvider{
		apiKey: apiKey,
		model:  model,
	}
}

func (p *ClaudeProvider) SetModel(name string) {
	p.model = name
}

func (p *ClaudeProvider) ListModels(ctx context.Context) ([]string, error) {
	return []string{
		"claude-3-5-sonnet-20240620",
		"claude-3-opus-20240229",
		"claude-3-sonnet-20240229",
		"claude-3-haiku-20240307",
	}, nil
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
	out := make(chan domain.Chunk)

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

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	go func() {
		defer resp.Body.Close()
		defer close(out)

		scanner := bufio.NewScanner(resp.Body)
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

		if toolName != "" {
			select {
			case <-ctx.Done():
				return
			case out <- domain.Chunk{
				ToolCall: &domain.ToolCall{
					ID:        toolCallID,
					Name:      toolName,
					Arguments: json.RawMessage(toolArgs.String()),
				},
			}:
			}
		}
	}()

	return out, nil
}
