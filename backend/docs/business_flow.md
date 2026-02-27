# Business Flow: AI Oak Orchestrator

## 1. User Authentication
- **Step:** User opens the UI.
- **Action:** Redirected to Keycloak for OIDC login.
- **Outcome:** UI receives a JWT, which is sent in the `Authorization: Bearer` header for all backend requests.

## 2. Dynamic Model Selection
- **Discovery:** UI calls `GET /api/models/available` to retrieve available models allowed by the API Key.
- **Selection:** User selects a model from the dropdown.
- **Switching:** UI sends a `set_model` command via WebSocket.
- **Outcome:** The backend updates the active model for the provider session thread-safely.

## 3. MCP Tool Management
- **Step:** Administrator adds a new MCP Tool.
- **Action:** REST POST to `/api/mcp/tools` with Docker image and configuration.
- **Outcome:** Backend persists config to Valkey and optionally pulls the Docker image.

## 4. Real-time Interaction (The Chat Flow)
- **Initiation:** User connects via WebSocket and sends a `message` directive.
- **Agent Initialization:** 
    - Orchestrator retrieves the conversation history from Valkey using the session ID.
    - Orchestrator retrieves active tools from the Registry.
    - Orchestrator initializes/connects to required MCP Docker containers (reusing existing ones if available).
- **The Thinking Loop:**
    1. **Signal:** Backend sends `agent:thinking` to notify the UI.
    2. **Prompting:** Full history (loaded from Valkey) + new message + tools sent to LLM.
    3. **Streaming:** Tokens streamed to UI as `agent:response` in real-time.
    4. **Tool Detection:** LLM requests a tool call.
    5. **Approval Request:** Backend pauses and sends `tool:approval_required` to the User.
    6. **User Decision:** User clicks "Approve" or "Reject" in UI, sending an `approval` command.
    7. **Execution:** If approved, Backend executes tool in Docker and sends result to UI via `tool:output` before feeding it back to the LLM.
- **Conclusion:** LLM provides a final answer. The updated history is persisted back to Valkey.

## 5. Resource Cleanup
- **Trigger:** Browser window closes or session expires.
- **Action:** Backend cancels the session context.
- **Outcome:** Active Docker exec sessions are terminated, and goroutines are reaped.
