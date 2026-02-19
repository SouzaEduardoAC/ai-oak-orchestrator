# Technical Documentation: AI Oak Orchestrator

## System Architecture
The backend is a decoupled Go service designed for horizontal scalability (at the API level) and efficient resource management (at the tool host level).

### 1. Transport Layer (`internal/mcp/transport`)
- **Stdio Implementation:** Wraps an `io.Reader` and `io.Writer`. It uses a `bufio.Scanner` to read newline-delimited JSON messages.
- **Concurrency:** Thread-safe writing via `sync.Mutex`. Listening happens in a dedicated goroutine.

### 2. JSON-RPC Client (`internal/mcp/rpc`)
- Implements JSON-RPC 2.0.
- **Asynchrony:** Uses a `map[string]chan *Response` to correlate responses with requests based on the `id` field.
- **Timeouts:** All calls respect the `context.Context` timeout.

### 3. Agent Orchestrator (`internal/agent`)
- The `Run` method executes the core logic.
- **Thread Safety:** Uses a `sync.RWMutex` to allow safe switching of LLM models mid-session.
- **Event Bus:** Communicates with the API layer via two channels:
    - `output chan AgentEvent`: For outgoing tokens, logs, and approval requests.
    - `input chan AgentCommand`: For incoming user decisions (approve/reject) and commands (set_model).

### 4. Docker Integration (`internal/infrastructure/docker`)
- Uses the official Docker Engine API.
- **Security:** Containers are spawned with limited resource constraints.
- **Communication:** Uses `ContainerExecAttach` to get a bidirectional stream to the process inside the container.

### 5. API Reference

#### REST API
| Method | Path | Description | Auth Required |
| :--- | :--- | :--- | :--- |
| GET | `/health` | Service health status | No |
| GET | `/api/mcp/tools` | List configured MCP tools | Yes |
| POST | `/api/mcp/tools` | Register a new MCP tool | Yes |
| DELETE | `/api/mcp/tools/:name` | Remove an MCP tool and stop its container | Yes |
| GET | `/api/mcp/health` | Get aggregate health of tool ecosystem | Yes |
| GET | `/api/models/available` | List available models for active provider | Yes |

#### WebSocket API (`/ws`)
**Incoming Messages:**
- `{"type": "message", "payload": "User message"}`
- `{"type": "approval", "payload": {"callId": "...", "approved": true}}`
- `{"type": "set_model", "payload": "model-name"}`

**Outgoing Events:**
- `{"type": "agent:response", "payload": "Partial word"}`
- `{"type": "tool:approval_required", "payload": {...tool details...}}`
- `{"type": "agent:log", "payload": "Tool progress or log"}`
- `{"type": "agent:error", "payload": "Error message"}`

## Infrastructure Details
### Docker Container Naming
All containers spawned by the orchestrator follow the deterministic naming pattern: `ai-oak-mcp-[toolname]`. This allows the server to adopt existing containers upon restart or shared tool usage.

## Environment Variables
- `SERVER_PORT`: Default `8080`.
- `LLM_API_KEY`: Required for Gemini/OpenAI.
- `REDIS_URL`: Location of the Redis instance.
- `KEYCLOAK_JWKS_URL`: URL to fetch public keys for JWT validation.
