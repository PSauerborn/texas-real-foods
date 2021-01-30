"""module containing API codebase"""

import logging
import json
from typing import Union

import zipcodes
from bottle import Bottle, abort, request, response
from pydantic import ValidationError

from config import LISTEN_ADDRESS, LISTEN_PORT, COUNTY_MAPPINGS
from models import ValidationRequest
from validation import validate_numbers

LOGGER = logging.getLogger(__name__)

APP = Bottle()

@APP.route('/validate', method='POST')
def validate():
    """API route used to validate a collection of
    phone numbers"""

    LOGGER.debug('received request to validate phone numbers')
    # return 400 error if no request body is given
    if not request.json:
        LOGGER.error('received invalid JSON request')
        abort(400, 'invalid request body')

    # parse request body from JSON and convert to pydantic model
    try:
        body = ValidationRequest(**dict(request.json))
    except ValidationError:
        LOGGER.exception('unable to validate request body')
        abort(400, 'invalid request body')

    # validate phone numbers using python library
    valid, invalid = validate_numbers(body.numbers, body.country_code)
    return {'http_code': 200, 'data': {'valid': valid, 'invalid': invalid}}

def get_economic_region(county: str) -> Union[dict, None]:
    """Function used to extract economic region
    from mappings based on county"""

    # remove 'county' word from county
    county = county.lower().replace(' county', '')
    LOGGER.debug('fetching council mapping for county %s', county)
    return COUNTY_MAPPINGS.get(county, None)

@APP.route('/zipcode/<code>', method=['GET'])
def zipcode(code: str):
    """API Route to return data about zip codes"""

    try:
        LOGGER.debug('received request to analyze zip code %s', code)
        if not zipcodes.is_real(code):
            LOGGER.error('received invalid zip code \'%s\'', code)
            abort(400, 'invalid zip code')

        # get zip code data from database
        data = zipcodes.matching(code)
        if not data:
            LOGGER.warning('cannot find data for zipcode %s', code)
            abort(400, 'invalid zip code')

        if len(data) > 1:
            LOGGER.warning('found multiple data entries for single zip code')
            abort(422, 'multiple entries found for given zip code')
        data = data[0]

        # extract county and retrieve council mapping if exists
        economic_region = get_economic_region(data.get('county', ''))
        if economic_region is not None:
            data['economic_region'] = economic_region
        return {'http_code': 200, 'data': data}

    except (ValueError, TypeError):
        LOGGER.exception('unable to parse zipcode')
        abort(400, 'invalid zip code')


if __name__ == '__main__':

    def error_handler(error_details: str) -> dict:
        """Custom error handler to convert errors
        to JSON Responses"""

        code = response.status_code
        message = response.body

        response.content_type = 'application/json'
        if 'Origin' in request.headers:
            response.headers['Access-Control-Allow-Origin'] = request.headers['Origin']
        else:
            response.headers['Access-Control-Allow-Origin'] = '*'
        return json.dumps({'success': False, 'http_code': code, 'message': message})

    APP.default_error_handler = error_handler
    APP.run(host=LISTEN_ADDRESS, port=LISTEN_PORT, server='waitress')