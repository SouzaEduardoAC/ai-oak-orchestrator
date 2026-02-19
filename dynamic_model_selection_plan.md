# Technical Blueprint: Dynamic Model Discovery & Selection

## 1. Executive Summary
**Objective:** Enhance the AI Oak Orchestrator to support dynamic discovery of available LLM models allowed by the configured API Key and enable real-time model switching from the Vue.js frontend.

**Definition of Done:**
1.  **Discovery API:** A REST API endpoint `GET /api/llm/models` that returns a list of available models for the currently active LLM provider (Gemini, OpenAI, or Claude).
2.  **Real-time Switching:** A WebSocket command `set_model` that allows users to change the active model for their session without restarting the server.
3.  **Parity:** Support for all three major providers (Gemini, OpenAI, Claude) and the Mock provider.

## 2. Current State Analysis
*   **Provider Interface:** Currently only supports `GenerateStream`. It lacks methods to list or change models.
*   **Orchestrator:** Holds a static reference to a provider. It needs to handle model state transitions safely.
*   **Frontend Store (`ai-web-manager`):** The `chat.ts` store is already designed to hold `availableModels` and `selectedModel`, but the backend currently provides no data to populate these.

## 3. Step-by-Step Strategic Roadmap

### Phase 1: Contract Expansion (`internal/mcp/llm`)
1.  **Interface Update:** Add `ListModels(ctx context.Context) ([]string, error)` and `SetModel(name string)` to the `Provider` interface in `provider.go`.
2.  **Gemini Implementation:** 
    *   Use `p.client.ListModels(ctx)` from the official SDK.
    *   Filter for models that support `generateContent`.
3.  **OpenAI Implementation:** 
    *   Implement an HTTP GET call to `https://api.openai.com/v1/models`.
    *   Filter for `gpt-*` models to keep the list clean.
4.  **Claude Implementation:** 
    *   Since Anthropic doesn't have a public "List Models" endpoint, provide a curated list of Claude 3.5 and 3.0 identifiers.
5.  **Mock Implementation:** 
    *   Return `["mock-model-1", "mock-model-2"]` for UI testing.

### Phase 2: Model Discovery API (`internal/api`)
1.  **LLM Handler:** Create `internal/api/llm.go` to house LLM-specific REST endpoints.
2.  **Discovery Endpoint:** Implement `GET /api/llm/models`.
    *   Logic: Calls the underlying provider's `ListModels` method.
    *   Response Format: `[{"id": "model-id", "name": "Friendly Name"}]`.
3.  **Registration:** Wire the handler in `backend/cmd/server/main.go`.

### Phase 3: Real-time Model Switching (`internal/api/websocket` & `internal/agent`)
1.  **Command Type:** Update `domain/models.go` to include `CommandSetModel AgentCommandType = "set_model"`.
2.  **Hub Integration:** 
    *   In `HandleWebSocket`, detect the `set_model` message.
    *   Route the payload (model name) to the active agent run's input channel.
3.  **Orchestrator Logic:**
    *   Add a `mu sync.RWMutex` to the `Orchestrator` to protect the provider model state.
    *   In the `Run` loop, process the `CommandSetModel` signal and call `provider.SetModel()`.

## 4. Verification & Testing Plan
*   **REST Discovery:** Curl `/api/llm/models` and verify the JSON array matches the provider's capabilities.
*   **WebSocket Switching:** 
    *   Send `{"type": "set_model", "payload": "gpt-3.5-turbo"}`.
    *   Verify via backend logs that subsequent `chat` messages use the new model name.
*   **Error Handling:** Verify that selecting an invalid model name returns an `error` event via WebSocket.

## 5. Risk Assessment
*   **Shared State:** The current singleton `Orchestrator` means a model change affects the global provider. 
    *   *Mitigation:* Document as MVP behavior; move to session-scoped providers in the next iteration.
*   **Rate Limiting:** Frequent polling of model lists.
    *   *Mitigation:* Implement a 5-minute memory cache for the model list results.
