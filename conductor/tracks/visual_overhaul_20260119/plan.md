# Implementation Plan - Visual & UX Overhaul (TUI Modernization)

This plan outlines the steps to modernize the TUI, focusing on a cleaner, "Claude Code/Gemini CLI" inspired aesthetic using `lipgloss` and `bubbletea`.

## Phase 1: Design System & Foundation
Establish the core visual language (colors, typography, spacing) and refactor the global layout structure.

- [x] Task: Audit and Refactor `tui/styles.go` <!-- id: 0 --> [f2a5905]
    - [ ] Analyze existing styles in `tui/styles.go`.
    - [ ] Define a new, modern color palette (primary, secondary, subtle, error, success) in a dedicated theme struct or constant block.
    - [ ] Create standardized Lipgloss styles for:
        - [ ] Headers/Titles (bold, colored, padded)
        - [ ] Containers/Boxes (borders, rounded corners if supported/appropriate, padding)
        - [ ] Text (body, subtle/dimmed, highlighted)
        - [ ] Status indicators (badges, icons)
    - [ ] Verify the new styles compile and look correct in a simple test view.

- [x] Task: Redesign Global Layout (App Shell) <!-- id: 1 --> [dc1ce01]
    - [ ] Update the `View()` method in `tui/model.go` to use a cleaner "App Shell" layout.
    - [ ] Implement a consistent Header (App Title/Logo) and Footer (Status + Help).
    - [ ] Ensure the main content area has appropriate padding and centering.
    - [ ] Update `tui/rendering.go` to utilize the new layout structure.

- [x] Task: Conductor - User Manual Verification 'Phase 1: Design System & Foundation' (Protocol in workflow.md) <!-- id: 2 --> [checkpoint: df58cdd]

## Phase 2: Component Modernization
Update specific views and components to align with the new design system.

- [x] Task: Modernize Main Menu & Navigation <!-- id: 3 --> [3d48a36]
    - [ ] Refactor `NewMenu` in `tui/components.go` to use `bubbles/list` with a custom, high-fidelity delegate.
    - [ ] Style the list items to look like "cards" or cleaner rows with icons.
    - [ ] Update `tui/views.go` to render the menu with the new style.

- [x] Task: Overhaul Media Browser <!-- id: 4 --> [c79126c]
    - [ ] Refactor `NewBrowserTable` in `tui/components.go`.
    - [ ] Apply new table styles (custom headers, row styling, selection highlight) using `lipgloss`.
    - [ ] Improve the visual feedback for selected items (e.g., checkbox icon or distinct color change).
    - [ ] Update the file picker integration to match the new aesthetic.

- [x] Task: Refine Post Composer <!-- id: 5 --> [69d766b]
    - [ ] Update `viewComposer` in `tui/views.go`.
    - [ ] Style the text input (`NewCaptionInput`) to feel more like a modern editor (borders, focus state).
    - [ ] Improve the layout of the "Selected Media" summary in the composer.
    - [ ] Add clear visual cues for "Draft" vs "Schedule" actions.

- [x] Task: Conductor - User Manual Verification 'Phase 2: Component Modernization' (Protocol in workflow.md) <!-- id: 6 --> [checkpoint: a8e79b8]

## Phase 3: Polish & Interaction
Focus on the finer details, animations, and status feedback to make the app feel responsive and "alive".

- [ ] Task: Enhance Dashboard & Status
    - [ ] Redesign `viewDashboard` to use "stat cards" (boxes with big numbers and labels) for quota usage.
    - [ ] Update the "Processing" view in `tui/rendering.go` to use a modern spinner and cleaner log output.
    - [ ] Improve the `viewFooter` to cleanly separate the Status Message from the Help Keybindings.

- [ ] Task: Final Review & consistency Check
    - [ ] Walk through the entire app to ensure consistent spacing and alignment.
    - [ ] Verify that all text is legible and has good contrast.
    - [ ] Ensure keyboard navigation remains intuitive and the help footer is accurate for every view.

- [ ] Task: Conductor - User Manual Verification 'Phase 3: Polish & Interaction' (Protocol in workflow.md)
