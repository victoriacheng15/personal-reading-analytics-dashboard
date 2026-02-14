import pytest
from unittest.mock import Mock, patch
from bs4 import BeautifulSoup
from utils import (
    extract_substack_articles,
    get_articles,
    get_strategy_handler,
    wrap_with_error_handler,
)


# Helper to create a soup element from string
def create_element(html):
    soup = BeautifulSoup(html, "html.parser")
    # Return the first tag found, as extractors expect an element, not the full soup doc
    return soup.find() if soup.find() else soup


# Tests for extract_substack_articles
def test_extract_substack_articles_success():
    html = """
    <div class="post-preview">
        <a data-testid="post-preview-title" href="https://example.substack.com/p/test">Test Substack</a>
        <time datetime="2025-01-15T10:00:00.000Z">Jan 15</time>
    </div>
    """
    element = create_element(html)

    result = extract_substack_articles(element)

    assert result == (
        "2025-01-15",
        "Test Substack",
        "https://example.substack.com/p/test",
        0,
    )


# Tests for wrap_with_error_handler
def test_wrap_with_error_handler_logs_error():
    mock_logger = Mock()

    def faulty_extractor(article):
        raise ValueError("Extraction failed")

    wrapped = wrap_with_error_handler(faulty_extractor, "TestSite")

    with patch("utils.extractors.logger", mock_logger):
        with pytest.raises(ValueError):
            wrapped("some html content")

    mock_logger.error.assert_called_once()
    args = mock_logger.error.call_args[0][0]
    assert "Error extracting TestSite article" in args
    assert "some html content" in args


# Tests for get_articles
def test_get_articles_yields_new_articles():
    mock_extractor = Mock(
        return_value=("2025-01-01", "New Title", "http://link.com", 1)
    )
    elements = ["elem1"]
    existing_titles = {"Old Title"}

    generator = get_articles(elements, mock_extractor, existing_titles, "Sheet Source")
    results = list(generator)

    assert len(results) == 1
    assert results[0] == ("2025-01-01", "New Title", "http://link.com", "Sheet Source", 1)


def test_get_articles_skips_existing_titles():
    mock_extractor = Mock(
        return_value=("2025-01-01", "Existing Title", "http://link.com", "source")
    )
    elements = ["elem1"]
    existing_titles = {"Existing Title"}

    generator = get_articles(elements, mock_extractor, existing_titles, "Sheet Source")
    results = list(generator)

    assert len(results) == 0


def test_get_articles_handles_exceptions():
    mock_extractor = Mock(side_effect=Exception("Extraction error"))
    elements = ["elem1"]
    existing_titles = set()

    generator = get_articles(elements, mock_extractor, existing_titles, "Sheet Source")
    results = list(generator)

    assert len(results) == 0


# Tests for get_strategy_handler
def test_get_strategy_handler_html():
    handler = get_strategy_handler("Shopify", "html", "test-class")
    assert handler["element"]() == "test-class"
    assert callable(handler["extractor"])


def test_get_strategy_handler_rss():
    handler = get_strategy_handler("GitHub", "rss", "test-class")
    # Should include both element and "item"
    assert handler["element"]() == ["test-class", "item"]
    assert callable(handler["extractor"])


def test_get_strategy_handler_rss_default_element():
    handler = get_strategy_handler("GitHub", "rss", None)
    assert handler["element"]() == ["item"]


def test_get_strategy_handler_substack():
    handler = get_strategy_handler("Substack", "substack", "test-class")
    element_args = handler["element"]()
    assert isinstance(element_args, dict)
    assert "class_" in element_args
    # Check if it's a regex pattern
    assert element_args["class_"].pattern == "test-class"
    assert callable(handler["extractor"])


def test_get_strategy_handler_unknown_provider_html():
    """HTML strategy now fallbacks to universal extractor even for unknown providers."""
    handler = get_strategy_handler("Unknown", "html", "test-class")
    assert handler is not None
    assert handler["element"]() == "test-class"
    assert callable(handler["extractor"])


def test_get_strategy_handler_generic_rss():
    # If strategy is RSS but provider is unknown, it should return a generic RSS extractor
    handler = get_strategy_handler("GenericBlog", "rss", "test-class")
    assert handler is not None
    assert handler["element"]() == ["test-class", "item"]
    assert callable(handler["extractor"])
    # It should be a lambda or partial wrapping extract_rss_item


def test_get_strategy_handler_html_json_config():
    """Test get_strategy_handler with JSON configuration for HTML strategy."""
    provider_name = "CustomBlog"
    strategy = "html"
    element_json = (
        '{"container": "div.article", "title_selector": "h3", "date_selector": ".date"}'
    )

    handler = get_strategy_handler(provider_name, strategy, element_json)

    assert handler is not None
    assert handler["element"]() == "div.article"
    # Extractor is wrapped, so we can't easily check internal lambda name,
    # but we can verify it's callable.
    assert callable(handler["extractor"])


def test_get_strategy_handler_migration_github():
    """Test that GitHub correctly migrates to universal extractor via handler."""
    handler = get_strategy_handler("GitHub", "html", "article")
    assert handler is not None
    assert callable(handler["extractor"])
