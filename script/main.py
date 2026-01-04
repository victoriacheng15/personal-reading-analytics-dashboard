import logging
import sys
import asyncio
import traceback
from utils import (
    # Sheet operations
    get_client,
    get_worksheet,
    get_all_providers,
    get_all_titles,
    batch_append_articles,
    SHEET_ID,
    # MongoDB operations
    get_mongo_client,
    batch_insert_articles_to_mongo,
    insert_error_event_to_mongo,
    close_mongo_client,
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
    datefmt="%Y-%m-%d %H:%M:%S",
    stream=sys.stdout,
)
logging.getLogger("httpx").setLevel(logging.CRITICAL)


async def process_provider(fetcher_state, provider, existing_titles):
    """Process a single provider asynchronously and return articles"""
    provider_name = provider["name"]
    provider_url = provider["url"]
    provider_element = provider["element"]

    handlers = provider_dict(provider_element)
    handler = handlers.get(provider_name.lower())

    if not handler:
        logger.info(f"Unknown provider: {provider_name}")
        return [], fetcher_state

    try:
        soup, fetcher_state = await fetch_page(fetcher_state, provider_url)
        if not soup:
            error_msg = f"Failed to fetch page for {provider_name} from {provider_url}"
            logger.warning(error_msg)

            # Capture fetch failure event to MongoDB
            mongo_client = get_mongo_client()
            if mongo_client:
                insert_error_event_to_mongo(
                    client=mongo_client,
                    source=provider_name,
                    error_type="fetch_failed",
                    error_message="Failed to fetch page",
                    url=provider_url,
                    metadata={"provider_element": provider_element, "retry_count": 0},
                )
                # Client is singleton, do not close

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
            f"Processing {provider_url} - {len(articles_found)} new articles found"
        )
        return articles_found, fetcher_state

    except Exception as e:
        logger.error(f"Error processing {provider_name}: {str(e)}", exc_info=True)

        # Capture provider-level failure event to MongoDB
        mongo_client = get_mongo_client()
        if mongo_client:
            insert_error_event_to_mongo(
                client=mongo_client,
                source=provider_name,
                error_type="provider_failed",
                error_message=str(e),
                url=provider_url,
                metadata={
                    "provider_element": provider_element,
                    "phase": "article_extraction",
                    "exception_type": type(e).__name__,
                },
                traceback_str=traceback.format_exc(),
            )
            # Client is singleton, do not close

        return [], fetcher_state


async def async_main(timestamp):
    client = get_client()
    articles_sheet = get_worksheet(client, SHEET_ID, ARTICLES_WORKSHEET)
    providers_sheet = get_worksheet(client, SHEET_ID, PROVIDERS_WORKSHEET)

    existing_titles = get_all_titles(articles_sheet)
    providers = get_all_providers(providers_sheet)

    fetcher_state = init_fetcher_state()
    all_articles = []

    # Create tasks for all providers to run concurrently
    tasks = [
        process_provider(fetcher_state, provider, existing_titles)
        for provider in providers
    ]

    # Execute all tasks concurrently
    if tasks:
        results = await asyncio.gather(*tasks)
        for articles, _ in results:
            all_articles.extend(articles)

    # Batch write all articles at once
    if all_articles:
        # Write to Google Sheets
        batch_append_articles(articles_sheet, all_articles)
        logger.info(
            f"Batch write complete: {len(all_articles)} articles added to the sheet."
        )

        # Write to MongoDB
        mongo_client = get_mongo_client()
        if mongo_client:
            batch_insert_articles_to_mongo(mongo_client, all_articles)
            # Client is singleton, do not close
    else:
        logger.info("\n✅ No new articles found\n")

    articles_sheet.sort((1, "des"))
    articles_sheet.update_cell(1, 6, f"Updated at\n{timestamp}")
    await close_fetcher(fetcher_state)


def main(timestamp):
    """Sync wrapper for async code"""
    try:
        asyncio.run(async_main(timestamp))
    finally:
        # Ensure MongoDB connection is closed at the very end
        close_mongo_client()


if __name__ == "__main__":
    date, time_str = current_time()
    timestamp = f"{date} - {time_str}"
    print(f"The process is starting at {timestamp}\n")
    main(timestamp)
    print("\nThe process is completed")
