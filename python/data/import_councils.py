"""script used to import JSON file of councils into Postgres database"""

import json
import os
import logging
logging.basicConfig(level=logging.DEBUG)

from typing import List

from contextlib import contextmanager
from urllib.request import urlopen

import requests
import psycopg2
import psycopg2.extras
from bs4 import BeautifulSoup

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

def get_postcodes(url: str = 'https://www.zipcodestogo.com/Texas/') -> List:
    """Function used to scrape postcodes from
    web"""

    zipcodes = []
    page = urlopen(url)
    html = page.read().decode("utf-8")
    # generate new instance of beautiful soup
    soup = BeautifulSoup(html, "html.parser")
    for row in soup.find_all('tr'):
        for link in row.find_all('a'):
            try:
                zipcodes.append(int(link.text))
            except:
                pass

    return zipcodes


def test_mappings():
    """Function used to test mappings"""

    data = {
        'succeeded': [],
        'failed': [],
        'partial': []
    }

    zipcodes = get_postcodes()
    LOGGER.info('analyzing %d zip codes', len(zipcodes))

    session = requests.Session()
    for code in zipcodes:
        try:
            response = session.get('http://localhost:10847/zipcode/' + str(code))
            response.raise_for_status()

            payload = response.json()['data']
            if 'economic_region' not in payload:
                data['partial'].append(code)
            else:
                data['succeeded'].append(code)
        except requests.HTTPError:
            LOGGER.exception('unable to validate zip code')
            data['failed'].append(code)

    with open('./results.json', 'w') as f:
        json.dump(data, f)


if __name__ == '__main__':

    test_mappings()











