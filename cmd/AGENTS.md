# Go Services (`/cmd`)

This directory contains the Go-based services for generating metrics and the analytics dashboard. These services consume data processed by the Python extraction script and are designed for long-term maintenance and extensibility.

## Operational Context

- **Metrics Generation:** This service (`cmd/metrics`) fetches raw data (from Google Sheets), calculates analytical metrics, and persists them as JSON files (`metrics/*.json`). This is typically run on a schedule (e.g., weekly via GitHub Actions).
- **Dashboard Generation:** This service (`cmd/dashboard`) reads the latest generated metrics JSON, prepares data for charting, and renders the static HTML analytics dashboard into the `site/` directory. This is run after metrics generation.
- **Environment:** All Go services are built and run within a Nix environment (`shell.nix`) for reproducible builds and dependency management.

## Directory Structure

### `cmd/metrics/`

Responsible for orchestrating the metrics calculation pipeline.

- `main.go`: The primary entry point for the metrics generation application.
- `internal/metrics/`: Contains the core business logic for:
  - Interacting with the Google Sheets API.
  - Parsing raw article data.
  - Calculating various statistics (read rates, unread articles by age, etc.).
- `internal/schema.go`: Defines the shared Go data structures (structs) used across both the `metrics` and `dashboard` services for representing articles, metrics, and other data.

### `cmd/dashboard/`

Responsible for rendering the user-facing analytics dashboard.

- `main.go`: The primary entry point for the dashboard generation application.
- `internal/dashboard/`: Contains logic for:
  - Loading the latest metrics JSON.
  - Preparing data into `ViewModel` structures suitable for HTML templates.
  - Rendering HTML templates.
- `internal/dashboard/templates/`: Stores all HTML templates and CSS stylesheets.
  - `analytics.html`, `base.html`, `evolution.html`, `footer.html`, `header.html`, `index.html`: Go HTML templates for the dashboard pages.
  - `css/styles.css`: Centralized stylesheet for the dashboard, adhering to defined CSS standards.

## Key Dependencies & Technologies

- **Runtime:** Go (GoLang)
- **Google Sheets API:** `google.golang.org/api/sheets/v4` for reading data.
- **Templating:** Go's built-in `html/template` package for dynamic HTML generation.
- **JSON Processing:** Go's `encoding/json` package for reading/writing metrics data.
- **Dependency Management:** Nix (`shell.nix`) ensures a consistent and reproducible Go toolchain.
- **Testing:** Go's native `testing` framework for unit and integration tests.

## Workflow Integration

- **Environment Setup:** Always use `nix-shell` to enter the Go development environment, ensuring all tools and dependencies are correctly set up.
- **Build & Run:**
  - `make run-metrics` (executed via `nix-shell --run "make run-metrics"`) to generate the latest metrics JSON.
  - `make run-dashboard` (executed via `nix-shell --run "make run-dashboard"`) to generate the HTML dashboard.
- **Testing:** `nix-shell --run "make go-test"` to run all Go unit tests.
- **Code Quality:** `nix-shell --run "make go-format"` to automatically format Go source files.
