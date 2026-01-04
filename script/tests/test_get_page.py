import asyncio
import time
from unittest.mock import patch, MagicMock, AsyncMock
from utils import init_fetcher_state, fetch_page, close_fetcher


def test_init_fetcher_state_returns_dict():
    state = init_fetcher_state()
    assert isinstance(state, dict)
    assert "last_request_time" in state
    assert "request_interval" in state
    assert "client" in state
    assert "lock" in state
    # The client should be an httpx.AsyncClient
    import httpx

    assert isinstance(state["client"], httpx.AsyncClient)
    assert isinstance(state["lock"], asyncio.Lock)


@patch("utils.get_page.httpx.AsyncClient.get", new_callable=AsyncMock)
def test_fetch_page_success(mock_get):
    # Mock response
    mock_response = MagicMock()
    mock_response.status_code = 200
    mock_response.text = "<html><body><h1>Test</h1></body></html>"
    mock_get.return_value = mock_response

    state = init_fetcher_state()
    soup, new_state = asyncio.run(fetch_page(state, "http://example.com"))
    assert soup is not None
    assert soup.h1.text == "Test"
    assert new_state["last_request_time"] >= state["last_request_time"]


@patch("utils.get_page.httpx.AsyncClient.get", new_callable=AsyncMock)
def test_fetch_page_non_200(mock_get):
    mock_response = MagicMock()
    mock_response.status_code = 404
    mock_response.text = "Not found"
    mock_get.return_value = mock_response

    state = init_fetcher_state()
    soup, new_state = asyncio.run(fetch_page(state, "http://example.com"))
    assert soup is None
    assert new_state["last_request_time"] >= state["last_request_time"]


@patch("utils.get_page.httpx.AsyncClient.get", new_callable=AsyncMock)
def test_fetch_page_exception(mock_get):
    mock_get.side_effect = Exception("Network error")
    state = init_fetcher_state()
    soup, new_state = asyncio.run(fetch_page(state, "http://example.com"))
    assert soup is None
    assert new_state == state


def test_close_fetcher_closes_client():
    state = init_fetcher_state()
    # Patch the aclose method
    state["client"].aclose = AsyncMock()
    asyncio.run(close_fetcher(state))
    state["client"].aclose.assert_awaited_once()


@patch("utils.get_page.httpx.AsyncClient.get", new_callable=AsyncMock)
def test_fetch_page_concurrency_lock_release(mock_get):
    """
    Verify that the lock is released BEFORE the network request.
    If the lock is held during the request, 3 requests taking 0.2s each would take >0.6s total.
    If parallel, they should take ~0.2s + minimal staggering.
    """

    # Simulate a slow network request
    async def slow_request(*args, **kwargs):
        await asyncio.sleep(0.2)
        mock_res = MagicMock()
        mock_res.status_code = 200
        mock_res.text = "<html></html>"
        return mock_res

    mock_get.side_effect = slow_request

    state = init_fetcher_state()
    # Set a tiny interval to allow fast firing but ensure rate limiting logic runs
    state["request_interval"] = 0.05

    async def run_concurrent():
        start = time.time()
        # Launch 3 requests
        tasks = [
            fetch_page(state, "http://ex.com/1"),
            fetch_page(state, "http://ex.com/2"),
            fetch_page(state, "http://ex.com/3"),
        ]
        await asyncio.gather(*tasks)
        return time.time() - start

    duration = asyncio.run(run_concurrent())

    # Analysis:
    # Req 1: Starts T+0.00, Network T+0.00 to T+0.20
    # Req 2: Starts T+0.00, Sleeps 0.05, Network T+0.05 to T+0.25
    # Req 3: Starts T+0.00, Sleeps 0.10, Network T+0.10 to T+0.30
    # Total expected time: ~0.30s
    #
    # If sequential (Buggy):
    # Req 1: T+0.00 to T+0.20
    # Req 2: T+0.20 to T+0.40 (wait satisfied by prev duration)
    # Req 3: T+0.40 to T+0.60
    # Total buggy time: ~0.60s

    assert duration < 0.55, (
        f"Requests took {duration}s, suggesting sequential execution (lock held during IO)"
    )
