# AI Context: The Oak - Web Manager

## Project Intent
This project is the frontend command-and-control center for **The Oak**. Its primary purpose is to provide a user-friendly interface for human-in-the-loop tool approvals, real-time agent communication, and dynamic protocol server registry management. This is the **frontend component** of the larger AI Oak Orchestrator monorepo.

## Core Domain Logic
1. **Human-in-the-Loop (HITL):** The system often requires explicit user approval before executing tools. The frontend must handle these requests sequentially, even if the backend processes them in parallel.
2. **Protocol Abstraction:** The interface abstracts different transport protocols (Stdio, Docker, HTTP, SSE) into a unified configuration and monitoring view.
3. **Session Persistence:** State is managed via Pinia to ensure that chat history and connection status remain stable during navigation.

## Development Constraints for AI Agents
- **Type Safety:** Always prefer strict TypeScript interfaces for API and WebSocket payloads.
- **Styling:** Adhere to the **"Oak Theme"** (Nature-inspired) defined in `tailwind.config.js`. Use utility classes like `bg-oak-primary` or `text-oak-accent`.
- **Composition API:** Use `<script setup lang="ts">` for all new Vue components.
- **Immutable Contracts:** Public APIs from the backend are fixed. Match the `WsMessage` protocol exactly.

## Integration Points
- **Backend:** Golang-based AI Orchestrator.
- **WebSocket:** `ws://[host]/ws` for real-time events.
- **REST:** `/api/models/available`, `/api/mcp/health`, `/api/mcp/add`, etc.
- **Auth:** Keycloak OIDC flow.
- **Containerization:** `Dockerfile` and `docker-compose.yml` for standardized deployment.