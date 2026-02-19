# Technical Blueprint: Final Polish & Protocol Alignment

## 1. Executive Summary
**Objective:** Finalize the Go backend migration by ensuring perfect protocol alignment with the `ai-web-manager` frontend and implementing the logic to dynamically manage and bootstrap MCP tool containers.

**Definition of Done:**
1.  **Protocol Sync:** Backend WebSocket events (`agent:thinking`, `agent:response`, `tool:approval_required`) match the Vue.js frontend expectations exactly.
2.  **Connection Manager:** A central service manages the lifecycle of MCP clients, mapping Registry configs to running Docker containers.
3.  **Auto-Bootstrap:** All tools registered in Redis are automatically started/warmed-up when the Go server initializes.

## 2. Current State Analysis & Findings
*   **Protocol Mismatch:** 
    *   Frontend expects `agent:response` for text and `tool:approval_required` for approvals.
    *   Backend currently uses `token` and `tool_approval_request`.
    *   Frontend calls `/api/models/available` for LLM discovery.
*   **Missing Orchestration:** 
    *   The backend can save tool configs but doesn't yet have the logic to "spawn on demand" or "spawn all on boot."
    *   The `Orchestrator` uses a static map of clients instead of a dynamic manager.

## 3. Step-by-Step Strategic Roadmap

### Phase 1: Protocol Alignment (`internal/api` & `internal/domain`)
1.  **Event Renaming:** Update `AgentEventType` in `domain/models.go`:
    *   `EventToken` -> `agent:response` (or `agent:token` if adding streaming support to UI).
    *   `EventApprovalRequest` -> `tool:approval_required`.
2.  **REST API Sync:**
    *   Update `LLMHandler` routes: `GET /api/llm/models` -> `GET /api/models/available`.
3.  **WebSocket Command Sync:**
    *   Update `Hub` to handle `type: "message"` for chat instead of `type: "chat"`.
    *   Handle `type: "approval"` for user decisions.

### Phase 2: MCP Connection Manager (`internal/mcp`)
1.  **Create Manager:** Implement `internal/mcp/manager.go`.
    *   Struct `ToolManager` holding references to `DockerManager`, `Registry`, and a cache of active `mcp.Client` instances.
2.  **Method: `EnsureTool(ctx, name)`:**
    *   Logic: Check cache -> Check Docker -> If missing, load from Registry -> Pull/Start Container -> Attach RPC -> Cache & Return.
3.  **Method: `InitializeAll(ctx)`:**
    *   Logic: Fetch all configs from Registry -> Start all active tools in parallel goroutines.

### Phase 3: Bootstrap & Wiring
1.  **Startup Logic:** Update `main.go` to initialize the `ToolManager` and call `InitializeAll`.
2.  **Orchestrator Update:** Update `agent.Orchestrator` to depend on `ToolManager.GetTools()` or similar dynamic discovery during the thinking loop.

## 4. Verification & Testing Plan
*   **Manual Integration:** Run the Go backend and the Vue frontend. 
    *   Verify the "Initializing" screen disappears once the model list is fetched.
    *   Verify chat messages trigger the "Processing" state in the UI.
*   **Reboot Test:** 
    *   Add a tool -> Stop backend -> Start backend.
    *   Verify `docker ps` shows the tool container is running without manual intervention.

## 5. Risk Assessment
*   **Payload Structure:** The frontend `tool:approval_required` handler expects specific sub-fields (`callId`, `name`, `args`).
    *   *Mitigation:* Double-check `domain.ToolCall` field names match the TypeScript interface in `ChatView.vue`.
*   **Concurrency:** Multiple users might trigger `EnsureTool` for the same tool simultaneously.
    *   *Mitigation:* Use a `sync.Map` or a mutex-protected map with "singleflight" pattern for container startup.
