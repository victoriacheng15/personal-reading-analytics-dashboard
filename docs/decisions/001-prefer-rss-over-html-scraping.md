# ADR 001: Prefer RSS/Atom Feeds over HTML Scraping

- **Status:** Accepted
- **Date:** 2026-01-16
- **Author:** Victoria Cheng

## Context and Problem Statement

The current extraction pipeline relies on scraping HTML DOM structures. This is brittle because CSS class names and nesting change frequently, breaking extractors without warning. HTML pages are also significantly larger than XML feeds, increasing bandwidth and parsing time. Storing specific CSS selectors in Google Sheets makes the system hard to audit and prone to manual entry errors.

## Decision Outcome

The project will prioritize **RSS/Atom feeds** as the primary extraction method for all providers that support them.

- **Primary Method:** Connect to the provider's `rss.xml` or `atom.xml` endpoint (configured via the existing `url` column in the providers sheet).
- **Fallback Method:** HTML scraping will only be used for providers that do not publish a discovery feed (e.g., Stripe, Shopify).
- **Implementation:** Instead of separate URL keys, the system uses **Dual-Mode Extractors**. The `provider_dict` defines multiple targets (e.g., `[provider_element, "item"]`), and the extractor function dynamically branches logic based on whether it encounters an RSS `<item>` or a standard HTML element.

## Consequences

- **Positive:**
  - **Stability:** RSS feeds are stable API contracts, unlike DOM structures.
  - **Performance:** Reduced bandwidth and parsing overhead.
  - **Simplicity:** Removes the need for complex CSS selector maintenance for supported providers.
- **Negative/Trade-offs:**
  - Limited by provider feed availability (not all engineering blogs provide full-text feeds or any feed at all).

## Verification

- [x] **Manual Check:** Inspect `script/utils/extractors.py` to verify `provider_dict` contains dual-mode configurations (e.g., for `freeCodeCamp` and `GitHub`).
- [x] **Automated Tests:** Run `pytest script/tests/test_extractors.py` and verify `test_extract_fcc_articles_rss_success` and `test_extract_github_articles_rss_success` pass.
