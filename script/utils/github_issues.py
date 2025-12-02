"""
GitHub issue creation utility for tracking extraction failures.
"""

import os
import requests
import logging
from datetime import datetime
from dotenv import load_dotenv

logger = logging.getLogger(__name__)

load_dotenv()
TOKEN = os.environ.get("TOKEN")
if not TOKEN:
    raise ValueError("TOKEN environment variable is required")


def create_github_issue(site_name: str, error_msg: str, article_snippet: str = None):
    """
    Create a GitHub issue when extraction fails.

    Args:
        site_name (str): Name of the site where extraction failed
        error_msg (str): Error message or exception details
        article_snippet (str): Optional snippet of the article that failed
    """
    token = os.getenv("GITHUB_TOKEN")
    if not token:
        logger.warning(
            "GITHUB_TOKEN not set, skipping GitHub issue creation for "
            + site_name
        )
        return

    repo_owner = "victoriacheng15"
    repo_name = "personal-reading-analytics-dashboard"

    title = f"🚨 Extraction failed for {site_name}"
    body = f"""
**Site**: {site_name}

**Error**: 
```
{error_msg}
```

**Article snippet**: 
```
{article_snippet or "N/A"}
```

**Time**: {datetime.now().isoformat()}

---
*Auto-created by extraction script*
"""

    headers = {
        "Authorization": f"token {TOKEN}",
        "Accept": "application/vnd.github.v3+json",
    }

    data = {
        "title": title,
        "body": body,
        "labels": ["extraction-error", site_name.lower()],
    }

    url = f"https://api.github.com/repos/{repo_owner}/{repo_name}/issues"

    try:
        response = requests.post(url, json=data, headers=headers, timeout=10)
        if response.status_code == 201:
            issue_url = response.json().get("html_url")
            logger.info(f"Created GitHub issue for {site_name}: {issue_url}")
        else:
            logger.error(
                f"Failed to create GitHub issue for {site_name}: "
                f"HTTP {response.status_code}"
            )
    except Exception as e:
        logger.error(f"Error creating GitHub issue for {site_name}: {e}")
