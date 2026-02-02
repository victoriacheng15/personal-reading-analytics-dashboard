# 3. Static Generation for Historical Metrics

- **Status:** Accepted
- **Date:** 2026-01-30
- **Author:** Victoria Cheng

## Context and Problem Statement

The current analytics dashboard (`site/index.html`) is ephemeral; it only displays the state of the reading list based on the most recent metrics generation (the latest `YYYY-MM-DD.json`). When a new report is generated, the previous state is overwritten and lost to the user interface.

However, the system already archives the raw data in the `metrics/` directory as individual JSON files. Users (and the system owner) desire the ability to browse past weeks to track trends, review previous "AI Delta Analyses," and see the state of their reading backlog at specific points in time.

We need a way to expose this historical data without introducing complex server-side infrastructure (e.g., a database + dynamic backend server), maintaining the project's core philosophy of being a lightweight, self-contained static site hosted on GitHub Pages.

## Decision Outcome

We will extend the static site generation process to build a **full historical archive** for every available metrics file.

- **Multi-Pass Generation**: Instead of processing only the latest JSON file, the generator will iterate through all `metrics/*.json` files.
- **Directory Structure**: We will adopt a nested folder structure for history to keep the root clean:

    ```text
    site/
    ├── analytics.html      # (Latest)
    └── history/
        ├── 2026-01-23/
        │   └── analytics.html  # Snapshot for Jan 23
        └── 2025-11-28/
            └── analytics.html  # Snapshot for Nov 28
    ```

- **Asset Handling**: Static assets (CSS/JS) will be copied to each historical subdirectory. Templates will use relative paths via a `BaseURL` field in the ViewModel (e.g., `../../analytics.html`) to ensure cross-linking works.
- **Navigation**: A context-aware snapshot selector (dropdown) will be added to the navigation bar on Analytics pages, allowing users to switch between the latest view and specific historical snapshots seamlessly.

## Consequences

- **Positive:**
  - **Immutable History**: Creates a permanent, browsable record of reading habits.
  - **Zero Infrastructure Cost**: Continues to rely on GitHub Pages and simple file storage. No database or application server required.
  - **Offline Capable**: The entire site remains a collection of static files that can be viewed locally without a server.
  - **SEO/Linkability**: Every report gets a permanent URL (e.g., `/history/2026-01-23/analytics.html`), allowing specific weeks to be bookmarked or shared.

- **Negative/Trade-offs:**
  - **Build Time Growth**: The generation time will increase linearly as the number of weekly reports grows. However, since metrics are generated only once per week, the total number of files (e.g., ~52 per year) will remain small enough for Go's static generation to handle with negligible performance impact for several years.
  - **Template Complexity**: Templates must handle dynamic "root paths" via `BaseURL` to ensure links and assets work correctly across different directory depths.

## Verification

- [x] **Manual Check:** Run `make nix-run-analytics` and inspect the `site/history/` directory to ensure subfolders are created for past dates.
- [x] **Automated Tests:** `make nix-go-test` passes, including updated service and main tests for multi-pass generation.
