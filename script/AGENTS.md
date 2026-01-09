# Python Extraction Script (`/script`)

This directory contains the Python-based extraction pipeline responsible for scraping article metadata and persisting it to Google Sheets and MongoDB. The system is designed to run daily via the `Daily Extraction` GitHub Actions workflow.

## Operational Context

- **Schedule:** Daily at 06:00 UTC (via `.github/workflows/extraction.yml`).
- **Primary Goal:** Scrape configured sources, extract article metadata (Date, Title, Link, Category), and sync new entries to Google Sheets.
- **Observability:** Critical events (extractions, failures) are logged to MongoDB for historical tracking and debugging.

## Directory Structure

### `main.py`

The entry point for the extraction process. It orchestrates:

- Initializing connections (Google Sheets, MongoDB).

- Fetching configured providers asynchronously.

- Batch processing and deduplicating articles.

- Persisting results.

### `utils/`

Modularized utility functions organized by responsibility:

- **`extractors.py`**: The core logic for parsing HTML from various sources (e.g., Substack, freeCodeCamp) using `BeautifulSoup`.
- **`sheet.py`**: Handles all interactions with the Google Sheets API via `gspread`.
- **`mongo.py`**: Manages MongoDB connections and event logging via `pymongo`.
- **`get_page.py`**: Handles asynchronous web requests using `httpx` and `asyncio`.
- **`format_date.py`**: Standardizes date formats across different providers.
- **`constants.py`**: Central repository for configuration constants and scopes.

### `tests/`

Comprehensive unit tests using `pytest`.

- Mocks external services (Google Sheets, MongoDB, HTTP requests) to ensure deterministic testing.
- Validates extractor logic against sample HTML snippets.

### `migrations/`

Stores database migration scripts (e.g., schema changes for MongoDB documents).

- **Naming Convention:** `XXX_description.py` (e.g., `001_standardize_event_schema.py`).
- **Usage:** Run manually or via CI when schema evolution is required.

## Key Dependencies & Technologies

- **Runtime:** Python 3.12+
- **Async I/O:** `asyncio` + `httpx` for concurrent fetching of multiple sources.
- **Parsing:** `BeautifulSoup4` for robust HTML scraping.
- **Data Persistence:**
  - `gspread`: For Google Sheets integration.
  - `pymongo`: For event logging and error tracking in MongoDB.
- **Testing:** `pytest` for unit testing and coverage.

## Workflow Integration

The `extraction.yml` workflow manages the environment:

- Sets up Python 3.12.
- Installs dependencies from `requirements.txt`.
- Injects secrets (`SHEET_ID`, `MONGO_URI`, credentials) into the runtime environment.
- Executes `script/main.py`.
- **Local Docker Testing:** To safely test the Docker image (`make docker-run`) without writing to Google Sheets, temporarily comment out line 93 (`sheet.append_rows(rows)`) in `script/utils/sheet.py`. This allows validation of extraction and processing logic in a dry-run fashion.
