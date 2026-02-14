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
    if _MONGO_CLIENT is not None:
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
    if client is None:
        logger.warning("MongoDB client is None. Cannot access collection.")
        return None

    db = client[MONGO_DB_NAME]
    return db[MONGO_COLLECTION_NAME]


def _create_event_doc(source, event_type, payload, meta=None):
    """
    Standardizes the construction of an event document.

    Args:
        source (str): Origin of the event.
        event_type (str): Type of event.
        payload (dict): Event-specific data.
        meta (dict, optional): Operational metadata.

    Returns:
        dict: Standardized event document.
    """
    doc = {
        "timestamp": datetime.now(UTC).isoformat(),
        "source": source,
        "event_type": event_type,
        "status": "ingested",
        "payload": payload,
    }
    if meta:
        doc["meta"] = meta
    return doc


def insert_articles_event_mongo(client, articles):
    """
    Transforms and inserts a batch of articles into MongoDB.

    Args:
        client (MongoClient): The MongoDB client.
        articles (list): List of tuples (date, title, link, source).
    """
    if not articles:
        return

    collection = _get_collection(client)
    if collection is None:
        return

    documents = []

    for date, title, link, source, tier in articles:
        # Extract domain from link
        try:
            parsed_uri = urlparse(link)
            domain = parsed_uri.netloc
        except Exception:
            domain = "unknown"

        payload = {
            "title": title,
            "link": link,
            "published_date": date,
            "domain": domain,
        }
        meta = {"discovery_tier": tier}
        # Ensure source is lowercased for consistency
        doc = _create_event_doc(source.lower(), "extraction", payload, meta=meta)
        documents.append(doc)

    if documents:
        try:
            result = collection.insert_many(documents)
            logger.info(
                f"Successfully inserted {len(result.inserted_ids)} articles into MongoDB."
            )
        except Exception as e:
            logger.error(f"Failed to insert articles into MongoDB: {e}")


def insert_error_event_mongo(
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
        source (str): The source provider name.
        error_type (str): Type of error.
        error_message (str): Error message description.
        url (str): The URL where the error occurred.
        domain (str, optional): Domain extracted from URL. If None, will be extracted from url.
        metadata (dict, optional): Additional metadata about the error (e.g. retry_count).
        traceback_str (str, optional): Full Python traceback string.
    """
    collection = _get_collection(client)
    if collection is None:
        return

    # Extract domain if not provided
    if not domain and url:
        try:
            parsed_uri = urlparse(url)
            domain = parsed_uri.netloc
        except Exception:
            domain = "unknown"

    # Build payload
    payload = {
        "domain": domain or "unknown",
        "message": error_message,
        "url": url,
    }

    if traceback_str:
        payload["traceback"] = traceback_str

    # Ensure source is lowercased for consistency
    doc = _create_event_doc(
        source.lower(),
        "extraction_failed" if error_type == "extraction_failed" else error_type,
        payload,
        meta=metadata,
    )

    try:
        result = collection.insert_one(doc)
        logger.info(
            f"Inserted error event ({error_type}) for {source} into MongoDB: {result.inserted_id}"
        )
    except Exception as e:
        logger.error(f"Failed to insert error event into MongoDB: {e}")


def insert_summary_event_mongo(client, articles_count):
    """
    Inserts a summary event into MongoDB.

    Args:
        client (MongoClient): The MongoDB client.
        articles_count (int): Number of new articles extracted.
    """
    collection = _get_collection(client)
    if collection is None:
        return

    payload = {"articles_count": articles_count}
    doc = _create_event_doc("system", "extraction_summary", payload)

    try:
        result = collection.insert_one(doc)
        logger.info(f"Inserted summary event into MongoDB: {result.inserted_id}")
    except Exception as e:
        logger.error(f"Failed to insert summary event into MongoDB: {e}")
