# Instagram Auto-Post TUI Tool 📸

A professional-grade Terminal User Interface (TUI) tool built in Go for managing and scheduling Instagram carousel posts. This tool is designed for personal usage to automate the publishing of photos while ensuring rate limits are respected and content isn't duplicated.

## ✨ Features

- **Professional TUI**: Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea), offering a sleek, responsive terminal experience.
- **Smart Media Tracking**: Uses SQLite to hash and track local files, ensuring you never post the same photo twice.
- **Carousel Support**: Easily select 2-10 photos to create high-engagement carousel posts via the Instagram Content Publishing API.
- **Interactive Composer**: Write and edit captions directly in your terminal.
- **Background Scheduler**: poll-based scheduler that manages the publishing queue and respects Instagram's rate limits.
- **Single Binary**: Compiles to a single lightweight binary for easy execution on macOS.

## 🛠 Tech Stack

- **Language**: Go (Golang)
- **TUI Framework**: Bubble Tea, Bubbles, Lip Gloss
- **Database**: SQLite3
- **API**: Instagram Graph API (v24.0)

## 🚀 Getting Started

### Prerequisites

1. **Go 1.21+** installed.
2. **Instagram Business Account** linked to a Facebook Page.
3. **Facebook App** with `instagram_content_publishing` permissions.

### Build

```bash
go build -o insta-auto-post cmd/insta-auto-post/main.go
```

### Run

```bash
./insta-auto-post
```

## 📖 Usage Guide

1. **Scan Media**: Open the **Media Browser** to see unposted images in your directory.
2. **Select for Carousel**: Use `Space` or `Enter` to select between 2 and 10 images.
3. **Compose**: Press `c` to enter the **Composer**. Input your caption and confirm.
4. **Draft/Schedule**: The post is saved to the SQLite database. The background scheduler will handle the upload and publishing process.

## ⚙️ Configuration

To fully enable publishing, update the `api/instagram.go` or use environment variables for:
- `ACCESS_TOKEN`: Your Meta Graph API User Token.
- `INSTAGRAM_BUSINESS_ACCOUNT_ID`: Your IG ID.

> [!IMPORTANT]
> The Instagram API requires media to be hosted on a **publicly accessible server**. For local usage, you should use a tunnel like [Ngrok](https://ngrok.com/) or [Cloudflare Tunnel] to expose your local media directory to the internet temporarily.

## 📂 Project Structure

- `cmd/`: Application entry point.
- `tui/`: Bubble Tea models, views, and update logic.
- `api/`: Instagram client and background scheduler.
- `db/`: SQLite schema and data persistence layer.

---
Built with ❤️ using Go and Bubble Tea.
