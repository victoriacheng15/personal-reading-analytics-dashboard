import asyncio
from unittest.mock import Mock, patch, AsyncMock, MagicMock
from main import process_provider, async_main, main

# Mock data
MOCK_PROVIDER = {
    "name": "test_provider",
    "url": "http://test.com",
    "element": "article",
}

MOCK_FETCHER_STATE = {"client": "mock_client"}


@patch("main.get_articles")
@patch("main.fetch_page", new_callable=AsyncMock)
@patch("main.get_strategy_handler")
def test_process_provider_success(
    mock_get_strategy_handler, mock_fetch_page, mock_get_articles
):
    """Test successful processing of a provider"""
    mock_soup = Mock()
    mock_soup.find_all.return_value = ["element1", "element2"]

    # Setup mocks
    mock_handler = {"element": lambda: "article", "extractor": Mock()}
    mock_get_strategy_handler.return_value = mock_handler
    mock_fetch_page.return_value = (mock_soup, MOCK_FETCHER_STATE)
    mock_get_articles.return_value = [("2025-01-01", "Title", "Link", "Source", 1)]

    # Execute
    semaphore = asyncio.Semaphore(1)
    articles, state = asyncio.run(
        process_provider(MOCK_FETCHER_STATE, MOCK_PROVIDER, set(), semaphore)
    )

    # Verify
    assert len(articles) == 1
    assert state == MOCK_FETCHER_STATE
    mock_fetch_page.assert_called_once_with(MOCK_FETCHER_STATE, "http://test.com")
    mock_soup.find_all.assert_called_once()
    mock_get_articles.assert_called_once()


@patch("main.get_strategy_handler")
def test_process_provider_unknown_provider(mock_get_strategy_handler):
    """Test processing with unknown provider"""
    mock_get_strategy_handler.return_value = None

    semaphore = asyncio.Semaphore(1)
    articles, state = asyncio.run(
        process_provider(MOCK_FETCHER_STATE, MOCK_PROVIDER, set(), semaphore)
    )

    assert articles == []
    assert state == MOCK_FETCHER_STATE


@patch("main.DRY_RUN", False)
@patch("main.get_mongo_client")
@patch("main.fetch_page", new_callable=AsyncMock)
@patch("main.get_strategy_handler")
def test_process_provider_fetch_failure(
    mock_get_strategy_handler, mock_fetch_page, mock_get_mongo_client
):
    """Test processing when page fetch fails"""
    mock_handler = {"element": lambda: "article", "extractor": Mock()}
    mock_get_strategy_handler.return_value = mock_handler
    mock_fetch_page.return_value = (None, MOCK_FETCHER_STATE)
    mock_get_mongo_client.return_value = None  # Prevent MongoDB connection

    semaphore = asyncio.Semaphore(1)
    articles, state = asyncio.run(
        process_provider(MOCK_FETCHER_STATE, MOCK_PROVIDER, set(), semaphore)
    )

    assert articles == []
    assert state == MOCK_FETCHER_STATE


@patch("main.DRY_RUN", False)
@patch("main.get_mongo_client")
@patch("main.insert_summary_event_mongo")
@patch("main.close_fetcher", new_callable=AsyncMock)
@patch("main.batch_append_articles")
@patch("main.process_provider", new_callable=AsyncMock)
@patch("main.init_fetcher_state")
@patch("main.get_all_providers")
@patch("main.get_all_titles")
@patch("main.get_worksheet")
@patch("main.get_client")
def test_async_main_success(
    mock_get_client,
    mock_get_worksheet,
    mock_get_titles,
    mock_get_providers,
    mock_init_state,
    mock_process,
    mock_batch_append,
    mock_close,
    mock_insert_summary,
    mock_get_mongo_client,
):
    """Test the main async flow with new articles"""
    # Setup mocks
    mock_sheet = Mock()
    mock_get_worksheet.return_value = mock_sheet
    mock_get_providers.return_value = [MOCK_PROVIDER]
    mock_process.return_value = (
        [("2025-01-01", "Title", "Link", "Source", 1)],
        MOCK_FETCHER_STATE,
    )
    mock_client = MagicMock()  # Use MagicMock which supports __getitem__ better
    # Configure mock_client to return a mock database when accessed via []
    mock_db = MagicMock()
    mock_collection = Mock()
    mock_db.__getitem__.return_value = mock_collection  # For db[collection_name]
    mock_client.__getitem__.return_value = mock_db  # For client[db_name]

    mock_get_mongo_client.return_value = mock_client

    # Execute
    asyncio.run(async_main("2025-01-01 - 12:00"))

    # Verify
    mock_batch_append.assert_called_once()
    mock_sheet.sort.assert_called_once_with((1, "des"))
    mock_sheet.update_cell.assert_called_once()
    mock_close.assert_called_once()
    mock_insert_summary.assert_called_once_with(mock_client, 1)


@patch("main.DRY_RUN", False)
@patch("main.get_mongo_client")
@patch("main.insert_summary_event_mongo")
@patch("main.close_fetcher", new_callable=AsyncMock)
@patch("main.batch_append_articles")
@patch("main.process_provider", new_callable=AsyncMock)
@patch("main.init_fetcher_state")
@patch("main.get_all_providers")
@patch("main.get_all_titles")
@patch("main.get_worksheet")
@patch("main.get_client")
def test_async_main_no_articles(
    mock_get_client,
    mock_get_worksheet,
    mock_get_titles,
    mock_get_providers,
    mock_init_state,
    mock_process,
    mock_batch_append,
    mock_close,
    mock_insert_summary,
    mock_get_mongo_client,
):
    """Test the main async flow with no new articles"""
    mock_sheet = Mock()
    mock_get_worksheet.return_value = mock_sheet
    mock_get_providers.return_value = [MOCK_PROVIDER]
    mock_process.return_value = ([], MOCK_FETCHER_STATE)
    mock_client = Mock()
    mock_get_mongo_client.return_value = mock_client

    asyncio.run(async_main("timestamp"))

    mock_batch_append.assert_not_called()
    mock_sheet.sort.assert_called_once()
    mock_insert_summary.assert_called_once_with(mock_client, 0)


@patch("main.async_main")
@patch("main.asyncio.run")
def test_main_wrapper(mock_run, mock_async_main):
    """Test the synchronous main wrapper"""
    main("timestamp")

    mock_run.assert_called_once()
    mock_async_main.assert_called_once_with("timestamp")
