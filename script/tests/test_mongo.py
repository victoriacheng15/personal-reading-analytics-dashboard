from unittest.mock import patch, Mock, MagicMock
from utils import (
    get_mongo_client,
    batch_insert_articles_to_mongo,
)


# Tests for get_mongo_client function
@patch("utils.mongo.MongoClient")
def test_get_mongo_client_success(mock_mongo_client):
    """Test that get_mongo_client successfully creates a MongoClient"""
    mock_client = Mock()
    mock_mongo_client.return_value = mock_client

    with patch("utils.mongo.MONGO_URI", "mongodb://localhost:27017"):
        result = get_mongo_client()

    assert result == mock_client
    mock_mongo_client.assert_called_once_with("mongodb://localhost:27017")


@patch("utils.mongo.MongoClient")
def test_get_mongo_client_with_mongodb_atlas_uri(mock_mongo_client):
    """Test that get_mongo_client works with MongoDB Atlas URI"""
    mock_client = Mock()
    mock_mongo_client.return_value = mock_client
    atlas_uri = "mongodb+srv://user:pass@cluster.mongodb.net/?retryWrites=true"

    with patch("utils.mongo.MONGO_URI", atlas_uri):
        result = get_mongo_client()

    assert result == mock_client
    mock_mongo_client.assert_called_once_with(atlas_uri)


# Tests for batch_insert_articles_to_mongo function
@patch("utils.mongo.datetime")
def test_batch_insert_articles_to_mongo_success(mock_datetime):
    """Test successful insertion of articles into MongoDB"""
    # Mock datetime
    mock_now = Mock()
    mock_now.isoformat.return_value = "2025-12-22T20:51:59.123456+00:00"
    mock_datetime.now.return_value = mock_now

    # Mock MongoDB client using MagicMock for bracket notation support
    mock_collection = Mock()
    mock_result = Mock()
    mock_result.inserted_ids = [1, 2, 3]
    mock_collection.insert_many.return_value = mock_result

    mock_db = MagicMock()
    mock_db.__getitem__.return_value = mock_collection

    mock_client = MagicMock()
    mock_client.__getitem__.return_value = mock_db

    articles = [
        ("2025-12-20", "Test Article 1", "https://example.com/article1", "github"),
        ("2025-12-21", "Test Article 2", "https://stripe.com/article2", "stripe"),
        ("2025-12-22", "Test Article 3", "https://substack.com/article3", "substack"),
    ]

    with patch("utils.mongo.MONGO_DB_NAME", "test_db"):
        with patch("utils.mongo.MONGO_COLLECTION_NAME", "articles"):
            batch_insert_articles_to_mongo(mock_client, articles)

    # Verify database and collection were accessed correctly
    mock_client.__getitem__.assert_called_with("test_db")
    mock_db.__getitem__.assert_called_with("articles")

    # Verify insert_many was called with correct documents
    call_args = mock_collection.insert_many.call_args
    documents = call_args[0][0]

    assert len(documents) == 3
    assert documents[0]["article"]["title"] == "Test Article 1"
    assert documents[0]["domain"] == "example.com"
    assert documents[0]["status"] == "ingested"
    assert documents[1]["domain"] == "stripe.com"
    assert documents[2]["domain"] == "substack.com"


@patch("utils.mongo.logger")
def test_batch_insert_articles_to_mongo_no_client(mock_logger):
    """Test that function returns early when client is None"""
    articles = [("2025-12-20", "Test", "https://example.com", "github")]

    batch_insert_articles_to_mongo(None, articles)

    # Should return without logging errors
    mock_logger.error.assert_not_called()


@patch("utils.mongo.logger")
def test_batch_insert_articles_to_mongo_empty_articles(mock_logger):
    """Test that function returns early when articles list is empty"""
    mock_client = MagicMock()

    batch_insert_articles_to_mongo(mock_client, [])

    # Should return without any database operations
    mock_client.__getitem__.assert_not_called()
    mock_logger.error.assert_not_called()


@patch("utils.mongo.logger")
@patch("utils.mongo.datetime")
def test_batch_insert_articles_to_mongo_extraction_error(mock_datetime, mock_logger):
    """Test that function handles domain extraction errors gracefully"""
    mock_now = Mock()
    mock_now.isoformat.return_value = "2025-12-22T20:51:59.123456+00:00"
    mock_datetime.now.return_value = mock_now

    mock_collection = Mock()
    mock_result = Mock()
    mock_result.inserted_ids = [1]
    mock_collection.insert_many.return_value = mock_result

    mock_db = MagicMock()
    mock_db.__getitem__.return_value = mock_collection

    mock_client = MagicMock()
    mock_client.__getitem__.return_value = mock_db

    # Article with malformed link that causes urlparse issue
    articles = [
        ("2025-12-20", "Test", "not-a-valid-url", "github"),
    ]

    with patch("utils.mongo.MONGO_DB_NAME", "test_db"):
        with patch("utils.mongo.MONGO_COLLECTION_NAME", "articles"):
            batch_insert_articles_to_mongo(mock_client, articles)

    call_args = mock_collection.insert_many.call_args
    documents = call_args[0][0]

    # Domain should be "not-a-valid-url" (the netloc of an invalid URL)
    assert "domain" in documents[0]
    # Should still insert without raising an error
    mock_collection.insert_many.assert_called_once()


