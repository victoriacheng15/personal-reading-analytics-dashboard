import pytest
from unittest.mock import Mock, patch
from bs4 import BeautifulSoup
from utils import (
    extract_fcc_articles,
    extract_substack_articles,
    extract_github_articles,
    extract_shopify_articles,
    extract_stripe_articles,
    get_articles,
    provider_dict,
    extractor_error_handler,
)


# Helper to create a soup element from string
def create_element(html):
    soup = BeautifulSoup(html, "html.parser")
    # Return the first tag found, as extractors expect an element, not the full soup doc
    return soup.find() if soup.find() else soup


# Tests for extract_fcc_articles
def test_extract_fcc_articles_success():
    html = """
    <article>
        <h2>Test Title</h2>
        <a href="/news/test-article"></a>
        <time datetime="2025-01-15T10:00:00Z">Jan 15, 2025</time>
    </article>
    """
    element = create_element(html)

    result = extract_fcc_articles(element)

    assert result == (
        "2025-01-15",
        "Test Title",
        "https://www.freecodecamp.org/news/test-article",
        "freeCodeCamp",
    )


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
        "substack",
    )


# Tests for extract_github_articles
def test_extract_github_articles_success():
    html = """
    <article>
        <h3>GitHub News</h3>
        <a class="Link--primary" href="https://github.blog/2025-01-15-news"></a>
        <time datetime="2025-01-15">Jan 15, 2025</time>
    </article>
    """
    element = create_element(html)

    result = extract_github_articles(element)

    assert result == (
        "2025-01-15",
        "GitHub News",
        "https://github.blog/2025-01-15-news",
        "github",
    )


# Tests for extract_shopify_articles
def test_extract_shopify_articles_success():
    # Note: The class names in extract_shopify_articles are specific.
    # We need to match what the lambda looks for: "tracking-[-.02em]", "pb-4", "hover:underline"
    html = """
    <div>
        <div class="tracking-[-.02em] pb-4 hover:underline">
            <a href="/engineering/shopify-article">Shopify Article</a>
        </div>
        <p class="richtext text-body-sm font-normal text-engineering-dark-author-text font-sans">
            Jan 15, 2025
        </p>
    </div>
    """
    element = create_element(html)

    result = extract_shopify_articles(element)

    assert result == (
        "2025-01-15",
        "Shopify Article",
        "https://shopify.engineering/engineering/shopify-article",
        "shopify",
    )


# Tests for extract_stripe_articles
def test_extract_stripe_articles_success():
    html = """
    <article>
        <h1>
            <a class="BlogIndexPost__titleLink" href="/blog/stripe-news">Stripe News</a>
        </h1>
        <time datetime="2025-01-15">Jan 15, 2025</time>
    </article>
    """
    element = create_element(html)

    result = extract_stripe_articles(element)

    assert result == (
        "2025-01-15",
        "Stripe News",
        "https://stripe.com/blog/stripe-news",
        "stripe",
    )


def test_extract_stripe_articles_fallback():
    html = """
    <article>
        <h1>
            <a href="/blog/stripe-news-fallback">Stripe Fallback</a>
        </h1>
        <time datetime="2025-01-15">Jan 15, 2025</time>
    </article>
    """
    element = create_element(html)

    result = extract_stripe_articles(element)

    assert result == (
        "2025-01-15",
        "Stripe Fallback",
        "https://stripe.com/blog/stripe-news-fallback",
        "stripe",
    )


# Tests for extractor_error_handler
def test_extractor_error_handler_logs_error():
    mock_logger = Mock()

    @extractor_error_handler("TestSite")
    def faulty_extractor(article):
        raise ValueError("Extraction failed")

    with patch("utils.extractors.logger", mock_logger):
        with pytest.raises(ValueError):
            faulty_extractor("some html content")

    mock_logger.error.assert_called_once()
    args = mock_logger.error.call_args[0][0]
    assert "Error extracting TestSite article" in args
    assert "some html content" in args


# Tests for get_articles
def test_get_articles_yields_new_articles():
    mock_extractor = Mock(
        return_value=("2025-01-01", "New Title", "http://link.com", "source")
    )
    elements = ["elem1"]
    existing_titles = {"Old Title"}

    generator = get_articles(elements, mock_extractor, existing_titles)
    results = list(generator)

    assert len(results) == 1
    assert results[0] == ("2025-01-01", "New Title", "http://link.com", "source")


def test_get_articles_skips_existing_titles():
    mock_extractor = Mock(
        return_value=("2025-01-01", "Existing Title", "http://link.com", "source")
    )
    elements = ["elem1"]
    existing_titles = {"Existing Title"}

    generator = get_articles(elements, mock_extractor, existing_titles)
    results = list(generator)

    assert len(results) == 0


def test_get_articles_handles_exceptions():
    mock_extractor = Mock(side_effect=Exception("Extraction error"))
    elements = ["elem1"]
    existing_titles = set()

    generator = get_articles(elements, mock_extractor, existing_titles)
    results = list(generator)

    assert len(results) == 0


# Tests for provider_dict
def test_provider_dict_structure():
    providers = provider_dict("test-element")
    assert "freecodecamp" in providers
    assert "substack" in providers
    assert "github" in providers
    assert "shopify" in providers
    assert "stripe" in providers

    for key, value in providers.items():
        assert "element" in value
        assert "extractor" in value
        assert callable(value["extractor"])
