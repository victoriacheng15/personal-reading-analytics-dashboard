import logging
import sys
import asyncio
import time
from utils import (
    # Sheet operations
    get_client,
    get_worksheet,
    get_all_providers,
    get_all_titles,
    batch_append_articles,
    SHEET_ID,
    # Web scraping
    init_fetcher_state,
    fetch_page,
    close_fetcher,
    # Article extraction
    provider_dict,
    get_articles,
    # Date utilities
    current_time,
    # Constants
    ARTICLES_WORKSHEET,
    PROVIDERS_WORKSHEET,
)

logger = logging.getLogger(__name__)
# Configure logging to write to stdout so it gets captured in log files
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    stream=sys.stdout,
)


async def process_provider(fetcher_state, provider, existing_titles):
    """Process a single provider asynchronously and return articles"""
    provider_name = provider["name"]
    provider_url = provider["url"]
    provider_element = provider["element"]

    handlers = provider_dict(provider_element)
    handler = handlers.get(provider_name)

    if not handler:
        logger.info(f"Unknown provider: {provider_name}")
        return [], fetcher_state

    try:
        soup, fetcher_state = await fetch_page(fetcher_state, provider_url)
        if not soup:
            logger.warning(f"Failed to fetch page for {provider_name}")
            return [], fetcher_state

        element_args = handler["element"]()
        elements = (
            soup.find_all(**element_args)
            if isinstance(element_args, dict)
            else soup.find_all(element_args)
        )

        articles_found = list(
            get_articles(elements, handler["extractor"], existing_titles)
        )
        logger.info(
            f"Processed {provider_name}: {len(articles_found)} new articles found"
        )
        return articles_found, fetcher_state

    except Exception as e:
        logger.error(f"Error processing {provider_name}: {str(e)}", exc_info=True)
        return [], fetcher_state


async def async_main(timestamp):
    client = get_client()
    articles_sheet = get_worksheet(client, SHEET_ID, ARTICLES_WORKSHEET)
    providers_sheet = get_worksheet(client, SHEET_ID, PROVIDERS_WORKSHEET)

    existing_titles = get_all_titles(articles_sheet)
    providers = get_all_providers(providers_sheet)

    fetcher_state = init_fetcher_state()
    all_articles = []

    for provider in providers:
        articles, fetcher_state = await process_provider(
            fetcher_state, provider, existing_titles
        )
        all_articles.extend(articles)

    # Batch write all articles at once
    if all_articles:
        batch_start = time.time()
        batch_append_articles(articles_sheet, all_articles)
        batch_time = time.time() - batch_start
        logger.info(
            f"Batch write complete: {len(all_articles)} articles written in {batch_time:.2f}s"
        )
    else:
        logger.info("\n✅ No new articles found\n")

    articles_sheet.sort((1, "des"))
    articles_sheet.update_cell(1, 6, f"Updated at\n{timestamp}")
    await close_fetcher(fetcher_state)


def main(timestamp):
    """Sync wrapper for async code"""
    asyncio.run(async_main(timestamp))


if __name__ == "__main__":
    date, time_str = current_time()
    timestamp = f"{date} - {time_str}"
    print(f"The process is starting at {timestamp}\n")
    main(timestamp)
    print("\nThe process is completed")