@patch("utils.mongo.logger")
@patch("utils.mongo.datetime")
def test_batch_insert_articles_to_mongo_insertion_error(mock_datetime, mock_logger):
    """Test that function logs errors when insertion fails"""
    mock_now = Mock()
    mock_now.isoformat.return_value = "2025-12-22T20:51:59.123456+00:00"
    mock_datetime.now.return_value = mock_now

    mock_collection = Mock()
    mock_collection.insert_many.side_effect = Exception("Connection failed")

    mock_db = MagicMock()
    mock_db.__getitem__.return_value = mock_collection

    mock_client = MagicMock()
    mock_client.__getitem__.return_value = mock_db

    articles = [
        ("2025-12-20", "Test Article", "https://example.com/article", "github"),
    ]

    with patch("utils.mongo.MONGO_DB_NAME", "test_db"):
        with patch("utils.mongo.MONGO_COLLECTION_NAME", "articles"):
            batch_insert_articles_to_mongo(mock_client, articles)

    # Verify error was logged
    mock_logger.error.assert_called_once()
    assert "Failed to insert articles into MongoDB" in str(mock_logger.error.call_args)


@patch("utils.mongo.datetime")
def test_batch_insert_articles_to_mongo_document_structure(mock_datetime):
    """Test that documents are created with correct structure"""
    mock_now = Mock()
    mock_now.isoformat.return_value = "2025-12-22T20:51:59.123456+00:00"
    mock_datetime.now.return_value = mock_now

    mock_collection = Mock()
    mock_result = Mock()
    mock_result.inserted_ids = [1]
    mock_collection.insert_many.return_value = mock_result

    mock_db = MagicMock()
    mock_db.__getitem__.return_value = mock_collection

    mock_client = MagicMock()
    mock_client.__getitem__.return_value = mock_db

    articles = [
        ("2025-12-20", "Why Observability Matters", "https://stripe.com/blog/observability", "stripe"),
    ]

    with patch("utils.mongo.MONGO_DB_NAME", "test_db"):
        with patch("utils.mongo.MONGO_COLLECTION_NAME", "articles"):
            batch_insert_articles_to_mongo(mock_client, articles)

    call_args = mock_collection.insert_many.call_args
    documents = call_args[0][0]
    doc = documents[0]

    # Verify document structure matches the spec
    assert doc["extracted_at"] == "2025-12-22T20:51:59.123456+00:00"
    assert doc["source"] == "stripe"
    assert doc["article"]["title"] == "Why Observability Matters"
    assert doc["article"]["link"] == "https://stripe.com/blog/observability"
    assert doc["article"]["published_date"] == "2025-12-20"
    assert doc["domain"] == "stripe.com"
    assert doc["status"] == "ingested"


@patch("utils.mongo.datetime")
def test_batch_insert_articles_to_mongo_multiple_articles(mock_datetime):
    """Test insertion of multiple articles in batch"""
    mock_now = Mock()
    mock_now.isoformat.return_value = "2025-12-22T20:51:59.123456+00:00"
    mock_datetime.now.return_value = mock_now

    mock_collection = Mock()
    mock_result = Mock()
    mock_result.inserted_ids = [1, 2, 3, 4, 5]
    mock_collection.insert_many.return_value = mock_result

    mock_db = MagicMock()
    mock_db.__getitem__.return_value = mock_collection

    mock_client = MagicMock()
    mock_client.__getitem__.return_value = mock_db

    articles = [
        ("2025-12-20", "Article 1", "https://github.com/1", "github"),
        ("2025-12-21", "Article 2", "https://stripe.com/2", "stripe"),
        ("2025-12-22", "Article 3", "https://substack.com/3", "substack"),
        ("2025-12-20", "Article 4", "https://shopify.engineering/4", "shopify"),
        ("2025-12-21", "Article 5", "https://freecodecamp.org/5", "freecodecamp"),
    ]

    with patch("utils.mongo.MONGO_DB_NAME", "test_db"):
        with patch("utils.mongo.MONGO_COLLECTION_NAME", "articles"):
            batch_insert_articles_to_mongo(mock_client, articles)

    call_args = mock_collection.insert_many.call_args
    documents = call_args[0][0]

    assert len(documents) == 5
    assert documents[0]["domain"] == "github.com"
    assert documents[1]["domain"] == "stripe.com"
    assert documents[2]["domain"] == "substack.com"
    assert documents[3]["domain"] == "shopify.engineering"
    assert documents[4]["domain"] == "freecodecamp.org"
