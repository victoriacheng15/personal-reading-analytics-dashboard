# 📚 Personal Reading Analytics Dashboard

A self-built fully automated reading analytics dashboard with zero infrastructure, refreshed automatically to turn personal data into actionable insights.

---

## 🌿 Design Philosophy

This project is built to reflect how I believe small, personal tools should work:

- **Zero infrastructure** → No servers, databases, or cloud costs. Runs entirely on GitHub (Actions + Pages).  
- **Fully automated** → Scheduled GitHub Actions keep data fresh—no manual runs or home servers.  
- **Cost-effective** → Uses only free tiers (GitHub, Google Sheets API)—proving powerful automation doesn’t require budget.

---

## 📊 What It Shows

- **Total articles**: 3000+ articles with read rate percentage  
- **Read vs. unread breakdown**: By year, by month, and by source (Substack, freeCodeCamp, GitHub, Shopify, Stripe)  
- **Yearly trends**: Year-by-year distribution with 5-year default view (adjustable)  
- **Monthly trends**: Monthly distribution across all sources or by individual source  
- **Interactive charts**: Year chart (bar/line toggle), monthly breakdown, and per-source statistics  
- **Source insights**: Per-source read rates, Substack author averages, and unread article tracking

### Supported Sources

Currently extracting articles from:

- freeCodeCamp
- Substack
- GitHub (Added 2024-03-18)
- Shopify (Added 2025-03-05)
- Stripe (Added 2025-11-19)  

---

## 🔗 Live Dashboard

👉 [Live Demo](https://victoriacheng15.github.io/personal-reading-analytics-dashboard/)

---

## 🛠 Tech Stacks

![Go](https://img.shields.io/badge/Go-00ADD8.svg?style=for-the-badge&logo=Go&logoColor=white)
![Python](https://img.shields.io/badge/Python-3776AB.svg?style=for-the-badge&logo=Python&logoColor=white)
![Google Sheets API](https://img.shields.io/badge/Google%20Sheets-34A853.svg?style=for-the-badge&logo=Google-Sheets&logoColor=white)
![Chart.js](https://img.shields.io/badge/Chart.js-FF6384.svg?style=for-the-badge&logo=Chart.js&logoColor=white)
![GitHub Actions](https://img.shields.io/badge/GitHub%20Actions-2088FF.svg?style=for-the-badge&logo=GitHub-Actions&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2496ED.svg?style=for-the-badge&logo=Docker&logoColor=white)

---

## 🛠️ Technical Overview

### Data Pipeline: Articles → Metrics → Dashboard

```mermaid
graph TD
    A["Article Extraction - Python"] -->|Extract & Deduplicate| B["Google Sheets"]
    B -->|Read Articles| C["Metrics Calculation - Go"]
    C -->|Generate JSON| D["metrics/YYYY-MM-DD.json"]
    D -->|Read Metrics| E["Dashboard Generation - Go"]
    E -->|Generate HTML| F["index.html with Chart.js"]
    F -->|Deploy| G["GitHub Pages for Live Dashboard"]
```

**Article Extraction**: Python web scrapers extract articles from engineering sources, deduplicate, and store in Google Sheets

**Metrics Calculation**: Go program reads articles from Google Sheets, calculates metrics, outputs JSON

**Dashboard Generation**: Go program reads metrics JSON, generates interactive HTML with Chart.js visualizations, deploys to GitHub Pages

### Documentation

For deep technical details, see the architecture docs:  

- [Extraction Pipeline Design](docs/extraction_architecture.md)  
- [Dashboard Pipeline Design](docs/dashboard_architecture.md)  
- [GitHub Actions Workflows](docs/github_actions.md)  
- [Jenkins CI/CD (Learning Experiment)](docs/jenkins.md) - Educational implementation only; production relies entirely on GitHub Actions.

---

## 📖 How This Project Evolved

Learn about the journey of this project: from local-only execution, to Docker containerization, to fully automated GitHub Actions workflows.

[Read Part 1: Article Extraction Pipeline](https://victoriacheng15.vercel.app/blog/from-pi-to-cloud-automation)

**Part 2: Dashboard & Metrics Pipeline** (Coming soon) - The evolution to metrics calculation and interactive visualization on GitHub Pages
