import httpx
import asyncio
import logging
import sys
import time
from bs4 import BeautifulSoup
from .constants import DEFAULT_REQUEST_INTERVAL, DEFAULT_TIMEOUT

logger = logging.getLogger(__name__)
# Configure logging to write to stdout for log file capture
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    stream=sys.stdout,
)


def init_fetcher_state():
    """
    Initialize the fetcher state with last request time, request interval, and an HTTP client.

    Returns:
        dict: A dictionary containing the fetcher state.
    """
    return {
        "last_request_time": 0.0,
        "request_interval": DEFAULT_REQUEST_INTERVAL,
        "client": httpx.AsyncClient(timeout=DEFAULT_TIMEOUT, http2=True),
    }


async def fetch_page(state, url):
    """
    Fetch and parse a webpage with rate limiting.

    Args:
        state (dict): Current fetcher state.
        url (str): URL to fetch.

    Returns:
        tuple: (BeautifulSoup object or None, updated state)
    """
    elapsed = time.time() - state["last_request_time"]
    if elapsed < state["request_interval"]:
        await asyncio.sleep(state["request_interval"] - elapsed)

    try:
        response = await state["client"].get(url)
        state["last_request_time"] = time.time()

        if response.status_code == 200:
            soup = BeautifulSoup(response.text, "html.parser")
            return soup, state

        logger.error(f"HTTP {response.status_code} from {url}")
        return None, state

    except Exception as e:
        logger.error(f"Error fetching {url}: {str(e)}")
        return None, state


async def close_fetcher(state):
    """
    Close the HTTP client stored in the state.

    Args:
        state (dict): Fetcher state containing the client.
    """
    await state["client"].aclose()
