# Instagram Auto-Post TUI 📸

A powerful, high-performance terminal tool to manage, schedule, and automate Instagram carousel posts directly from your workstation. Built with Go and the Bubble Tea framework for a refined developer experience.

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)
![TUI Framework](https://img.shields.io/badge/TUI-BubbleTea-00ADD8?style=flat-square)
![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)

## ✨ Features

- 🖥️ **Premium TUI**: A modular, tabbed interface featuring a dashboard, media browser, and scheduler.
- 📁 **Advanced Media Browser**: Traverse directories, select multiple photos (up to 10 for carousels), and manage files with ease.
- ⏳ **Smart Scheduler**: Background processing engine that polls Instagram container status to ensure reliable media publishing.
- 📝 **Post Composer**: Dynamic captioning with support for both immediate scheduling and draft persistence.
- 💾 **SQLite Persistence**: Reliable storage for posts, drafts, and application settings.
- 🧹 **Auto-Cleanup**: Optional feature to delete posted media after 30 days, keeping your storage lean.
- 🧪 **Dry Run Mode**: Validate your entire workflow and API interactions without actually publishing to Instagram.

## 🚀 Getting Started

### 1. Prerequisites

- **Go 1.21+** installed on your system.
- **Instagram Business Account** linked to a Facebook Page.
- **Facebook Developer App** with the `instagram_content_publishing` permission.
- A public URL (e.g., via `ngrok` or a VPS) if you are not running in DRY_RUN mode, as Instagram needs to pull media from a public link.

### 2. Configuration

Clone the repository and create a `.env` file from the example:

```bash
cp .env.example .env
```

Edit your `.env` with the following parameters:

```env
# Instagram API Credentials
INSTA_ACCESS_TOKEN=your_access_token
INSTA_IG_ID=your_instagram_business_account_id
INSTA_APP_ID=your_app_id
INSTA_APP_SECRET=your_app_secret

# App Settings
DRY_RUN=true                    # Set to false for production
PUBLIC_URL_PREFIX=https://.../  # Public prefix for media files
```

### 3. Installation & Usage

```bash
# Install dependencies
go mod tidy

# Run the application
go run ./cmd/post-creator
```

## 🎮 Navigation & Shortcuts

| Key                | Action                                           |
| ------------------ | ------------------------------------------------ |
| `j`/`k` or `↑`/`↓` | Navigate/Scroll                                  |
| `Enter`            | Select/Confirm/Enter Folder                      |
| `s`                | Pick current directory as Home in Settings/Setup |
| `c`                | Continue to Composer (from Browser)              |
| `d`                | Save as Draft (in Composer)                      |
| `Tab`              | Switch focus (in Auth/Settings)                  |
| `q` / `Esc`        | Back / Exit view                                 |
| `?`                | Toggle Help                                      |
| `Ctrl+C`           | Force Quit                                       |

## 🏗️ Technical Architecture

The project follows a modular Go architecture:

- `cmd/`: Application entry point.
- `tui/`: Bubble Tea components, handlers, and rendering logic (split into modular views).
- `api/`: Instagram Content Publishing API client with custom retry logic and status polling.
- `db/`: SQLite layer for persistent storage of settings and posts.
- `scheduler/`: Background routine for managing scheduled tasks.

## 🛠️ Development

To contribute or modify the tool:

1. **Database Schema**: Managed via `db/schema.sql`.
2. **Components**: UI elements are defined in `tui/components.go`.
3. **Styles**: Global theme and colors are in `tui/styles.go`.

---

Built with ❤️ for creators who love the command line.
