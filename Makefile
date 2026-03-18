GO ?= go
GOLANGCI_LINT ?= golangci-lint
FRONTEND_PM ?= pnpm
DOCKER_COMPOSE ?= docker compose
DEV_COMPOSE_FILE ?= docker-compose.dev.yaml
PROD_COMPOSE_FILE ?= docker-compose.prod.yaml
DEV_ENV_TEMPLATE ?= ./.env.dev.example
PROD_ENV_TEMPLATE ?= ./.env.prod.example
DEV_ENV_FILE ?= ./.env.dev.local
PROD_ENV_FILE ?= ./.env.prod.local
DEV_CONFIG ?= ./config.dev.local.yml
PROD_CONFIG ?= ./config.prod.local.yml

.PHONY: help lint lint-go lint-frontend test test-unit test-integration test-frontend test-e2e test-all build-api build-worker frontend-build dev-init dev-up dev-down dev-api dev-worker dev-frontend prod-init prod-up prod-down prod-api prod-worker

help:
	@printf '%s\n' \
	'开发环境:' \
	'  make dev-init      生成 .env.dev.local 和 config.dev.local.yml' \
	'  make dev-up        启动开发依赖' \
	'  make dev-down      停止开发依赖' \
	'  make dev-api       启动 API（读取 config.dev.local.yml）' \
	'  make dev-worker    启动 Worker（读取 config.dev.local.yml）' \
	'  make dev-frontend  启动前端开发服务器' \
	'' \
	'单机生产基线:' \
	'  make prod-init     生成 .env.prod.local 和 config.prod.local.yml' \
	'  make prod-up       启动单机生产依赖' \
	'  make prod-down     停止单机生产依赖' \
	'  make prod-api      启动 API（读取 config.prod.local.yml）' \
	'  make prod-worker   启动 Worker（读取 config.prod.local.yml）' \
	'' \
	'校验与构建:' \
	'  make lint' \
	'  make test' \
	'  make build-api' \
	'  make build-worker' \
	'  make frontend-build'

lint: lint-go lint-frontend

lint-go:
	$(GOLANGCI_LINT) run ./...

lint-frontend:
	cd "frontend" && $(FRONTEND_PM) lint

test: test-unit

test-unit:
	$(GO) test ./...

test-integration:
	$(GO) test -tags=integration ./internal/integration/...

test-frontend:
	cd "frontend" && $(FRONTEND_PM) test:unit

test-e2e:
	cd "frontend" && $(FRONTEND_PM) test:e2e

test-all: test-unit test-integration test-frontend

build-api:
	$(GO) build ./cmd/api

build-worker:
	$(GO) build ./cmd/worker

frontend-build:
	cd "frontend" && $(FRONTEND_PM) build

dev-init:
	@if [ ! -f "$(DEV_ENV_FILE)" ]; then cp "$(DEV_ENV_TEMPLATE)" "$(DEV_ENV_FILE)"; fi
	@if [ ! -f "$(DEV_CONFIG)" ]; then cp "config.dev.yml.example" "$(DEV_CONFIG)"; fi

dev-up: dev-init
	$(DOCKER_COMPOSE) --env-file "$(DEV_ENV_FILE)" -f "$(DEV_COMPOSE_FILE)" up -d

dev-down:
	$(DOCKER_COMPOSE) --env-file "$(DEV_ENV_FILE)" -f "$(DEV_COMPOSE_FILE)" down

dev-api: dev-init
	SNEAKERFLASH_CONFIG="$(DEV_CONFIG)" $(GO) run ./cmd/api

dev-worker: dev-init
	SNEAKERFLASH_CONFIG="$(DEV_CONFIG)" $(GO) run ./cmd/worker

dev-frontend:
	cd "frontend" && $(FRONTEND_PM) dev

prod-init:
	@if [ ! -f "$(PROD_ENV_FILE)" ]; then cp "$(PROD_ENV_TEMPLATE)" "$(PROD_ENV_FILE)"; fi
	@if [ ! -f "$(PROD_CONFIG)" ]; then cp "config.prod.yml.example" "$(PROD_CONFIG)"; fi

prod-up: prod-init
	$(DOCKER_COMPOSE) --env-file "$(PROD_ENV_FILE)" -f "$(PROD_COMPOSE_FILE)" up -d

prod-down:
	$(DOCKER_COMPOSE) --env-file "$(PROD_ENV_FILE)" -f "$(PROD_COMPOSE_FILE)" down

prod-api: prod-init
	SNEAKERFLASH_CONFIG="$(PROD_CONFIG)" $(GO) run ./cmd/api

prod-worker: prod-init
	SNEAKERFLASH_CONFIG="$(PROD_CONFIG)" $(GO) run ./cmd/worker
