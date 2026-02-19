# Architecture Blueprint: Production Readiness & Final Refinement

## 1. Executive Summary
**Objective:** Close the remaining architectural gaps identified in the `stability_phase_completion_report.md` to ensure the Go backend is operational, secure, and self-healing before full-scale integration with the frontend.

**Definition of Done:**
1.  **Lifecycle Management:** The `JanitorService` actively prunes orphaned MCP containers and monitors health.
2.  **Security Alignment:** All REST endpoints are correctly grouped under JWT-protected middleware.
3.  **Protocol Completeness:** The JSON-RPC client supports "Notifications" (one-way messages) as required by the MCP specification.
4.  **Ecosystem Orchestration:** A root-level `docker-compose.yml` is provided to spin up the entire stack (Go + Redis + Keycloak + Vue).

## 2. Current State Analysis & Findings
*   **Infrastructure (`janitor.go`):** Found skeleton code with placeholders for Docker container listing. We must implement label-based filtering to avoid killing non-system containers.
*   **API (`main.go`):** The `MCPHandler` currently registers routes directly on the Echo instance instead of using the `apiGroup`, causing a security bypass for tool management.
*   **Protocol (`rpc/client.go`):** The `handleIncoming` method silently ignores messages without IDs. MCP servers often send notifications (e.g., progress updates) that we need to capture.
*   **Deployment:** The project is modularized into `backend/`, but lacks a unified orchestration layer to link with `~/Projects/ai-web-manager`.

## 3. Step-by-Step Strategic Roadmap

### Phase 1: Lifecycle & Self-Healing
1.  **Labeling Strategy:** Update `ContainerManager.CreateContainer` to apply a standard label: `com.ai-oak.mcp-tool=true`.
2.  **Janitor Logic:** Implement `Cleanup()` using `cli.ContainerList` with filters.
    *   **Action:** cross-reference running containers with the Redis `Registry`.
    *   **Action:** Kill any container with the system label that is not in the Registry or has timed out.
3.  **Health Tracking:** Update `Registry` to store `LastSeen` timestamps for active tool sessions.

### Phase 2: Protocol & Security Hardening
1.  **Notification Bus:** Add an `OnNotification(func(method string, params json.RawMessage))` callback to the RPC `Client`.
2.  **Group Alignment:** Refactor `api.MCPHandler` and `api.LLMHandler` to accept an `*echo.Group` in their `RegisterRoutes` methods.
    *   **Outcome:** Ensures `AuthMiddleware` is applied to `/api/mcp/*` and `/api/llm/*`.
3.  **Stream Robustness:** Implement a 30-second read timeout for Docker stdio streams to prevent goroutine hangs on crashed containers.

### Phase 3: Ecosystem Unification
1.  **Unified Docker Compose:** Create a `docker-compose.yml` in the root directory.
    *   **Services:** `orchestrator` (backend), `web-manager` (frontend), `redis`, `keycloak`, `postgres` (for Keycloak).
2.  **Networking:** Define an `oak-network` to allow the Orchestrator to resolve containers it spawns by their name.

## 4. Verification & Testing Plan
*   **Orphan Pruning:** Manually start a container with the `com.ai-oak.mcp-tool` label. Wait 5 mins. Verify Janitor kills it.
*   **Security Check:** Attempt to `GET /api/mcp/tools` without a Bearer token. Expect `401 Unauthorized`.
*   **Notification Test:** Mock an MCP server sending a `notifications/progress` message. Verify the Go backend logs it via the new callback.
*   **Compose Test:** Run `docker-compose up` and verify the Go backend can ping Redis and Keycloak within the internal network.

## 5. Risk Assessment
*   **Docker Socket Security:** Mounting `docker.sock` in the backend container is powerful.
    *   *Mitigation:* Run the Go process as a non-root user with specific GID access to the socket.
*   **Janitor Aggression:** If the Registry is temporarily unreachable, the Janitor might kill all tools.
    *   *Mitigation:* Implement a "Grace Period" and retry logic for the Registry before taking destructive action.
*   **Frontend Sync:** The Vue app might expect specific status codes for tool failures.
    *   *Mitigation:* Map Go error types to existing Node.js error codes for seamless UI compatibility.
