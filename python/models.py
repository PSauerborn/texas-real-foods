"""module containing data models"""

from typing import List

from pydantic import BaseModel


class ValidationRequest(BaseModel):
    """Request model containing validation request
    data fields"""

    country_code: str
    numbers: List[str]