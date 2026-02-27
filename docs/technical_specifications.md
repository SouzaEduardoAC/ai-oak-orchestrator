# Technical Specifications: AI Oak Orchestrator

## 1. Entry Points
- **Backend Core**: `backend/cmd/server/main.go` - Initializes Echo server, Valkey infrastructure (RESP2), and LLM providers.
- **Frontend Gateway**: `frontend/nginx.conf` - Acts as the primary entry point for browsers, proxying `/api` and `/ws` internally to the orchestrator.
- **UI Store**: `frontend/src/stores/chat.ts` - Manages reactive state for the human-in-the-loop reasoning loop.

## 2. State Changes
- **Valkey Persistence**: Session history is stored in Valkey using standard key-value pairs. Authentication uses explicit password handling to bypass handshake ambiguities.
- **Docker Tooling**: The `ToolManager` dynamically spins up containers for MCP tools, mapped via the internal Docker bridge network.
- **Real-time Synchronization**: The backend `Hub` maintains WebSocket client maps and broadcasts tokens/errors using non-blocking goroutines.

## 3. Error Matrix
| Scenario | Detection | Mitigation / UX |
| :--- | :--- | :--- |
| **API Quota Hit (429)** | `GeminiProvider` detects error in stream. | Propagates `agent:error` to UI for user visibility. |
| **Model Not Found (404)** | Internal validation in `GenerateStream`. | Automatic `models/` prefixing and explicit logging. |
| **Valkey NOAUTH** | `NewClient` initialization. | Explicit `RESP2` protocol and password trimming. |
| **WS Disconnect** | `useWebSocket.ts` onerror/onclose. | Exponential backoff reconnect strategy. |

## 4. Complexity & Performance
- **Prompt Synthesis**: O(M) where M is total messages in session. Session history is pruned/synced asynchronously.
- **WebSocket I/O**: High-concurrency handling using Echo's middleware and Go's native goroutines for each session loop.
- **Network Latency**: Minimized via Nginx internal proxying (`orchestrator:8080`) within the Docker network.
