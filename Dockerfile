FROM python:3.12-slim

WORKDIR /usr/app

COPY requirements.txt ./

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && pip install --no-cache-dir -r requirements.txt

COPY script ./script
COPY .env ./
COPY credentials.json ./

WORKDIR /usr/app/script

CMD ["python", "main.py"]
