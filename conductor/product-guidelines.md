# Product Guidelines

## Prose Style & Tone
- **Professional & Minimalist:** Use clear, direct, and concise language. Communication should prioritize efficiency and utility, avoiding unnecessary fluff.
- **Actionable & Precise:** Error messages and instructions must be specific. Tell the user exactly what happened and provide a clear path to resolution.

## Visual Identity (TUI)
- **Modern & Information-Dense:** Utilize the full capabilities of Lipgloss for a vibrant, colorful interface with clear visual hierarchies. 
- **Structured Data:** Prioritize grids, tables, and borders to organize large amounts of information (like media lists and analytics) in a readable, professional format.
- **Standardized Elements:** Use consistent color coding for status (e.g., green for published, yellow for scheduled, red for failed).

## User Experience (UX) Principles
- **Keyboard-First Design:** Ensure every functional element of the app is reachable via intuitive, well-documented hotkeys.
- **Persistent Context:** A dynamic footer should always display the available keybindings for the current view, reducing cognitive load.
- **Fail-Safe Interaction:** Implement confirmation prompts for any action that could lead to data loss or unintended public exposure (e.g., canceling a scheduled post, deleting media).

## Intelligent Insights & Trust
- **Transparent Reasoning:** When the app makes a suggestion (like an optimal posting time), it should briefly explain the underlying data or logic to build user confidence.
- **Frictionless Override:** Intelligent defaults should be helpful but never restrictive. The user must always be one keystroke away from overriding an automated suggestion.

## Error Handling & Reliability
- **Comprehensive Activity Logging:** Maintain a persistent, human-readable log of all background operations (polling, container creation, publishing) so the user can audit the app's performance.
- **Self-Healing Mechanics:** The background engine should implement robust retry logic for transient API or network errors before escalating to a "failed" state.
