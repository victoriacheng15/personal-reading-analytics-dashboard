from datetime import datetime
from email.utils import parsedate_to_datetime
from dateutil import parser


def clean_and_convert_date(date_str):
    """
    Cleans and converts a date string to the format "%Y-%m-%d".
    Supports ISO 8601, RFC 822 (RSS), and fuzzy text parsing.

    Args:
        date_str (str): The date string to clean and convert.

    Returns:
        str: The cleaned and converted date string in the format "%Y-%m-%d".
    """
    if not date_str:
        return ""

    date_str = date_str.strip()

    # Tier 1: ISO 8601 (starts with YYYY-MM-DD)
    if len(date_str) >= 10 and date_str[:10].replace("-", "").isdigit():
        try:
            date_obj = datetime.strptime(date_str[:10], "%Y-%m-%d")
            return date_obj.strftime("%Y-%m-%d")
        except ValueError:
            pass

    # Tier 2: RFC 822 (RSS) format
    try:
        date_obj = parsedate_to_datetime(date_str)
        return date_obj.strftime("%Y-%m-%d")
    except Exception:
        pass

    # Tier 3: Fuzzy Parsing with dateutil
    try:
        date_obj = parser.parse(date_str, fuzzy=True)
        return date_obj.strftime("%Y-%m-%d")
    except (ValueError, OverflowError):
        pass

    # Tier 4: Legacy/Custom formats
    try:
        date_obj = datetime.strptime(date_str[4:16].strip(), "%b %d %Y")
        return date_obj.strftime("%Y-%m-%d")
    except (ValueError, IndexError):
        pass

    return ""


def current_time():
    """
    Retrieves the current date and time.

    Returns:
        tuple: A tuple containing the current date and time in the format (date, time).
    """
    date = datetime.now().date()
    time = datetime.now().time().strftime("%H:%M")
    return (date, time)
