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
	// Anthropic does not have a public list models endpoint yet, so we return a curated list.
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
	Stream    bool            `json:"stream"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
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
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			line = strings.TrimPrefix(line, "data: ")

			var event struct {
				Type string `json:"type"`
				Delta struct {
					Text string `json:"text"`
				} `json:"delta"`
			}

			if err := json.Unmarshal([]byte(line), &event); err != nil {
				continue
			}

			if event.Type == "content_block_delta" && event.Delta.Text != "" {
				select {
				case <-ctx.Done():
					return
				case out <- domain.Chunk{Text: event.Delta.Text}:
				}
			}
		}
	}()

	return out, nil
}