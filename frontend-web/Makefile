IMAGE_NAME ?= frontend-web
CONTAINER_NAME ?= frontend-web
HOST_PORT ?= 3000
CONTAINER_PORT ?= 3000
ENV_FILE ?= .env
ENV_ARGS := $(if $(wildcard $(ENV_FILE)),--env-file $(ENV_FILE),)

.DEFAULT_GOAL := help

.PHONY: help build up down restart logs status shell install-hooks dev lint

help:
	@echo "Frontend Web — Docker + dev commands"
	@echo ""
	@echo "Local:"
	@echo "  make dev          Run next dev on localhost:$(HOST_PORT)"
	@echo "  make lint         Run eslint"
	@echo ""
	@echo "Docker:"
	@echo "  make up           Build image and run container on 127.0.0.1:$(HOST_PORT)"
	@echo "  make down         Stop and remove container"
	@echo "  make restart      Down + up"
	@echo "  make build        Build the Docker image only"
	@echo "  make logs         Follow container logs"
	@echo "  make status       Show container status"
	@echo "  make shell        Open a shell in the running container"
	@echo ""
	@echo "Git hooks:"
	@echo "  make install-hooks  git config core.hooksPath .githooks"

dev:
	pnpm dev

lint:
	pnpm lint

build:
	docker build -t $(IMAGE_NAME) .

up: build
	docker rm -f $(CONTAINER_NAME) 2>/dev/null || true
	docker run -d \
		--name $(CONTAINER_NAME) \
		--restart unless-stopped \
		-p 127.0.0.1:$(HOST_PORT):$(CONTAINER_PORT) \
		$(ENV_ARGS) \
		-e PORT=$(CONTAINER_PORT) \
		$(IMAGE_NAME)

down:
	docker rm -f $(CONTAINER_NAME) 2>/dev/null || true

restart: down up

logs:
	docker logs -f $(CONTAINER_NAME)

status:
	docker ps -a --filter "name=$(CONTAINER_NAME)"

shell:
	docker exec -it $(CONTAINER_NAME) sh

install-hooks:
	@git config core.hooksPath .githooks
	@chmod +x .githooks/pre-push .githooks/commit-msg
	@echo "Git hooks installed from .githooks"
