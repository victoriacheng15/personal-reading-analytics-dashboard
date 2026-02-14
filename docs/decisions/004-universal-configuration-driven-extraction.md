# ADR 004: Universal Configuration-Driven Extraction

- **Status:** Accepted
- **Date:** 2026-02-13
- **Author:** Victoria Cheng

## Context and Problem Statement

As the project scaled from a handful of data sources to dozens of engineering blogs, the "specific-case" architecture—where each blog required its own Python extraction function—became a significant bottleneck. This model required a code deployment for every new provider and increased technical debt through duplicated parsing logic. We needed a way to onboard new blogs with "Zero-Code" effort while maintaining high data fidelity.

## Decision Outcome

The project will transition to a **Universal Configuration-Driven Extraction Engine** to decouple site-specific logic from the application codebase.

- **Heuristic Engine:** Implement a functional universal handler (`universal_html_extractor`) that uses "Link-First" heuristics for title detection and a 5-tier discovery strategy for publication dates.
- **Metadata-Driven Governance:** Shift the "Source of Truth" for extraction parameters from hardcoded Python logic to the `providers` worksheet in the Google Sheets SSOT.
- **Dynamic Normalization:** Refactor the Go backend to dynamically map and capitalize source names based on live SSOT metadata, replacing static Go maps.
- **Operational Hardening:** Integrate `asyncio.Semaphore` for concurrency control and capture "Discovery Tier" metadata in MongoDB to audit heuristic performance in production.

## Consequences

- **Positive:**
  - **Rapid Scalability:** New engineering blogs can be onboarded instantly via a metadata update in the Google Sheet with zero code changes.
  - **Maintenance Reduction:** Deleted over 40% of technical debt by removing site-specific Python functions and hardcoded mappings.
  - **System Resilience:** Heuristic-driven discovery is more resilient to minor DOM changes compared to rigid, hardcoded CSS selectors.
- **Negative/Trade-offs:**
  - **Heuristic Ambiguity:** Reliance on discovery heuristics can lead to misidentification in highly irregular layouts, necessitating explicit JSON selector overrides in the SSOT.

## Verification

- [x] **Automated Tests:** Verified all 77 Python tests pass, including the new `test_universal_extractor.py` suite.
- [x] **Containerized Integration:** Successfully executed `make run` (Docker) to confirm that existing providers (Stripe, Shopify, GitHub, FCC) are correctly processed via the universal engine.
- [x] **Observability Audit:** Confirmed that MongoDB event documents now include `meta.discovery_tier` for fine-grained auditing of extraction methods.
