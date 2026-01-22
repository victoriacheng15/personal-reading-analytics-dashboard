# RFC [001]: Prefer RSS/Atom Feeds over HTML Scraping

- **Status:** Proposed
- **Date:** 2026-01-16
- **Author:** Victoria Cheng

## The Problem

The current extraction pipeline relies on scraping HTML DOM structures. This is brittle because:

- **High Maintenance:** CSS class names and nesting change frequently, breaking extractors without warning.
- **Execution Overhead:** HTML pages are significantly larger than XML feeds, increasing bandwidth and parsing time.
- **Complexity:** Storing specific CSS selectors in Google Sheets makes the system hard to audit and prone to manual entry errors.

## Proposed Solution

The project will prioritize **RSS/Atom feeds** as the primary extraction method for all providers that support them.

- **Primary Method:** Connect to the provider's `rss.xml` or `atom.xml` endpoint.
- **Fallback Method:** HTML scraping will only be used for providers that do not publish a discovery feed (e.g., Stripe, Shopify).
- **Implementation:** The system will check for a configured `rss_url` key in the provider handler. If present, it takes precedence over the standard URL.

## Comparison / Alternatives Considered

- **Alternative 1: Headless Browsing (Playwright/Puppeteer):** Could handle dynamic content better than simple scraping, but adds massive execution overhead and doesn't solve the "brittle CSS" problem.
- **Alternative 2: Professional Scraping APIs:** Reliable but introduces recurring costs and external vendor dependency.
- **Why this path?** RSS/Atom feeds are published API contracts. They provide a standardized, low-overhead way to fetch metadata that is inherently more stable than HTML structures.

## Failure Modes (Operational Excellence)

- **Feed Removal:** If a provider removes their RSS feed, the system will trigger a `fetch_failed` event in MongoDB.
- **Schema Mismatch:** If the XML structure deviates from RSS 2.0/Atom standards, the extractor will raise an `extraction_failed` event.
- **Observability:** These will be monitored via the `event_type` in the MongoDB `articles` collection to identify broken feeds immediately.

## Conclusion

Migrating to RSS improves the systemic reliability of the extraction pipeline. Migration will begin with GitHub and freeCodeCamp, while keeping HTML scraping as a legacy fallback for providers without feeds.
