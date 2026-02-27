# AI Oak Orchestrator (Frontend): The Oak UI

The frontend management console for **The Oak** AI Orchestrator. This modular Vue 3 SPA allows users to interact with AI agents, manage protocol server registries, and approve sensitive tool executions. This is the **frontend component** of the AI Oak Orchestrator monorepo.

## 📁 Project Structure

- **`/src`**: The core Vue 3 + TypeScript source code.
- **`/public`**: Static assets.
- **`/docs`**: System documentation and architectural details.

## 📖 Documentation

- **[Technical Documentation](./docs/technical_documentation.md)**: Deep dive into the multi-theme architecture, auth, and protocols.
- **[AI Context](./docs/ai_context.md)**: Essential reading for AI agents or developers new to the project.

## 🚀 Getting Started

### Prerequisites
- Node.js (v18+)
- AI Orchestrator Backend (running)

### Installation
1. Ensure you are in the `frontend` directory.
2. Install dependencies:
```bash
npm install
```

### Local Development
```bash
# The .env file is typically managed at the root of the monorepo (see ../.env.example)
# You can override variables here if needed:
# cp .env.example .env

# Start the dev server
npm run dev
```

### Running with Docker
It is recommended to run the full stack from the root directory using `docker compose up`. If you need to run just the frontend container:
```bash
docker compose up -d --build web-interface
```
The interface will be available at `http://localhost:5173`.

## 🛠 Features
- **Real-time Agent Chat**: Powered by native WebSockets.
- **Multi-Transport Registry**: Support for Stdio, SSE, HTTP, and Docker containers.
- **HITL Approvals**: Sequential queue for human-in-the-loop tool execution.
- **Multi-Theme Engine**: Dynamic switching between **Oak** and **MCP** visual identities.
- **Modular Auth**: Keycloak integration with a local bypass for development.
