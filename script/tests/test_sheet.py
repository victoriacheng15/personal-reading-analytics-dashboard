import pytest
from unittest.mock import patch, Mock

# Import functions from utils package
from utils import (
    get_creds_path,
    get_client,
    get_worksheet,
    get_all_titles,
    get_all_providers,
    batch_append_articles,
)


# Tests for get_creds_path function
def test_get_creds_path_returns_string():
    """Test that get_creds_path returns a string"""
    result = get_creds_path()
    assert isinstance(result, str)


def test_get_creds_path_ends_with_credentials_json():
    """Test that the returned path ends with credentials.json"""
    result = get_creds_path()
    assert result.endswith("credentials.json")


def test_get_creds_path_contains_correct_path():
    """Test that the path is constructed correctly relative to the script"""
    result = get_creds_path()
    assert "credentials.json" in result


# Tests for get_client function
@patch("utils.sheet.Credentials.from_service_account_file")
@patch("utils.sheet.gspread.authorize")
def test_get_client_success(mock_authorize, mock_creds):
    """Test that get_client successfully creates a gspread client"""
    mock_creds.return_value = Mock()
    mock_client = Mock()
    mock_authorize.return_value = mock_client

    result = get_client()

    assert result == mock_client
    mock_creds.assert_called_once()
    mock_authorize.assert_called_once()


@patch("utils.sheet.Credentials.from_service_account_file")
@patch("utils.sheet.gspread.authorize")
def test_get_client_uses_correct_scopes(mock_authorize, mock_creds):
    """Test that get_client uses the correct scopes"""
    mock_creds.return_value = Mock()
    mock_authorize.return_value = Mock()

    get_client()

    # Verify that from_service_account_file was called with scopes
    args, kwargs = mock_creds.call_args
    assert "scopes" in kwargs


# Tests for get_worksheet function
@patch("utils.sheet.gspread.Client.open_by_key")
def test_get_worksheet_success(mock_open):
    """Test that get_worksheet successfully opens a worksheet"""
    mock_sheet = Mock()
    mock_worksheet = Mock()
    mock_sheet.worksheet.return_value = mock_worksheet
    mock_open.return_value = mock_sheet

    client = Mock()
    client.open_by_key = mock_open

    result = get_worksheet(client, "sheet_id_123", "Articles")

    assert result == mock_worksheet
    mock_open.assert_called_once_with("sheet_id_123")
    mock_sheet.worksheet.assert_called_once_with("Articles")


def test_get_worksheet_raises_error_with_empty_sheet_id():
    """Test that get_worksheet raises ValueError with empty sheet_id"""
    client = Mock()
    with pytest.raises(ValueError, match="sheet_id and sheet_name cannot be empty"):
        get_worksheet(client, "", "Articles")


def test_get_worksheet_raises_error_with_empty_sheet_name():
    """Test that get_worksheet raises ValueError with empty sheet_name"""
    client = Mock()
    with pytest.raises(ValueError, match="sheet_id and sheet_name cannot be empty"):
        get_worksheet(client, "sheet_id_123", "")


def test_get_worksheet_raises_error_with_none_sheet_id():
    """Test that get_worksheet raises ValueError with None sheet_id"""
    client = Mock()
    with pytest.raises(ValueError, match="sheet_id and sheet_name cannot be empty"):
        get_worksheet(client, None, "Articles")


def test_get_worksheet_raises_error_with_none_sheet_name():
    """Test that get_worksheet raises ValueError with None sheet_name"""
    client = Mock()
    with pytest.raises(ValueError, match="sheet_id and sheet_name cannot be empty"):
        get_worksheet(client, "sheet_id_123", None)


# Tests for get_all_titles function
def test_get_all_titles_extracts_second_column():
    """Test that get_all_titles extracts titles from the second column"""
    mock_sheet = Mock()
    mock_sheet.get_all_values.return_value = [
        ["ID", "Title", "Link"],
        ["1", "Article One", "http://example.com/1"],
        ["2", "Article Two", "http://example.com/2"],
        ["3", "Article Three", "http://example.com/3"],
    ]

    result = get_all_titles(mock_sheet)

    assert result == {"Article One", "Article Two", "Article Three"}


