# Gemini TUI Interaction Guide 🤖

This guide explains how to run, interact with, and test the `insta-auto-post` TUI tool. As a headless agent, you can simulate user interactions using `send_command_input` to verify the logic and flow.

## 🚀 How to Run

To start the TUI, use the following command:

```bash
go run cmd/insta-auto-post/main.go
```

## 🎮 Interaction Patterns

The TUI is built with Bubble Tea. Since you cannot see the screen directly, follow these input patterns for testing different states:

### 1. First-Time Setup
On a clean database (or if `photos_dir` is not set):
- **Set Directory**: Type the path (e.g., `tests/photos`) and press `\n`.
- **Auto-Cleanup**: Type `y` or `n` to toggle the 30-day cleanup feature.

### 2. Main Menu Navigation
Once setup is complete, you will be at the main menu:
- **Move Selection**: Use `j` (down) or `k` (up).
- **Enter View**: Press `\n` on the selected item.
- **Back to Menu**: Press `q` from any sub-view.

### 3. Media Browser (The Core Loop)
- **Select Media**: Navigate with `j`/`k`, then press `Space` or `\n` to toggle selection `[x]`.
- **Continue to Composer**: Press `c` once you have selected your photos.

### 4. Post Composer
- **Write Caption**: Type your caption text.
- **Save Draft**: Press `\n` to save the post to the database and return to the main menu.

## 🧪 Testing Scenarios

### Scenario A: Verify Photo Limit Warning
1. Create a directory with 1,001 empty files:
   ```bash
   mkdir -p test_warn && for i in {1..1001}; do touch test_warn/img_$i.jpg; done
   ```
2. Run the TUI and set the directory to `test_warn`.
3. Verify the TUI displays the warning (you can check logs or simulate an `\n` to bypass it).

### Scenario B: Testing Auto-Cleanup
1. Manually insert a "posted" record into the database with a date > 30 days ago.
2. Ensure the file exists on disk.
3. Start the TUI with `auto_cleanup` enabled.
4. Verify the file is deleted from the filesystem.

### Scenario C: Post Persistence
1. Select 3 photos in the Browser.
2. Enter a caption "Test Caption" in the Composer.
3. Press `\n`.
4. Run a SQL query to verify the post exists:
   ```bash
   sqlite3 insta_auto_post.db "SELECT caption FROM posts ORDER BY id DESC LIMIT 1;"
   ```

## ⚠️ Important Notes
- **Input Lag**: When using `run_command` and `send_command_input`, wait for the process to process the input.
- **Database**: The database is `insta_auto_post.db`. You can reset the state by deleting this file.
- **Logs**: If the TUI crashes, check the terminal output for panic messages.
