.PHONY: help install freeze update format gofmt \
        go-update go-test go-cov go-cov-html go-cov-log \
        py-test py-cov \
        run-metrics run-dashboard cleanup \
        run clean \
        docker-build docker-run

# === Nix Integration ===
# Helper to run any target inside nix-shell (e.g., make nix-go-test)
nix-%:
	nix-shell --run "make $*"

# === Help ===
help:
	@echo "Available commands:"
	@echo "  make help             - Show this help message"
	@echo "  make install          - Create .venv and install Python dependencies"
	@echo "  make freeze           - Freeze Python dependencies to requirements.txt"
	@echo "  make update           - Update Python dependencies"
	@echo ""
	@echo "  make check            - Run ruff check on Python files"
	@echo "  make py-format        - Format Python files with ruff"
	@echo "  make py-test          - Run Python tests"
	@echo "  make py-cov           - Run Python coverage with missing-line report"
	@echo "  make run              - Run Python main script (via .venv)"
	@echo "  make clean            - Remove Python __pycache__ and .pyc files"
	@echo ""
	@echo "  make go-format        - Format Go files with gofmt"
	@echo "  make go-update        - Update Go dependencies"
	@echo "  make go-test          - Run Go tests"
	@echo "  make go-cov           - Run Go tests with coverage summary"
	@echo "  make go-cov-html      - Generate Go HTML coverage report (coverage.html)"
	@echo "  make go-cov-log       - Show uncovered Go functions in terminal"
	@echo "  make run-metrics      - Build and run metrics binary (metricsjson)"
	@echo "  make run-analytics    - Build and run analytics binary"
	@echo ""
	@echo "  make docker-run       - Build, run, and remove Docker image"

# === Python venv and package Management ===
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

# === Python Linting, Testing, and Running ===
check:
	.venv/bin/python -m ruff check script/ --diff

py-format:
	.venv/bin/python -m ruff format script/

py-test:
	.venv/bin/python -m pytest script/

py-cov:
	.venv/bin/python -m pytest --cov=script --cov-report=term-missing

run:
	.venv/bin/python script/main.py

clean:
	find . -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null
	find . -type f -name "*.py[co]" -delete 2>/dev/null

# === Go Linting, Testing, and Running ===
go-format:
	gofmt -w ./cmd

go-update:
	go get -u ./...
	go mod tidy

go-test:
	go test -v ./cmd/...

go-cov:
	go test -cover ./cmd/...

go-cov-html:
	go test -coverprofile=coverage.out ./cmd/... && \
	go tool cover -html=coverage.out -o coverage.html

go-cov-log:
	go test -coverprofile=coverage.out ./cmd/... && \
	go tool cover -func=coverage.out | awk '$$3 != "100.0%"'

run-metrics:
	go build -o ./metricsjson.exe ./cmd/metrics && ./metricsjson.exe && rm ./metricsjson.exe

run-analytics:
	go build -o ./analytics.exe ./cmd/analytics && ./analytics.exe && rm ./analytics.exe

# === Docker Management ===
docker-run:
	docker build -t articles-extractor .
	docker run --rm articles-extractor
	docker image rm articles-extractor
