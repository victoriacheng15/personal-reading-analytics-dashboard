import re
import logging
import traceback
from datetime import datetime
from utils.format_date import clean_and_convert_date
from utils.mongo import insert_error_event_mongo, get_mongo_client
from utils.constants import (
    STRATEGY_RSS,
    STRATEGY_HTML,
    STRATEGY_SUBSTACK,
)


logger = logging.getLogger(__name__)


# Error handling wrapper for extractors
def wrap_with_error_handler(func, site_name):
    def wrapper(article):
        try:
            return func(article)
        except Exception as e:
            # Try to get a snippet of the article HTML for context
            snippet = None
            article_url = "unknown"
            try:
                snippet = str(article)[:300].replace("\n", " ")
            except Exception:
                snippet = "<unavailable>"

            # Try to extract URL from article
            try:
                # For standard HTML
                link = article.find("a")
                if link and link.get("href"):
                    article_url = link.get("href")
                # For RSS items
                elif article.find("link"):
                    article_url = article.find("link").get_text().strip()
            except Exception:
                pass

            tb = traceback.format_exc()
            logger.error(
                f"Error extracting {site_name} article: {e}\n"
                f"Article snippet: {snippet}\n"
                f"Traceback: {tb}"
            )

            # Capture extraction failure event to MongoDB
            try:
                mongo_client = get_mongo_client()
                if mongo_client:
                    insert_error_event_mongo(
                        client=mongo_client,
                        source=site_name.lower(),
                        error_type="extraction_failed",
                        error_message=f"{type(e).__name__}: {str(e)}",
                        url=article_url,
                        metadata={
                            "extractor_function": func.__name__,
                            "article_snippet": snippet,
                        },
                        traceback_str=tb,
                    )
            except Exception as mongo_error:
                logger.warning(
                    f"Failed to log extraction error to MongoDB: {mongo_error}"
                )

            raise

    return wrapper


def universal_html_extractor(element, config=None, provider_url=None):
    """
    Universal extractor for HTML-based blogs driven by configuration.
    Uses 'Link-First' heuristics for titles and a 5-tier discovery for dates.
    """
    config = config or {}

    # 1. Title & Link (Link-First Heuristic)
    title, link = _extract_title_and_link(element, config, provider_url)

    # 2. Date (Multi-Tier Discovery)
    date = _extract_date(element, config)

    return (date, title, link)


def _extract_title_and_link(element, config, provider_url=None):
    # Step A: Find Primary Anchor
    # Heuristic: First <a> tag with text length > 10
    anchors = element.find_all("a")
    primary_anchor = None
    for a in anchors:
        if len(a.get_text().strip()) > 10:
            primary_anchor = a
            break

    # Fallback to first anchor if none meet length requirement
    if not primary_anchor and anchors:
        primary_anchor = anchors[0]

    if not primary_anchor:
        return "<untitled>", ""

    # Step B: Extract Link & Normalize
    href = primary_anchor.get("href", "")
    link = href
    if provider_url and not (href.startswith("http://") or href.startswith("https://")):
        from urllib.parse import urljoin

        link = urljoin(provider_url, href)

    # Step C: Extract Title
    title_selector = config.get("title_selector")
    if title_selector:
        title_elem = element.select_one(title_selector)
        title = (
            title_elem.get_text().strip()
            if title_elem
            else primary_anchor.get_text().strip()
        )
    else:
        title = primary_anchor.get_text().strip()

    return title, link


