GO ?= go
GOLANGCI_LINT ?= golangci-lint
FRONTEND_PM ?= pnpm

.PHONY: lint lint-go lint-frontend test build-api build-worker frontend-build

lint: lint-go lint-frontend

lint-go:
	$(GOLANGCI_LINT) run ./...

lint-frontend:
	cd "frontend" && $(FRONTEND_PM) lint

test:
	$(GO) test ./...

build-api:
	$(GO) build ./cmd/api

build-worker:
	$(GO) build ./cmd/worker

frontend-build:
	cd "frontend" && $(FRONTEND_PM) build
