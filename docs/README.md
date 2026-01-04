# 📚 Project Documentation

Welcome to the technical documentation hub for the **Personal Reading Analytics Dashboard**. This collection of guides covers everything from high-level architecture to operational details and data schemas.

## 🏗 Architecture & Design

- **[Extraction Pipeline](architecture/extraction.md)**
  - Details the async Python ETL pipeline that scrapes articles and loads them into Google Sheets and MongoDB.
  - Covers the orchestration, extraction, transformation, and load layers.

- **[Dashboard Pipeline](architecture/dashboard.md)**
  - Explains the Go-based metrics generation and static site building process.
  - Includes the multi-page generation flow (Landing, Analytics, Evolution) and data inputs.

- **[Observability & Event Logging](architecture/event_logging.md)**
  - Outlines the standardized event architecture used for pipeline auditability and downstream ingestion.
  - Defines the structured "Envelope" schema and event lifecycle.

- **[Data Schemas](architecture/schemas.md)**
  - The single source of truth for all data models.
  - Defines the JSON contracts for Metrics, MongoDB documents, and the Evolution timeline configuration.

## ⚙️ Operations & DevOps

- **[Operations & CI/CD Guide](operations.md)**
  - The primary guide for maintaining the project.
  - Covers local development commands (`Makefile`), GitHub Actions workflows, secrets management, and failure recovery.

## 🧪 Experiments

- **[Jenkins CI/CD Pipeline](experiments/jenkins.md)**
  - A comparative study on implementing the build pipeline using Jenkins.
  - Documentation of a self-hosted alternative to the production GitHub Actions setup.
