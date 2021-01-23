"""module containing API codebase"""

import logging
import json

from bottle import Bottle, abort, request, response
from pydantic import ValidationError

from config import LISTEN_ADDRESS, LISTEN_PORT
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