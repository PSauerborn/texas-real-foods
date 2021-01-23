"""Module containing validation and processing functions"""

import logging
from typing import List

import phonenumbers

LOGGER = logging.getLogger(__name__)


def validate_numbers(numbers: List[str], area: str) -> tuple:
    """Function used to validate a list
    of phone numbers

    Arguments:
        numbers: list containing numbers to validate
        area: str area of numbers
    Returns:
        tuple of (valid, invalid) numbers
    """

    LOGGER.debug('validating %s numbers for code %s', len(numbers), area)
    valid, invalid = [], []

    # iterate over numbers and attempt to parse
    for num in numbers:
        try:
            parsed = phonenumbers.parse(num, area)
            # check if number is both possible and valid
            is_possible = phonenumbers.is_possible_number(parsed)
            is_valid = phonenumbers.is_valid_number(parsed)
            if is_possible and is_valid:
                valid.append(num)
            else:
                invalid.append(num)
        except Exception:
            invalid.append(num)
    return valid, invalid

