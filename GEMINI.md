# Project: Instagram Post Creator TUI

This document provides comprehensive context for the `insta_auto_post` project, serving as a guide for developers and AI agents.

## 1. Project Overview

**Name:** Instagram Post Creator TUI (`insta_auto_post`)
**Purpose:** A terminal-based tool to manage, compose, and schedule Instagram carousel posts using local media files.
**Tech Stack:**
- **Language:** Go (1.21+)
- **TUI Framework:** [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **Database:** SQLite
- **API:** Instagram Graph API (v24.0)

## 2. Architecture

The application is structured into four main layers:

### A. Entry Point (`cmd/post-creator/main.go`)
- Loads environment variables (`.env`).
- Initializes the SQLite database.
- Sets up the Instagram API client.
- Starts a background **Scheduler** for post processing.
- Starts a local **HTTP File Server** (port 8080) to expose local photos to the Instagram API (requires a public tunnel like `ngrok` if not in `DRY_RUN`).
- Launches the **TUI** program.

### B. TUI Layer (`tui/`)
Follows the Model-View-Update (ELM) architecture:
- **Model (`model.go`)**: Holds application state (current view, selected media, inputs).
- **Update (`handlers.go`, `actions.go`)**: Handles key presses and messages (window resize, API responses).
- **View (`views.go`, `rendering.go`)**: Renders the UI strings based on the current state.
- **Components (`components.go`)**: Reusable UI parts (tables, lists).

### C. Database Layer (`db/`)
- Uses `database/sql` with SQLite.
- **`media.go`**: Manages local file scanning, hashing, and "posted" status.
- **`posts.go`**: Manages post creation, status updates, and retrieval.
- **`settings.go`**: Key-value store for user preferences (e.g., `photos_dir`).

### D. API Layer (`api/`)
- **`instagram.go`**: Client for the Instagram Graph API.
- **`publisher.go`**: Logic to create media containers and publish them.
- **`scheduler.go`**: A background routine that polls the database for `scheduled` posts and advances them through the publishing lifecycle (`scheduled` -> `publishing` -> `published`).

## 3. Data Model (SQLite)

### Tables
- **`media`**: Tracks local image files.
  - `path`: Unique file path.
  - `hash`: Content hash to prevent duplicate uploads.
  - `is_posted`: Boolean flag.
- **`posts`**: Represents a carousel or single post.
  - `status`: `draft`, `scheduled`, `publishing`, `published`, `failed`.
  - `ig_container_id`: Instagram API ID for the container.
- **`post_media`**: Junction table linking `posts` and `media`.
  - `display_order`: Order of images in the carousel.
- **`settings`**: Configuration.
  - `photos_dir`: Root directory for scanning images.

## 4. Key Workflows

### Post Creation Flow
1.  **Scan**: User selects a directory. App scans for valid extensions (`.jpg`, `.png`, etc.).
2.  **Select**: User picks 1-10 images in the TUI.
3.  **Compose**: User enters a caption.
4.  **Schedule**: Post is saved to DB with status `scheduled`.
5.  **Publish (Background)**:
    - Scheduler finds `scheduled` post.
    - Uploads media to Instagram (creating containers).
    - Checks container status until `FINISHED`.
    - Calls `media_publish` endpoint.
    - Updates DB status to `published`.

## 5. Developer Guide

### Environment Setup
Create a `.env` file:
```env
INSTA_ACCESS_TOKEN=...
INSTA_IG_ID=...
INSTA_APP_ID=...
INSTA_APP_SECRET=...
DRY_RUN=true  # Set to false to actually post
PUBLIC_URL_PREFIX=https://your-tunnel.ngrok.io/ # Required for non-dry-run
```

### Running the App
```bash
go run cmd/post-creator/main.go
```

### Testing
- **Unit Tests**: `go test ./...`
- **Manual TUI Testing**: See "TUI Interaction Guide" below.

## 6. TUI Interaction Guide 🤖

(Original content preserved for reference)

### 🚀 How to Run
```bash
go run cmd/post-creator/main.go
```

### 🎮 Interaction Patterns

#### 1. First-Time Setup
- **Set Directory**: 
  - `j`/`k` to navigate. `Enter` to open. `s` to select root.
- **Auto-Cleanup**: `y`/`n`.

#### 2. Main Menu
- **Move**: `j`/`k`. **Enter**: Select.
- **Views**: Dashboard, Media Browser, Scheduled Posts, Settings.

#### 3. Media Browser
- **Select**: `Enter` on file (`📄`). `Enter` on folder (`📁`).
- **Continue**: Press `c` after selection.

#### 4. Post Composer
- **Draft**: Press `d`.
- **Schedule**: Press `Enter`.
- **Cancel**: Press `q`.

#### 5. Settings
- **Change Directory**: Navigate and press `s`.

## 7. Troubleshooting

- **"Photo Limit Reached"**: The directory has too many files. Delete old ones or enable Auto-Cleanup.
- **API Errors**: Check `.env` credentials and `debug.log`.
- **Images not loading on Instagram**: Ensure `PUBLIC_URL_PREFIX` is reachable from the internet.