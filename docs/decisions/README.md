# 🏗 Architectural Decisions

This directory contains the history of significant architectural shifts and technical decisions for the **Personal Reading Analytics Dashboard**. This project uses a combined ADR (Architecture Decision Record) and RFC (Request for Comments) approach to propose and record the "why" behind the technical evolution.

## 📋 Decision Log

| ID | Title | Status |
| :--- | :--- | :--- |
| **004** | [Universal Configuration-Driven Extraction](004-universal-configuration-driven-extraction.md) | `Accepted` |
| **003** | [Static Generation for Historical Metrics](003-static-historical-metrics.md) | `Accepted` |
| **002** | [Integrate AI Delta Analysis](002-integrate-ai-delta-analysis.md) | `Accepted` |
| **001** | [Prefer RSS/Atom Feeds over HTML Scraping](001-prefer-rss-over-html-scraping.md) | `Accepted` |

---

## 📄 ADR Template

New architectural decisions should follow the structure below:

```markdown
# ADR [00X]: [Descriptive Title]

- **Status:** Proposed | Accepted | Superseded
- **Date:** YYYY-MM-DD
- **Author:** Victoria Cheng

## Context and Problem Statement

What specific issue triggered this change?

## Decision Outcome

What was the chosen architectural path?

## Consequences

- **Positive:** (e.g., Faster development, resolved dependency drift).
- **Negative/Trade-offs:** (e.g., Added complexity to the CI/CD pipeline).

## Verification

- [ ] **Manual Check:** (e.g., Verified logs/UI locally).
- [ ] **Automated Tests:** (e.g., `make nix-go-test` passed).
```
