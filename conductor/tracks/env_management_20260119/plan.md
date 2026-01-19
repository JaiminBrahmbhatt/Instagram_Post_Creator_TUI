# Implementation Plan - Enhanced Environment & Settings Management

This plan expands the Settings view to manage comprehensive environment configurations, backed by the database.

## Phase 1: Foundation & Renaming
Rename the menu item and prepare the database/configuration layer to support new settings.

- [x] Task: Rename Menu Item <!-- id: 0 --> [5dd8d72]
    - [ ] Update `tui/constants.go`: Rename `MenuTitleSettings` (or specific sub-item) constants.
    - [ ] Update `tui/components.go`: Change "Manage API Credentials" to "Environment Configuration".
    - [ ] Verify the menu displays the new name.

- [x] Task: Update Configuration Logic <!-- id: 1 --> [d594b60]
    - [ ] Review `cmd/post-creator/main.go` and `api/client.go` (or wherever config is loaded).
    - [ ] Ensure `DRY_RUN` and `PUBLIC_URL_PREFIX` can be loaded from the `db.Settings` table, falling back to `os.Getenv`.
    - [ ] Create helper methods in `db/settings.go` if needed to get boolean/string settings with defaults.

- [x] Task: Conductor - User Manual Verification 'Phase 1: Foundation & Renaming' (Protocol in workflow.md) <!-- id: 2 --> [checkpoint: 3982016]

## Phase 2: UI Implementation
Expand the Settings form to include the new fields.

- [x] Task: Expand Settings Model <!-- id: 3 --> [d92c040]
    - [x] Update `NewAuthInputs` in `tui/components.go` (or create `NewEnvInputs`) to include:
        - [x] Public URL Prefix (Text)
        - [x] Dry Run (Text/Toggle - use "true"/"false" text for simplicity initially)
    - [x] Update `tui/model.go` to hold these new inputs.

- [x] Task: Update Settings View <!-- id: 4 --> [d92c040]
    - [x] Modify `viewSettingsAuth` (rename to `viewSettingsEnv`?) in `tui/views.go` to render the expanded list of inputs.
    - [x] Ensure the layout handles the increased number of fields gracefully (scrolling or just fitting).

- [x] Task: Update Save Logic <!-- id: 5 --> [d92c040]
    - [x] Modify `tui/settings_handlers.go` (`updateSettingsAuthView`) to:
        - [x] Load initial values from DB/Env when entering the view.
        - [x] Save `PUBLIC_URL_PREFIX` and `DRY_RUN` to the SQLite `settings` table on Enter/Save.
        - [x] Continue saving Secrets to Keyring as before.

- [x] Task: Conductor - User Manual Verification 'Phase 2: UI Implementation' (Protocol in workflow.md) <!-- id: 6 --> [checkpoint: e2f41d2]

## Phase 3: Verification & Cleanup
Ensure the application actually uses the new settings.

- [x] Task: Verify Config Usage <!-- id: 7 --> [ce78b49]
    - [ ] Verify that the `api` package or `main` loop actually queries the DB for these values (via the `Update Configuration Logic` work in Phase 1, but verifying here).
    - [ ] If `main.go` initializes things *before* the DB is queryable for these, we might need to adjust initialization order or hot-reloading.
    - [ ] *Correction:* `DRY_RUN` might be checked deeply. We need to ensure dynamic checking or restart requirement.
    - [ ] Add a "Restart Required" note if dynamic update isn't feasible for some settings.

- [ ] Task: Conductor - User Manual Verification 'Phase 3: Verification & Cleanup' (Protocol in workflow.md)
