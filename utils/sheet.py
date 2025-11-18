import os
from typing import Set, List, Dict, Any, Callable
import gspread
from gspread import Worksheet
from google.oauth2.service_account import Credentials
from dotenv import load_dotenv
from .constants import GOOGLE_SHEETS_SCOPES

load_dotenv()
SHEET_ID = os.environ.get("SHEET_ID")
if not SHEET_ID:
    raise ValueError("SHEET_ID environment variable is required")
scopes = GOOGLE_SHEETS_SCOPES


def get_creds_path() -> str:
    """
    Returns the path to the credentials.json file.
    """
    return os.path.join(os.path.dirname(__file__), "..", "credentials.json")


def get_client() -> gspread.Client:
    creds_path = get_creds_path()
    creds = Credentials.from_service_account_file(creds_path, scopes=scopes)
    return gspread.authorize(creds)


def get_worksheet(client: gspread.Client, sheet_id: str, sheet_name: str) -> Worksheet:
    """
    Opens and returns a specific worksheet by name.

    Args:
        client (gspread.Client): The authenticated gspread client.
        sheet_id (str): The ID of the Google Sheet.
        sheet_name (str): The name of the worksheet to open.

    Returns:
        Worksheet: The worksheet object

    Raises:
        ValueError: If sheet_id or sheet_name is empty
        gspread.exceptions.SpreadsheetNotFound: If the sheet doesn't exist
    """
    if not sheet_id or not sheet_name:
        raise ValueError("sheet_id and sheet_name cannot be empty")

    sheet = client.open_by_key(sheet_id)
    return sheet.worksheet(sheet_name)


def get_all_titles(articles_sheet: Worksheet) -> Set[str]:
    """
    Retrieves all article titles from the 'articles' sheet, excluding the header.

    Args:
        articles_sheet (Worksheet): The worksheet for articles.

    Returns:
        set: A set of all titles for O(1) lookup.
    """
    all_titles = articles_sheet.get_all_values()[1:]
    return {row[1] for row in all_titles}


def get_all_providers(providers_sheet: Worksheet) -> List[Dict[str, Any]]:
    """
    Retrieves all provider data from the 'providers' sheet.

    Args:
        providers_sheet (Worksheet): The worksheet for providers.

    Returns:
        list: List of dicts, each representing one row (provider info).
    """
    return providers_sheet.get_all_records()


# def append_article(
#     sheet: Worksheet, article_info: tuple, log_func: Callable = print
# ) -> None:
#     """
#     Appends a new article row to the given sheet and logs it.

#     Args:
#         sheet (Worksheet): The worksheet to update.
#         article_info (tuple): The article data as (date, title, link, source).
#         log_func (function, optional): Function used to log output. Defaults to print.
#     """
#     date = article_info[0]
#     title = article_info[1]
#     link = article_info[2]
#     log_func(f"==> {title} - {date}\n{link}\n")
#     sheet.append_row(list(article_info))


def batch_append_articles(
    sheet: Worksheet, articles: List[tuple], log_func: Callable = print
) -> None:
    """
    Appends multiple article rows to the given sheet in a single batch operation.

    Args:
        sheet (Worksheet): The worksheet to update.
        articles (list): List of article tuples as (date, title, link, source).
        log_func (function, optional): Function used to log output. Defaults to print.
    """
    if not articles:
        return

    # Log all articles
    for article_info in articles:
        date = article_info[0]
        title = article_info[1]
        link = article_info[2]
        log_func(f"==> {title} - {date}\n{link}\n")

    # Batch append all rows at once
    rows = [list(article) for article in articles]
    # log_func(f"\n📋 Batch rows to append:\n{rows}\n")
    # TODO: Uncomment to actually append to Google Sheets
    sheet.append_rows(rows)
