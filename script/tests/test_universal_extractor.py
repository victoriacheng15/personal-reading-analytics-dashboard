from bs4 import BeautifulSoup
from utils.extractors import universal_html_extractor


def test_universal_extractor_link_first_heuristic():
    """Verify title and link extraction using the link-first heuristic."""
    html = """
    <article>
        <div class="icon">🔍</div>
        <a href="/short">Short</a>
        <a class="main-link" href="/target-article">How to Build a Production-Grade Chatroom in Go</a>
        <p>Some summary text here...</p>
    </article>
    """
    soup = BeautifulSoup(html, "html.parser")

    date, title, link, tier = universal_html_extractor(
        soup, provider_url="https://example.com"
    )

    assert title == "How to Build a Production-Grade Chatroom in Go"
    assert link == "https://example.com/target-article"


def test_universal_extractor_explicit_selectors():
    """Verify that explicit selectors from config override heuristics."""
    html = """
    <article>
        <h2 class="custom-title">Explicit Title</h2>
        <a href="/wrong-link">Heuristic Link</a>
        <span class="custom-date">2025-05-20</span>
    </article>
    """
    soup = BeautifulSoup(html, "html.parser")
    config = {"title_selector": ".custom-title", "date_selector": ".custom-date"}

    date, title, link, tier = universal_html_extractor(soup, config=config)

    assert title == "Explicit Title"
    assert date == "2025-05-20"
    assert tier == 1


def test_universal_extractor_date_tier_2_semantic():
    """Verify Tier 2: Semantic <time> tag discovery."""
    html = """
    <article>
        <a href="/link">This is a very long title for the article</a>
        <time datetime="2025-01-15T10:00:00Z">Jan 15, 2025</time>
    </article>
    """
    soup = BeautifulSoup(html, "html.parser")

    date, _, _, tier = universal_html_extractor(soup)
    assert date == "2025-01-15"
    assert tier == 2


def test_universal_extractor_date_tier_3_attribute():
    """Verify Tier 3: Attribute Search (data-date)."""
    html = """
    <article>
        <a href="/link">This is a very long title for the article</a>
        <div data-date="2025-12-21">Dec 21</div>
    </article>
    """
    soup = BeautifulSoup(html, "html.parser")

    date, _, _, tier = universal_html_extractor(soup)
    assert date == "2025-12-21"
    assert tier == 3


def test_universal_extractor_date_tier_4_class():
    """Verify Tier 4: Class name search (e.g., 'post-date')."""
    html = """
    <article>
        <a href="/link">This is a very long title for the article</a>
        <span class="post-date">March 10, 2025</span>
    </article>
    """
    soup = BeautifulSoup(html, "html.parser")

    date, _, _, tier = universal_html_extractor(soup)
    assert date == "2025-03-10"
    assert tier == 4


def test_universal_extractor_date_tier_5_pattern():
    """Verify Tier 5: Heuristic pattern scan in text nodes."""
    html = """
    <article>
        <a href="/link">This is a very long title for the article</a>
        <p>Published on 2025.05.15 by Author</p>
    </article>
    """
    soup = BeautifulSoup(html, "html.parser")

    date, _, _, tier = universal_html_extractor(soup)
    assert date == "2025-05-15"
    assert tier == 5


def test_universal_extractor_absolute_url_normalization():
    """Verify that relative links are correctly joined with provider URL."""
    html = '<article><a href="blog/post-1">Long enough title here</a></article>'
    soup = BeautifulSoup(html, "html.parser")

    _, _, link, _ = universal_html_extractor(soup, provider_url="https://github.blog/")

    assert link == "https://github.blog/blog/post-1"
