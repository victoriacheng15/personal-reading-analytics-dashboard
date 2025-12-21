import asyncio
from unittest.mock import patch, MagicMock, AsyncMock
from utils import init_fetcher_state, fetch_page, close_fetcher


def test_init_fetcher_state_returns_dict():
    state = init_fetcher_state()
    assert isinstance(state, dict)
    assert "last_request_time" in state
    assert "request_interval" in state
    assert "client" in state
    # The client should be an httpx.AsyncClient
    import httpx

    assert isinstance(state["client"], httpx.AsyncClient)


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
