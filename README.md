# AI Oak Orchestrator

## Overview
A high-performance orchestration platform that bridges Large Language Models (LLMs) with isolated tools via the **Model Context Protocol (MCP)**. This project implements a recursive reasoning loop with integrated **Human-in-the-Loop (HITL)** approvals, ensuring all autonomous actions remain under user oversight.

## Project Structure
- **`backend/`**: Go-based orchestration engine.
- **`frontend/`**: Vue 3 + TypeScript web interface.
- **`docs/`**: Comprehensive system documentation.

## Quick Start
```bash
# Run the entire stack with Docker Compose
docker compose up
```

## Local Development
To run the orchestrator and web interface while using existing external dependencies (Valkey, Keycloak, etc.):
1. Create a `.env` file from `.env.example`.
2. Run using the dedicated local compose file:
   ```bash
   docker compose -f local-compose-env.yml up
   ```

### LLM Authentication
The **Gemini/Google AI** provider is designed for flexibility:
- **API Key**: Set `LLM_API_KEY` in your `.env` for standard key-based authentication.
- **Application Default Credentials (ADC)**: Leave `LLM_API_KEY` empty to use your existing environment authentication (e.g., from the Gemini CLI or `gcloud`).

## Documentation
For deep dives, see the full documentation suite:
- [Business Flow](./docs/business_flow.md): High-level logic and Mermaid diagrams.
- [Technical Specifications](./docs/technical_specifications.md): Implementation details, entry points, and error matrices.
- [AI Context](./docs/ai_context.md): Machine-readable summary for AI agents and developer onboarding.

---
*Last Synced: 2026-02-27*
