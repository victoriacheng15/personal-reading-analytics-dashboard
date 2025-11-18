"""
Utils package for the articles extractor application.

This module provides a clean interface for importing all utility functions
and constants used throughout the application.
"""

# Sheet operations
from .sheet import (
    get_client,
    get_worksheet,
    get_all_providers,
    get_all_titles,
    # append_article,
    batch_append_articles,
    SHEET_ID,
)

# Web scraping and page fetching
from .get_page import (
    init_fetcher_state,
    fetch_page,
    close_fetcher,
)

# Article extraction
from .extractors import (
    provider_dict,
    get_articles,
)

# Date and time utilities
from .format_date import current_time

# Constants
from .constants import (
    ARTICLES_WORKSHEET,
    PROVIDERS_WORKSHEET,
    DEFAULT_REQUEST_INTERVAL,
    DEFAULT_TIMEOUT,
    GOOGLE_SHEETS_SCOPES,
)

__all__ = [
    # Sheet operations
    "get_client",
    "get_worksheet",
    "get_all_providers",
    "get_all_titles",
    # "append_article",
    "batch_append_articles",
    "SHEET_ID",
    # Web scraping
    "init_fetcher_state",
    "fetch_page",
    "close_fetcher",
    # Article extraction
    "provider_dict",
    "get_articles",
    # Date utilities
    "current_time",
    # Constants
    "ARTICLES_WORKSHEET",
    "PROVIDERS_WORKSHEET",
    "DEFAULT_REQUEST_INTERVAL",
    "DEFAULT_TIMEOUT",
    "GOOGLE_SHEETS_SCOPES",
]
