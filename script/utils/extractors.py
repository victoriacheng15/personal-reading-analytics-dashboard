import re
import logging
import traceback
from datetime import datetime
from utils.format_date import clean_and_convert_date
from utils.mongo import insert_error_event_mongo, get_mongo_client
from utils.constants import (
    SOURCE_FREECODECAMP,
    SOURCE_SUBSTACK,
    SOURCE_GITHUB,
    SOURCE_SHOPIFY,
    SOURCE_STRIPE,
)


logger = logging.getLogger(__name__)


# Error handling decorator for extractors
def extractor_error_handler(site_name):
    def decorator(func):
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
                        # Client is now a singleton managed globally, do not close here
                except Exception as mongo_error:
                    logger.warning(
                        f"Failed to log extraction error to MongoDB: {mongo_error}"
                    )

                raise

        return wrapper

    return decorator


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

def extract_rss_item(article, source_name):
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

    return (date, title, link, source_name)


@extractor_error_handler(SOURCE_FREECODECAMP)
def extract_fcc_articles(article):
    """
    Extracts article information from a freeCodeCamp article element.
    Handles both HTML articles and RSS items.
    """
    if article.name == "item":
        return extract_rss_item(article, SOURCE_FREECODECAMP)
    else:
        # Legacy HTML Scraping
        title = article.find("h2").get_text().strip()
        href = article.find("a").get("href")
        link = f"https://www.freecodecamp.org{href}"
        date = clean_and_convert_date(article.find("time").get("datetime"))
        return (date, title, link, SOURCE_FREECODECAMP)


@extractor_error_handler(SOURCE_SUBSTACK)
def extract_substack_articles(article):
    """
    Extracts article information from a Substack article element.
    """
    title = article.find(attrs={"data-testid": "post-preview-title"}).get_text().strip()
    link = article.find(attrs={"data-testid": "post-preview-title"}).get("href")
    # Date is assumed to be in a format like "YYYY-MM-DD"
    date = article.find("time").get("datetime").split("T")[0]
    return (date, title, link, SOURCE_SUBSTACK)


@extractor_error_handler(SOURCE_GITHUB)
def extract_github_articles(article):
    """
    Extracts article information from a GitHub article element.
    Handles both HTML articles and RSS items.
    """
    if article.name == "item":
        return extract_rss_item(article, SOURCE_GITHUB)
    else:
        # Legacy HTML Scraping
        title = article.find("h3").get_text().strip()
        link = article.find(class_="Link--primary").get("href")
        date = article.find("time").get("datetime")
        return (date, title, link, SOURCE_GITHUB)


@extractor_error_handler(SOURCE_SHOPIFY)
def extract_shopify_articles(article):
    """
    Extracts article information from a Shopify article element.
    """
    title_div = article.find(
        "div",
        class_=lambda x: x
        and "tracking-[-.02em]" in x
        and "pb-4" in x
        and "hover:underline" in x,
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
    return (date, title, link, SOURCE_SHOPIFY)


@extractor_error_handler(SOURCE_STRIPE)
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

    return (date, title, link, SOURCE_STRIPE)


def get_articles(elements, extract_func, existing_titles):
    """
    Extracts articles from a given provider.

    Args:
        elements (list): A list of BeautifulSoup elements representing articles.
        extract_func (function): The function to use for extracting article information.

    Yields:
        tuple: A tuple containing the extracted article information.
    """
    # Normalize existing titles for comparison
    normalized_existing_titles = set(t.strip().lower() for t in existing_titles)
    for article in elements:
        try:
            article_info = extract_func(article)
            title = article_info[1]
            normalized_title = title.strip().lower()
            # article_info tuple now: (date, title, link, source)
            if normalized_title not in normalized_existing_titles:
                yield article_info
        except Exception as _:
            pass


def provider_dict(provider_element):
    """
    Returns a dictionary mapping provider names to their corresponding elements and extractor functions.

    Args:
        provider_element (str): The element or class name used to identify articles from the provider.

    Returns:
        dict: A dictionary containing the provider's element and extractor function.
    """
    return {
        SOURCE_FREECODECAMP.lower(): {
            # Search for BOTH the HTML class (provider_element) AND the RSS tag "item"
            "element": lambda: [provider_element, "item"],
            "extractor": extract_fcc_articles,
        },
        SOURCE_SUBSTACK.lower(): {
            "element": lambda: {"class_": re.compile(provider_element)},
            "extractor": extract_substack_articles,
        },
        SOURCE_GITHUB.lower(): {
            # Search for BOTH the HTML class (provider_element) AND the RSS tag "item"
            "element": lambda: [provider_element, "item"],
            "extractor": extract_github_articles,
        },
        SOURCE_SHOPIFY.lower(): {
            "element": lambda: provider_element,
            "extractor": extract_shopify_articles,
        },
        SOURCE_STRIPE.lower(): {
            "element": lambda: provider_element,
            "extractor": extract_stripe_articles,
        },
    }
