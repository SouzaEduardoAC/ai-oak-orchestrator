# Technical Blueprint: Server-Side Migration (Node.js to Golang)

## 1. Executive Summary
**Objective:** Port the `mcp-orchestrator` from Node.js/TypeScript to a high-performance Golang backend. The system must maintain full functional parity with the existing MVP while improving concurrency management, type safety, and integration with the standalone Vue.js frontend and Keycloak authentication.

**Definition of Done:**
*   **Performance:** Go backend handles multiple concurrent agent sessions with lower memory overhead than Node.js.
*   **Parity:** All MCP tool discovery, execution, and LLM streaming functionalities work as before.
*   **Protocol:** Standard JSON-over-WebSocket protocol established between Go and Vue frontend.
*   **Security:** Every request/socket connection is validated against Keycloak JWTs.
*   **Infrastructure:** Service is fully containerized and managed via `docker-compose`.

## 2. Current State Analysis
The investigation of `~/Projects/mcp-orchestrator` reveals a modular TypeScript architecture:
*   **Transport Layer:** `StdioDockerTransport.ts` uses `dockerode` to stream stdin/stdout.
*   **Agent Logic:** `MCPAgent.ts` manages a complex state machine (Thinking -> Tool Call -> Approval -> Result).
*   **Registry:** `ConfigStore.ts` and `MCPRegistry.ts` manage tool definitions.
*   **Frontend Integration:** The Vue app in `ai-web-manager` uses `useWebSocket.ts`.

## 3. Detailed Strategic Roadmap

### Phase 1: Foundation & Infrastructure Scaffold
1.  **Module Initialization:** 
    *   Command: `go mod init github.com/ecoza/ai-oak-orchestrator`
    *   Dependencies: `echo`, `viper`, `zap`, `golang-jwt`.
2.  **Environment Configuration:** 
    *   Use `viper` to handle `.env` and `config.yaml`.
    *   Required Keys: `LLM_API_KEY`, `DOCKER_HOST`, `KEYCLOAK_URL`, `REDIS_URL`.
3.  **Logging:** 
    *   Implement structured logging with `uber-go/zap`.
4.  **Domain Models:** 
    *   Port TypeScript interfaces to Go structs in `/internal/domain`.
    *   Include: `Session`, `Message`, `MCPTool`, `ToolCall`, `ToolResult`.

### Phase 2: Docker & Persistence Layer
1.  **Docker SDK Integration:**
    *   Package: `github.com/docker/docker/client`.
    *   Implement `ContainerManager` for image pulling, container creation, and exec management.
2.  **Redis Repository:**
    *   Package: `github.com/redis/go-redis/v9`.
    *   Implement `ConversationRepository` for chat history.
    *   Implement `SessionRepository` for volatile state.

### Phase 3: MCP Protocol Implementation (The Host)
1.  **JSON-RPC 2.0 Core:**
    *   Implement a generic JSON-RPC client in `/internal/mcp/rpc`.
    *   Features: Request ID mapping and asynchronous response matching using Go channels.
2.  **Stdio Transport:**
    *   Implement `io.ReadWriter` wrapper for `docker.ContainerExecAttach`.
    *   Use goroutines to pump `stdout` to a "Receive" channel and `stdin` from a "Send" channel.
3.  **Tool Discovery:**
    *   Port `list_tools` and `call_tool` logic.
    *   Add automated cleanup for stagnant MCP containers.

### Phase 4: LLM Provider Layer
1.  **Provider Interface:**
    *   `type LLMProvider interface { GenerateStream(ctx context.Context, prompt string, tools []domain.Tool) (<-chan domain.Chunk, error) }`
2.  **Implementation Adapters:**
    *   **Gemini:** Use `github.com/google/generative-ai-go`.
    *   **OpenAI/Claude:** Use standard HTTP clients or official Go SDKs.
3.  **Tool Call Parsing:**
    *   Implement regex or JSON-based parsing of LLM chunks to detect tool execution intents.

### Phase 5: The Agent State Machine
1.  **Agent Orchestrator:**
    *   Location: `/internal/agent/orchestrator.go`.
    *   Manage the recursive "Think-Act-Observe" loop.
2.  **Approval Queue:**
    *   Implement a thread-safe "Pause" mechanism.
    *   Wait for WebSocket signal (`tool:approved`) before executing high-risk tools.

### Phase 6: Real-time Communication (WebSocket Hub)
1.  **Echo + Gorilla:**
    *   Setup `labstack/echo` for REST.
    *   Integrate `github.com/gorilla/websocket` for the Hub.
2.  **The Hub Pattern:**
    *   Centralized registry of active client connections.
    *   Routing: `Broadcast`, `SendToUser`, `ReceiveFromUser`.
3.  **Protocol Parity:**
    *   Match frontend expectations: `{"event": "agent:thinking", "data": {...}}`.

### Phase 7: Security & Auth (Keycloak)
1.  **JWT Middleware:**
    *   Fetch Keycloak JWKS for public key validation.
    *   Validate `Authorization: Bearer <token>` on all entry points.
2.  **Claims Extraction:**
    *   Inject user identity into `context.Context`.

## 4. Verification & Testing Plan

### Automated Testing
*   **Unit Tests:**
    *   `mcp/rpc`: Mock `io.ReadWriter` to verify JSON-RPC state.
    *   `llm/parser`: Test various LLM formats for tool call extraction.
*   **Integration Tests:**
    *   `docker/integration`: Verify Go can start and communicate with a test MCP container.
    *   `redis/integration`: Verify session persistence.

### Manual Parity Checklist
1.  **Auth:** Frontend redirects to Keycloak; backend rejects invalid tokens.
2.  **Discovery:** UI populates tool list from Go backend.
3.  **Streaming:** Tokens stream word-by-word into the Vue chat.
4.  **Cleanup:** Refreshing browser kills associated Docker exec sessions.

## 5. Risk Assessment & Mitigation

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| **Goroutine Leakage** | High | Use `context.WithCancel` for all long-running sessions. |
| **Stdio Deadlock** | Medium | Use buffered channels and `select` with timeouts for all IO. |
| **Protocol Mismatch** | Medium | Maintain shared JSON Schema definitions for WS messages. |
| **Socket Permissions** | Low | Validate Docker GID in the Go container runtime. |
