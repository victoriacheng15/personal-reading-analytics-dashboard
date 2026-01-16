# Operations & CI/CD Guide

This document covers the operational aspects of the project, including local development commands (Makefile) and the automated CI/CD pipeline (GitHub Actions).

## 1. Local Development (Makefile)

The `Makefile` is the primary entry point for running tasks locally.

| Command | Description |
| :--- | :--- |
| `make run-metrics` | Fetches data from Google Sheets and generates `metrics/YYYY-MM-DD.json`. |
| `make run-dashboard` | Generates the HTML dashboard in `site/index.html` using the latest metrics. |
| `make cleanup` | Removes compiled binaries (`metricsjson`, `dashboard`) and test coverage files. |
| `make go-test` | Runs all Go unit tests with verbose output. |
| `make go-coverage` | Runs Go tests and generates a coverage report. |
| `make gofmt` | Formats all Go code in `cmd/`. |

## 2. CI/CD Pipeline Overview

The project uses five automated workflows to handle quality control, data extraction, metrics generation, and deployment.

```mermaid
graph TD
    subgraph "1. Quality Gate (Validation)"
        A[Push/PR] --> B{Go Lint & Test}
        A --> C{Python Lint & Test}
    end

    subgraph "2. Data Ingestion (Python)"
        D[Scheduled Workflow] -->|Extract| E[Update Google Sheets]
        D -->|Log Events| F[Update MongoDB]
    end

    subgraph "3. Metrics & Deployment (Go)"
        G[Scheduled Workflow] -->|Fetch from Sheets| H[Metrics JSON]
        H -->|PR| I[Review & Merge]
        I -->|Push to Main| J[Deploy Dashboard]
        J --> K[GitHub Pages]
    end

    %% Tech Stack Flow
    B --> G
    C --> D
    
    %% Data Dependency
    E -.-> G
```

### Workflow Reference Table

| Workflow | File | Trigger | Purpose |
| :--- | :--- | :--- | :--- |
| **Go Check** | `go_lint.yml` | PR, Push (`cmd/**`) | Validates Go formatting (`gofmt`), static analysis (`go vet`), and tests. |
| **Python Check** | `py_lint.yml` | PR, Push (`script/**`) | Validates Python code using `ruff` and runs `pytest`. |
| **Daily Extraction** | `extraction.yml` | Schedule (Daily 6am), Manual | Scrapes articles and updates Google Sheets/MongoDB. |
| **Weekly Metrics** | `metrics_generation.yml` | Schedule (Fri 1am), Manual | Calculates metrics and opens a PR with a new JSON file. |
| **Deploy** | `deployment.yml` | Push (`metrics/**`), Manual | Builds static HTML and deploys to GitHub Pages. |

## 3. Configuration & Secrets

All sensitive configuration is managed via GitHub Secrets.

| Secret Name | Required | Description |
| :--- | :--- | :--- |
| `CREDENTIALS` | **Yes** | Google Service Account JSON (full content). |
| `SHEET_ID` | **Yes** | ID of the Google Sheet used for storage. |
| `MONGO_URI` | **Yes** | Connection string for MongoDB (Event Logging). |
| `MONGO_DB_NAME` | **Yes** | MongoDB Database Name. |
| `MONGO_COLLECTION_NAME` | **Yes** | MongoDB Collection Name. |

## 4. Failure Recovery

| Issue | Resolution |
| :--- | :--- |
| **Extraction Fails** | Check `extraction.yml` logs for API errors. Retry manually via `workflow_dispatch`. |
| **Metrics PR Missing** | Check `metrics_generation.yml` logs. Verify `SHEET_ID` access. Run `make run-metrics` locally to debug. |
| **Deploy Fails** | Ensure `metrics/` folder has JSON files. Check `deployment.yml` logs for template errors. |
| **Linting Fails** | Run `make gofmt` or `ruff check script/` locally and commit fixes. |
