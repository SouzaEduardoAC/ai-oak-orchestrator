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

type OpenAIProvider struct {
	apiKey string
	model  string
}

func NewOpenAIProvider(apiKey, model string) *OpenAIProvider {
	return &OpenAIProvider{
		apiKey: apiKey,
		model:  model,
	}
}

func (p *OpenAIProvider) SetModel(name string) {
	p.model = name
}

func (p *OpenAIProvider) ListModels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.openai.com/v1/models", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
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

	var models []string
	for _, m := range result.Data {
		if strings.HasPrefix(m.ID, "gpt-") {
			models = append(models, m.ID)
		}
	}
	return models, nil
}

type openAIRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
	Tools    []openAITool    `json:"tools,omitempty"`
	Stream   bool            `json:"stream"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAITool struct {
	Type     string `json:"type"`
	Function struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		Parameters  json.RawMessage `json:"parameters"`
	} `json:"function"`
}

func (p *OpenAIProvider) GenerateStream(ctx context.Context, prompt string, tools []domain.Tool) (<-chan domain.Chunk, error) {
	out := make(chan domain.Chunk)

	reqBody := openAIRequest{
		Model: p.model,
		Messages: []openAIMessage{
			{Role: "user", Content: prompt},
		},
		Stream: true,
	}

	if len(tools) > 0 {
		for _, t := range tools {
			ot := openAITool{Type: "function"}
			ot.Function.Name = t.Name
			ot.Function.Description = t.Description
			ot.Function.Parameters = t.Schema
			reqBody.Tools = append(reqBody.Tools, ot)
		}
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

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
			if line == "[DONE]" {
				break
			}

			var streamResp struct {
				Choices []struct {
					Delta struct {
						Content   string `json:"content"`
						ToolCalls []struct {
							ID       string `json:"id"`
							Function struct {
								Name      string `json:"name"`
								Arguments string `json:"arguments"`
							} `json:"function"`
						} `json:"tool_calls"`
					} `json:"delta"`
				} `json:"choices"`
			}

			if err := json.Unmarshal([]byte(line), &streamResp); err != nil {
				continue
			}

			if len(streamResp.Choices) == 0 {
				continue
			}

			delta := streamResp.Choices[0].Delta

			// Handle Text
			if delta.Content != "" {
				select {
				case <-ctx.Done():
					return
				case out <- domain.Chunk{Text: delta.Content}:
				}
			}

			// Handle Tool Calls (Accumulate fragments)
			if len(delta.ToolCalls) > 0 {
				tc := delta.ToolCalls[0]
				if tc.ID != "" {
					toolCallID = tc.ID
				}
				if tc.Function.Name != "" {
					toolName = tc.Function.Name
				}
				if tc.Function.Arguments != "" {
					toolArgs.WriteString(tc.Function.Arguments)
				}
			}
		}

		// Emit accumulated tool call at the end of the stream
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
