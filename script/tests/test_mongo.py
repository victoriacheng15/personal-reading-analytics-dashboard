from unittest.mock import patch, Mock, MagicMock
from utils import (
    get_mongo_client,
    insert_articles_event_mongo,
    insert_error_event_mongo,
    insert_summary_event_mongo,
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


# Tests for insert_articles_event_mongo function
@patch("utils.mongo._get_collection")
@patch("utils.mongo.datetime")
def test_insert_articles_event_mongo_success(mock_datetime, mock_get_collection):
    """Test successful insertion of articles into MongoDB"""
    # Mock datetime
    mock_now = Mock()
    mock_now.isoformat.return_value = "2025-12-22T20:51:59.123456+00:00"
    mock_datetime.now.return_value = mock_now

    # Mock MongoDB collection
    mock_collection = Mock()
    mock_result = Mock()
    mock_result.inserted_ids = [1, 2, 3]
    mock_collection.insert_many.return_value = mock_result
    mock_get_collection.return_value = mock_collection

    mock_client = Mock()

    articles = [
        ("2025-12-20", "Test Article 1", "https://example.com/article1", "GitHub", 1),
        ("2025-12-21", "Test Article 2", "https://stripe.com/article2", "Stripe", 2),
        ("2025-12-22", "Test Article 3", "https://substack.com/article3", "Substack", 0),
    ]

    insert_articles_event_mongo(mock_client, articles)

    # Verify _get_collection was called with client
    mock_get_collection.assert_called_once_with(mock_client)

    # Verify insert_many was called with correct documents
    call_args = mock_collection.insert_many.call_args
    documents = call_args[0][0]

    assert len(documents) == 3
    assert documents[0]["payload"]["title"] == "Test Article 1"
    assert documents[0]["payload"]["domain"] == "example.com"
    assert documents[0]["meta"]["discovery_tier"] == 1
    assert documents[0]["status"] == "ingested"
    assert documents[0]["event_type"] == "extraction"
    assert documents[1]["payload"]["domain"] == "stripe.com"
    assert documents[1]["meta"]["discovery_tier"] == 2
    assert documents[1]["event_type"] == "extraction"
    assert documents[2]["payload"]["domain"] == "substack.com"
    assert documents[2]["meta"]["discovery_tier"] == 0
    assert documents[2]["event_type"] == "extraction"


@patch("utils.mongo.logger")
def test_insert_articles_event_mongo_no_client(mock_logger):
    """Test that function returns early when client is None"""
    articles = [("2025-12-20", "Test", "https://example.com", "GitHub", 1)]

    insert_articles_event_mongo(None, articles)

    # Should return without logging errors
    mock_logger.error.assert_not_called()


@patch("utils.mongo.logger")
def test_insert_articles_event_mongo_empty_articles(mock_logger):
    """Test that function returns early when articles list is empty"""
    mock_client = MagicMock()

    insert_articles_event_mongo(mock_client, [])

    # Should return without any database operations
    mock_client.__getitem__.assert_not_called()
    mock_logger.error.assert_not_called()


@patch("utils.mongo._get_collection")
@patch("utils.mongo.logger")
@patch("utils.mongo.datetime")
def test_insert_articles_event_mongo_insertion_error(
    mock_datetime, mock_logger, mock_get_collection
):
    """Test that function logs errors when insertion fails"""
    mock_now = Mock()
    mock_now.isoformat.return_value = "2025-12-22T20:51:59.123456+00:00"
    mock_datetime.now.return_value = mock_now

    mock_collection = Mock()
    mock_collection.insert_many.side_effect = Exception("Connection failed")
    mock_get_collection.return_value = mock_collection

    mock_client = Mock()

    articles = [
        ("2025-12-20", "Test Article", "https://example.com/article", "GitHub", 1),
    ]

    insert_articles_event_mongo(mock_client, articles)

    # Verify error was logged
    mock_logger.error.assert_called_once()
    assert "Failed to insert articles into MongoDB" in str(mock_logger.error.call_args)


@patch("utils.mongo._get_collection")
@patch("utils.mongo.datetime")
def test_insert_articles_event_mongo_document_structure(
    mock_datetime, mock_get_collection
):
    """Test that documents are created with correct structure"""
    mock_now = Mock()
    mock_now.isoformat.return_value = "2025-12-22T20:51:59.123456+00:00"
    mock_datetime.now.return_value = mock_now

    mock_collection = Mock()
    mock_result = Mock()
    mock_result.inserted_ids = [1]
    mock_collection.insert_many.return_value = mock_result
    mock_get_collection.return_value = mock_collection

    mock_client = Mock()

    articles = [
        (
            "2025-12-20",
            "Why Observability Matters",
            "https://stripe.com/blog/observability",
            "Stripe",
            4,
        ),
    ]

    insert_articles_event_mongo(mock_client, articles)

    call_args = mock_collection.insert_many.call_args
    documents = call_args[0][0]
    doc = documents[0]

    # Verify document structure matches the spec
    assert doc["timestamp"] == "2025-12-22T20:51:59.123456+00:00"
    assert doc["source"] == "stripe"
    assert doc["payload"]["title"] == "Why Observability Matters"
    assert doc["payload"]["link"] == "https://stripe.com/blog/observability"
    assert doc["payload"]["published_date"] == "2025-12-20"
    assert doc["payload"]["domain"] == "stripe.com"
    assert doc["meta"]["discovery_tier"] == 4
    assert doc["status"] == "ingested"
    assert doc["event_type"] == "extraction"


# Tests for insert_error_event_mongo function
@patch("utils.mongo._get_collection")
@patch("utils.mongo.datetime")
def test_insert_error_event_mongo_success(mock_datetime, mock_get_collection):
    """Test successful insertion of error event into MongoDB"""
    mock_now = Mock()
    mock_now.isoformat.return_value = "2025-12-23T10:30:00.000000+00:00"
    mock_datetime.now.return_value = mock_now

    mock_collection = Mock()
    mock_result = Mock()
    mock_result.inserted_id = "error_id_123"
    mock_collection.insert_one.return_value = mock_result
    mock_get_collection.return_value = mock_collection

    mock_client = Mock()

    insert_error_event_mongo(
        client=mock_client,
        source="freeCodeCamp",
        error_type="fetch_failed",
        error_message="Failed to fetch page",
        url="https://freecodecamp.org/blog",
        domain="freecodecamp.org",
        metadata={"http_status": 503, "retry_count": 0},
    )

    call_args = mock_collection.insert_one.call_args
    doc = call_args[0][0]

    assert doc["timestamp"] == "2025-12-23T10:30:00.000000+00:00"
    assert doc["source"] == "freecodecamp"
    assert doc["payload"]["message"] == "Failed to fetch page"
    assert doc["payload"]["url"] == "https://freecodecamp.org/blog"
    assert doc["payload"]["domain"] == "freecodecamp.org"
    assert doc["status"] == "ingested"
    assert doc["event_type"] == "fetch_failed"
    assert doc["meta"]["http_status"] == 503
    assert doc["meta"]["retry_count"] == 0


@patch("utils.mongo._get_collection")
@patch("utils.mongo.datetime")
def test_insert_error_event_mongo_with_traceback(mock_datetime, mock_get_collection):
    """Test error event insertion with traceback string"""
    mock_now = Mock()
    mock_now.isoformat.return_value = "2025-12-23T10:30:00.000000+00:00"
    mock_datetime.now.return_value = mock_now

    mock_collection = Mock()
    mock_result = Mock()
    mock_result.inserted_id = "error_id_456"
    mock_collection.insert_one.return_value = mock_result
    mock_get_collection.return_value = mock_collection

    mock_client = Mock()

    traceback_str = "Traceback (most recent call last):\n  File 'test.py', line 10\n    raise Exception"

    insert_error_event_mongo(
        client=mock_client,
        source="Shopify",
        error_type="extraction_failed",
        error_message="AttributeError: 'NoneType' object has no attribute 'get_text'",
        url="https://shopify.engineering/post123",
        traceback_str=traceback_str,
    )

    call_args = mock_collection.insert_one.call_args
    doc = call_args[0][0]

    assert doc["event_type"] == "extraction_failed"
    assert doc["payload"]["traceback"] == traceback_str


@patch("utils.mongo._get_collection")
@patch("utils.mongo.datetime")
def test_insert_error_event_mongo_extracts_domain(mock_datetime, mock_get_collection):
    """Test that domain is extracted from URL when not provided"""
    mock_now = Mock()
    mock_now.isoformat.return_value = "2025-12-23T10:30:00.000000+00:00"
    mock_datetime.now.return_value = mock_now

    mock_collection = Mock()
    mock_result = Mock()
    mock_result.inserted_id = "error_id_789"
    mock_collection.insert_one.return_value = mock_result
    mock_get_collection.return_value = mock_collection

    mock_client = Mock()

    insert_error_event_mongo(
        client=mock_client,
        source="Stripe",
        error_type="provider_failed",
        error_message="KeyError: 'extractor'",
        url="https://stripe.com/blog",
    )

    call_args = mock_collection.insert_one.call_args
    doc = call_args[0][0]

    assert doc["payload"]["domain"] == "stripe.com"
    assert doc["event_type"] == "provider_failed"


@patch("utils.mongo.logger")
def test_insert_error_event_mongo_no_client(mock_logger):
    """Test that function returns early when client is None"""
    insert_error_event_mongo(
        client=None,
        source="GitHub",
        error_type="fetch_failed",
        error_message="Failed to fetch",
        url="https://github.com/blog",
    )

    mock_logger.error.assert_not_called()


@patch("utils.mongo._get_collection")
@patch("utils.mongo.datetime")
def test_insert_error_event_mongo_invalid_url(mock_datetime, mock_get_collection):
    """Test that function handles invalid URLs gracefully"""
    mock_now = Mock()
    mock_now.isoformat.return_value = "2025-12-23T10:30:00.000000+00:00"
    mock_datetime.now.return_value = mock_now

    mock_collection = Mock()
    mock_result = Mock()
    mock_result.inserted_id = "error_id_999"
    mock_collection.insert_one.return_value = mock_result
    mock_get_collection.return_value = mock_collection

    mock_client = Mock()

    insert_error_event_mongo(
        client=mock_client,
        source="GitHub",
        error_type="fetch_failed",
        error_message="Failed to fetch",
        url="not-a-valid-url",
    )

    call_args = mock_collection.insert_one.call_args
    doc = call_args[0][0]

    # Should set domain to "unknown" or the invalid URL result
    assert "domain" in doc["payload"]
    mock_collection.insert_one.assert_called_once()


@patch("utils.mongo._get_collection")
@patch("utils.mongo.logger")
@patch("utils.mongo.datetime")
def test_insert_error_event_mongo_insertion_error(
    mock_datetime, mock_logger, mock_get_collection
):
    """Test that function logs errors when insertion fails"""
    mock_now = Mock()
    mock_now.isoformat.return_value = "2025-12-23T10:30:00.000000+00:00"
    mock_datetime.now.return_value = mock_now

    mock_collection = Mock()
    mock_collection.insert_one.side_effect = Exception("Connection timeout")
    mock_get_collection.return_value = mock_collection

    mock_client = Mock()

    insert_error_event_mongo(
        client=mock_client,
        source="Substack",
        error_type="extraction_failed",
        error_message="Parsing error",
        url="https://substack.com/post",
    )

    mock_logger.error.assert_called_once()
    assert "Failed to insert error event into MongoDB" in str(
        mock_logger.error.call_args
    )


# Tests for insert_summary_event_mongo function
@patch("utils.mongo._get_collection")
@patch("utils.mongo.datetime")
def test_insert_summary_event_mongo_success(mock_datetime, mock_get_collection):
    """Test successful insertion of summary event"""
    mock_now = Mock()
    mock_now.isoformat.return_value = "2025-12-23T10:30:00.000000+00:00"
    mock_datetime.now.return_value = mock_now

    mock_collection = Mock()
    mock_result = Mock()
    mock_result.inserted_id = "summary_id_123"
    mock_collection.insert_one.return_value = mock_result
    mock_get_collection.return_value = mock_collection

    mock_client = Mock()

    insert_summary_event_mongo(client=mock_client, articles_count=5)

    call_args = mock_collection.insert_one.call_args
    doc = call_args[0][0]

    assert doc["timestamp"] == "2025-12-23T10:30:00.000000+00:00"
    assert doc["source"] == "system"
    assert doc["event_type"] == "extraction_summary"
    assert doc["status"] == "ingested"
    assert doc["payload"]["articles_count"] == 5