def _extract_date(element, config):
    # Tier 1: Explicit selector from SSOT
    date_selector = config.get("date_selector")
    if date_selector:
        date_elem = element.select_one(date_selector)
        if date_elem:
            # Check datetime attribute first, then text
            date_raw = date_elem.get("datetime") or date_elem.get_text()
            date = clean_and_convert_date(date_raw)
            if date:
                return date

    # Tier 2: Semantic <time> tag
    time_tag = element.find("time")
    if time_tag:
        date_raw = time_tag.get("datetime") or time_tag.get_text()
        date = clean_and_convert_date(date_raw)
        if date:
            return date

    # Tier 3: Attribute Search (common meta patterns)
    date_attrs = ["pubdate", "data-date", "data-published", "content"]
    for attr in date_attrs:
        elem = element.find(attrs={attr: True})
        if elem:
            date = clean_and_convert_date(elem.get(attr))
            if date:
                return date

    # Tier 4: Class/Meta Search
    meta_classes = ["date", "time", "meta", "published", "post-date"]
    for cls in meta_classes:
        elem = element.find(class_=re.compile(cls, re.I))
        if elem:
            date = clean_and_convert_date(elem.get_text())
            if date:
                return date

    # Tier 5: Pattern Scan (Heuristic Regex)
    # Search all text for something that looks like a date
    text = element.get_text(separator=" ", strip=True)
    date = clean_and_convert_date(text)
    if date:
        return date

    return ""


def clean_text(text):
    """
    Cleans text by removing CDATA tags and whitespace.
    """
    if not text:
        return ""
    # Remove common CDATA patterns
    text = text.replace("<![CDATA[", "").replace("]]>", "").replace("]]", "")
    # Final strip of any stray closing brackets/arrows that html.parser might leave
    # and handle any remaining newlines or whitespace
    return text.strip().strip(" >[]").strip()


def extract_rss_item(article):
    """
    Generic RSS item extractor.
    Parses <item> tags (RSS 2.0) using BeautifulSoup.
    Handles <title>, <link>, and <pubDate>.
    """
    # Title
    title = clean_text(article.find("title").get_text())

    # Link
    # BeautifulSoup's html.parser can be tricky with <link> in RSS
    link_elem = article.find("link")
    link = ""
    if link_elem:
        link = clean_text(link_elem.get_text())
        # If link is empty, it might be due to self-closing tag behavior in html.parser
        if not link and link_elem.next_sibling:
            link = clean_text(str(link_elem.next_sibling))

    # Final strip to handle any remaining newlines or whitespace
    link = link.strip()

    # Date
    # RSS 2.0 uses <pubDate>, which html.parser may lowercase to <pubdate>
    date_elem = article.find("pubdate") or article.find("pubDate")
    date_raw = date_elem.get_text() if date_elem else ""
    date = clean_and_convert_date(date_raw)

    return (date, title, link)


def extract_fcc_articles(article):
    """
    Extracts article information from a freeCodeCamp article element.
    Handles both HTML articles and RSS items.
    """
    if article.name == "item":
        return extract_rss_item(article)
    else:
        # Legacy HTML Scraping
        title = article.find("h2").get_text().strip()
        href = article.find("a").get("href")
        link = f"https://www.freecodecamp.org{href}"
        date = clean_and_convert_date(article.find("time").get("datetime"))
        return (date, title, link)


def extract_substack_articles(article):
    """
    Extracts article information from a Substack article element.
    """
    title = article.find(attrs={"data-testid": "post-preview-title"}).get_text().strip()
    link = article.find(attrs={"data-testid": "post-preview-title"}).get("href")
    # Date is assumed to be in a format like "YYYY-MM-DD"
    date = article.find("time").get("datetime").split("T")[0]
    return (date, title, link)


def extract_github_articles(article):
    """
    Extracts article information from a GitHub article element.
    Handles both HTML articles and RSS items.
    """
    if article.name == "item":
        return extract_rss_item(article)
    else:
        # Legacy HTML Scraping
        title = article.find("h3").get_text().strip()
        link = article.find(class_="Link--primary").get("href")
        date = article.find("time").get("datetime")
        return (date, title, link)


