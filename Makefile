GO ?= go
GOLANGCI_LINT ?= golangci-lint
NPM ?= npm

.PHONY: lint lint-go lint-frontend test build-api build-worker frontend-build

lint: lint-go lint-frontend

lint-go:
	$(GOLANGCI_LINT) run ./...

lint-frontend:
	cd "frontend" && $(NPM) run lint

test:
	$(GO) test ./...

build-api:
	$(GO) build ./cmd/api

build-worker:
	$(GO) build ./cmd/worker

frontend-build:
	cd "frontend" && $(NPM) run build

