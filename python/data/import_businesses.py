"""Module containing functions to import businesses from CSV files"""

import logging
logging.basicConfig(level=logging.DEBUG)
import time

import requests
import pandas as pd


LOGGER = logging.getLogger(__name__)

def import_businesses(path: str = './python/data/businesses.csv') -> pd.DataFrame:
    """Function used to import businesses from
    CSV file"""

    df = pd.read_csv(path)
    # remove businesses that are duplicates
    df = df[~df['Business Name'].str.contains('&#')]
    return df

def upload_business(session: requests.Session, business: dict):
    """Function to import business to Texas Real Foods API"""

    url = 'https://trf.project-gateway.app/api/texas-real-foods/business'
    try:
        r = session.post(url, json=business)
        LOGGER.info(r.text)
        r.raise_for_status()

        return True
    except requests.HTTPError:
        LOGGER.exception('unable to create new businesses')
    return False

if __name__ == '__main__':

    def json_from_row(row):
        return {
            'business_name': row['Business Name'],
            'business_uri': row['Business URL'],
            'metadata': {
                'yelp_business_id': row['Yelp Place ID'],
                'google_place_id': row['Google Place ID']
            }
        }

    # convert businesses into JSON format
    businesses = [json_from_row(row) for _, row in import_businesses().iterrows()]
    with requests.Session() as conn:
        conn.headers.update({'X-ApiKey': 'DuMpDLRGvJhC4YgLgrA2GrvksD9QJVgs'})
        for business in businesses:
            LOGGER.info(business)
            # upload business to texas real foods API
            upload_business(conn, business)
            time.sleep(0.25)
