.PHONY: help install update format run up logs down gofmt build-metrics build-dashboard

help:
	@echo "Available commands:"
	@echo "  make install          - Install Python dependencies"
	@echo "  make update           - Update Python dependencies"
# formatting
	@echo "  make format           - Format Python files with ruff"
	@echo "  make gofmt            - Format Go files with gofmt"
#  go build binary
	@echo "  make run-metrics      - Build and run metrics binary (metricsjson)"
	@echo "  make run-dashboard    - Build and run dashboard binary"
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
	ruff format script/

gofmt:
	gofmt -w ./cmd

test:
	go test ./cmd/...

coverage:
	go test -cover ./cmd/...

coverage-html:
	go test -coverprofile=coverage.out ./cmd/... && go tool cover -html=coverage.out

run-metrics:
	go build -o ./metricsjson ./cmd/metrics && ./metricsjson

run-dashboard:
	go build -o ./dashboard ./cmd/dashboard && ./dashboard

cleanup:
	rm -f ./metricsjson ./dashboard

run:
	cd script && python main.py

up:
	docker compose up --build

down:
	docker compose down

logs:
	docker logs extractor > logs.txt
