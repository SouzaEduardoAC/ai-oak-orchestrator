# Technical Blueprint: Stability & Feature Parity

## 1. Executive Summary
**Objective:** Finalize the transition to the Go backend by achieving 100% feature parity with the Node.js MVP. This includes enabling tool-calling across all LLM providers, completing the persistent registry, and implementing automated health monitoring for MCP containers.

**Definition of Done:**
1.  **Tool Parity:** OpenAI and Claude providers correctly detect and return tool execution requests in their streams.
2.  **Persistent Registry:** MCP tool configurations are fully retrievable from Redis and survive server restarts.
3.  **Self-Healing:** A background service monitors MCP container health and cleans up orphaned resources.

## 2. Current State Analysis
*   **LLM Providers:** `gemini.go` supports tool calling. `openai.go` and `claude.go` are currently text-only.
*   **Registry:** `SaveConfig` is implemented, but `GetConfigs` is a placeholder. 
*   **Resource Management:** No automated cleanup or health-check logic currently exists in the Go implementation.

## 3. Step-by-Step Strategic Roadmap

### Phase 1: Tool Calling Parity (`internal/mcp/llm`)
1.  **OpenAI Tool Support:**
    *   Update `openAIRequest` to include the `tools` field.
    *   Update the stream parser to detect `tool_calls` in the delta chunks.
    *   Map OpenAI tool call format to `domain.ToolCall`.
2.  **Claude Tool Support:**
    *   Update `claudeRequest` to include the `tools` field.
    *   Implement parsing for `tool_use` events in the Anthropic SSE stream.
    *   Map Anthropic tool call format to `domain.ToolCall`.

### Phase 2: Registry Persistence (`internal/mcp`)
1.  **Redis SCAN Implementation:** Implement `GetConfigs` in `registry.go` using Redis keys/scanning.
2.  **Auto-Initialization:** Update `main.go` or a new `Manager` service to scan the registry on startup and ensure required Docker containers are running.

### Phase 3: Health & Stability (`internal/infrastructure/docker`)
1.  **Health Check Endpoint:** Implement `GET /api/mcp/health` to report status of all active tool containers.
2.  **Janitor Service:**
    *   Create `backend/internal/services/janitor.go`.
    *   Periodically verify that `docker.sock` still sees the managed containers.
    *   Clean up `HijackedResponse` sessions that haven't sent activity in X minutes.

## 4. Verification & Testing Plan
*   **Multi-Model Tool Test:** 
    *   Switch to OpenAI -> Ask "List files" -> Verify tool call is triggered.
    *   Switch to Claude -> Ask "List files" -> Verify tool call is triggered.
*   **Persistence Test:** Save a tool -> Restart backend -> Verify tool is still listed in `/api/mcp/tools`.
*   **Health Test:** Manually kill an MCP container via CLI -> Verify the backend health API reflects the failure.

## 5. Risk Assessment
*   **Stream Complexity:** OpenAI tool-call deltas can be fragmented (e.g., arguments arrive in multiple small JSON chunks).
    *   *Mitigation:* Implement an accumulator logic in the provider to buffer tool arguments before emitting the `domain.Chunk`.
*   **API Breaking Changes:** LLM providers frequently update their tool-calling schemas.
    *   *Mitigation:* Keep provider implementations modular and focused on raw API mapping to local domain models.