def extract_shopify_articles(article):
    """
    Extracts article information from a Shopify article element.
    """
    title_div = article.find(
        "div",
        class_=lambda x: (
            x and "tracking-[-.02em]" in x and "pb-4" in x and "hover:underline" in x
        ),
    )
    title_a = title_div.find("a")
    title = title_a.get_text().strip()
    blog_address = title_a.get("href")
    link = f"https://shopify.engineering{blog_address}"
    date_element = (
        article.find(
            "p",
            class_="richtext text-body-sm font-normal text-engineering-dark-author-text font-sans",
        )
        .get_text()
        .strip()
    )
    before_format_date = datetime.strptime(date_element, "%b %d, %Y")
    date = before_format_date.strftime("%Y-%m-%d")
    return (date, title, link)


def extract_stripe_articles(article):
    """
    Extracts article information from a Stripe Engineering blog article element.
    """
    # Title link is inside an <h1> with a nested <a class="BlogIndexPost__titleLink">...
    title_a = article.find("a", class_=lambda x: x and "BlogIndexPost__titleLink" in x)
    if not title_a:
        # fallback to any h1 > a
        h1 = article.find("h1")
        title_a = h1.find("a") if h1 else None

    title = title_a.get_text().strip() if title_a else "<untitled>"
    href = title_a.get("href") if title_a else None
    # Normalize to absolute URL when possible
    link = f"https://stripe.com{href}" if href and href.startswith("/") else href

    # Date is in a <time datetime="..." element
    time_elem = article.find("time")
    date_raw = time_elem.get("datetime") if time_elem else None
    date = clean_and_convert_date(date_raw) if date_raw else ""

    return (date, title, link)


def get_articles(elements, extract_func, existing_titles, source_name):
    """
    Extracts articles from a given provider.

    Args:
        elements (list): A list of BeautifulSoup elements representing articles.
        extract_func (function): The function to use for extracting article information.
        existing_titles (set): Set of titles already in the database to skip.
        source_name (str): The canonical name of the source from the providers sheet.

    Yields:
        tuple: A tuple containing (date, title, link, source_name).
    """
    # Normalize existing titles for comparison
    normalized_existing_titles = set(t.strip().lower() for t in existing_titles)
    for article in elements:
        try:
            article_info = extract_func(article)
            # Unpack first three elements and ignore the 4th (hardcoded) source if present
            date, title, link = article_info[0], article_info[1], article_info[2]

            normalized_title = title.strip().lower()
            if normalized_title not in normalized_existing_titles:
                # Always use the source_name provided from the sheet
                yield (date, title, link, source_name)
        except Exception as _:
            pass


# Mapping of provider names to their specialized extractor functions
EXTRACTOR_MAPPING = {
    "freecodecamp": extract_fcc_articles,
    "substack": extract_substack_articles,
    "github": extract_github_articles,
    "shopify": extract_shopify_articles,
    "stripe": extract_stripe_articles,
}


def get_strategy_handler(provider_name, strategy, element):
    """
    Factory that returns the appropriate element search criteria and extractor
    function based on the provider's strategy.

    Args:
        provider_name (str): The name of the provider.
        strategy (str): The extraction strategy (rss, html, substack).
        element (str): The primary element or class to search for.

    Returns:
        dict: A dictionary containing 'element' (lambda returning search criteria)
              and 'extractor' (function). Returns None if no extractor found.
    """
    strategy = (strategy or STRATEGY_HTML).lower()

    # 1. Determine the extractor function
    extractor = EXTRACTOR_MAPPING.get(provider_name.lower())

    # Fallback to generic RSS extractor if strategy is RSS and no specialized extractor exists
    if not extractor and strategy == STRATEGY_RSS:
        extractor = lambda art: extract_rss_item(art)

    if not extractor:
        return None

    # Wrap the extractor with the dynamic error handler using the provider_name
    wrapped_extractor = wrap_with_error_handler(extractor, provider_name)

    # 2. Determine the element search criteria
    if strategy == STRATEGY_RSS:
        # Default to "item" for RSS if element is missing
        search_element = [element, "item"] if element else ["item"]
    elif strategy == STRATEGY_SUBSTACK:
        search_element = {"class_": re.compile(element)}
    else:  # STRATEGY_HTML
        search_element = element

    return {"element": lambda: search_element, "extractor": wrapped_extractor}
