# GitHub Actions Workflows Guide

This document describes all GitHub Actions workflows that automate the Personal Reading Analytics Dashboard project.

## Overview

The project uses five automated workflows organized by development pipeline: **code quality validation** → **data collection** → **metrics calculation** → **visualization deployment**:

1. **Go Lint & Test Checking** - Validates Go code formatting and runs tests (on PR & push to main)
2. **Python Ruff and Test Checking** - Validates Python code quality and runs tests (on PR & push to main)
3. **Daily Extraction** - Extracts articles from web sources (daily at 6am UTC)
4. **Weekly Metrics Generation** - Calculates metrics from Google Sheets (Friday at 1am UTC)
5. **Dashboard Deployment** - Generates and deploys dashboard (on push to main when metrics/** or cmd/** change)

## 1. Go Lint & Test Checking Workflow

**File**: `.github/workflows/go_lint.yml`

### Purpose 1

Validates Go code formatting, runs static analysis checks, and executes unit tests on pull requests and main branch pushes. First quality gate for Go code. Uses **Go 1.24**.

### Trigger 1

- **Pull request**: Triggered when PR targets main branch with changes to `cmd/**` paths
- **Push to main**: Triggered when code pushed to main with `cmd/**` changes

### Execution Steps 1

**Two Jobs**: `lint` (runs first) → `test` (runs after lint passes)

```mermaid
graph TD
    A["Job: lint<br/>Run Tests"] --> B{"Pass?"}
    B -->|No| C["Fail"]
    B -->|Yes| D["Check gofmt<br/>Format check"]
    D --> E{"Issues?"}
    E -->|Yes| C
    E -->|No| F["Run go vet<br/>Static analysis"]
    F --> G{"Issues?"}
    G -->|Yes| C
    G -->|No| H["Job: test<br/>Run Tests Again"]
    H --> I{"Pass?"}
    I -->|No| C
    I -->|Yes| J["Pass"]
```

### Output 1

- **Tests**: ✅ All tests pass (or ❌ test failures with output)
- **Format**: ✅ Go code is properly formatted
- **Analysis**: ✅ No static analysis issues found
- **Failure**: ❌ Detailed error output + suggestion to run `gofmt -w ./cmd/` or fix failing tests
- **Exit code**: 0 (pass) or 1 (fail, blocks merge)

### Validation Checks 1

- **go test**: Unit tests for `cmd/internal/metrics` and other packages
- **gofmt**: Code style consistency (spacing, indentation, naming conventions)
- **go vet**: Static analysis for common bugs and issues

### Key Features 1

- **Pre-merge validation**: Prevents failing tests or improperly formatted code from merging
- **Test-first**: Tests run before format/lint checks (fail fast)
- **Auto-suggestion**: Provides fix command in error message
- **Scope-specific**: Only runs when Go files in `cmd/**` are modified
- **No modifications**: Only checks, doesn't auto-fix
- **Comprehensive coverage**: Validates functionality, formatting, and code quality in one workflow

---

## 2. Python Ruff and Test Checking Workflow

**File**: `.github/workflows/py_lint.yml`

### Purpose 2

Validates Python code quality using Ruff linter and runs pytest tests on pull requests and main branch pushes. Second quality gate for Python code. Uses **Python 3.12**.

### Trigger 2

- **Pull request**: Triggered when PR targets main branch with changes to `script/**` paths
- **Push to main**: Triggered when code pushed to main with `script/**` changes

### Execution Steps 2

**Two Jobs**: `ruff` (runs first) → `test` (runs after ruff passes)

```mermaid
graph TD
    A["Job: ruff<br/>Install Ruff"] --> B["Check Quality<br/>ruff check script/ --diff"]
    B --> C{"Issues?"}
    C -->|Yes| D["Fail<br/>Show diff"]
    C -->|No| E["Job: test<br/>Install dependencies"]
    E --> F["Run pytest<br/>script/ tests"]
    F --> G{"Pass?"}
    G -->|No| D
    G -->|Yes| H["Pass"]
```

### Output 2

- **Success**: Exit code 0 (no violations)
- **Exit code**: 1 + diff showing violations (blocks merge)

### Validation Checks 2

- **Code quality** (Ruff): Unused imports, undefined names, syntax errors
- **Style violations** (Ruff): PEP 8 compliance issues
- **Common pitfalls** (Ruff): Potential bugs and anti-patterns
- **Unit tests** (pytest): All tests in `script/tests/` must pass

### Key Features 2

- **Diff output**: Shows exact violations with context
- **Scope-specific**: Only runs when Python files in `src/` are modified
- **Pre-merge validation**: Prevents code quality regression
- **No modifications**: Only checks, doesn't auto-fix

---

## 3. Daily Extraction Workflow

**File**: `.github/workflows/extraction.yml`

### Purpose 3

Runs Python web scraping script to extract articles from configured providers and loads into dual stores (Google Sheets + MongoDB if configured). Keeps data fresh daily. Uses **Python 3.12**.

### Trigger 3

- **Schedule**: Daily at 6:00 AM UTC
- **Manual trigger**: `workflow_dispatch` button in GitHub UI

### Execution Steps 3

```mermaid
graph TD
    A["Install Deps<br/>requirements.txt"] --> B["Setup Auth<br/>credentials.json"]
    B --> C["Verify Auth<br/>Check file exists"]
    C --> D["Run Extraction<br/>python3 script/main.py"]
    D --> E{"Success?"}
    E -->|Yes| F["Update Sheets<br/>Google Sheets"]
    E -->|No| G["Error"]
    F --> H{"MongoDB<br/>configured?"}
    H -->|Yes| I["Update MongoDB<br/>Batch insert"]
    H -->|No| J["Cleanup<br/>Delete credentials.json"]
    I --> J
    G --> J
```

### Output 3

- **Google Sheets**: Updated with new articles (primary storage)
- **MongoDB**: Updated with new article documents (if `MONGO_URI` configured)
- **Logs**: Available in GitHub Actions workflow run logs (stdout)
- **No artifacts**: Logs not uploaded as files, visible in workflow UI only

### Key Features 3

- **Daily automation**: Runs without manual intervention at 6am UTC
- **Dual-store writes**: Updates Google Sheets (always) and MongoDB (if configured)
- **Async extraction**: Uses Python asyncio with HTTP/2 for efficient scraping
- **Credentials cleanup**: Deletes credentials.json after use (never committed)
- **Verification step**: Checks credentials.json exists before running script
- **Error visibility**: Script failures visible in GitHub Actions workflow logs

### Environment Variables 3

Required:
- `SHEET_ID` - From `secrets.SHEET_ID` (Google Sheet ID for articles)
- `CREDENTIALS` - From `secrets.CREDENTIALS` (Google service account JSON, written to credentials.json)

Optional (for MongoDB dual-store):
- `MONGO_URI` - From `secrets.MONGO_URI` (MongoDB connection string, if set)
- `MONGO_DB_NAME` - From `secrets.MONGO_DB_NAME` (MongoDB database name, if set)
- `MONGO_COLLECTION_NAME` - From `secrets.MONGO_COLLECTION_NAME` (MongoDB collection name, if set)

---

## 4. Weekly Metrics Generation Workflow

**File**: `.github/workflows/metrics_generation.yml`

### Purpose 4

Fetches articles from Google Sheets API, calculates comprehensive metrics, and creates a pull request with the results. Happens after articles have been collected throughout the week.

### Trigger 4

- **Schedule**: Friday at 1:00 AM UTC (weekly)
- **Manual trigger**: `workflow_dispatch` button in GitHub UI
- **No automatic push**: Uses PR for review before merging to main

### Permissions 4

```yaml
permissions:
  contents: write         # Create/push branches
  pull-requests: write    # Create pull requests
```

### Execution Steps 4

```mermaid
graph TD
    A["Setup Auth<br/>credentials.json"] --> B["Verify Auth<br/>Check file exists"]
    B --> C["Create .env<br/>SHEET_ID + CREDENTIALS_PATH"]
    C --> D["Build & Run<br/>make run-metrics"]
    D --> E["Cleanup<br/>Delete credentials.json"]
    E --> F["Git Config<br/>github-actions[bot]"]
    F --> G["Create Branch<br/>metrics/weekly-update"]
    G --> H["Commit & Push<br/>metrics/ folder"]
    H --> I["Create PR<br/>gh pr create"]
```

### Output 4

- **Branch**: `metrics/weekly-update` (recreated each run)
- **Files changed**: `metrics/YYYY-MM-DD.json` (new metrics file)
- **Pull Request**: Auto-created with title "chore: weekly metrics update"
- **Artifact**: None (results stored in branch)

### Key Features 4

- **Credentials handling**: Creates file from secret, deletes after use (never committed)
- **Safe push**: Uses `--force-with-lease` to prevent overwriting concurrent updates
- **PR-based**: Requires approval before merging (respects branch protection)
- **Idempotent PR creation**: Uses `|| true` to avoid error if PR already exists
- **Error recovery**: Continues pipeline even if no changes (via `|| echo`)

### Environment Variables 4

Set as GitHub Secrets:

- `CREDENTIALS` - Google service account JSON (entire file as string)
- `SHEET_ID` - Google Sheet ID to fetch articles from

---

## 5. Dashboard Deployment Workflow

**File**: `.github/workflows/deployment.yml`

### Purpose 5

Generates interactive HTML dashboard from latest metrics JSON and deploys to GitHub Pages. Final step in the pipeline—makes data publicly visible. Uses **Go 1.23**.

### Trigger 5

- **On push to main**: Triggers when changes pushed to `main` branch affecting:
  - `metrics/**` (new metrics files)
  - `cmd/**` (dashboard code changes)
  - `.github/workflows/deployment.yml` (workflow changes)
- **Manual trigger**: `workflow_dispatch` button in GitHub UI
- **No schedule**: Runs only on push, not on a weekly schedule

### Permissions 5

```yaml
permissions:
  contents: read         # Read repository code
  pages: write           # Deploy to GitHub Pages
  id-token: write        # OpenID Connect token for deployment
```

### Concurrency

```yaml
concurrency:
  group: "pages"
  cancel-in-progress: false    # Wait for previous deploy to complete
```

### Execution Steps 5

```mermaid
graph TD
    A["Build Dashboard<br/>dashboard"] --> B["Read Metrics<br/>Latest JSON"]
    B --> C{"Metrics?"}
    C -->|No| D["Fail"]
    C -->|Yes| E["Generate HTML<br/>index.html"]
    E --> F["Setup Pages<br/>Config"]
    F --> G["Upload Artifact<br/>site/"]
    G --> H["Deploy Pages<br/>Live"]
```

### Output 5

- **Dashboard**: `site/index.html` (generated locally, uploaded as artifact)
- **GitHub Pages URL**: `https://victoriacheng15.github.io/personal-reading-analytics/`
- **Deployment time**: ~1-2 minutes
- **Artifact**: Temporary (auto-deleted after deployment)

### Key Features 5

- **Latest metrics**: Automatically uses most recent `metrics/YYYY-MM-DD.json`
- **No credentials needed**: Dashboard generator reads only local JSON files
- **Concurrency protection**: Only one deploy at a time (prevents race conditions)
- **Automatic environment**: GitHub Pages environment auto-configured
- **Safe deployment**: Previous deployments are not cancelled mid-stream

### Deployment Requirements

Dashboard deployment triggers automatically when metrics PR is merged to main (which updates `metrics/**` files). This ensures latest data is visualized immediately after metrics are available.

---

## Workflow Coordination

```mermaid
graph TD
    DailyExt["☀️ Daily Extraction<br/>6am UTC daily"]
    WeeklyMetrics["📊 Weekly Metrics<br/>Friday 1am UTC"]
    AutoDeploy["🚀 Deploy Dashboard<br/>On metrics/ push"]
    CodeQA["✓ Code QA<br/>On PR & push to main"]
    
    DailyExt -->|Updates Google Sheets| WeeklyMetrics
    WeeklyMetrics -->|Creates PR| Review["Manual Review<br/>& Merge"]
    Review -->|Merges to main| AutoDeploy
    CodeQA -->|Validates| WeeklyMetrics
    CodeQA -->|Validates| DailyExt
    AutoDeploy -->|Live Dashboard| Pages["🌐 GitHub Pages"]
```

### Data Flow Timeline

**Week cycle:**

1. **Daily** (6am UTC): Articles extracted → Google Sheets + MongoDB updated
2. **Friday 1am UTC**: Metrics calculated → PR created (`metrics/weekly-update` branch) → awaits review
3. **Manual review**: Developer reviews and merges metrics PR to main
4. **Immediately on merge**: Dashboard deployment triggered automatically (push to main with `metrics/**` changes)
5. **~2 minutes later**: Live dashboard with latest data available at GitHub Pages URL

### Failure Modes & Recovery

| Workflow | Failure | Recovery |
|----------|---------|----------|
| Extraction | API error | Manual retry via `workflow_dispatch` or wait for next daily run |
| Metrics generation | Google Sheets unreachable | PR not created, retry next Friday or manually trigger |
| Dashboard deploy | Metrics file missing | Fails with error, waits for metrics PR to be merged |
| Code QA | Formatting issue | Fix code and push, workflow re-runs automatically |

---

## GitHub Secrets Configuration

Required secrets in GitHub repository Settings → Secrets and variables → Actions:

| Secret | Example | Purpose | Required |
|--------|---------|---------|----------|
| `CREDENTIALS` | `{"type": "service_account", ...}` | Google service account JSON (entire JSON as string) | Yes |
| `SHEET_ID` | `1a2b3c4d5e6f7g8h` | Google Sheet ID for articles data | Yes |
| `MONGO_URI` | `mongodb+srv://user:pass@cluster.mongodb.net/` | MongoDB connection string | Optional |
| `MONGO_DB_NAME` | `reading_analytics` | MongoDB database name | Optional |
| `MONGO_COLLECTION_NAME` | `articles` | MongoDB collection name | Optional |

### Setup Instructions

1. Navigate to repository Settings → Secrets and variables → Actions
2. Click "New repository secret"
3. Add `CREDENTIALS`: Paste entire Google service account JSON
4. Add `SHEET_ID`: Paste Google Sheet ID
5. Save and workflows will use them automatically

---

## Monitoring & Debugging

### View Workflow Runs

1. Go to **Actions** tab in GitHub
2. Select workflow name
3. Click run to see detailed logs

### Common Issues

**Metrics workflow fails to create PR:**

- Check `CREDENTIALS` and `SHEET_ID` secrets are set
- Verify Google service account has Sheets API access
- Check branch protection rules allow github-actions[bot] to push
- Ensure `make run-metrics` command works locally

**Dashboard fails to deploy:**

- Ensure metrics PR is merged to main (triggers deployment automatically)
- Check GitHub Pages is enabled in Settings → Pages → Source: GitHub Actions
- Verify `site/index.html` was generated correctly
- Check that `metrics/` folder has at least one JSON file

**Extraction workflow runs but no data appears:**

- Check workflow logs for "Failed to fetch page" warnings
- Verify providers worksheet in Google Sheets has correct URLs
- For MongoDB: Check `MONGO_URI`, `MONGO_DB_NAME`, `MONGO_COLLECTION_NAME` secrets if dual-store expected

**Code QA fails:**

- Go linting: Run `make gofmt` locally to format code
- Go tests: Run `go test ./cmd/... -v` locally
- Python linting: Run `ruff check script/` locally to see violations
- Python tests: Run `python3 -m pytest script/` locally

### Logs Location

- **GitHub UI**: Actions tab → workflow run → step logs
- **Artifacts**: Actions tab → workflow run → Artifacts section
- **Retention**: 7 days
