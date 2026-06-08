# Backend – common local and Docker tasks.
BACKEND_ROOT := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
COMPOSE := docker compose -f deployments/docker/docker-compose.yaml

.PHONY: help deps build run run-api run-worker reindex-search \
	test test-integration lint fmt clean kill-8080 \
	migrate apply-sql seed seeds swagger \
	install-hooks _docker-check docker-redis docker-deps stack up down docker-up docker-down

help:
	@echo "Targets:"
	@echo "  deps               go mod tidy"
	@echo "  build              build api, worker, and reindex-search binaries"
	@echo "  run                run API server (8080) and worker together"
	@echo "  run-api            run API server only"
	@echo "  run-worker         run worker only"
	@echo "  reindex-search     rebuild Meilisearch indexes from the database"
	@echo "  kill-8080          kill process using port 8080"
	@echo "  swagger            generate swagger docs (requires swag CLI)"
	@echo "  test               run unit tests"
	@echo "  test-integration   run integration tests"
	@echo "  lint               run golangci-lint"
	@echo "  fmt                format Go code"
	@echo "  clean              remove binaries and cache"
	@echo "  install-hooks      install local Git hooks for this repo"
	@echo "  docker-redis       start Redis only (port 6379)"
	@echo "  docker-deps        start Redis (set MARIADB_* in .env for your DB)"
	@echo "  migrate            apply SQL migrations"
	@echo "  apply-sql          run one SQL file: make apply-sql FILE=migrations/seeds/001_init.sql"
	@echo "  seed / seeds       apply pending migrations + idempotent seeds (never drops data)"
	@echo "  up                 rebuild and restart Docker stack with latest code"
	@echo "  down               docker compose down"

migrate:
	go run ./cmd/migrate

apply-sql:
	@test -n "$(FILE)" || (echo "usage: make apply-sql FILE=migrations/seeds/001_init.sql"; exit 1)
	go run ./cmd/apply-sql $(FILE)

seed:
	go run ./cmd/seed

seeds:
	@$(MAKE) seed

deps:
	go mod tidy

install-hooks:
	@git config core.hooksPath .githooks
	@chmod +x .githooks/pre-push .githooks/commit-msg
	@echo "Git hooks installed from .githooks"

build:
	@cd "$(BACKEND_ROOT)" && \
		go build -o bin/api ./cmd/api && \
		go build -o bin/worker ./cmd/worker && \
		go build -o bin/reindex-search ./cmd/reindex-search

run:
	@bash -c 'set -e; cd "$(BACKEND_ROOT)"; \
		echo "Building API and worker from $$(pwd) (stop with Ctrl+C to pick up code changes)"; \
		go build -trimpath -o bin/api-dev ./cmd/api && \
		go build -trimpath -o bin/worker-dev ./cmd/worker && \
		bin/worker-dev & w=$$!; trap "kill $$w 2>/dev/null" EXIT INT TERM; exec bin/api-dev'

run-api:
	@bash -c 'set -e; cd "$(BACKEND_ROOT)"; \
		echo "Building API from $$(pwd) (stop with Ctrl+C to pick up code changes)"; \
		go build -trimpath -o bin/api-dev ./cmd/api && exec bin/api-dev'

kill-8080:
	@lsof -ti :8080 | xargs kill 2>/dev/null && echo "Freed port 8080." || echo "Nothing was using port 8080."

run-worker:
	@bash -c 'set -e; cd "$(BACKEND_ROOT)"; \
		echo "Building worker from $$(pwd)"; \
		go build -trimpath -o bin/worker-dev ./cmd/worker && exec bin/worker-dev'

reindex-search:
	@cd "$(BACKEND_ROOT)" && go run ./cmd/reindex-search

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

docker-redis: _docker-check
	docker run -d --name app-redis -p 6379:6379 redis:7-alpine 2>/dev/null || docker start app-redis
	@echo "Redis running on localhost:6379"

docker-deps: docker-redis
	@echo "Redis started. Set MARIADB_* in .env for your database, then: make run-api"

stack: _docker-check
	$(COMPOSE) up -d --build --force-recreate --remove-orphans

up: stack

down: _docker-check
	$(COMPOSE) down --remove-orphans

docker-up: up

docker-down: down
