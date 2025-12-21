import sys
import os
from unittest.mock import MagicMock

# Add the script directory to sys.path so we can import utils package
script_path = os.path.join(os.path.dirname(__file__), "..")
sys.path.insert(0, script_path)

# Mock the constants module
sys.modules["constants"] = MagicMock()
sys.modules["constants"].GOOGLE_SHEETS_SCOPES = ["scope1", "scope2"]

# Mock environment before any imports
os.environ["SHEET_ID"] = "test_sheet_id"
