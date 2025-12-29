.PHONY: help install freeze update format gofmt \
        go-test go-coverage go-coverage-html go-coverage-log \
        py-test py-coverage \
        run-metrics run-dashboard cleanup \
        run clean \
        up down logs

# === Help ===
help:
	@echo "Available commands:"
	@echo "  make install          - Create .venv and install Python dependencies"
	@echo "  make freeze           - Freeze Python dependencies to requirements.txt"
	@echo "  make update           - Update Python dependencies"
	@echo ""
	@echo "  make format           - Format Python files with ruff"
	@echo "  make gofmt            - Format Go files with gofmt"
	@echo ""
	@echo "  make go-test          - Run Go tests"
	@echo "  make go-coverage      - Run Go tests with coverage summary"
	@echo "  make go-coverage-html - Generate Go HTML coverage report (coverage.html)"
	@echo "  make go-coverage-log  - Show uncovered Go functions in terminal"
	@echo ""
	@echo "  make py-test          - Run Python tests"
	@echo "  make py-coverage      - Run Python coverage with missing-line report"
	@echo ""
	@echo "  make run-metrics      - Build and run metrics binary (metricsjson)"
	@echo "  make run-dashboard    - Build and run dashboard binary"
	@echo "  make cleanup          - Remove Go binaries and coverage files"
	@echo ""
	@echo "  make run              - Run Python main script (via .venv)"
	@echo "  make clean            - Remove Python __pycache__ and .pyc files"
	@echo ""
	@echo "  make up               - Start Docker containers"
	@echo "  make down             - Stop Docker containers"
	@echo "  make logs             - Export Docker logs to logs.txt"
	@echo ""
	@echo "  make help             - Show this help message"

# === Python ===
install:
	@if [ ! -d .venv ]; then \
		echo "Creating virtual environment..."; \
		python3 -m venv .venv; \
	fi
	.venv/bin/pip install --upgrade pip
	.venv/bin/pip install -r requirements.txt

freeze:
	.venv/bin/pip freeze > requirements.txt

update:
	.venv/bin/pip install --upgrade -r requirements.txt

check:
	.venv/bin/python -m ruff check script/ --diff

format:
	.venv/bin/python -m ruff format script/

py-test:
	.venv/bin/python -m pytest script/

py-coverage:
	.venv/bin/python -m pytest --cov=script --cov-report=term-missing

run:
	.venv/bin/python script/main.py

clean:
	find . -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null
	find . -type f -name "*.py[co]" -delete 2>/dev/null

# === Go ===
gofmt:
	gofmt -w ./cmd

go-test:
	go test -v ./cmd/...

go-coverage:
	go test -cover ./cmd/...

go-coverage-html:
	go test -coverprofile=coverage.out ./cmd/... && \
	go tool cover -html=coverage.out -o coverage.html

go-coverage-log:
	go test -coverprofile=coverage.out ./cmd/... && \
	go tool cover -func=coverage.out | awk '$$3 != "100.0%"'

run-metrics:
	go build -o ./metricsjson.exe ./cmd/metrics && ./metricsjson.exe

run-dashboard:
	go build -o ./dashboard.exe ./cmd/dashboard && ./dashboard.exe

cleanup:
	rm -f ./metricsjson.exe ./dashboard.exe coverage.out coverage.html

# === Docker ===
up:
	docker compose up --build

down:
	docker compose down

logs:
	docker logs extractor > logs.txt