# Documentation Structure (`/docs`)

This directory (`/docs`) serves as the central repository for all project documentation, including architectural overviews, operational guides, and historical context. The goal is to provide high-signal, concise, and structured information to facilitate understanding and maintenance of the Personal Reading Analytics Dashboard.

## Purpose

The `docs/` folder is designed to:

- Provide comprehensive architectural diagrams and explanations.
- Detail operational procedures for deployment, monitoring, and maintenance.
- Store historical design decisions and archived documentation for context.
- Guide new contributors and remind existing ones about project conventions and standards.

## Directory Structure

- **`README.md`**: Top-level overview of the project, quick start guide, and general information.
- **`operations.md`**: Detailed guides for deploying, running, and troubleshooting the system.
- **`architecture/`**: Contains sub-documents detailing specific architectural components.
  - `dashboard.md`: Architecture of the Go-based dashboard service.
  - `event_logging.md`: Details on the MongoDB event logging system.
  - `extraction.md`: Architecture of the Python-based extraction pipeline.
  - `schemas.md`: Documentation for data schemas used across the system.
- **`archive/`**: Stores outdated or superseded documentation versions for historical reference. This ensures that old but potentially useful context is not lost.
  - `architecture-v1.md`, `github_actions-v1.md`, `installation-v1.md`, `README-v1.md`: Examples of archived documents.
- **`experiments/`**: Documentation related to experimental features, proof-of-concepts, or abandoned ideas that might still hold valuable insights.
  - `jenkins.md`: Example of an experimental document.
- **`agents.md` (this file)**: Describes the structure and purpose of the `/docs` directory itself.

## Documentation Principles

- **High-Signal & Concise:** Information is presented directly, without unnecessary fluff.
- **Structured:** Uses clear headings, bullet points, and code blocks for readability.
- **Up-to-Date:** Regularly reviewed and updated to reflect the current state of the project.
- **Markdown-centric:** All documentation is written in Markdown format.