def test_get_all_titles_returns_set():
    """Test that get_all_titles returns a set"""
    mock_sheet = Mock()
    mock_sheet.get_all_values.return_value = [
        ["ID", "Title"],
        ["1", "Article One"],
    ]

    result = get_all_titles(mock_sheet)

    assert isinstance(result, set)


def test_get_all_titles_skips_header():
    """Test that get_all_titles skips the header row"""
    mock_sheet = Mock()
    mock_sheet.get_all_values.return_value = [
        ["ID", "Title"],
        ["1", "Article One"],
    ]

    result = get_all_titles(mock_sheet)

    assert "ID" not in result
    assert "Title" not in result


def test_get_all_titles_empty_sheet():
    """Test that get_all_titles handles empty sheet with only header"""
    mock_sheet = Mock()
    mock_sheet.get_all_values.return_value = [["ID", "Title"]]

    result = get_all_titles(mock_sheet)

    assert result == set()


# Tests for get_all_providers function
def test_get_all_providers_returns_list_of_dicts():
    """Test that get_all_providers returns a list of dictionaries"""
    mock_sheet = Mock()
    expected_providers = [
        {"name": "Provider A", "url": "http://provider-a.com"},
        {"name": "Provider B", "url": "http://provider-b.com"},
    ]
    mock_sheet.get_all_records.return_value = expected_providers

    result = get_all_providers(mock_sheet)

    assert result == expected_providers
    assert isinstance(result, list)
    assert all(isinstance(item, dict) for item in result)


def test_get_all_providers_calls_get_all_records():
    """Test that get_all_providers calls get_all_records method"""
    mock_sheet = Mock()
    mock_sheet.get_all_records.return_value = []

    get_all_providers(mock_sheet)

    mock_sheet.get_all_records.assert_called_once()


def test_get_all_providers_empty_sheet():
    """Test that get_all_providers handles empty sheet"""
    mock_sheet = Mock()
    mock_sheet.get_all_records.return_value = []

    result = get_all_providers(mock_sheet)

    assert result == []


# Tests for batch_append_articles function
def test_batch_append_articles_calls_append_rows():
    """Test that batch_append_articles calls append_rows with correct format"""
    mock_sheet = Mock()
    articles = [
        ("2025-01-15", "Article One", "http://example.com/1", "Source A"),
        ("2025-01-16", "Article Two", "http://example.com/2", "Source B"),
    ]

    batch_append_articles(mock_sheet, articles)

    mock_sheet.append_rows.assert_called_once()
    args, kwargs = mock_sheet.append_rows.call_args
    rows = args[0]

    assert len(rows) == 2
    assert rows[0] == ["2025-01-15", "Article One", "http://example.com/1", "Source A"]
    assert rows[1] == ["2025-01-16", "Article Two", "http://example.com/2", "Source B"]


def test_batch_append_articles_returns_none_on_empty():
    """Test that batch_append_articles returns early with empty list"""
    mock_sheet = Mock()
    articles = []

    result = batch_append_articles(mock_sheet, articles)

    assert result is None
    mock_sheet.append_rows.assert_not_called()


def test_batch_append_articles_returns_none_on_none():
    """Test that batch_append_articles returns early with None"""
    mock_sheet = Mock()

    result = batch_append_articles(mock_sheet, None)

    assert result is None
    mock_sheet.append_rows.assert_not_called()


def test_batch_append_articles_single_article():
    """Test batch_append_articles with a single article"""
    mock_sheet = Mock()
    articles = [("2025-01-15", "Article One", "http://example.com/1", "Source A")]

    batch_append_articles(mock_sheet, articles)

    mock_sheet.append_rows.assert_called_once()
    args, kwargs = mock_sheet.append_rows.call_args
    rows = args[0]

    assert len(rows) == 1
    assert rows[0] == ["2025-01-15", "Article One", "http://example.com/1", "Source A"]
