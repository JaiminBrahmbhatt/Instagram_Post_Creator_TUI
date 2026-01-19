# Initial Concept
A terminal-based tool to manage, compose, and schedule Instagram carousel posts using local media files, with intelligent insights for engagement optimization.

# Product Definition

## Target Audience
- Software Developers who want to automate their social media management through the command line.

## Core Problem Statement
The current process of managing and posting content to Instagram is manual and time-consuming. Developers lack a tool that integrates with their local workflow, understands media relationships (what to post together), and optimizes posting times based on engagement data, all while running reliably in the background.

## Key Goals
- Automate the Instagram posting workflow entirely from the terminal.
- Provide intelligent media management that suggests optimal groupings (single vs. carousel).
- Use engagement insights to determine and recommend the best times for posting.
- Enable robust, asynchronous background scheduling and publishing.

## Main Features
- **Intelligent Media Selection:** Suggestions for grouping photos together or posting them individually based on content and historical engagement.
- **Engagement Optimization Engine:** Analyzes past post performance to suggest the "best time to post" for future content.
- **Async Scheduler & Publisher:** A background engine that handles the multi-step Instagram publishing process (container creation -> status polling -> final publishing) without user intervention.
- **Analytics Dashboard:** A TUI view displaying performance metrics (likes, comments) and trends.
- **Duplicate Prevention:** Tracks media state (e.g., via SQLite) to ensure the same files are not inadvertently reposted.

## Success Metrics
- **Reliable Background Execution:** High success rate for scheduled posts once triggered by the background engine without manual intervention.
- **Workflow Efficiency:** Significant reduction in time spent between media selection and final publication compared to manual or GUI-based methods.
