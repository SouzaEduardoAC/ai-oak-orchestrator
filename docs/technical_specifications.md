# Technical Specifications: AI Oak Orchestrator

## 1. Entry Points
- **Backend Service**: `main.go` initializes the Echo (v4) web server on the configured `SERVER_PORT`.
- **WebSocket Gateway**: `hub.HandleWebSocket` (Ref: `internal/api/websocket/hub.go`) serves as the bidirectional communication channel for real-time agent interaction.
- **REST API Registry**: `MCPHandler.RegisterRoutes` (Ref: `internal/api/mcp.go`) exposes endpoints for tool configuration and health management.
- **Frontend SPA**: `main.ts` initializes the Vue 3 application, mounting the root component and connecting to Pinia/Vue Router.

## 2. State Changes
- **Session Context**: The `Orchestrator` maintains message history in memory during a reasoning run, which is periodically synced with **Valkey** (Ref: `internal/infrastructure/valkey`).
- **Tool Lifecycle**: The `ToolManager` performs `CreateContainer` and `StartContainer` calls to the Docker Engine (Ref: `internal/infrastructure/docker/manager.go`), maintaining a singleton instance of each active tool.
- **Frontend State**: Pinia stores (`chat.ts`, `mcp.ts`) manage the reactive state of messages, tool approval queues, and real-time connectivity status.

## 3. Error Matrix
| Failure Scenario | Mitigation Strategy | Ref |
| :--- | :--- | :--- |
| **Tool Execution Timeout** | Context-aware `context.WithTimeout` on JSON-RPC calls. | `internal/mcp/rpc/client.go` |
| **Docker Engine Unreachable** | Service start-up failure + Health check reporting. | `main.go` |
| **WebSocket Disconnect** | Frontend exponential backoff reconnection logic. | `useWebSocket.ts` |
| **LLM Rate Limit** | Echo middleware and application-level retry-on-error. | `internal/agent/orchestrator.go` |

## 4. Complexity & Performance
- **Agent Reasoning**: The loop is recursive and `O(N)` where N is the number of thinking steps required by the LLM. It is I/O-bound.
- **Tool Discovery**: `ListTools` iterates through the active tool map, maintaining efficient access via hash maps.
- **Docker I/O**: Bidirectional streams for Stdio use non-blocking reads and goroutines to prevent main-thread blockage during high-volume tool output.
