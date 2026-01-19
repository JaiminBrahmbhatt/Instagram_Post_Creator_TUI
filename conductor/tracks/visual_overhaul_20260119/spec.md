# Track Specification: Visual & UX Overhaul (TUI Modernization)

## 1. Goal
Modernize the Terminal User Interface (TUI) to be cleaner, more beautiful, and user-friendly, drawing inspiration from high-quality CLI tools like Claude Code and Gemini CLI. The goal is to elevate the aesthetic and usability using advanced `lipgloss` styling and `bubbletea` patterns.

## 2. Core Requirements
- **Visual Aesthetic:**
    - Implement a polished, minimalist design with clear visual hierarchy.
    - Use a refined color palette (likely borrowing from the existing scheme but applied more effectively) for consistent branding.
    - Enhance borders, padding, and spacing to reduce visual clutter.
    - Introduce subtle animations (spinners, transitions) where appropriate to make the app feel "alive".
- **UX Improvements:**
    - Redesign the main navigation (Menu) to be more intuitive.
    - Improve the "Media Browser" with a better grid/list layout and clearer selection indicators.
    - Refine the "Composer" view for a better writing experience.
    - Enhance the "Dashboard" and "Status" displays to clearly communicate background activity.
    - Ensure persistent, context-aware help (keybindings) at the bottom of the screen.

## 3. Design References
- **Inspiration:** Claude Code, Gemini CLI, Charmbracelet examples.
- **Key Libraries:** `bubbletea`, `lipgloss`, `bubbles`.

## 4. Scope
- **In Scope:**
    - Refactoring `tui/styles.go` for a unified design system.
    - Updating `tui/views.go` and `tui/components.go` to use new styles.
    - Improving the layout of the Main Menu, Media Browser, Composer, and Dashboard.
    - Enhancing global elements like the status footer and help strip.
- **Out of Scope:**
    - Changing core backend logic (API, Database, Scheduler) unless strictly necessary for UI state.
    - Adding new major features (like Analytics engine) - this is strictly a visual/UX refactor.

## 5. Success Criteria
- The application looks significantly more modern and polished.
- Information is easier to scan and read.
- Navigation feels smoother and more intuitive.
- All existing functionality remains intact and accessible.
