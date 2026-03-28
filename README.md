# Instagram Auto-Post TUI 📸

A terminal tool to manage, schedule, and automate Instagram carousel posts from your workstation. Includes both an interactive TUI and a headless CLI for scripting and agent-driven workflows.

![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go)
![TUI Framework](https://img.shields.io/badge/TUI-BubbleTea-00ADD8?style=flat-square)
![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)

## Features

- **Interactive TUI** — Dashboard, media browser with filter/sort, scheduler view, post composer, and settings. Claude Code-inspired design language.
- **Headless CLI** (`post-creator-cli`) — Full feature parity with structured JSON output. Designed for scripting and AI agent use.
- **Media Browser** — Navigate directories, filter by name (`/`), cycle sort (`s`), go up a directory (`←`). Shows posted status per file.
- **Scheduler** — Background scheduler in TUI mode; `scheduler run` for one-shot cron use; `scheduler daemon` for long-running service.
- **SQLite Persistence** — Posts, media registry (with hash deduplication), and settings.
- **Ngrok Tunneling** — Built-in ngrok integration to serve local photos to the Instagram API.
- **Secret Visibility** — All credential fields masked by default; `Shift+Tab` toggles visibility with a `[visible]` badge.
- **Dry Run Mode** — Full workflow validation without publishing.

## Prerequisites

- **Go 1.25+**
- **Instagram Business Account** linked to a Facebook Page with `instagram_content_publishing` permission
- **[Ngrok Authtoken](https://dashboard.ngrok.com/get-started/your-authtoken)** (free tier works)

## Getting Started

```bash
git clone https://github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI
cd Instagram_Post_Creator_TUI
go mod tidy
```

Copy the example env file and fill in your credentials:

```bash
cp .env.example .env
```

```env
INSTA_ACCESS_TOKEN=your_access_token
INSTA_IG_ID=your_instagram_business_account_id
DRY_RUN=true   # set to false when ready to publish
```

> Ngrok authtoken and Instagram credentials can also be entered directly in the app's **Settings** menu and are stored securely in the system keyring.

## Running

```bash
# Interactive TUI
go run ./cmd/post-creator

# Headless CLI
go run ./cmd/cli -- --help
```

Or build binaries:

```bash
go build -o post-creator ./cmd/post-creator
go build -o post-creator-cli ./cmd/cli
```

## CLI Usage

All commands output JSON. Use `--pretty` for human-readable output.

```bash
post-creator-cli [--skip-tunnel] [--db <path>] [--pretty] <command>

post-creator-cli auth set --token <token> --ig-id <id>
post-creator-cli auth status

post-creator-cli post create --media beach.jpg --caption "Hello" [--schedule 2026-12-01T10:00:00Z]
post-creator-cli post list [--status draft|scheduled|published|failed] [--limit 20]
post-creator-cli post publish --id 3
post-creator-cli post delete --id 3

post-creator-cli media scan [--dir ./photos]
post-creator-cli media list [--unposted]
post-creator-cli media ignore --id 5

post-creator-cli scheduler run      # one-shot (use with cron)
post-creator-cli scheduler daemon   # blocking loop (use as service)
post-creator-cli scheduler status

post-creator-cli settings get --key photos_dir
post-creator-cli settings set --key photos_dir --value /path/to/photos
post-creator-cli settings list

post-creator-cli quota
```

**Cron example** — publish due posts every minute:
```
* * * * * /path/to/post-creator-cli --skip-tunnel scheduler run
```

## Keybindings (TUI)

| Key | Action |
|---|---|
| `↑`/`↓` or `k`/`j` | Navigate |
| `Enter` | Select / confirm / enter folder |
| `Esc` / `q` | Back |
| `Ctrl+C` | Quit |
| `Tab` | Next field (auth/settings forms) |
| `Shift+Tab` | Toggle secret field visibility |
| `/` | Enter filter mode (browser) |
| `s` | Cycle sort (browser) |
| `←` | Go up directory (browser) |
| `c` | Continue to composer (from browser) |
| `d` | Save as draft (composer) |
| `?` | Toggle help |

## Architecture

```
cmd/post-creator/   TUI binary
cmd/cli/            Headless CLI binary
internal/app/       Shared infrastructure (DB, client, HTTP server, tunnel)
api/                Instagram client, scheduler, publisher, auth, tunnel
db/                 SQLite layer — posts, media, settings
tui/                BubbleTea TUI (views, styles, keybindings, browser logic)
```

Credentials are stored in the **system keyring** (`zalando/go-keyring`) with fallback to environment variables.

---

Built for creators who live in the terminal.
