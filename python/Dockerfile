FROM python:3.8-slim as server

COPY *.py ./
COPY ./requirements.txt ./
COPY ./data ./data

RUN pip install --upgrade pip && pip install -r requirements.txt

CMD ["python", "api.py"]

