# Extraction Architecture

The extraction layer is an **async web scraping ETL pipeline** that aggregates articles from multiple sources into centralized data stores.

## High-Level System Design

```mermaid
graph TD
        Sources["External Sources<br/>(RSS Feeds, HTML Blogs)"] -->|Async HTTP/2| Scraper(Python ETL Script<br/>script/main.py)
        Scraper -->|Extract & Normalize| UniversalExtractor(Universal Extractor<br/>utils/extractors.py)
        UniversalExtractor -->|Configuration Driven| SheetsSSOT["Google Sheets<br/>(providers worksheet)"]
        UniversalExtractor -->|Transform| Deduper{Deduplication}
        Deduper -->|Load Primary| Sheets["Google Sheets<br/>(articles worksheet)"]
        Deduper -->|Load Secondary & Event Log| Mongo["MongoDB<br/>(events collection)"]
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

### 3. Universal Extractor (`utils/extractors.py`)

The core of the "Zero-Code" onboarding. This component performs configuration-driven, heuristic-based extraction.

- **Metadata-Driven Governance:** Reads site-specific selectors, strategies (RSS, HTML, Substack), and other extraction parameters directly from the Google Sheets `providers` worksheet (the SSOT).
- **Heuristic Extraction:** Employs a "Link-First" strategy for title detection and a 5-tier discovery strategy for publication dates, making it resilient to minor DOM changes.
- **Dynamic Normalization:** Standardizes dates to ISO 8601 and dynamically maps/capitalizes source names based on live SSOT metadata, replacing static Go maps.
- **Operational Hardening:** Captures "Discovery Tier" metadata in MongoDB to audit heuristic performance in production.

### 4. Load Layer (`utils/sheet.py` & `mongo.py`)

Dual-write strategy for data redundancy.

- **Primary:** Batch writes to Google Sheets for the dashboard.
- **Secondary:** Optional batch writes to MongoDB for structured querying.

## Extraction Sequence

```mermaid
sequenceDiagram
    participant Orchestrator as script/main.py
    participant Sheets as Google Sheets
    participant Fetcher as utils/get_page.py
    participant Extractor as Universal Extractor
    participant Deduper as Deduplication
    participant Mongo as MongoDB

    Orchestrator->>Sheets: Fetch All Providers (SSOT)
    Orchestrator->>Sheets: Fetch Existing Titles (Dedup Cache)

    loop For Each Provider from SSOT
        Orchestrator->>Fetcher: Request Page (URL from SSOT)
        Fetcher-->>Orchestrator: Return HTML/RSS
        Orchestrator->>Extractor: Extract Article Tuples (using SSOT config & heuristics)
        Extractor-->>Orchestrator: Return New Articles + Discovery Tier
        Orchestrator->>Deduper: Deduplicate vs Cache
        Deduper-->>Orchestrator: Return Unique Articles
    end

    Orchestrator->>Sheets: Batch Write New Articles
    Orchestrator->>Sheets: Update Timestamp
    Orchestrator->>Mongo: Log Extraction Events (including Discovery Tier)
    Orchestrator->>Mongo: Log Summary & Error Events
```

## References

- **Data Schemas:** See [schemas.md](schemas.md) for Article tuples and MongoDB document definitions.
- **Automation:** See [operations.md](operations.md) for the daily extraction schedule.
