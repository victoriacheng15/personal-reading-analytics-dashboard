.PHONY: help install update format run up logs down gofmt build-metrics build-dashboard

help:
	@echo "Available commands:"
	@echo "  make install          - Install Python dependencies"
	@echo "  make update           - Update Python dependencies"
# formatting
	@echo "  make format           - Format Python files with ruff"
	@echo "  make gofmt            - Format Go files with gofmt"
#  go build binary
	@echo "  make build-metrics    - Build metrics binary (metricsjson)"
	@echo "  make build-dashboard  - Build dashboard binary"
# Run python script
	@echo "  make run              - Run Python main script"
# docker compose
	@echo "  make up               - Start Docker containers"
	@echo "  make down             - Stop Docker containers"
	@echo "  make logs             - Export Docker logs"
	@echo "  make help             - Show this help message"

install:
	python -m pip install -r requirements.txt

update:
	pur -r requirements.txt

format:
	ruff format src/main.py src/utils

gofmt:
	gofmt -w ./cmd

build-metrics:
	go build -o ./metricsjson ./cmd/metrics

build-dashboard:
	go build -o ./dashboard ./cmd/dashboard

run:
	cd src && python main.py

up:
	docker compose up --build

down:
	docker compose down

logs:
	docker logs extractor > logs.txt
