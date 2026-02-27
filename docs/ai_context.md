# AI Context: AI Oak Orchestrator

## Machine-Readable Summary
- **Sync Date**: 2026-02-27
- **Architecture**: Monorepo (Go/Echo/Vite/Vue3/Nginx)
- **Primary Transport**: WebSocket (RESP2 for Valkey, JSON-RPC for MCP)
- **Auth Strategy**: Flexible (Keycloak OIDC or ADC/API Key for LLM)
- **Network Pattern**: Nginx Proxy (UI) -> Internal Docker Service (Backend)

## Dependency Map & Roles
### Backend (Go)
- `github.com/redis/go-redis/v9`: Valkey/Redis client (v9 chosen for explicit AUTH robustness).
- `github.com/labstack/echo/v4`: High-performance HTTP/WebSocket routing.
- `github.com/google/generative-ai-go/genai`: Google AI SDK for Gemini reasoning.
- `github.com/docker/docker/client`: Orchestrates MCP tool container lifecycle.

### Frontend (Vue 3 / TypeScript)
- `pinia`: Reactive global state for chat history and HITL queues.
- `tailwindcss`: Styled for "Oak" (earthy) and "MCP" (tech) themes.
- `useWebSocket.ts`: Custom composable handling relative-path WS proxying.

## Critical Constraints
- **Model Prefixing**: Gemini models *must* have `models/` or `tunedModels/` prefix before use in `GenerateStream`.
- **JSON Marshaling**: All WebSocket payloads must be explicitly marshaled via `json.Marshal` to ensure special characters don't break string boundaries.
- **Port Mapping**: Host port 3000 -> Backend 8080; Host port 5173 -> Nginx 80.
