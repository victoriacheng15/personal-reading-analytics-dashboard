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

_MONGO_CLIENT = None


def get_mongo_client():
    """
    Returns a shared MongoClient instance (Singleton).
    """
    global _MONGO_CLIENT
    if not MONGO_URI:
        # Avoid spamming logs if URI is missing; caller handles None check
        return None

    if _MONGO_CLIENT is None:
        try:
            _MONGO_CLIENT = MongoClient(MONGO_URI)
        except Exception as e:
            logger.error(f"Failed to create MongoDB client: {e}")
            return None

    return _MONGO_CLIENT


def close_mongo_client():
    """
    Closes the global MongoDB client connection.
    """
    global _MONGO_CLIENT
    if _MONGO_CLIENT:
        _MONGO_CLIENT.close()
        _MONGO_CLIENT = None


def _get_collection(client):
    """
    Returns the MongoDB collection for articles/events.

    Args:
        client (MongoClient): The MongoDB client.

    Returns:
        Collection: MongoDB collection or None if client is invalid.
    """
    if not client:
        logger.warning("MongoDB client is None. Cannot access collection.")
        return None

    db = client[MONGO_DB_NAME]
    return db[MONGO_COLLECTION_NAME]


def batch_insert_articles_to_mongo(client, articles):
    """
    Transforms and inserts a batch of articles into MongoDB.

    Args:
        client (MongoClient): The MongoDB client.
        articles (list): List of tuples (date, title, link, source).
    """
    if not articles:
        return

    collection = _get_collection(client)
    if not collection:
        return

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
            "event_type": "extraction",
        }
        documents.append(doc)

    if documents:
        try:
            result = collection.insert_many(documents)
            logger.info(
                f"Successfully inserted {len(result.inserted_ids)} articles into MongoDB."
            )
        except Exception as e:
            logger.error(f"Failed to insert articles into MongoDB: {e}")


def insert_error_event_to_mongo(
    client,
    source,
    error_type,
    error_message,
    url,
    domain=None,
    metadata=None,
    traceback_str=None,
):
    """
    Inserts an error event into MongoDB.

    Args:
        client (MongoClient): The MongoDB client.
        source (str): The source provider name (e.g., 'freecodecamp', 'shopify').
        error_type (str): Type of error ('fetch_failed', 'extraction_failed', 'provider_failed').
        error_message (str): Error message description.
        url (str): The URL where the error occurred.
        domain (str, optional): Domain extracted from URL. If None, will be extracted from url.
        metadata (dict, optional): Additional metadata about the error.
        traceback_str (str, optional): Full traceback string for debugging.
    """
    collection = _get_collection(client)
    if not collection:
        return

    # Extract domain if not provided
    if not domain and url:
        try:
            parsed_uri = urlparse(url)
            domain = parsed_uri.netloc
        except Exception:
            domain = "unknown"

    # Build error document
    error_doc = {
        "type": error_type,
        "message": error_message,
        "url": url,
    }

    if traceback_str:
        error_doc["traceback"] = traceback_str

    # Build full document
    doc = {
        "extracted_at": datetime.now(UTC).isoformat(),
        "source": source,
        "error": error_doc,
        "domain": domain or "unknown",
        "status": "ingested",
        "event_type": error_type,
    }

    if metadata:
        doc["metadata"] = metadata

    try:
        result = collection.insert_one(doc)
        logger.info(
            f"Inserted error event ({error_type}) for {source} into MongoDB: {result.inserted_id}"
        )
    except Exception as e:
        logger.error(f"Failed to insert error event into MongoDB: {e}")
