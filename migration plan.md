# Architecture Migration Plan: MCP Orchestrator (Go + Vue + Keycloak)

## 1. Executive Summary

**Objective:** Port the existing Node.js/Express MCP Orchestrator to a decoupled architecture consisting of a **Golang backend** and a standalone **Vue.js frontend**. Introduce **Keycloak** for robust authentication and authorization.

**Definition of Done:**
1.  **Backend:** fully ported to Go, exposing REST APIs for management and WebSockets for real-time chat.
2.  **Frontend:** Vue 3 (Vite-based) SPA replacing the single HTML file, fully integrated with the Go backend.
3.  **Auth:** Keycloak running in Docker; Frontend requires login; Backend validates JWT tokens.
4.  **Parity:** All existing features (Chat, MCP Management, Docker spawning, LLM integration) functioning identically.
5.  **Infrastructure:** Updated `docker-compose.yml` orchestrating all services.

## 2. Current State Analysis & Findings

*   **Architecture:** Monolithic Node.js app serving static assets and handling logic.
*   **Communication:**
    *   **REST:** `/api/mcp/*` for configuration and `/api/models/*` for diagnostics.
    *   **Socket.IO:** Primary channel for Chat (`message`, `agent:thinking`) and Tool Approval (`tool:approval`).
*   **State:** Redis used for session and conversation persistence.
*   **Infrastructure:** Depends on `var/run/docker.sock` to spawn sibling containers (`mcp-server`).
*   **Frontend:** A single `index.html` with embedded Vue 3 CDN and Socket.IO client.
*   **Complexity:**
    *   High complexity in `MCPAgent.ts` and `SocketRegistry.ts` (orchestration logic).
    *   Low complexity in REST endpoints (CRUD).

## 3. Strategic Roadmap

### Phase 1: Infrastructure & Scaffold
*   **Goal:** Establish the new environment and services.
*   **Actions:**
    1.  **Directory Structure:** Create `backend/` (Go) and `frontend/` (Vue) directories.
    2.  **Docker Compose:** Update `docker-compose.yml` to include:
        *   `postgres` (for Keycloak).
        *   `keycloak` (Auth service).
        *   `redis` (Existing).
        *   `backend` (Go).
        *   `frontend` (Nginx/Dev server).
    3.  **Keycloak Config:** Create a realm export or startup script to bootstrap the `mcp-orchestrator` realm and client.

### Phase 2: Backend Core & API (Go)
*   **Goal:** Functional Go server with Auth and REST endpoints.
*   **Tech Stack:** Go 1.22+, Echo (Framework), `golang-jwt` (Auth).
*   **Actions:**
    1.  **Init Module:** `go mod init mcp-orchestrator`.
    2.  **Auth Middleware:** Implement JWT validation middleware checking against Keycloak's public key/introspection endpoint.
    3.  **REST Port:** Port `/api/mcp` endpoints (Add/Remove/List/Health).
        *   *Adaptation:* Use `go-redis` for the `ConfigStore` equivalent.
    4.  **LLM Clients:** Implement the `LLMProvider` interface in Go (Gemini, Anthropic, OpenAI) using standard HTTP clients or official Go SDKs where available (e.g., `google-generative-ai-go`).

### Phase 3: The Orchestrator Engine (Go)
*   **Goal:** Port the complex Docker and Agent logic.
*   **Tech Stack:** `docker/docker/client` (Official Go SDK), `gorilla/websocket`.
*   **Actions:**
    1.  **Docker Client:** Port `ContainerPool` and `DockerClient` logic. Go's Docker SDK is robust and typed.
    2.  **WebSocket Handler:** Replace `socket.io` with standard WebSockets (`gorilla/websocket`).
        *   *Architectural Change:* Define a strict JSON structure for WS messages (e.g., `type: "chat" | "approval"`, `payload: {...}`).
    3.  **Agent Logic:** Port `MCPAgent.ts` and `JanitorService.ts`.
        *   *Challenge:* Replicating the "Tool Approval" state machine in Go channels/goroutines.

### Phase 4: Frontend Modernization (Vue.js)
*   **Goal:** Robust SPA with Auth.
*   **Tech Stack:** Vue 3, Vite, Pinia (State), Tailwind, `keycloak-js`.
*   **Actions:**
    1.  **Scaffold:** `npm create vue@latest`.
    2.  **Auth:** Integrate `keycloak-js`. Protect routes; add Login/Logout buttons.
    3.  **Components:** Refactor the giant HTML file into:
        *   `components/ChatWindow.vue`
        *   `components/MCPManager.vue`
        *   `components/ToolApproval.vue`
    4.  **WebSocket Client:** Implement a native `WebSocket` client store to replace the `socket.io-client` logic, matching the new Go backend protocol.

### Phase 5: Integration & Verification
*   **Actions:**
    1.  **End-to-End Test:** Login -> Add MCP (Docker) -> Chat -> Approve Tool -> Verify Output.
    2.  **Docker Networking:** Ensure the Go backend can talk to the sibling containers it spawns.

## 4. Verification & Testing Plan

### Automated Tests
*   **Backend:** Unit tests for `Agent` logic and `DockerClient` mocking.
*   **Integration:** Test that the Go backend correctly parses LLM streams and triggers tool calls.

### Manual Verification Checklist
1.  **Auth:** Accessing `/` redirects to Keycloak.
2.  **MCP Spawning:** Adding a "Stdio-Docker" MCP results in a container starting (check `docker ps`).
3.  **Chat:** Send "List files" -> Agent asks approval -> Click Approve -> Agent lists files.
4.  **Persistence:** Restart Backend -> Chat history remains (Redis).

## 5. Risk Assessment

1.  **Socket.IO vs. WebSockets:**
    *   *Risk:* The current app relies heavily on Socket.IO's event convenience.
    *   *Mitigation:* We will move to standard WebSockets. This requires rewriting the client networking logic, but results in a more portable and "Go-idiomatic" backend.
2.  **LLM Stream Parsing:**
    *   *Risk:* Parsing raw text streams from LLMs to detect tool calls can be brittle.
    *   *Mitigation:* Use robust regex/parsing logic in Go, potentially leveraging existing libraries like `langchaingo` if they fit, or carefully porting the existing regex logic.
3.  **Docker-in-Docker Networking:**
    *   *Risk:* The Go app (in Docker) spawning sibling containers needs correct network attachment.
    *   *Mitigation:* Reuse the `app_network` logic from the existing `docker-compose.yml`. Ensure the Go app container mounts `docker.sock`.
