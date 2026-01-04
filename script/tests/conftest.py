import sys
import os
from unittest.mock import MagicMock, patch
import pytest

# Add the script directory to sys.path so we can import utils package
script_path = os.path.join(os.path.dirname(__file__), "..")
sys.path.insert(0, script_path)

# Mock the constants module
sys.modules["constants"] = MagicMock()
sys.modules["constants"].GOOGLE_SHEETS_SCOPES = ["scope1", "scope2"]

# Mock environment before any imports - CRITICAL: Clear MONGO_URI before any imports
os.environ.pop("MONGO_URI", None)
os.environ["MONGO_URI"] = ""
os.environ["SHEET_ID"] = "test_sheet_id"


@pytest.fixture(autouse=True)
def mock_mongo_client():
    """
    Automatically mock MongoClient for all tests to prevent
    accidental connections to real MongoDB during testing.
    Also resets the Singleton client to ensure test isolation.
    """
    import utils.mongo

    utils.mongo._MONGO_CLIENT = None

    with patch("pymongo.MongoClient") as mock_client:
        # Make it return None or a mock so no real connection is made
        mock_client.return_value = None
        yield mock_client
        # Cleanup after test
        utils.mongo._MONGO_CLIENT = None
