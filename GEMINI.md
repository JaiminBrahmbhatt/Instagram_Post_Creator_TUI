# Gemini TUI Interaction Guide 🤖

This guide explains how to run, interact with, and test the `post-creator` TUI tool. As a headless agent, you can simulate user interactions using `send_command_input` to verify the logic and flow.

## 🚀 How to Run

To start the TUI, use the following command:

```bash
go run cmd/post-creator/main.go
```

## 🎮 Interaction Patterns

The TUI is built with Bubble Tea. Since you cannot see the screen directly, follow these input patterns for testing different states:

### 1. First-Time Setup
On a clean database (or if `photos_dir` is not set):
- **Set Directory**: 
  - Use `j` / `k` to navigate folders.
  - Press `Enter` to open a folder.
  - Press `s` to **select the current directory** as the root.
- **Auto-Cleanup**: Type `y` or `n` to toggle the 30-day cleanup feature.

### 2. Main Menu Navigation
Once setup is complete, you will be at the main menu:
- **Move Selection**: Use `j` (down) or `k` (up).
- **Enter View**: Press `Enter` on the selected item.
- **Views**:
  - **Dashboard**: View limits (Work in Progress).
  - **Media Browser**: Select photos for posting.
  - **Scheduled Posts**: View history and pending posts.
  - **Settings**: Configure app parameters.

### 3. Media Browser (The Core Loop)
- **Navigate**: Use `j`/`k` (classic) or `Enter` on `..` to go up.
- **Enter Folder**: Press `Enter` on a directory name (prefixed with `📁`).
- **Select Media**: Press `Enter` on a file name (prefixed with `📄`) to toggle selection (marked with `[x]`).
- **Continue to Composer**: Press `c` once you have selected your photos.

### 4. Settings Menu
- **Change Photos Directory**:
  - Navigate to the desired folder.
  - Press `s` to confirm it as the new root directory.
- **Auto Cleanup**: Triggers the `y/n` prompt for the 30-day cleanup feature.

### 5. Post Composer
- **Write Caption**: Type your caption text.
- **Schedule Post**: Press `Enter` to save with status `scheduled` (immediate post).
- **Save Draft**: Press `d` to save with status `draft`.
- **Cancel**: Press `q` to return to the menu without saving.

### 6. Scheduled Posts (Table View)
- **Navigate Table**: Use `j`/`k` to scroll through the list of posts.
- **Back to Menu**: Press `q`.

## 🧪 Testing Scenarios

### Scenario A: Verify Photo Limit Warning
1. Create a directory with 1,001 empty files:
   ```bash
   mkdir -p test_warn && for i in {1..1001}; do touch test_warn/img_$i.jpg; done
   ```
2. Run the TUI and set the directory to `test_warn`.
3. Verify the TUI displays the warning (requires an `Enter` to bypass).

### Scenario B: Testing Auto-Cleanup
1. Manually insert a "posted" record into the database with a date > 30 days ago.
2. Ensure the file exists on disk.
3. Start the TUI with `auto_cleanup` enabled.
4. Verify the file is deleted from the filesystem (check logs or disk).

### Scenario C: Post Persistence (Draft vs Schedule)
1. Select 3 photos in the Browser.
2. Enter a caption "Test Draft".
3. Press `d`.
4. Run: `sqlite3 post_creator.db "SELECT status FROM posts WHERE caption='Test Draft'"` -> Should be `draft`.
5. Repeat for "Test Schedule" and press `Enter` -> Should be `scheduled`.

### Scenario D: Changing Directory via Settings
1. Go to **Settings** -> **Change Photos Directory**.
2. Navigate to a subfolder.
3. Press `s`.
4. Verify `photos_dir` in `settings` table:
   ```bash
   sqlite3 post_creator.db "SELECT value FROM settings WHERE key='photos_dir';"
   ```

## ⚠️ Important Notes
- **Input Lag**: When using `run_command` and `send_command_input`, wait for the process to process the input.
- **Database**: The database is `post_creator.db`. You can reset the state by deleting this file.
- **Logs**: If the TUI crashes, check the terminal output for panic messages.

