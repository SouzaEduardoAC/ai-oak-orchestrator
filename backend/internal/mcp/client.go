package mcp

import (
	"context"
	"encoding/json"

	"github.com/ecoza/ai-oak-orchestrator/internal/domain"
	"github.com/ecoza/ai-oak-orchestrator/internal/mcp/rpc"
)

type Client struct {
	rpc *rpc.Client
}

func NewClient(t rpc.Transport) *Client {
	return &Client{
		rpc: rpc.NewClient(t),
	}
}

type InitializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ClientInfo      map[string]string      `json:"clientInfo"`
}

func (c *Client) Initialize(ctx context.Context) error {
	params := InitializeParams{
		ProtocolVersion: "2024-11-05",
		Capabilities:    make(map[string]interface{}),
		ClientInfo: map[string]string{
			"name":    "ai-oak-orchestrator",
			"version": "0.1.0",
		},
	}
	return c.rpc.Call(ctx, "initialize", params, nil)
}

type ListToolsResult struct {
	Tools []domain.Tool `json:"tools"`
}

func (c *Client) ListTools(ctx context.Context) ([]domain.Tool, error) {
	var result ListToolsResult
	if err := c.rpc.Call(ctx, "tools/list", nil, &result); err != nil {
		return nil, err
	}
	return result.Tools, nil
}

type CallToolParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type CallToolResult struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	IsError bool `json:"isError"`
}

func (c *Client) CallTool(ctx context.Context, name string, args json.RawMessage) (*domain.ToolResult, error) {
	params := CallToolParams{
		Name:      name,
		Arguments: args,
	}
	var result CallToolResult
	if err := c.rpc.Call(ctx, "tools/call", params, &result); err != nil {
		return nil, err
	}

	content := ""
	if len(result.Content) > 0 {
		content = result.Content[0].Text
	}

	return &domain.ToolResult{
		Content: content,
		IsError: result.IsError,
	}, nil
}
