# === Variables ===
IMAGE_NAME = extraction
LINT_IMAGE = ghcr.io/igorshubovych/markdownlint-cli:v0.44.0

# Nix wrapper logic: Use nix-shell if available and not already inside one
USE_NIX = $(shell command -v nix-shell >/dev/null 2>&1 && [ -z "$$IN_NIX_SHELL" ] && echo "yes" || echo "no")

ifeq ($(USE_NIX),yes)
    NIX_RUN = nix-shell --run
else
    NIX_RUN = 
endif

.PHONY: help run \
        install freeze update py-run py-check py-format py-test py-cov \
        go-check go-format go-update go-test go-cov \
        run-metrics run-analytics lint clean nix-%

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
	@echo "  make run-metrics      - [Go] Build and run metrics generator"
	@echo "  make run-analytics    - [Go] Build and run analytics engine"
	@echo ""
	@echo "  make lint             - [Quality] Run markdownlint via Docker"
	@echo "  make clean            - [Utils] Remove build artifacts and caches"

# === Docker (Python Application) ===
run:
	docker build -t $(IMAGE_NAME) .
	docker run --rm $(IMAGE_NAME)
	docker image rm $(IMAGE_NAME)

# === Python Development ===
install:
	$(NIX_RUN) "if [ ! -d .venv ]; then python3 -m venv .venv; fi && \
	.venv/bin/pip install --upgrade pip && \
	.venv/bin/pip install -r requirements.txt"

freeze:
	$(NIX_RUN) ".venv/bin/pip freeze > requirements.txt"

update:
	$(NIX_RUN) ".venv/bin/pip install --upgrade -r requirements.txt"

py-check:
	$(NIX_RUN) ".venv/bin/python -m ruff check script/ --diff"

py-format:
	$(NIX_RUN) ".venv/bin/python -m ruff format script/"

py-test:
	$(NIX_RUN) ".venv/bin/python -m pytest script/"

py-cov:
	$(NIX_RUN) ".venv/bin/python -m pytest --cov=script --cov-report=term-missing"

py-run:
	$(NIX_RUN) ".venv/bin/python script/main.py"

# === Go (Analytics & Metrics) ===
go-check:
	$(NIX_RUN) "if [ \$$(gofmt -l ./cmd/ | wc -l) -gt 0 ]; then exit 1; fi"

go-format:
	$(NIX_RUN) "gofmt -w ./cmd"

go-update:
	$(NIX_RUN) "go get -u ./... && go mod tidy"

go-test:
	$(NIX_RUN) "go test -v ./cmd/..."

go-cov:
	$(NIX_RUN) "go test -coverprofile=coverage.out ./cmd/... && go tool cover -func=coverage.out && rm coverage.out || exit 1"

run-metrics:
	$(NIX_RUN) "go build -o ./metricsjson.exe ./cmd/metrics && ./metricsjson.exe && rm ./metricsjson.exe"

run-analytics:
	$(NIX_RUN) "go build -o ./analytics.exe ./cmd/analytics && ./analytics.exe && rm ./analytics.exe"

# === Quality & Linting ===
lint:
	docker run --rm -v "$(PWD):/data" -w /data $(LINT_IMAGE) --fix "**/*.md"

# === Utilities ===
clean:
	find . -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null
	find . -type f -name "*.py[co]" -delete 2>/dev/null
	rm -f coverage.out coverage.html *.exe

# === Nix Integration (Backwards Compatibility) ===
nix-%:
	nix-shell --run "make $*"