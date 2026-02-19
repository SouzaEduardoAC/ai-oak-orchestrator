# Business Flow: AI Oak Orchestrator

## 1. User Authentication
- **Step:** User opens the UI.
- **Action:** Redirected to Keycloak for OIDC login.
- **Outcome:** UI receives a JWT, which is sent in the `Authorization: Bearer` header for all backend requests.

## 2. Dynamic Model Selection
- **Discovery:** UI calls `GET /api/llm/models` to retrieve available models allowed by the API Key.
- **Selection:** User selects a model from the dropdown.
- **Switching:** UI sends a `set_model` command via WebSocket.
- **Outcome:** The backend updates the active model for the provider session thread-safely.

## 3. MCP Tool Management
- **Step:** Administrator adds a new MCP Tool.
- **Action:** REST POST to `/api/mcp/tools` with Docker image and configuration.
- **Outcome:** Backend persists config to Redis and optionally pulls the Docker image.

## 3. Real-time Interaction (The Chat Flow)
- **Initiation:** User connects via WebSocket and sends a "Chat" message.
- **Agent Initialization:** 
    - Orchestrator retrieves active tools from the Registry.
    - Orchestrator initializes/connects to required MCP Docker containers.
- **The Thinking Loop:**
    1. **Prompting:** Current history + tools sent to LLM.
    2. **Streaming:** Tokens streamed to UI in real-time.
    3. **Tool Detection:** LLM requests a tool call.
    4. **Approval Request:** Backend pauses and sends `tool_approval_request` to the User.
    5. **User Decision:** User clicks "Approve" or "Reject" in UI.
    6. **Execution:** If approved, Backend executes tool in Docker and feeds result back to LLM.
- **Conclusion:** LLM provides a final answer once tool data is processed.

## 4. Resource Cleanup
- **Trigger:** Browser window closes or session expires.
- **Action:** Backend cancels the session context.
- **Outcome:** Active Docker exec sessions are terminated, and goroutines are reaped.
