"""script used to import JSON file of councils into Postgres database"""

import json
import os
import logging
logging.basicConfig(level=logging.DEBUG)

from contextlib import contextmanager

import psycopg2
import psycopg2.extras

LOGGER = logging.getLogger(__name__)

POSTGRES_HOST = os.getenv('POSTGRES_HOST', '192.168.99.100')
POSTGRES_PORT = int(os.getenv('POSTGRES_PORT', 5432))
POSTGRES_USER = os.getenv('POSTGRES_USER', 'postgres')
POSTGRES_PASSWORD = os.getenv('POSTGRES_PASSWORD', '')
POSTGRES_DB = os.getenv('POSTGRES_DB', '')

COUNCILS_FILE_PATH = os.getenv('COUNCILS_FILE_PATH', './councils.json')
with open(COUNCILS_FILE_PATH, 'r') as f:
    COUNCIL_DATA = json.load(f)


@contextmanager
def persistence():
    """Function used to create postgres persistence
    connection. Persistence connections are returned
    as conext managers"""
    connection = None
    LOGGER.debug(POSTGRES_USER)
    try:
        LOGGER.debug('connecting to postgres at %s:%s', POSTGRES_HOST, POSTGRES_PORT)
        connection = psycopg2.connect(host=POSTGRES_HOST, port=POSTGRES_PORT, dbname=POSTGRES_DB,
                                      user=POSTGRES_USER, password=POSTGRES_PASSWORD)
        yield connection
    except Exception:
        LOGGER.exception('unable to connect to postgres server')
        raise
    finally:
        if connection is not None:
            connection.close()

def import_counties():
    """Function to import counties into postgres
    database"""

    query = 'INSERT INTO texas_counties(county, region) VALUES(%s,%s)'
    with persistence() as db:
        # create new cursor instance
        cursor = db.cursor(cursor_factory=psycopg2.extras.RealDictCursor)

        for council, counties in COUNCIL_DATA.items():
            for county in counties:
                cursor.execute(query, (county, council))
                db.commit()

def invert_values(output_path: str = './python/data/counties.json'):
    """Function used to invert JSON file"""

    mappings = {}
    for council, counties in COUNCIL_DATA.items():
        for county in counties:
            mappings[county.lower()] = council.lower()

    with open(output_path, 'w') as f:
        json.dump(mappings, f)

if __name__ == '__main__':

    invert_values()









