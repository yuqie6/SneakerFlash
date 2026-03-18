GO ?= go
GOLANGCI_LINT ?= golangci-lint
FRONTEND_PM ?= pnpm

.PHONY: lint lint-go lint-frontend test test-unit test-integration test-frontend test-e2e test-all build-api build-worker frontend-build

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
