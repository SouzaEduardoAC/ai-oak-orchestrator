# AI Oak Orchestrator

## Overview
A high-performance orchestration platform bridging Large Language Models (LLMs) with containerized tools via the **Model Context Protocol (MCP)**. Featuring Human-in-the-Loop (HITL) approvals and a custom theme engine.

## Monorepo Structure
- **`backend/`**: Go orchestration service using Echo and Valkey.
- **`frontend/`**: Vue 3 management interface served via Nginx.
- **`docs/`**: Authoritative system documentation.

## Quick Start
```bash
# Using existing external containers (Valkey, Keycloak, Postgres)
docker compose -f local-compose-env.yml up -d --build
```

## Local Setup
1. Copy `.env.example` to `.env`.
2. Ensure your `GEMINI_API_KEY` is set in your host environment.
3. Access the UI at `http://localhost:5173`.

## Authoritative Documentation
- [Business Flow](./docs/business_flow.md): Logic paths and diagrams.
- [Technical Specifications](./docs/technical_specifications.md): Entry points and error handling.
- [AI Context](./docs/ai_context.md): Machine-readable technical metadata.

---
*Last Synced: 2026-02-27*
