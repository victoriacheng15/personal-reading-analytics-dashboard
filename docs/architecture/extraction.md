# Extraction Architecture

The extraction layer is an **async web scraping ETL pipeline** that aggregates articles from multiple sources into centralized data stores.

## High-Level System Design

```mermaid
graph TD
    Sources[External Sources<br/>freeCodeCamp, Substack, etc.] -->|Async HTTP/2| Scraper(Python ETL Script<br/>script/main.py)
    Scraper -->|Extract| Parsers[HTML Parsers<br/>BeautifulSoup]
    Parsers -->|Transform| Deduper{Deduplication}
    Deduper -->|Load| Sheets[Google Sheets<br/>Primary Store]
    Deduper -->|Load| Mongo[MongoDB<br/>Secondary Store]
```

## Core Components

### 1. Orchestration Layer (`script/main.py`)

Coordinates the end-to-end async pipeline.

- **Concurrency:** Uses `asyncio` to manage non-blocking I/O.
- **Lifecycle:** Manages authentication, fetcher state, and final database writes.

### 2. Extraction Layer (`utils/get_page.py`)

Handles network interactions with resilience.

- **Features:** HTTP/2 support, connection pooling, and stateful rate-limiting (1s intervals).
- **Safety:** Graceful degradation on timeouts or non-200 responses.

### 3. Transformation Layer (`utils/extractors.py`)

Provider-specific logic to turn raw HTML into structured data.

- **Routing:** Selects the correct parser based on the domain.
- **Normalization:** Standardizes dates to ISO 8601 and titles to lowercase for deduplication.

### 4. Load Layer (`utils/sheet.py` & `mongo.py`)

Dual-write strategy for data redundancy.

- **Primary:** Batch writes to Google Sheets for the dashboard.
- **Secondary:** Optional batch writes to MongoDB for structured querying.

## Extraction Sequence

```mermaid
sequenceDiagram
    participant Main as Orchestrator
    participant Fetch as Async Fetcher
    participant Parse as Extractor
    participant DB as Google Sheets

    Main->>DB: Fetch Existing Titles (Dedup Cache)
    loop For Each Provider
        Main->>Fetch: Request Page (HTTP/2)
        Fetch-->>Main: Return HTML DOM
        Main->>Parse: Extract Article Tuples
        Parse-->>Main: Return New Articles
        Main->>Main: Deduplicate vs Cache
    end
    Main->>DB: Batch Write New Articles
    Main->>DB: Update Timestamp
```

## References

- **Data Schemas:** See [schemas.md](schemas.md) for Article tuples and MongoDB document definitions.
- **Automation:** See [operations.md](operations.md) for the daily extraction schedule.
