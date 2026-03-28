# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build TUI binary
go build ./cmd/post-creator/...

# Build CLI binary
go build ./cmd/cli/...

# Run all tests
go test ./...

# Run tests for a single package
go test ./tui/... -v
go test ./db/... -v
go test ./cmd/cli/commands/... -v

# Run a single test
go test ./tui/... -run TestFilterBrowserRows -v

# Run TUI (requires ngrok token + Instagram credentials configured)
SKIP_TUNNEL=true ./post-creator

# Run CLI (headless)
./post-creator-cli --skip-tunnel auth status
./post-creator-cli --skip-tunnel --pretty post list
```

## Architecture

### Two binaries, one infrastructure

Both binaries share `internal/app.App` for startup: DB init, API client creation, HTTP file server (serves `photos/` dir to Instagram), and ngrok tunnel. The TUI runs the server on `:8080`; the CLI uses `:8081` to allow both to coexist.

```
cmd/post-creator/  → TUI binary (interactive)
cmd/cli/           → CLI binary (headless, JSON output)
internal/app/      → Shared: App struct, StartInfra, Shutdown
api/               → Instagram client, scheduler, auth, tunnel
db/                → SQLite layer (posts, media, settings)
tui/               → BubbleTea TUI
```

### Credential storage

Credentials (`INSTA_ACCESS_TOKEN`, `INSTA_IG_ID`, `NGROK_AUTH_TOKEN`) are stored in the **system keyring** via `api.GetCredential` / `api.SetCredential`. Fallback to environment variables / `.env` file. Never stored in the DB or source.

### CLI pattern: `**app.App`

All CLI command constructors in `cmd/cli/commands/` accept `**app.App` (pointer to pointer). The app is `nil` at command registration time and is set by cobra's `PersistentPreRunE` in `cmd/cli/main.go`. Commands dereference via `(*a).DB`, `(*a).Scheduler` at `RunE` time. All output goes through `cmd/cli/output.OK` / `output.Err` for consistent JSON.

### TUI structure

The `tui/` package is a single BubbleTea model (`Model` in `model.go`). View routing:

- `model.go` — `Update` dispatches to per-view handlers; `View` calls `renderCurrentView`
- `handlers.go` — `updateBrowserView`, `updateComposerView`, `updateMenuView`
- `settings_handlers.go` — auth and ngrok settings, Shift+Tab visibility toggle
- `setup_handlers.go` — first-time setup (directory picker using `browserTable`)
- `views.go` — all `view*()` rendering functions, one per `ViewState`
- `browser.go` — `filterBrowserRows`, `sortBrowserRows`, `buildBrowserRows`, `browserColumns`; extracted browser logic shared between BrowserView and SettingsDirView
- `table_actions.go` — `refreshBrowserTable`, `enterBrowserDirectory`, `parentDirectory`
- `styles.go` — Claude Code-inspired palette (`Theme.Primary = #DA7756`), all lipgloss styles
- `keys.go` — unified `KeyMap`; `Back = esc/q`, `Quit = ctrl+c`, `ShiftTab = toggle secret visibility`

### Scheduler

`api.Scheduler` runs as a background goroutine in the TUI (`scheduler.Start()`). In CLI mode it is never started as a goroutine — `scheduler run` calls `CheckAndPublish()` once (cron-friendly); `scheduler daemon` runs a 1-minute ticker loop. `ReportChan` must be drained in CLI mode via a goroutine to prevent blocking.

### Instagram publishing flow

`post create` → saves row to `posts` table with `status=draft` or `status=scheduled`. `scheduler.CheckAndPublish()` picks up due scheduled posts → `api.Publisher.PublishPost()` → uploads media via the ngrok tunnel URL → calls Instagram Graph API to create carousel → updates `status=published`.

### DB schema

SQLite via `go:embed schema.sql`. Three main tables: `posts`, `media` (file registry with hash + posted status), `settings` (key-value). `post_media` join table links posts to media files.
