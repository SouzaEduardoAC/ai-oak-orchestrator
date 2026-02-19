package llm

import (
	"context"
	"encoding/json"

	"github.com/ecoza/ai-oak-orchestrator/internal/domain"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type GeminiProvider struct {
	client *genai.Client
	model  string
}

func NewGeminiProvider(ctx context.Context, apiKey string, modelName string) (*GeminiProvider, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}
	return &GeminiProvider{
		client: client,
		model:  modelName,
	}, nil
}

func (p *GeminiProvider) GenerateStream(ctx context.Context, prompt string, tools []domain.Tool) (<-chan domain.Chunk, error) {
	model := p.client.GenerativeModel(p.model)
	
	if len(tools) > 0 {
		genaiTools := &genai.Tool{
			FunctionDeclarations: make([]*genai.FunctionDeclaration, len(tools)),
		}
		for i, t := range tools {
			genaiTools.FunctionDeclarations[i] = &genai.FunctionDeclaration{
				Name:        t.Name,
				Description: t.Description,
				// Parameters: mapMcpSchemaToGenaiSchema(t.Schema),
			}
		}
		model.Tools = []*genai.Tool{genaiTools}
	}
	
	iter := model.GenerateContentStream(ctx, genai.Text(prompt))
	out := make(chan domain.Chunk)

	go func() {
		defer close(out)
		for {
			resp, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return
			}

			for _, cand := range resp.Candidates {
				if cand.Content != nil {
					for _, part := range cand.Content.Parts {
						switch v := part.(type) {
						case genai.Text:
							select {
							case <-ctx.Done():
								return
							case out <- domain.Chunk{Text: string(v)}:
							}
						case genai.FunctionCall:
							args, _ := json.Marshal(v.Args)
							select {
							case <-ctx.Done():
								return
							case out <- domain.Chunk{
								ToolCall: &domain.ToolCall{
									Name:      v.Name,
									Arguments: args,
								},
							}:
							}
						}
					}
				}
			}
		}
	}()

	return out, nil
}

func (p *GeminiProvider) Close() error {
	return p.client.Close()
}
