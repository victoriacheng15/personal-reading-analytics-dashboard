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
    provider_dict,
    get_articles,
    extractor_error_handler,
    extract_fcc_articles,
    extract_substack_articles,
    extract_github_articles,
    extract_shopify_articles,
    extract_stripe_articles,
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
    "provider_dict",
    "get_articles",
    "extractor_error_handler",
    "extract_fcc_articles",
    "extract_substack_articles",
    "extract_github_articles",
    "extract_shopify_articles",
    "extract_stripe_articles",
    # Date utilities
    "clean_and_convert_date",
    "current_time",
    # Constants
    "ARTICLES_WORKSHEET",
    "PROVIDERS_WORKSHEET",
    "DEFAULT_REQUEST_INTERVAL",
    "DEFAULT_TIMEOUT",
    "GOOGLE_SHEETS_SCOPES",
]
