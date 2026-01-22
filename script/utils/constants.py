"""
Constants used throughout the articles extractor application.
"""

import os
from dotenv import load_dotenv

# Load environment variables from .env file
load_dotenv()

# Request settings
DEFAULT_REQUEST_INTERVAL = 1.0
DEFAULT_TIMEOUT = 30.0

# Execution mode
DRY_RUN = os.environ.get("DRY_RUN", "false").lower() == "true"

# Google Sheets settings
GOOGLE_SHEETS_SCOPES = ["https://www.googleapis.com/auth/spreadsheets"]

# Worksheet names
ARTICLES_WORKSHEET = "articles"
PROVIDERS_WORKSHEET = "providers"

# Canonical Source Names (Brand Case)
SOURCE_FREECODECAMP = "freeCodeCamp"
SOURCE_SUBSTACK = "Substack"
SOURCE_GITHUB = "GitHub"
SOURCE_SHOPIFY = "Shopify"
SOURCE_STRIPE = "Stripe"
