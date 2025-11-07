# Article Extractor

An **automated data pipeline** that orchestrates ETL (Extract, Transform, Load) workflows across multiple article sources—freeCodeCamp, Substack, GitHub Engineering, and Shopify Engineering. The project aggregates and deduplicates content into a centralized Google Sheet using **serverless architecture** (GitHub Actions), showcasing modern DataOps practices without requiring dedicated infrastructure.

**Deployment Options:**

- **Serverless**: Scheduled GitHub Actions workflows (infrastructure-free)
- **Traditional**: Cron jobs or Docker containers
- **Manual**: CLI execution for ad-hoc runs

## ✨ Key Features

- **Serverless Data Pipeline**: Scheduled GitHub Actions workflows eliminate infrastructure overhead
- **Multi-Source ETL**: Extract, transform, and aggregate articles from 4+ providers
- **Data Deduplication**: Intelligent duplicate detection and skipping for data quality
- **Google Sheets Integration**: OAuth-authenticated API for reliable data delivery
- **Flexible Deployment**: GitHub Actions (serverless), Docker, cron, or CLI execution
- **Production-Grade Error Handling**: Provider-level fault tolerance with comprehensive logging

## 🚀 Quick Start

Please refer to this [Installation Guide](docs/installation.md)

## 📚 Documentation

- [Architecture Guide](docs/architecture.md) - System design, components, and data flow
- [GitHub Actions](docs/github_actions.md) - Workflow automation and scheduling

## 🛠 Tech Stacks

![Python](https://img.shields.io/badge/Python-3.10+-3776AB.svg?style=for-the-badge&logo=Python&logoColor=white)
![Google Sheets API](https://img.shields.io/badge/Google%20Sheets-34A853.svg?style=for-the-badge&logo=Google-Sheets&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2496ED.svg?style=for-the-badge&logo=Docker&logoColor=white)
![Raspberry Pi](https://img.shields.io/badge/Raspberry%20Pi-A22846.svg?style=for-the-badge&logo=Raspberry-Pi&logoColor=white)
![GitHub Actions](https://img.shields.io/badge/GitHub%20Actions-2088FF.svg?style=for-the-badge&logo=GitHub-Actions&logoColor=white)

## 💡 Technical Insights

### Python Generators for Responsive Data Pipelines

I refactored the data flow to use Python generators instead of collecting articles in memory before processing. This architectural choice demonstrates understanding of streaming data principles:

- **Original approach**: Collected all articles in a list → processed → uploaded (batch processing)
- **Refactored approach**: Each article flows through the pipeline immediately after extraction → uploaded (streaming processing)

**Benefits realized:**

- **Memory efficiency**: Eliminates temporary storage; scales to large datasets
- **Responsive UX**: Data appears in Google Sheets incrementally, not in one batch dump
- **Fault resilience**: Pipeline interruptions don't waste prior work; partial runs produce value
- **Natural sequencing**: Generators inherently enforce sequential completion (extract → transform → load), making data flow explicit and maintainable

This demonstrates design thinking beyond "just getting it to work"—considering data flow architecture, memory management, and operational resilience.

## 📖 How This Project Evolved

Learn about the journey of this project: from local-only execution, to Docker containerization, to fully automated GitHub Actions workflows.

[Read the blog post](https://victoriacheng15.vercel.app/blog/from-pi-to-cloud-automation)
