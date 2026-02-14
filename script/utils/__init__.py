"""
Utils package for the articles extractor application.

This module provides a clean interface for importing all utility functions
and constants used throughout the application.
"""

# Sheet operations
from .sheet import (
    get_creds_path,
    get_client,
    get_worksheet,
    get_all_providers,
    get_all_titles,
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
    get_strategy_handler,
    get_articles,
    wrap_with_error_handler,
    extract_substack_articles,
)

# Date and time utilities
from .format_date import clean_and_convert_date, current_time

# Constants
from .constants import (
    ARTICLES_WORKSHEET,
    PROVIDERS_WORKSHEET,
    DEFAULT_REQUEST_INTERVAL,
    DEFAULT_TIMEOUT,
    GOOGLE_SHEETS_SCOPES,
    DRY_RUN,
)

from .mongo import (
    get_mongo_client,
    insert_articles_event_mongo,
    insert_error_event_mongo,
    insert_summary_event_mongo,
    close_mongo_client,
)

__all__ = [
    # Sheet operations
    "get_creds_path",
    "get_client",
    "get_worksheet",
    "get_all_providers",
    "get_all_titles",
    "batch_append_articles",
    "SHEET_ID",
    # Web scraping
    "init_fetcher_state",
    "fetch_page",
    "close_fetcher",
    # Article extraction
    "get_strategy_handler",
    "get_articles",
    "wrap_with_error_handler",
    "extract_substack_articles",
    # Date utilities
    "clean_and_convert_date",
    "current_time",
    # Constants
    "ARTICLES_WORKSHEET",
    "PROVIDERS_WORKSHEET",
    "DEFAULT_REQUEST_INTERVAL",
    "DEFAULT_TIMEOUT",
    "GOOGLE_SHEETS_SCOPES",
    "DRY_RUN",
    # MongoDB operations
    "get_mongo_client",
    "insert_articles_event_mongo",
    "insert_error_event_mongo",
    "insert_summary_event_mongo",
    "close_mongo_client",
]
