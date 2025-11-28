# Installation Guide

The app is currently capable of extracting articles from:

- freeCodeCamp
- Substack
- The GitHub Blog
- Shopify (Added 2025-03-05)
- Stripe (Added 2025-11-19)

Support for additional sites may be added in the future.

## Quick Start

Clone the repository and navigate to the project directory:

```bash
git clone git@github.com:victoriacheng15/articles-extractor.git

cd articles-extractor
```

## Prerequisites

### Set Up Credentials & Environment

- **Google Sheets API**:  
  - Follow Google’s [API quickstart guide](https://developers.google.com/sheets/api/quickstart/python) to download `credentials.json`.  
  - Place `credentials.json` in the project’s root directory.  

- **Configure `.env`**:  

 ```bash
 cp .env.example .env  # Copy template
 ```

 Edit `.env` with:  

 ```ini
 SHEET_ID="your_google_sheet_id_here"  # Found in your Sheet’s URL
 ```

---

### Add Providers to Google Sheet

- **Create a worksheet**:  
  - Name the sheet `providers` (exact spelling, lowercase).

- **Build your provider list**:

  Create a table with these columns in your Google Sheet:

  | name        | element (CSS selector/class)                        | url                                         |
  |-------------|-----------------------------------------------------|----------------------------------------------|
  | freecodecamp| article                                             | <https://www.freecodecamp.org/news/>           |
  | github      | article                                             | <https://github.blog/category/engineering/>    |
  | shopify     | article                                             | <https://shopify.engineering/latest>          |
  | substack    | pencraft pc-display-flex pc-flexDirection-column pc-gap-4 | <https://[your-substack].substack.com/archive> |
  | substack    | pencraft pc-display-flex pc-flexDirection-column pc-gap-4 | <https://[your-substack].substack.com/archive> |
  | substack    | pencraft pc-display-flex pc-flexDirection-column pc-gap-4 | <https://[your-substack].substack.com/archive> |

  **Note**: Replace `[your-substack]` with your actual Substack domain.

---

## Deployment

Choose **one** of these methods to run the app:

> ℹ️ **Tip**: Only use GitHub Actions if you want automated scheduling. For local or Docker deployments, you can disable the workflow in your repo's Actions settings to avoid unnecessary runs.

---

### 1. Manual Local Run (Simple & Direct)

Run the app directly on your machine using Python.

**Quick Setup**:

```bash
python3 -m venv venv
source venv/bin/activate  # Linux/Mac
# venv\Scripts\activate   # Windows
```

**Install and Run**:
Refer to the [Makefile](../Makefile) for available commands:

```bash
make install   # Install dependencies
make run       # Run the extraction script
```

Or run directly:

```bash
pip install -r requirements.txt
python3 main.py
```

---

### 2. Docker + Cron Scheduling (Self-Hosted)

Deploy using Docker and schedule with cron. Perfect for Raspberry Pi, NAS, or any always-on machine.

Refer to the [Makefile](../Makefile) for Docker commands:

```bash
make up       # Start Docker container
make logs     # View container logs
make down     # Stop Docker container
```

**Schedule with cron**:

```bash
crontab -e
```

Add a line for your desired schedule (example: daily at 9 AM):

```bash
0 9 * * * cd ~/path_to_project/articles-extractor && make up
```

Use [crontab.guru](https://crontab.guru/) to customize the timing.

> Replace `~/path_to_project` with your project's absolute path. Run `pwd` in the terminal to find it.

---

### 3. GitHub Actions (Cloud-Based)

Automatically run the app on GitHub's servers. No additional setup needed beyond configuring secrets.

**Add secrets** in repo settings (`Settings > Secrets and variables > Actions`):

- `CREDENTIALS`: Paste the entire content of your `credentials.json` file
- `SHEET_ID`: Your Google Sheet ID

**Workflow details** (already configured in `.github/workflows/scheduled_extraction.yml`):

The workflow will:

- Run automatically every day at **6 AM UTC** (or manually via `workflow_dispatch`)
- Set up Python 3.10 and install dependencies
- Create a temporary `credentials.json` from your secrets
- Run the extraction script (`make run`)
- Save extraction logs as workflow artifacts
- Clean up sensitive files

You can customize the schedule by editing the cron value in `.github/workflows/scheduled_extraction.yml`:

```yaml
on:
  schedule:
    - cron: '0 6 * * *'  # Daily at 6 AM UTC (change this as needed)
  workflow_dispatch:     # Manual trigger anytime
```

View logs in your repo's **Actions** tab → **Daily Extraction Schedule** → select a workflow run.

> **Security Note**: Never commit `credentials.json` or `.env` files!
