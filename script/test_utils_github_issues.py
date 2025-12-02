"""
Unit tests for GitHub issue creation functionality using pytest.
"""

import pytest
from unittest.mock import patch, MagicMock
from utils import create_github_issue


@pytest.fixture
def mock_getenv_token():
    """Fixture to mock os.getenv with token."""
    import os
    original_getenv = os.getenv
    
    def mock_getenv(key, default=None):
        if key == "GITHUB_TOKEN":
            return "test_token_12345"
        return original_getenv(key, default)
    
    with patch("utils.github_issues.os.getenv", side_effect=mock_getenv):
        yield


@pytest.fixture
def mock_post():
    """Fixture to mock requests.post."""
    with patch("utils.github_issues.requests.post") as mock:
        yield mock


def test_create_github_issue_success(mock_getenv_token, mock_post):
    """Test successful GitHub issue creation."""
    # Mock successful response
    mock_response = MagicMock()
    mock_response.status_code = 201
    mock_response.json.return_value = {
        "html_url": "https://github.com/victoriacheng15/personal-reading-analytics-dashboard/issues/42"
    }
    mock_post.return_value = mock_response

    create_github_issue("github", "Test extraction error", "Test snippet")

    # Verify POST was called
    mock_post.assert_called_once()

    # Get the call arguments
    call_args = mock_post.call_args
    url = call_args[0][0] if call_args[0] else call_args[1].get("url")
    data = call_args[1].get("json")

    # Verify the request details
    assert "github.com/repos/victoriacheng15" in url
    assert data["title"] == "🚨 Extraction failed for github"
    assert "github" in data["labels"]
    assert "extraction-error" in data["labels"]
    assert "Test extraction error" in data["body"]
    assert "Test snippet" in data["body"]


def test_create_github_issue_http_error(mock_getenv_token, mock_post):
    """Test handling of HTTP errors."""
    mock_response = MagicMock()
    mock_response.status_code = 401
    mock_post.return_value = mock_response

    # Should not raise, just log error
    create_github_issue("stripe", "Authentication failed")

    # Verify POST was called
    mock_post.assert_called_once()


def test_create_github_issue_network_error(mock_getenv_token, mock_post):
    """Test handling of network errors."""
    mock_post.side_effect = Exception("Network connection failed")

    # Should not raise, just log error
    create_github_issue("substack", "Network error")

    # Verify POST was called
    mock_post.assert_called_once()


def test_create_github_issue_no_token():
    """Test that function exits gracefully without GITHUB_TOKEN."""
    with patch("utils.github_issues.os.getenv", return_value=None):
        # Should not raise, just log warning
        create_github_issue("shopify", "Test error")


def test_create_github_issue_labels(mock_getenv_token, mock_post):
    """Test that labels are properly set."""
    mock_response = MagicMock()
    mock_response.status_code = 201
    mock_response.json.return_value = {"html_url": "test_url"}
    mock_post.return_value = mock_response

    create_github_issue("freeCodeCamp", "Test error")

    call_args = mock_post.call_args
    data = call_args[1].get("json")

    # Verify labels
    assert "extraction-error" in data["labels"]
    assert "freecodecamp" in data["labels"]  # Should be lowercase


def test_create_github_issue_with_snippet(mock_getenv_token, mock_post):
    """Test issue creation with article snippet."""
    mock_response = MagicMock()
    mock_response.status_code = 201
    mock_response.json.return_value = {"html_url": "test_url"}
    mock_post.return_value = mock_response

    snippet = "<div class='article'>Test Article</div>"
    create_github_issue("github", "Selector changed", snippet)

    call_args = mock_post.call_args
    data = call_args[1].get("json")

    # Verify snippet is in body
    assert snippet in data["body"]


def test_create_github_issue_without_snippet(mock_getenv_token, mock_post):
    """Test issue creation without article snippet."""
    mock_response = MagicMock()
    mock_response.status_code = 201
    mock_response.json.return_value = {"html_url": "test_url"}
    mock_post.return_value = mock_response

    create_github_issue("stripe", "Network timeout")

    call_args = mock_post.call_args
    data = call_args[1].get("json")

    # Verify N/A is used when no snippet
    assert "N/A" in data["body"]
