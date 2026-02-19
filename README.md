# AI Oak Orchestrator (Go Backend)

A high-performance orchestration engine for LLMs and Model Context Protocol (MCP) tools.

## Features
- **Fast & Lightweight:** Built with Golang for low latency and efficient stream handling.
- **MCP Native:** Host implementation of the Model Context Protocol.
- **Docker Integration:** Seamlessly executes tools inside isolated containers.
- **Real-time:** Bidirectional WebSocket communication for streaming and user-in-the-loop approvals.
- **Conversation Persistence:** Full multi-turn session memory backed by Redis.
- **Dynamic Model Selection:** Automatically discovers available models for your API key and allows switching on-the-fly.
- **Self-Healing Infrastructure:** Janitor service automatically prunes orphaned containers and maintains tool health.
- **Secure:** Integrated with Keycloak for modern identity management.

## Getting Started

### Prerequisites
- Go 1.22+
- Docker Engine
- Redis

### Setup
1. Clone the repository.
2. Create a `.env` file in the root:
   ```env
   LLM_PROVIDER=gemini
   LLM_API_KEY=your_key_here
   LLM_MODEL=gemini-1.5-pro
   REDIS_URL=redis://localhost:6379
   KEYCLOAK_JWKS_URL=http://keycloak:8080/realms/master/protocol/openid-connect/certs
   ```
3. Install dependencies:
   ```bash
   go mod tidy
   ```

### Running
```bash
go run cmd/server/main.go
```

## Documentation
For deeper dives, see the `docs/` folder:
- [AI Context](./docs/ai_context.md)
- [Business Flow](./docs/business_flow.md)
- [Technical Documentation](./docs/technical_documentation.md)

## License
MIT
