# Agent Guide for Personal Reading Analytics

This document provides context and instructions for AI agents working on the **Personal Reading Analytics** project, a 3-stage data pipeline and dashboard for tracking reading habits.

## 2. Project Overview

**Personal Reading Analytics** is a fully automated data pipeline with CI/CD governance to track, analyze, and visualize reading habits. While the 3-stage pipeline is designed for high automation, it follows a human-in-the-loop model via CI/CD governance for code reviews and merging Pull Requests to ensure architectural integrity.

It operates as a three-stage pipeline:

1. **Extraction (Python)**: Utilizes a universal, configuration-driven extractor for "Zero-Code" onboarding of new sources via Google Sheets, scraping article metadata, and syncing to Google Sheets/MongoDB.
2. **Metrics (Go)**: Calculates statistics from the data.
3. **Dashboard (Go)**: Generates a static HTML dashboard.

- **Core Tech**: Go (Golang), Python 3.x, Nix.
- **Data Store**: Google Sheets (primary), MongoDB (logs/events).
- **Styling**: Centralized CSS (`styles.css`), no frameworks like Tailwind.
- **Goal**: Resilient, observable data pipeline with a clean, static dashboard.
- **External Observability**: Integrates with a public **[Observability Hub](https://github.com/victoriacheng15/observability-hub)** (repo only) that consumes events (MongoDB to PostgreSQL) for Grafana visualization (not publicly exposed).
- **Provider Management**: Supported sources are managed exclusively via the Google Sheets SSOT; direct listings in `README.md` have been removed to prevent redundancy.

## 3. Build and Test Commands

The project uses **Nix** for the Go environment and a standard `.venv` for Python.

### Go (Nix-enabled)

The `Makefile` automatically uses `nix-shell` if available.

| Command | Description |
| :--- | :--- |
| `make run-analytics` | Builds and runs the dashboard generator (`analytics.exe`). |
| `make run-metrics` | Builds and runs the metrics calculator (`metricsjson.exe`). |
| `make go-test` | Runs all Go unit tests. |
| `make go-format` | Formats Go code using `gofmt`. |
| `make go-cov` | Runs Go tests with coverage summary in the terminal. |
| `make go-update` | Updates Go dependencies and runs `go mod tidy`. |

### Python (Local venv)

Ensure you have run `make install` first to set up the environment.

| Command | Description |
| :--- | :--- |
| `make run` | [Docker] Builds and runs extraction via Docker. |
| `make py-run` | [Local] Runs extraction via local venv. |
| `make install` | Create .venv and install dependencies. |
| `make freeze` | Freeze current Python dependencies to `requirements.txt`. |
| `make update` | Update Python dependencies in .venv from `requirements.txt`. |
| `make py-check` | Run ruff check (lint). |
| `make py-format` | Format files with ruff. |
| `make py-test` | Run Python tests via `pytest`. |
| `make py-cov` | Run Python coverage report with missing lines. |

### General Utilities

| Command | Description |
| :--- | :--- |
| `make lint` | Run markdownlint via Docker to check Markdown files. |
| `make clean` | Remove build artifacts and caches. |

## 4. Code Style Guidelines

### Go

- **Strict Adherence**: Code **must** pass `go fmt` and `go vet`.
- **Error Handling**: Wrap errors with context: `fmt.Errorf("failed to [action]: %w", err)`.
- **Dependencies**: Prioritize the Go standard library. Use external packages only when necessary.

### Python

- **Linting**: Must pass `ruff` checks.
- **Structure**: Modularize logic in `script/utils`.
- **Observability**: Critical events (extractions, failures) must be logged to MongoDB.

### HTML/CSS

- **CSS**: Use standard CSS variables in `cmd/internal/analytics/templates/css/styles.css`.
- **No Inline Styles**: All styles must reside in the centralized CSS file.
- **Layout**: Prefer `flex` or `grid` with `gap` for spacing.

### Markdown/Mermaid

- **Mermaid Syntax**: When using parentheses `()` within the text of square-bracketed `[]` Mermaid nodes (e.g., `Node["Text (with parentheses)"]`), always enclose the entire node text in double quotes to avoid parsing errors. For example, use `Node["Example (details)"]` instead of `Node[Example (details)]`.

## 5. Testing Instructions

- **Unit Tests**:
  - Go: `make go-test`
  - Python: `make py-test`
- **Coverage**:
  - Go: `make go-cov`
  - Python: `make py-cov`
- **New Features**: All new logic (extractors, metrics, templates) **must** include accompanying unit tests.

## 6. Security & Automation

- **CI/CD**: GitHub Actions handle extraction (`extraction.yml`) and linting/testing (`go_lint.yml`, `py_lint.yml`).
- **Data Integrity**: Handle unexpected data from external APIs (Google Sheets, Scrapers) gracefully.
- **Secrets**: API keys (Google Sheets, Mongo) are managed via environment variables. Do not commit secrets.
