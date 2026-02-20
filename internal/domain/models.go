package domain

import "encoding/json"

type Session struct {
        ID        string    `json:"id"`
        Messages  []Message `json:"messages"`
        CreatedAt int64     `json:"created_at"`
}

type Message struct {
        Role    string `json:"role"` // "user", "assistant", "system"
        Content string `json:"content"`
}

type Tool struct {
        Name        string          `json:"name"`
        Description string          `json:"description"`
        Schema      json.RawMessage `json:"inputSchema"`
}

type ToolCall struct {
        ID        string          `json:"id"`
        Name      string          `json:"name"`
        Arguments json.RawMessage `json:"arguments"`
}

type ToolResult struct {
        ID      string `json:"id"`
        Content string `json:"content"`
        IsError bool   `json:"isError"`
}

type Chunk struct {
        Text     string    `json:"text,omitempty"`
        ToolCall *ToolCall `json:"toolCall,omitempty"`
}

type AgentEventType string

const (
        EventToken           AgentEventType = "agent:response"
        EventApprovalRequest AgentEventType = "tool:approval_required"
        EventError           AgentEventType = "agent:error"
        EventLog             AgentEventType = "agent:log"
	EventThinking        AgentEventType = "agent:thinking"
	EventToolOutput      AgentEventType = "tool:output"
)

type AgentEvent struct {
        Type    AgentEventType  `json:"type"`
        Payload json.RawMessage `json:"payload"`
}

type AgentCommandType string

const (
        CommandApprove  AgentCommandType = "approve"
        CommandReject   AgentCommandType = "reject"
        CommandSetModel AgentCommandType = "set_model"
)

type AgentCommand struct {
        Type    AgentCommandType `json:"type"`
        Payload json.RawMessage  `json:"payload"`
}

type ToolConfig struct {
        Name   string            `json:"name"`
        Image  string            `json:"image"`
        Env    map[string]string `json:"env"`
        Active bool              `json:"active"`
}
