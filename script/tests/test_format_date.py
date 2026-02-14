from unittest.mock import patch, Mock
from datetime import date
from utils import clean_and_convert_date, current_time


# Tests for clean_and_convert_date function
def test_clean_and_convert_date_with_numeric_start():
    """Test clean_and_convert_date with date string starting with digit"""
    date_str = "2025-01-15"
    result = clean_and_convert_date(date_str)
    assert result == "2025-01-15"


def test_clean_and_convert_date_with_numeric_start_and_extra_chars():
    """Test clean_and_convert_date with date string starting with digit and extra characters"""
    date_str = "2025-12-21T10:30:00Z"
    result = clean_and_convert_date(date_str)
    assert result == "2025-12-21"


def test_clean_and_convert_date_with_text_start():
    """Test clean_and_convert_date with date string not starting with digit (month name format)"""
    date_str = "Mon Jan 15 2025"
    result = clean_and_convert_date(date_str)
    assert result == "2025-01-15"


def test_clean_and_convert_date_with_text_start_and_extra_chars():
    """Test clean_and_convert_date with month name format and extra characters"""
    date_str = "Fri Dec 21 2025 10:30:00"
    result = clean_and_convert_date(date_str)
    assert result == "2025-12-21"


def test_clean_and_convert_date_different_months():
    """Test clean_and_convert_date with different month values"""
    test_cases = [
        ("2025-03-10", "2025-03-10"),
        ("2025-11-30", "2025-11-30"),
        ("Wed Mar 10 2025", "2025-03-10"),
        ("Sat Nov 30 2024", "2024-11-30"),
    ]

    for input_date, expected in test_cases:
        result = clean_and_convert_date(input_date)
        assert result == expected


def test_clean_and_convert_date_fuzzy_parsing():
    """Test clean_and_convert_date with fuzzy strings using dateutil"""
    test_cases = [
        ("Published on January 15, 2025", "2025-01-15"),
        ("Posted: 2025/12/21", "2025-12-21"),
        ("Jan 10, 2025 10:30 PM", "2025-01-10"),
        ("21st December 2024", "2024-12-21"),
        ("2025.05.15", "2025-05-15"),
        ("Invalid Date String", ""),
    ]

    for input_date, expected in test_cases:
        result = clean_and_convert_date(input_date)
        assert result == expected


# Tests for current_time function
@patch("utils.format_date.datetime")
def test_current_time_returns_tuple(mock_datetime):
    """Test that current_time returns a tuple"""
    mock_now = Mock()
    mock_now.date.return_value = date(2025, 12, 21)
    mock_now.time.return_value.strftime.return_value = "14:30"
    mock_datetime.now.return_value = mock_now

    result = current_time()

    assert isinstance(result, tuple)
    assert len(result) == 2


@patch("utils.format_date.datetime")
def test_current_time_returns_correct_format(mock_datetime):
    """Test that current_time returns date and time in correct format"""
    mock_now = Mock()
    mock_now.date.return_value = date(2025, 12, 21)
    mock_now.time.return_value.strftime.return_value = "14:30"
    mock_datetime.now.return_value = mock_now

    result_date, result_time = current_time()

    assert result_date == date(2025, 12, 21)
    assert result_time == "14:30"
    assert isinstance(result_time, str)


@patch("utils.format_date.datetime")
def test_current_time_time_format_hh_mm(mock_datetime):
    """Test that current_time formats time as HH:MM"""
    mock_now = Mock()
    mock_now.date.return_value = date(2025, 1, 1)
    mock_now.time.return_value.strftime.return_value = "09:05"
    mock_datetime.now.return_value = mock_now

    result_date, result_time = current_time()

    assert result_time == "09:05"
    # Verify strftime was called with the correct format
    mock_now.time.return_value.strftime.assert_called_with("%H:%M")


@patch("utils.format_date.datetime")
def test_current_time_different_times(mock_datetime):
    """Test current_time with different time values"""
    test_cases = [
        (date(2025, 1, 1), "00:00"),
        (date(2025, 6, 15), "12:30"),
        (date(2025, 12, 31), "23:59"),
    ]

    for test_date, test_time in test_cases:
        mock_now = Mock()
        mock_now.date.return_value = test_date
        mock_now.time.return_value.strftime.return_value = test_time
        mock_datetime.now.return_value = mock_now

        result_date, result_time = current_time()

        assert result_date == test_date
        assert result_time == test_time
