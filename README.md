# Instagram Auto-Post TUI Tool 📸

A professional-grade Terminal User Interface (TUI) tool built in Go for managing and scheduling Instagram carousel posts. This tool is designed for personal usage to automate the publishing of photos while ensuring rate limits are respected and content isn't duplicated.

## ✨ Features

- **Professional TUI**: Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea), offering a sleek, responsive terminal experience.
- **First-Time Setup**: Guided setup on first run to configure your workspace.
- **Directory Manager**: Lock the TUI to a specific photos directory to keep your workspace organized.
- **Smart Media Tracking**: Uses SQLite to hash and track local files, ensuring you never post the same photo twice.
- **Auto-Cleanup**: Optional feature to automatically delete posted photos after 30 days, keeping your directory clean.
- **Photo Limit Protection**: Automatic warnings when your photo directory exceeds 1,000 items to maintain performance.
- **Carousel Support**: Easily select 2-10 photos to create high-engagement carousel posts via the Instagram Content Publishing API.
- **Interactive Composer**: Write captions directly in your terminal and choose to schedule NOW or save for later.
- **Scheduled Posts View**: A dedicated table view to track your post history, current status, and upcoming queue.
- **Background Scheduler**: A real-time background worker that manages the publishing queue and respects Instagram's rate limits.
- **Dry Run Mode**: A safe way to test your setup and scheduler without actually posting to Instagram.
- **Environment Sync**: Integration with `.env` files for easy configuration and cloud sync via tools like Dotenv Vault.
- **Single Binary**: Compiles to a single lightweight binary for easy execution on macOS.

## 🛠 Tech Stack

- **Language**: Go (Golang 1.25.6)
- **TUI Framework**: Bubble Tea, Bubbles (Table, FilePicker, List), Lip Gloss
- **Database**: SQLite3
- **API**: Instagram Graph API (v24.0)

## 🚀 Getting Started

### Prerequisites

1. **Go 1.25+** installed.
2. **Instagram Business Account** linked to a Facebook Page.
3. **Facebook App** with `instagram_content_publishing` permissions.

### Build

```bash
go build -o insta-auto-post cmd/insta-auto-post/main.go
```

### Run

```bash
go run cmd/insta-auto-post/main.go
```

## ⚙️ Configuration

Create a `.env` file in the root directory (use `.env.example` as a template):

```env
# Instagram API Credentials
INSTA_ACCESS_TOKEN=your_token
INSTA_IG_ID=your_account_id

# App Settings
DRY_RUN=true
PUBLIC_URL_PREFIX=https://your-public-host.com/
```

- **DRY_RUN**: If `true`, the scheduler will log its actions instead of calling the Instagram API. Perfect for initial testing.
- **PUBLIC_URL_PREFIX**: Instagram requires images to be publicly accessible. Set this to your local tunnel URL (Ngrok, Cloudflare, etc.).

> [!IMPORTANT]
> For local usage, your local `photos` folder must be served publicly. The scheduler appends the filename to this prefix to create the final URL for Instagram.

## 📖 Usage Guide

1. **First Run Setup**:
   - Enter the absolute path to your photos directory (e.g., `/Users/me/photos`).
   - Choose whether to enable **Auto-Cleanup**.
2. **Scan Media**: Open the **Media Browser**.
3. **Select for Carousel**: Use `Enter` to select images.
4. **Compose**: Press `c` to enter the **Composer**.
   - Input your caption.
   - Press **Enter** to schedule for IMMEDIATE posting.
   - Press **d** to save as a Draft.
5. **Monitor Queue**: Open **Scheduled Posts** to see the table of all posts and their current status (`SCHEDULED`, `DRAFT`, `PUBLISHED`).

## 📂 Project Structure

- `cmd/`: Application entry point.
- `tui/`: Bubble Tea models, views, and update logic for the terminal interface.
- `api/`: Instagram client and background scheduler logic.
- `db/`: SQLite schema, setting persistence, and media tracking layer.

---
Built with ❤️ using Go and Bubble Tea.
