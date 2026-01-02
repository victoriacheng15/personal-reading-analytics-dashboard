import httpx
import asyncio
import logging
import time
from bs4 import BeautifulSoup
from .constants import DEFAULT_REQUEST_INTERVAL, DEFAULT_TIMEOUT

logger = logging.getLogger(__name__)


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
        "lock": asyncio.Lock(),
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
    wait_time = 0
    async with state["lock"]:
        now = time.time()
        # Calculate time since the last reserved slot
        elapsed = now - state["last_request_time"]
        
        # If we are too fast, schedule a wait
        if elapsed < state["request_interval"]:
            wait_time = state["request_interval"] - elapsed
        
        # Reserve this slot by updating the time to when this request effectively 'starts'
        # If wait_time > 0, we push the 'last' time into the future
        state["last_request_time"] = now + wait_time

    # Sleep outside the lock to allow other tasks to schedule their slots
    if wait_time > 0:
        await asyncio.sleep(wait_time)

    try:
        response = await state["client"].get(url)
        
        if response.status_code == 200:
            soup = BeautifulSoup(response.text, "html.parser")
            return soup, state

        return None, state

    except Exception as _:
        return None, state


async def close_fetcher(state):
    """
    Close the HTTP client stored in the state.

    Args:
        state (dict): Fetcher state containing the client.
    """
    await state["client"].aclose()
