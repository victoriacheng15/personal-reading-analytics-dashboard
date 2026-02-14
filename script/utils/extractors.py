import re
import json
import logging
import traceback
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
    date, tier = _extract_date(element, config)

    return (date, title, link, tier)


def _extract_title_and_link(element, config, provider_url=None):
    # Step A: Find Primary Anchor
    # 1. Explicit title selector from config
    title_selector = config.get("title_selector")
    if title_selector:
        title_elem = element.select_one(title_selector)
        if title_elem:
            # If the selected element isn't an anchor, look for one inside it
            anchor = title_elem if title_elem.name == "a" else title_elem.find("a")
            if anchor:
                return title_elem.get_text().strip(), _normalize_url(
                    anchor.get("href"), provider_url
                )
            return title_elem.get_text().strip(), ""

    # 2. Prioritize headers (h1-h4) that contain links
    for tag in ["h1", "h2", "h3", "h4"]:
        for header in element.find_all(tag):
            anchor = header.find("a")
            if anchor and len(anchor.get_text().strip()) > 5:
                return anchor.get_text().strip(), _normalize_url(
                    anchor.get("href"), provider_url
                )

    # 3. Look for anchors with "title" or "link" in their class names
    title_anchors = element.find_all(
        "a", class_=re.compile(r"title|headline|link|entry", re.I)
    )
    for anchor in title_anchors:
        text = anchor.get_text().strip()
        # Filter out category links and short noise
        if (
            len(text) > 10
            and "&" not in text
            and not re.search(r"category|tag|topic", anchor.get("href", ""), re.I)
        ):
            return text, _normalize_url(anchor.get("href"), provider_url)

    # 4. Fallback: First <a> tag with substantial text
    anchors = element.find_all("a")
    for a in anchors:
        text = a.get_text().strip()
        href = a.get("href", "")
        # Avoid common generic links, category links, and breadcrumbs
        if (
            len(text) > 15
            and text.lower() not in ["read more", "continue reading"]
            and not re.search(r"category|tag|topic|author", href, re.I)
            and "&" not in text
        ):
            return text, _normalize_url(href, provider_url)

    # absolute fallback
    if anchors:
        return anchors[0].get_text().strip(), _normalize_url(
            anchors[0].get("href"), provider_url
        )

    return "<untitled>", ""


def _normalize_url(href, provider_url):
    if not href:
        return ""
    if provider_url and not (href.startswith("http://") or href.startswith("https://")):
        from urllib.parse import urljoin

        return urljoin(provider_url, href)
    return href


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
                return date, 1

    # Tier 2: Semantic <time> tag
    time_tag = element.find("time")
    if time_tag:
        date_raw = time_tag.get("datetime") or time_tag.get_text()
        date = clean_and_convert_date(date_raw)
        if date:
            return date, 2

    # Tier 3: Attribute Search (common meta patterns)
    date_attrs = ["pubdate", "data-date", "data-published", "content"]
    for attr in date_attrs:
        elem = element.find(attrs={attr: True})
        if elem:
            date = clean_and_convert_date(elem.get(attr))
            if date:
                return date, 3

    # Tier 4: Class/Meta Search
    meta_classes = ["date", "time", "meta", "published", "post-date"]
    for cls in meta_classes:
        elem = element.find(class_=re.compile(cls, re.I))
        if elem:
            date = clean_and_convert_date(elem.get_text())
            if date:
                return date, 4

    # Tier 5: Pattern Scan (Heuristic Regex)
    # Search all text for something that looks like a date
    text = element.get_text(separator=" ", strip=True)
    date = clean_and_convert_date(text)
    if date:
        return date, 5

    return "", 0


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

    return (date, title, link, 0)  # 0 indicates standard RSS strategy


def extract_substack_articles(article):
    """
    Extracts article information from a Substack article element.
    """
    title = article.find(attrs={"data-testid": "post-preview-title"}).get_text().strip()
    link = article.find(attrs={"data-testid": "post-preview-title"}).get("href")
    # Date is assumed to be in a format like "YYYY-MM-DD"
    date = article.find("time").get("datetime").split("T")[0]
    return (date, title, link, 0)


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
            # Unpack first four elements (date, title, link, tier)
            date, title, link, tier = (
                article_info[0],
                article_info[1],
                article_info[2],
                article_info[3],
            )

            normalized_title = title.strip().lower()
            if normalized_title not in normalized_existing_titles:
                # Always use the source_name provided from the sheet
                yield (date, title, link, source_name, tier)
        except Exception as _:
            pass


# Substack is kept as it uses a very specific platform layout
# All other HTML blogs use the universal extractor


def get_strategy_handler(provider_name, strategy, element, provider_url=None):
    """
    Factory that returns the appropriate element search criteria and extractor
    function based on the provider's strategy.

    Args:
        provider_name (str): The name of the provider.
        strategy (str): The extraction strategy (rss, html, substack).
        element (str): The primary element or class to search for, or a JSON config.
        provider_url (str, optional): The base URL of the provider for relative link normalization.

    Returns:
        dict: A dictionary containing 'element' (lambda returning search criteria)
              and 'extractor' (function). Returns None if no extractor found.
    """
    strategy = (strategy or STRATEGY_HTML).lower()
    config = {}
    search_element = element

    # Try to parse element as JSON for HTML strategy
    if element and (element.startswith("{") or strategy == STRATEGY_HTML):
        try:
            config = json.loads(element)
            search_element = config.get("container", element)
        except (json.JSONDecodeError, TypeError):
            # Fallback to treating element as a plain string selector/tag
            pass

    # 1. Determine the extractor function
    extractor = None

    if strategy == STRATEGY_HTML:
        extractor = lambda art: universal_html_extractor(art, config, provider_url)
    elif strategy == STRATEGY_SUBSTACK:
        extractor = extract_substack_articles
    elif strategy == STRATEGY_RSS:
        extractor = extract_rss_item

    if not extractor:
        return None

    # Wrap the extractor with the dynamic error handler using the provider_name
    wrapped_extractor = wrap_with_error_handler(extractor, provider_name)

    # 2. Determine the element search criteria
    if strategy == STRATEGY_RSS:
        # Default to "item" for RSS if element is missing
        search_element = [search_element, "item"] if search_element else ["item"]
    elif strategy == STRATEGY_SUBSTACK:
        search_element = {"class_": re.compile(search_element)}
    else:  # STRATEGY_HTML
        # search_element is already set from config or passed element
        pass

    return {"element": lambda: search_element, "extractor": wrapped_extractor}
