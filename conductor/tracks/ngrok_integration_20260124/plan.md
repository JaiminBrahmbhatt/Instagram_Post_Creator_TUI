# Plan: Native Ngrok Integration

This plan outlines the integration of `ngrok-go` into the `insta_auto_post` project to replace external tunnel binaries.

## Phase 1: Secure Token Management [checkpoint: 38162ec]
Goal: Implement secure storage and retrieval of the Ngrok Auth Token using `go-keyring`.

- [x] Task: Implement Ngrok Token Storage Logic 3cede30
    - [ ] Create `api/auth.go` (if not exists) or update it to handle Ngrok tokens.
    - [ ] Write tests for saving and retrieving the token from the keyring.
    - [ ] Implement `SaveNgrokToken(token string)` and `GetNgrokToken()`.
- [x] Task: Update Settings TUI for Ngrok Token 62d7b5a
    - [ ] Add `NgrokToken` field to the Settings model in `tui/model.go`.
    - [ ] Update `tui/views.go` to render a password-masked input for the token.
    - [ ] Implement handler in `tui/settings_handlers.go` to save the token to the keyring.
- [x] Task: Conductor - User Manual Verification 'Phase 1: Secure Token Management' (Protocol in workflow.md) d6da22c

## Phase 2: Native Tunnel Implementation
Goal: Integrate `ngrok-go` library and implement tunnel lifecycle management.

- [x] Task: Implement Ngrok Tunnel Manager f98040d
    - [ ] Add `github.com/ngrok/ngrok-go` to `go.mod`.
    - [ ] Create `api/tunnel.go` to manage the ngrok session and tunnel.
    - [ ] Write tests for tunnel lifecycle (Mocking ngrok session if possible).
    - [ ] Implement `StartTunnel(ctx, token)` and `StopTunnel()`.
- [x] Task: Integrate Tunnel with Application Startup 2b6e14e
    - [ ] Update `cmd/post-creator/main.go` to initialize the tunnel on startup if a token exists.
    - [ ] Ensure the tunnel URL is captured and stored in the application state.
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Native Tunnel Implementation' (Protocol in workflow.md)

## Phase 3: Integration & UI Visibility [checkpoint: 862b4f3]
Goal: Use the tunnel URL in the publishing flow and display status in the TUI.

- [x] Task: Update Publisher to Use Tunnel URL 878d390
    - [ ] Modify `api/publisher.go` to use the dynamic ngrok URL instead of a hardcoded or env-based host.
    - [ ] Update media container creation logic to point to the ngrok-provided public endpoint.
- [x] Task: Display Tunnel Status in Dashboard ac9172a
    - [ ] Update `tui/model.go` to track tunnel status and public URL.
    - [ ] Update `tui/views.go` (Dashboard) to display "Tunnel: Active [URL]" or "Tunnel: Inactive".
- [x] Task: Conductor - User Manual Verification 'Phase 3: Integration & UI Visibility' (Protocol in workflow.md) bdfb317

## Phase 4: Final Polish & Cleanup
Goal: Ensure graceful shutdown and remove obsolete environment variables.

- [x] Task: Graceful Shutdown Implementation 6f3f63f
    - [ ] Ensure `StopTunnel()` is called during the application's cleanup phase in `main.go`.
- [ ] Task: Documentation & Cleanup
    - [ ] Update `README.md` and `.env.example` to reflect the new Ngrok token requirement.
    - [ ] Remove any logic related to `DRY_RUN` host overrides that are now redundant.
- [ ] Task: Conductor - User Manual Verification 'Phase 4: Final Polish & Cleanup' (Protocol in workflow.md)
