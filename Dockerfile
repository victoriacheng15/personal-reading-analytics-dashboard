FROM python:3.12-slim

WORKDIR /usr/app

COPY requirements.txt ./

RUN pip install --no-cache-dir -r requirements.txt

COPY script ./script
COPY .env ./
COPY credentials.json ./

WORKDIR /usr/app/script

CMD ["python", "main.py"]
