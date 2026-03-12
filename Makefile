# === Variables ===
IMAGE_NAME = extraction
LINT_IMAGE = ghcr.io/igorshubovych/markdownlint-cli:v0.44.0
# Container Engine (Default to Podman)
DOCKER ?= podman

# Nix wrapper logic: Use nix-shell if available and not already inside one
# Also check if we are in a CI environment where we usually want to use system tools
USE_NIX = $(shell if command -v nix-shell >/dev/null 2>&1 && [ -z "$$IN_NIX_SHELL" ] && [ "$$GITHUB_ACTIONS" != "true" ]; then echo "yes"; else echo "no"; fi)

ifeq ($(USE_NIX),yes)
    NIX_RUN = nix-shell --run
else
    NIX_RUN = bash -c
endif

.PHONY: help run \
        install freeze update py-run py-check py-format py-test py-cov \
        go-check go-format go-update go-test go-cov \
        metrics-build web-build lint clean

# === Help ===
help:
	@echo "Available commands:"
	@echo "  make help             - Show this help message"
	@echo ""
	@echo "  make run              - [Python] Build and run extraction via Docker"
	@echo "  make py-run           - [Python] Run extraction via local venv"
	@echo "  make install          - [Python] Create .venv and install dependencies"
	@echo "  make py-check         - [Python] Run ruff check (lint)"
	@echo "  make py-format        - [Python] Format files with ruff"
	@echo "  make py-test          - [Python] Run tests"
	@echo ""
	@echo "  make go-check         - [Go] Check formatting (no changes)"
	@echo "  make go-format        - [Go] Format files with gofmt"
	@echo "  make go-test          - [Go] Run tests"
	@echo "  make go-cov           - [Go] Run tests with coverage summary"
	@echo "  make metrics-build    - [Go] Build metrics json"
	@echo "  make web-build        - [Go] Build web site"
	@echo ""
	@echo "  make lint             - [Quality] Run markdownlint via Docker"
	@echo "  make clean            - [Utils] Remove build artifacts and caches"

# === Docker (Python Application) ===
run:
	$(DOCKER) build -t $(IMAGE_NAME) .
	$(DOCKER) run --rm $(IMAGE_NAME)
	$(DOCKER) image rm $(IMAGE_NAME)

# === Python Development ===
install:
	if [ ! -d .venv ]; then python3 -m venv .venv; fi && \
	.venv/bin/pip install --upgrade pip && \
	.venv/bin/pip install -r requirements.txt

freeze:
	.venv/bin/pip freeze > requirements.txt

update:
	.venv/bin/pip install --upgrade -r requirements.txt

py-check:
	.venv/bin/python -m ruff check script/ --diff

py-format:
	.venv/bin/python -m ruff format script/

py-test:
	.venv/bin/python -m pytest script/

py-cov:
	.venv/bin/python -m pytest --cov=script --cov-report=term-missing

py-run:
	.venv/bin/python script/main.py

# === Go (Analytics & Metrics) ===
go-check:
	if [ \$$(gofmt -l ./cmd/ ./internal/ | wc -l) -gt 0 ]; then exit 1; fi 

go-format:
	gofmt -w ./cmd ./internal 

go-update:
	go get -u ./... && go mod tidy 

go-test:
	go test -v ./cmd/... ./internal/... 

go-cov:
	go test -coverprofile=coverage.out ./cmd/... ./internal/... && go tool cover -func=coverage.out && rm coverage.out || exit 1

metrics-build:
	go build -o ./metricsjson.exe ./cmd/metrics && ./metricsjson.exe && rm ./metricsjson.exe 

setup-tailwind:
	@echo "Downloading tailwind css cli v4..."
	@curl -sL https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64 -o tailwindcss
	@chmod +x tailwindcss

web-build: setup-tailwind
	echo 'Running analytics build...' && \
	rm -rf dist && \
	mkdir -p dist && \
	go build -o ./web-ssg ./cmd/web && \
	./web-ssg && \
	mkdir -p dist/css && \
	./tailwindcss -i ./internal/web/templates/css/input.css -o ./dist/css/styles.css --minify && \
	rm ./web-ssg && \
	rm tailwindcss

# === Quality & Linting ===
lint:
	$(DOCKER) run --rm -v "$(PWD):/data:Z" -w /data $(LINT_IMAGE) --fix "**/*.md"

# === Utilities ===
clean:
	find . -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null
	find . -type f -name "*.py[co]" -delete 2>/dev/null
	rm -f coverage.out coverage.html *.exe
