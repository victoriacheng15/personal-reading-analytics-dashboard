# RFC [002]: Integrate AI-Powered Weekly Metrics Summary

- **Status:** Proposed
- **Date:** 2026-01-22
- **Author:** Victoria Cheng

## The Problem

The current analytics dashboard (`analytics.exe`) is **stateless regarding history**. It loads only the single most recent JSON metrics file (`loadLatestMetrics`) and renders a snapshot of the present moment.

- **Missing Feature:** There is **zero capability** for the visitor to see week-over-week changes or trends (e.g., "Read rate increased by 5%"). The data exists on disk, but the product provides no interface to access it.

## Proposed Solution

The project will integrate **Google Gemini (Generative AI)** to process the raw metrics and generate a qualitative weekly summary. This will be the **first and only** mechanism in the system to provide historical context.

- **Mechanism:** The `metrics.exe` binary will support two distinct operational modes via flags:
  - **`--fetch` (Workflow A):** Connects to Google Sheets, calculates stats, and saves the raw `YYYY-MM-DD.json`.
  - **`--summarize` (Workflow B):** Reads the latest local JSON, compares it with the previous week's file, generates the AI summary, and appends it to the *existing* JSON file.

    *Default behavior (no flags) runs A then B sequentially.*

- **Architecture:**
  - A new package `cmd/internal/ai` will isolate external API interactions.
  - The `metrics` package will remain the source of truth for data structure.
  - The system will use the official `google-generative-ai-go` SDK.

## Comparison / Alternatives Considered

- **Alternative 1: Manual JSON Diffing:** Requires the visitor to manually compare raw files. This is not user-friendly and doesn't scale for a public dashboard.
- **Alternative 2: Database-backed Trends:** Implementing a database (PostgreSQL/MongoDB) for historical stats. While robust, it adds significant infra complexity (hosting, migrations) to a project currently designed as a static pipeline.
- **Why this path?** AI-generated summaries provide high-signal narrative context without requiring a database or complex frontend charting for every possible delta.

## Failure Modes (Operational Excellence)

- **API Failure:** If Gemini is unreachable or the quota is exceeded, the system **must log a warning and proceed**. The core JSON metrics must still be saved. The pipeline cannot fail solely due to the AI step.
- **Missing History:** If no previous metrics file exists (first run), the system will generate a summary based solely on the current snapshot.
- **Privacy:** Only aggregated statistics are sent to the LLM. No full article content or personal identifiers are transmitted.

## Conclusion

This integration transforms the dashboard from a passive data display into an active feedback loop, providing visitors with actual analysis instead of just accounting. Next steps involve implementing the `ai` package and updating the `metrics` CLI.
