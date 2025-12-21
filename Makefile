.PHONY: help install update format gofmt \
        run-metrics run-dashboard cleanup \
        run clean test coverage coverage-html \
        up down logs

# === Help ===
help:
	@echo "Available commands:"
	@echo "  make install          - Install Python dependencies"
	@echo "  make update           - Update Python dependencies"
	@echo ""
	@echo "  make format           - Format Python files with ruff"
	@echo "  make gofmt            - Format Go files with gofmt"
	@echo ""
	@echo "  make run-metrics      - Build and run metrics binary (metricsjson)"
	@echo "  make run-dashboard    - Build and run dashboard binary"
	@echo "  make cleanup          - Remove Go binaries"
	@echo ""
	@echo "  make run              - Run Python main script"
	@echo "  make clean            - Remove Python __pycache__ and .pyc files"
	@echo ""
	@echo "  make test             - Run Go tests"
	@echo "  make coverage         - Run Go tests with coverage"
	@echo "  make coverage-html    - Generate HTML coverage report"
	@echo "  make coverage-log     - Show coverage summary in console"
	@echo ""
	@echo "  make up               - Start Docker containers"
	@echo "  make down             - Stop Docker containers"
	@echo "  make logs             - Export Docker logs to logs.txt"
	@echo ""
	@echo "  make help             - Show this help message"

# === Python ===
install:
	python -m pip install -r requirements.txt

update:
	pur -r requirements.txt

format:
	ruff format script/

run:
	cd script && python main.py

clean:
	find . -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null
	find . -type f -name "*.py[co]" -delete 2>/dev/null

# === Go ===
gofmt:
	gofmt -w ./cmd

test:
	go test ./cmd/...

coverage:
	go test -cover ./cmd/...

coverage-html:
	go test -coverprofile=coverage.out ./cmd/... && go tool cover -html=coverage.out

coverage-log:
	go test -coverprofile=coverage.out ./cmd/... && \
	go tool cover -func=coverage.out | grep -v '100.0%'

run-metrics:
	go build -o ./metricsjson ./cmd/metrics && ./metricsjson

run-dashboard:
	go build -o ./dashboard ./cmd/dashboard && ./dashboard

cleanup:
	rm -f ./metricsjson ./dashboard

# === Docker ===
up:
	docker compose up --build

down:
	docker compose down

logs:
	docker logs extractor > logs.txt