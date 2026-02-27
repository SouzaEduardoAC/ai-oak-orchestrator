# AI Context: AI Oak Orchestrator

## Machine-Readable Summary
- **Project Name**: AI Oak Orchestrator
- **Architecture**: Go (Echo) / Vue 3 (Pinia) / Docker / Valkey
- **Role**: High-performance LLM-to-Tool middleware via Model Context Protocol (MCP).
- **Core Engine**: Recursive Orchestrator loop with real-time HITL approval gate.
- **Analysis Date**: 2026-02-27 (Sync Mode: Full Synthesis)
- **Last Synced Hash**: AUTO-GEN-2026

## Dependency Map
### Backend (Go)
- `github.com/labstack/echo/v4`: Core HTTP/Routing engine.
- `github.com/docker/docker/client`: Direct Docker Engine SDK for tool isolation.
- `github.com/valkey-io/valkey-go`: High-performance session/config storage.
- `go.uber.org/zap`: Structured, high-performance logging throughout the system.
- `internal/agent`: Orchestrator logic and recursive LLM/Tool reasoning loop.
- `internal/mcp`: Implementation of MCP (Model Context Protocol) client and tool registry.

### Frontend (Vue 3 / TypeScript)
- `vue`: Core UI framework (Composition API).
- `pinia`: State management for Chat, Auth, and MCP tool health.
- `vite`: Build tool and development server.
- `tailwindcss`: Utility-first CSS for "Oak" and "MCP" themes.
- `axios`: RESTful API interaction with the backend.

## Knowledge Constraints
- **HITL Enforcement**: All tool calls *must* pass through the `EventApprovalRequest` flow before execution.
- **MCP Version**: Supports JSON-RPC 2.0 over Stdio transport.
- **Session ID**: Required for every WebSocket connection to map history in Valkey.
