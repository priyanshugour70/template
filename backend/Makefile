# Backend – common local and Docker tasks.
BACKEND_ROOT := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
COMPOSE := docker compose -f deployments/docker/docker-compose.yaml

.PHONY: help deps build run run-api run-worker \
	test test-integration lint fmt clean kill-8080 \
	migrate swagger \
	install-hooks _docker-check docker-deps stack up down docker-up docker-down

help:
	@echo "Targets:"
	@echo "  deps               go mod tidy"
	@echo "  build              build api and worker binaries"
	@echo "  run                run API server (8080) and worker together"
	@echo "  run-api            run API server only"
	@echo "  run-worker         run worker only"
	@echo "  kill-8080          kill process using port 8080"
	@echo "  swagger            generate swagger docs (requires swag CLI)"
	@echo "  test               run unit tests"
	@echo "  test-integration   run integration tests"
	@echo "  lint               run golangci-lint"
	@echo "  fmt                format Go code"
	@echo "  clean              remove binaries and cache"
	@echo "  install-hooks      install local Git hooks for this repo"
	@echo "  docker-deps        start PostgreSQL + Redis only"
	@echo "  migrate            apply SQL migrations via psql (see scripts/migrate.sh)"
	@echo "  up / docker-up     rebuild and restart full Docker stack"
	@echo "  down / docker-down docker compose down"

migrate:
	@bash scripts/migrate.sh

deps:
	go mod tidy

install-hooks:
	@git config core.hooksPath .githooks
	@chmod +x .githooks/pre-push .githooks/commit-msg
	@echo "Git hooks installed from .githooks"

build:
	@cd "$(BACKEND_ROOT)" && \
		go build -o bin/api ./cmd/api && \
		go build -o bin/worker ./cmd/worker

run:
	@bash -c 'set -e; cd "$(BACKEND_ROOT)"; \
		echo "Building API and worker (stop with Ctrl+C to pick up code changes)"; \
		go build -trimpath -o bin/api-dev ./cmd/api && \
		go build -trimpath -o bin/worker-dev ./cmd/worker && \
		bin/worker-dev & w=$$!; \
		trap "kill $$w 2>/dev/null; wait $$w 2>/dev/null" EXIT INT TERM; \
		bin/api-dev; wait $$w 2>/dev/null'

run-api:
	@bash -c 'set -e; cd "$(BACKEND_ROOT)"; \
		echo "Building API"; \
		go build -trimpath -o bin/api-dev ./cmd/api && exec bin/api-dev'

run-worker:
	@bash -c 'set -e; cd "$(BACKEND_ROOT)"; \
		echo "Building worker"; \
		go build -trimpath -o bin/worker-dev ./cmd/worker && exec bin/worker-dev'

kill-8080:
	@lsof -ti :8080 | xargs kill 2>/dev/null && echo "Freed port 8080." || echo "Nothing was using port 8080."

SWAG_VERSION ?= v1.16.6
swagger:
	@command -v swag >/dev/null 2>&1 && swag init -g cmd/api/main.go -o docs/swagger --parseDependency --parseInternal || \
	go run github.com/swaggo/swag/cmd/swag@$(SWAG_VERSION) init -g cmd/api/main.go -o docs/swagger --parseDependency --parseInternal

test:
	go test ./...

test-integration:
	go test -tags=integration ./test/...

lint:
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed; see https://golangci-lint.run/"; exit 1; }
	golangci-lint run ./...

fmt:
	gofmt -s -w .
	@command -v goimports >/dev/null 2>&1 && goimports -w . || true

clean:
	rm -rf bin/
	go clean -cache -testcache

_docker-check:
	@docker info >/dev/null 2>&1 || (echo "Docker is not running. Start Docker Desktop and try again."; exit 1)

docker-deps: _docker-check
	$(COMPOSE) up -d postgres redis
	@echo "Postgres on localhost:5432, Redis on localhost:6379"

stack: _docker-check
	$(COMPOSE) up -d --build --force-recreate --remove-orphans

up: stack
down: _docker-check
	$(COMPOSE) down --remove-orphans

docker-up: up
docker-down: down
