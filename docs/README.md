# 📚 Project Documentation

Welcome to the technical documentation hub for the **Personal Reading Analytics Dashboard**. This collection of guides covers everything from high-level architecture to operational details and data schemas.

## 🏗 Architecture & Design

- **[Extraction Pipeline Design](extraction_architecture.md)**
  - Details the async Python ETL pipeline that scrapes articles and loads them into Google Sheets and MongoDB.
  - Covers the orchestration, extraction, transformation, and load layers.

- **[Dashboard Pipeline Design](dashboard_architecture.md)**
  - Explains the Go-based metrics generation and static site building process.
  - Includes the multi-page generation flow (Landing, Analytics, Evolution) and data inputs.

- **[Data Schemas](schemas.md)**
  - The single source of truth for all data models.
  - Defines the JSON contracts for Metrics, MongoDB documents, and the Evolution timeline configuration.

## ⚙️ Operations & DevOps

- **[Operations & CI/CD Guide](operations.md)**
  - The primary guide for maintaining the project.
  - Covers local development commands (`Makefile`), GitHub Actions workflows, secrets management, and failure recovery.

- **[Jenkins CI/CD Experiment](jenkins.md)**
  - A comparative study on implementing the build pipeline using Jenkins.
  - Documentation of a self-hosted alternative to the production GitHub Actions setup.
