.PHONY: all build run test test-race test-cover lint fmt tidy docker-up docker-down docker-logs clean

BIN_DIR := bin
APP_BIN := $(BIN_DIR)/httpBack
MIGRATE_BIN := $(BIN_DIR)/migrate

all: lint test build

test:
	go test ./...

test-cover:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

lint:
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "golangci-lint not installed. See https://golangci-lint.run/welcome/install/"; exit 1; }
	golangci-lint run ./...

docker-logs:
	docker compose logs -f
