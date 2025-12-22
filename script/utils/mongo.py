import os
import logging
from datetime import datetime, UTC
from urllib.parse import urlparse
from pymongo import MongoClient
from dotenv import load_dotenv

load_dotenv()

logger = logging.getLogger(__name__)

MONGO_URI = os.environ.get("MONGO_URI")
MONGO_DB_NAME = os.environ.get("MONGO_DB_NAME", "articles_db")
MONGO_COLLECTION_NAME = os.environ.get("MONGO_COLLECTION_NAME", "articles")


def get_mongo_client():
    """
    Returns a MongoClient instance.
    """
    if not MONGO_URI:
        logger.warning("MONGO_URI not found in environment variables.")
        return None
    return MongoClient(MONGO_URI)


def batch_insert_articles_to_mongo(client, articles):
    """
    Transforms and inserts a batch of articles into MongoDB.

    Args:
        client (MongoClient): The MongoDB client.
        articles (list): List of tuples (date, title, link, source).
    """
    if not client or not articles:
        return

    db = client[MONGO_DB_NAME]
    collection = db[MONGO_COLLECTION_NAME]

    documents = []
    current_utc_time = datetime.now(UTC).isoformat()

    for date, title, link, source in articles:
        # Extract domain from link
        try:
            parsed_uri = urlparse(link)
            domain = parsed_uri.netloc
        except Exception:
            domain = "unknown"

        doc = {
            "extracted_at": current_utc_time,
            "source": source,
            "article": {
                "title": title,
                "link": link,
                "published_date": date,
            },
            "domain": domain,
            "status": "ingested",
        }
        documents.append(doc)

    if documents:
        try:
            result = collection.insert_many(documents)
            logger.info(f"Successfully inserted {len(result.inserted_ids)} articles into MongoDB.")
        except Exception as e:
            logger.error(f"Failed to insert articles into MongoDB: {e}")
