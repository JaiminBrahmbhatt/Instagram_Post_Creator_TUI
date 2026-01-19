# Track Specification: Resource Management & Graceful Shutdown

## 1. Goal
Ensure that the application manages system resources (specifically network ports and external tunnels) responsibly. The HTTP server used for serving photos should shut down gracefully when the application exits, releasing the port immediately. The Cloudflare tunnel should also be reliably terminated.

## 2. Core Requirements
- **Graceful HTTP Shutdown:** Refactor `main.go` to use `http.Server` with a `Shutdown` context instead of `http.ListenAndServe`.
- **Port Release:** Ensure the port (default 8080) is closed when the TUI quits (Esc/q).
- **Tunnel Cleanup:** Verify and ensure the tunnel cleanup function is always called.

## 3. Scope
- **In Scope:**
    -   Modifying `cmd/post-creator/main.go`.
    -   Adding a shutdown signal handling mechanism (or hooking into Bubble Tea's quit).
- **Out of Scope:**
    -   Changing the TUI logic itself (except for exit handling).

## 4. Success Criteria
-   Port 8080 is free immediately after the app closes.
-   No "address already in use" errors when restarting the app quickly.
-   `cloudflared` processes are terminated.
