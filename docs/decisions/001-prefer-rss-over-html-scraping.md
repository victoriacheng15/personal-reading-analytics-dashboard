# RFC [001]: Prefer RSS/Atom Feeds over HTML Scraping

- **Status:** Proposed
- **Date:** 2026-01-16
- **Author:** Victoria Cheng

## The Problem

The current extraction pipeline relies on scraping HTML DOM structures. This is brittle because:

- **High Maintenance:** CSS class names and nesting change frequently, breaking extractors without warning.
- **Execution Overhead:** HTML pages are significantly larger than XML feeds, increasing bandwidth and parsing time.
- **Complexity:** Storing specific CSS selectors in Google Sheets makes the system hard to audit and prone to manual entry errors.

## Decision

The project will prioritize **RSS/Atom feeds** as the primary extraction method for all providers that support them.

- **Primary Method:** Connect to the provider's `rss.xml` or `atom.xml` endpoint.
- **Fallback Method:** HTML scraping will only be used for providers that do not publish a discovery feed (e.g., Stripe, Shopify).
- **Implementation:** The system will check for a configured `rss_url` key in the provider handler. If present, it takes precedence over the standard URL.

## Consequences

### Positive

- **Reliability:** RSS is a published API contract. It rarely changes structure, significantly reducing "random" failures.
- **Performance:** XML payloads are smaller and faster to parse than full HTML pages.
- **Standardization:** Date formats in RSS (RFC 822) are more consistent than arbitrary HTML text (e.g., "Jan 15", "2 days ago").

### Negative

- **Inconsistency:** The system will maintain two parallel extraction logic paths (XML parsing vs. HTML scraping) indefinitely, as not all providers support RSS.
- **Metadata Limitations:** RSS feeds sometimes provide limited content summaries compared to the full metadata available on the HTML page (though usually sufficient for Title/Link/Date).

## Failure Modes (Operational Excellence)

- **Feed Removal:** If a provider removes their RSS feed, the system will trigger a `fetch_failed` event in MongoDB.
- **Schema Mismatch:** If the XML structure deviates from RSS 2.0/Atom standards, the extractor will raise an `extraction_failed` event.
- **Observability:** These will be monitored via the `event_type` in the MongoDB `articles` collection to identify broken feeds immediately.

## Conclusion

Migrating to RSS improves the systemic reliability of the extraction pipeline. Migration will begin with GitHub and freeCodeCamp, while keeping HTML scraping as a legacy fallback for providers without feeds.
