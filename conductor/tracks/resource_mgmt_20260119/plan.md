# Implementation Plan - Resource Management & Graceful Shutdown

This plan focuses on refactoring the `main` entry point to manage the HTTP server lifecycle.

## Phase 1: Server Refactoring
Refactor the global `http.ListenAndServe` into a managed `http.Server` instance.

- [x] Task: Create Server Manager <!-- id: 0 --> [a74fbf0]
    - [x] Modify `cmd/post-creator/main.go`.
    - [x] Create an `http.Server` struct.
    - [x] Run `server.ListenAndServe()` in a goroutine.
    - [x] Implement a shutdown channel to coordinate exit.

- [x] Task: Hook into Exit <!-- id: 1 --> [a74fbf0]
    - [x] Ensure that when `p.Run()` returns (app exit), we call `server.Shutdown(context.Background())`.
    - [x] Ensure the tunnel cleanup is also called.

- [ ] Task: Conductor - User Manual Verification 'Phase 1: Server Refactoring' (Protocol in workflow.md)
