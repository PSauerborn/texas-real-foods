"""Module containing configuration settings for the phonenumber API"""

import logging
import os
import json

from typing import Any


LOGGER = logging.getLogger(__name__)

TRUE_CONVERSIONS = ['true', 't', '1']

def override_value(key: str, default: Any, secret: bool = False) -> Any:
    """Helper function used to override local configuration
    settings with values set in environment variables

    Arguments:
        key: str name of environment variable to override
        default: Any default value to use if not set
        secret: bool hide value from logs if True
    Returns:
        default value if not set in environs, else value from
            environment variables
    """

    value = os.environ.get(key.upper(), None)

    if value is not None:
        LOGGER.info('overriding variable %s with value %s', key, value if not secret else '*' * len(value))

        # cast to boolean if default is of instance boolean
        if isinstance(default, bool):
            LOGGER.info('default value for %s is boolean. casting to boolean', key)
            value = value.lower() in TRUE_CONVERSIONS
    else:
        value = default
    return type(default)(value)

#####################################
# configure log level for application
#####################################

LOG_LEVELS = {
    'DEBUG': logging.DEBUG,
    'INFO': logging.INFO,
    'WARNING': logging.WARN,
    'ERROR': logging.ERROR,
    'CRITICAL': logging.CRITICAL
}

LOG_LEVEL = LOG_LEVELS.get(override_value('log_level', 'INFO'), logging.DEBUG)
logging.basicConfig(level=LOG_LEVEL)

LISTEN_PORT = override_value('LISTEN_PORT', 10847)
LISTEN_ADDRESS = override_value('LISTEN_ADDRESS', '0.0.0.0')

COUNTIES_FILE_PATH = override_value('COUNTIES_FILE_PATH', './data/counties.json')
with open(COUNTIES_FILE_PATH, 'r') as f:
    COUNTY_MAPPINGS = json.load(f)