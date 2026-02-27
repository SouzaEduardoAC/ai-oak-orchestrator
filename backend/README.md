# AI Oak Orchestrator (Backend)

A high-performance orchestration engine for LLMs and Model Context Protocol (MCP) tools. This is the **backend component** of the AI Oak Orchestrator monorepo.

## Features
- **Fast & Lightweight:** Built with Golang for low latency and efficient stream handling.
- **MCP Native:** Host implementation of the Model Context Protocol.
- **Docker Integration:** Seamlessly executes tools inside isolated containers.
- **Real-time:** Bidirectional WebSocket communication for streaming and user-in-the-loop approvals.
- **Conversation Persistence:** Full multi-turn session memory backed by Valkey.
- **Dynamic Model Selection:** Automatically discovers available models for your API key and allows switching on-the-fly.
- **Self-Healing Infrastructure:** Janitor service automatically prunes orphaned containers and maintains tool health.
- **Secure:** Integrated with Keycloak for modern identity management.

## Getting Started

### Prerequisites
- Go 1.22+
- Docker Engine
- Valkey

### Setup
1. Ensure you are in the `backend` directory.
2. The `.env` file is typically managed at the root of the monorepo (see `../.env.example` for details).
3. Install dependencies:
   ```bash
   go mod tidy
   ```

### Running
```bash
go run cmd/server/main.go
```
*Note: It is recommended to run the full stack from the root directory using `docker compose up`.*

## Documentation
For deeper dives, see the `docs/` folder:
- [AI Context](./docs/ai_context.md)
- [Business Flow](./docs/business_flow.md)
- [Technical Documentation](./docs/technical_documentation.md)

## License
MIT
