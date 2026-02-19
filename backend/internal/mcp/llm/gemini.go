package llm

import (
	"context"

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

func (p *GeminiProvider) GenerateStream(ctx context.Context, prompt string, tools []domain.Tool) (<-chan string, error) {
	model := p.client.GenerativeModel(p.model)
	
	if len(tools) > 0 {
		genaiTools := &genai.Tool{
			FunctionDeclarations: make([]*genai.FunctionDeclaration, len(tools)),
		}
		for i, t := range tools {
			genaiTools.FunctionDeclarations[i] = &genai.FunctionDeclaration{
				Name:        t.Name,
				Description: t.Description,
				// Placeholder for Parameters mapping
				// Parameters: mapMcpSchemaToGenaiSchema(t.Schema),
			}
		}
		model.Tools = []*genai.Tool{genaiTools}
	}
	
	iter := model.GenerateContentStream(ctx, genai.Text(prompt))
	out := make(chan string)

	go func() {
		defer close(out)
		for {
			resp, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				// TODO: Better error handling in stream
				return
			}

			for _, cand := range resp.Candidates {
				if cand.Content != nil {
					for _, part := range cand.Content.Parts {
						if text, ok := part.(genai.Text); ok {
							select {
							case <-ctx.Done():
								return
							case out <- string(text):
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
