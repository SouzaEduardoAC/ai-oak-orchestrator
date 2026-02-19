# AI Context: AI Oak Orchestrator (Go Backend)

## Project Overview
AI Oak Orchestrator is a high-performance Golang backend designed to bridge Large Language Models (LLMs) with local/remote tools via the **Model Context Protocol (MCP)**. It acts as the "brain" of the ecosystem, managing conversation state, tool execution lifecycle in Docker containers, and real-time streaming to frontends.

## Core Architecture
- **Language:** Go 1.22+
- **HTTP Framework:** Echo (v4)
- **Real-time:** Gorilla WebSocket
- **Auth:** Keycloak (OIDC/JWKS)
- **Protocol:** MCP (JSON-RPC 2.0 over Stdio/Docker Exec)
- **Infrastructure:** Docker SDK for containerized tool hosts, Redis for session/config persistence.

## Key Components for AI Interaction
1. **Agent Orchestrator (`internal/agent`)**: Implements the recursive reasoning loop. It manages the transition between thinking, requesting tool approvals, and processing tool results. It depends on the `ToolProvider` interface for decoupling.
2. **MCP Host (`internal/mcp`)**: Implementation of the MCP client side. It communicates with tools via JSON-RPC. Includes the `ToolManager` which manages container lifecycles.
3. **Docker Manager (`internal/infrastructure/docker`)**: Handles the lifecycle of tool-providing containers.
4. **WebSocket Hub (`internal/api/websocket`)**: Manages bidirectional communication with the frontend and session persistence via the `RedisStore` interface.

## Critical Patterns
- **Concurrency:** Uses goroutines and channels heavily for pumping Docker Stdio streams and LLM response tokens.
- **Context:** `context.Context` is used for all operations to ensure clean teardown of agent runs and Docker exec sessions.
- **Typing:** Strict domain models in `internal/domain` must be respected to maintain parity with the frontend.
- **Testability:** Core services depend on interfaces (`LLMProvider`, `ToolProvider`, `RedisStore`) to allow mocking in unit and integration tests.

## Guidelines for AI Modifications
- When adding new LLM providers, implement the `Provider` interface in `internal/mcp/llm`, including `ListModels` and `SetModel`.
- New MCP features should be added to the `Client` in `internal/mcp/client.go`.
- Always ensure JSON-RPC request/response consistency.
- Use the `Orchestrator.mu` mutex when accessing shared provider state in the agent loop.
