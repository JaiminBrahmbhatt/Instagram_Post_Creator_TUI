# Track Specification: Enhanced Environment & Settings Management

## 1. Goal
Centralize all environment variable and configuration management within the TUI's "Settings" tab. Rename "Manage API Credentials" to "Environment Configuration" (or similar) to reflect its expanded scope, allowing users to manage not just API keys but all `.env` related settings (like `DRY_RUN`, `PUBLIC_URL_PREFIX`, etc.) directly from the application.

## 2. Core Requirements
- **Rename Settings Item:** Change "Manage API Credentials" to "Environment Configuration".
- **Expand Configuration Scope:**
    -   Allow editing of `DRY_RUN` (boolean toggle).
    -   Allow editing of `PUBLIC_URL_PREFIX` (text input).
    -   Maintain existing editing of `INSTA_ACCESS_TOKEN` and `INSTA_IG_ID`.
- **Persistence:** Ensure all changes are correctly saved to the `.env` file (or keyring where appropriate) and reloaded by the application.
- **UI Updates:** Update the settings form to handle different input types (text vs boolean/toggle).

## 3. Design References
- **Existing Settings:** Follow the pattern of the current `SettingsAuthView` but expand the list of inputs.
- **Input Types:** Use `bubbles/textinput` for strings. For `DRY_RUN`, consider a text input accepting "true"/"false" or a custom toggle component if feasible (simplest is text input first).

## 4. Scope
- **In Scope:**
    -   Modifying `tui/constants.go` for the menu title.
    -   Modifying `tui/components.go` to add new inputs.
    -   Modifying `tui/views.go` to render the expanded form.
    -   Modifying `tui/model.go` and `tui/settings_handlers.go` to handle the new fields and save logic.
    -   Ensuring `.env` file updates (using `godotenv` or manual writing if needed, though `db` settings are distinct from `.env`).
    -   *Clarification:* The current app uses `go-keyring` for secrets and `.env` for initial config. We need to decide if we are writing back to `.env` or storing these overrides in the SQLite `settings` table.
    -   *Decision:* The prompt implies "ENV management". Writing to `.env` at runtime is tricky. Storing in SQLite `settings` table (which overrides `.env`) is the robust pattern already used for `photos_dir`. We will migrate `DRY_RUN` and `PUBLIC_URL_PREFIX` to be loadable from DB Settings, falling back to ENV.

- **Out of Scope:**
    -   Major architectural changes to how secrets are stored (keep using Keyring for tokens).

## 5. Success Criteria
-   User can change `DRY_RUN` and `PUBLIC_URL_PREFIX` from the TUI.
-   "Manage API Credentials" is renamed.
-   Application respects the new values immediately or after restart.
