# Technology Stack

## Core Language & Runtime
- **Go (1.21+):** Primary programming language chosen for its performance, strong concurrency primitives (essential for background scheduling), and excellent support for CLI/TUI tools.

## Terminal User Interface (TUI)
- **Bubble Tea (The Elm Architecture for Go):** The foundational framework for managing the application state and UI updates.
- **Lipgloss:** Used for sophisticated UI styling, layouts, and colors.
- **Bubbles:** Reusable TUI components (text inputs, lists, tables, file pickers) for a consistent user experience.

## Data Persistence & Storage
- **SQLite:** A lightweight, serverless relational database used for local storage of post data, media metadata, and application settings.
- **Go-sqlite3:** The CGO-based SQLite driver for Go.

## API & Networking
- **Instagram Graph API (v24.0+):** The official interface for content publishing and retrieving engagement metrics.
- **Standard Library `net/http`:** For making robust, concurrent API requests.
- **Ngrok-Go:** Native integration for creating secure tunnels to expose local media files to the internet without external binaries.

## Security & Configuration
- **Go-keyring:** Cross-platform library used to securely store sensitive API credentials (Access Tokens, App Secrets) in the OS-level keychain.
- **Godotenv:** For loading environment-specific configurations during development and dry runs.
