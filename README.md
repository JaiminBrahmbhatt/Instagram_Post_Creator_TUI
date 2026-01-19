# Instagram Auto-Post TUI 📸

A simple terminal tool to manage and schedule Instagram carousel posts from your computer.

## ✨ Features

- **Easy Setup**: Guided walkthrough on your first run.
- **Media Browser**: Navigate and select photos directly in your terminal.
- **Carousel Support**: Post up to 10 photos in a single carousel.
- **Scheduled Posts**: Manage your queue and view past posts.
- **Auto-Cleanup**: Optionally delete photos 30 days after they are posted.
- **Dry Run Mode**: Test everything without actually posting to Instagram.

## 🚀 Getting Started

### 1. Prerequisites
- [Go](https://go.dev/doc/install) installed.
- An Instagram Business account and a Facebook App for API access.

### 2. Configuration
Create a `.env` file in the folder (see `.env.example`):
```env
INSTA_ACCESS_TOKEN=your_token
INSTA_IG_ID=your_account_id
DRY_RUN=true
```
*Set `DRY_RUN=false` only when you are ready to post for real.*

### 3. Run
```bash
go run cmd/insta-auto-post/main.go
```

## 📖 How to Use

1. **Setup**: On first launch, pick your photos folder and toggle "Auto-Cleanup".
2. **Select**: Go to **Media Browser**, navigate to your photos, and press `Enter` to select them.
3. **Draft/Schedule**: Press `c` to write a caption.
   - Press `Enter` to post/schedule.
   - Press `d` to save as a draft.
4. **Manage**: Use the **Scheduled Posts** view to check the status of your queue.

---
Built with ❤️ using Go.

