# 📚 Personal Reading Analytics Dashboard

A self-built fully automated reading analytics dashboard with zero infrastructure, refreshed automatically to turn personal data into actionable insights.

---

## 🔗 Live Dashboard

👉 [See Live Dashbaord](https://victoriacheng15.github.io/personal-reading-analytics-dashboard/)

---

## 🌿 Design Philosophy

This project is built to reflect how I believe small, personal tools should work:

- **Zero infrastructure** → No servers or hosting costs. Runs entirely on GitHub (Actions + Pages).  
- **Fully automated** → Scheduled GitHub Actions keep data fresh—no manual runs or home servers.  
- **Observability first** → Uses an Event Sourcing pattern (MongoDB) to decouple extraction from analytics, ensuring full auditability and health monitoring.
- **Cost-effective** → Uses only free tiers (GitHub, Google Sheets API, MongoDB Atlas)—proving powerful automation doesn’t require budget.

---

## 🏗 Architecture & Documentation

Unlike typical "script-based" scrapers, this system is architected for scale and maintenance using an **Event Sourcing** pattern.

- **Ingestion**: Python scripts harvest content and emit standardized events to a MongoDB immutable log.
- **Observability**: An external hub consumes these events to populate Grafana dashboards for system health monitoring.
- **Visualization**: Go binaries process the event stream to generate the static site.

### 📈 System Observability

To demonstrate operational maturity, I maintain a public **[Observability Hub](https://victoriacheng15.github.io/observability-hub/dashboards.html)**.
This separate dashboard visualizes the "health" of this pipeline (ETL status, error rates, latencies) without requiring Grafana authentication.

For deep technical details, architectural diagrams, and operational guides, please visit the **[Documentation Hub](docs/README.md)**.

---

## 🛠 Tech Stacks

![Go](https://img.shields.io/badge/Go-00ADD8.svg?style=for-the-badge&logo=Go&logoColor=white)
![Python](https://img.shields.io/badge/Python-3776AB.svg?style=for-the-badge&logo=Python&logoColor=white)
![Google Sheets API](https://img.shields.io/badge/Google%20Sheets-34A853.svg?style=for-the-badge&logo=Google-Sheets&logoColor=white)
![MongoDB](https://img.shields.io/badge/MongoDB-47A248.svg?style=for-the-badge&logo=MongoDB&logoColor=white)
![Chart.js](https://img.shields.io/badge/Chart.js-FF6384.svg?style=for-the-badge&logo=Chart.js&logoColor=white)
![GitHub Actions](https://img.shields.io/badge/GitHub%20Actions-2088FF.svg?style=for-the-badge&logo=GitHub-Actions&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2496ED.svg?style=for-the-badge&logo=Docker&logoColor=white)

---

## 📊 What It Shows

**Key Metrics Section:**

- **Total articles**: Tracking total articles across currently supported sources
- **Read rate**: Percentage of articles completed with visual highlighting
- **Reading statistics**: Read count, unread count, and average articles per month
- **Highlight badges**: Top read rate source, most unread source, current month's read articles

**7 Interactive Visualizations (Chart.js):**

1. **Year Breakdown**: Bar chart showing article distribution by publication year
2. **Read/Unread by Year**: Stacked bar chart with reading progress across years
3. **Monthly Breakdown**: Toggle between total articles (line chart) and by-source distribution (stacked bar)
4. **Read/Unread by Month**: Seasonal reading patterns across all months
5. **Read/Unread by Source**: Horizontal stacked bars comparing progress per provider
6. **Unread Age Distribution**: Age buckets (<1 month, 1-3 months, 3-6 months, 6-12 months, >1 year)
7. **Unread by Year**: Identifies which years have the most unread backlog

**Source Analytics:**

- Per-source statistics with read/unread split and read percentages
- Substack per-author average calculation (total articles ÷ author count)
- Top 3 oldest unread articles with clickable links, dates, and age calculations
- Source metadata showing when each provider was added to tracking

### Supported Sources

Currently extracting articles from:

- freeCodeCamp
- Substack
- GitHub (Added 2024-03-18)
- Shopify (Added 2025-03-05)
- Stripe (Added 2025-11-19)

---

## 📖 How This Project Evolved

Learn about the journey of this project: from local-only execution, to Docker containerization, to fully automated GitHub Actions workflows.

- [Read Part 1: From Pi to Cloud Automation](https://victoriacheng15.vercel.app/blog/from-pi-to-cloud-automation)
- **Part 2: From Links to Reading Insights** (Coming soon) - The evolution to metrics calculation and interactive visualization on GitHub Pages
- **Part 3: From Metrics to Milestones** (Coming soon) - Visualizing technical growth and engineering milestones through a dynamic evolution timeline

---

## 🚀 Ready to Explore?

Don't just take my word for it—interact with the real data.

👉 **[Launch Personal Reading Analytics](https://victoriacheng15.github.io/personal-reading-analytics-dashboard/)**
