SHELL := /bin/bash

MODULE     := github.com/snc-software/go-template-service
BINARY     := api
VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS    := -s -w -X main.version=$(VERSION)
MIGRATIONS := migrations
IMAGE      ?= go-template-service:$(VERSION)

GOOSE_VERSION     := v3.27.0
SWAG_VERSION      := v2.0.0-rc5.0.20260624025242-85616068ff6b
GOLANGCI_VERSION  := v2.12.2

# The service reads .env then .env.local; make reads the same files so that
# migrations and the running service can never disagree about the database.
ifneq (,$(wildcard .env))
include .env
export
endif
ifneq (,$(wildcard .env.local))
include .env.local
export
endif

DB_SSLMODE ?= require
GOOSE_DBSTRING := host=$(DB_HOST) port=$(DB_PORT) dbname=$(DB_NAME) user=$(DB_USER) password=$(DB_PASSWORD) sslmode=$(DB_SSLMODE)

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(firstword $(MAKEFILE_LIST)) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

.PHONY: run
run: ## Run the service
	go run -ldflags '$(LDFLAGS)' ./cmd/$(BINARY)

.PHONY: build
build: ## Build the binary into bin/
	go build -trimpath -ldflags '$(LDFLAGS)' -o bin/$(BINARY) ./cmd/$(BINARY)

.PHONY: test
test: ## Run unit tests with the race detector
	go test -race -coverprofile=coverage.out ./...

.PHONY: test-integration
test-integration: ## Run tests that need a database
	go test -race -tags=integration ./...

.PHONY: cover
cover: test ## Open the coverage report
	go tool cover -html=coverage.out

.PHONY: fmt
fmt: ## Format and fix imports
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION) fmt

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: lint
lint: ## Run golangci-lint
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION) run ./...

.PHONY: vuln
vuln: ## Scan dependencies for known vulnerabilities
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

.PHONY: docs
docs: ## Regenerate the OpenAPI document from annotations
	go run github.com/swaggo/swag/v2/cmd/swag@$(SWAG_VERSION) init \
		-g cmd/$(BINARY)/main.go --parseInternal --v3.1 --outputTypes json,yaml -o docs

.PHONY: migrate
migrate: ## Apply all pending migrations
	go run github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION) -dir $(MIGRATIONS) postgres "$(GOOSE_DBSTRING)" up

.PHONY: migrate-down
migrate-down: ## Roll back the most recent migration
	go run github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION) -dir $(MIGRATIONS) postgres "$(GOOSE_DBSTRING)" down

.PHONY: migrate-status
migrate-status: ## Show migration status
	go run github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION) -dir $(MIGRATIONS) postgres "$(GOOSE_DBSTRING)" status

.PHONY: migrate-create
migrate-create: ## Create a migration: make migrate-create name=add_widgets
	@test -n "$(name)" || { echo "usage: make migrate-create name=<name>"; exit 1; }
	go run github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION) -dir $(MIGRATIONS) create $(name) sql

.PHONY: docker
docker: ## Build the container image
	docker build --build-arg VERSION=$(VERSION) -t $(IMAGE) .

.PHONY: up
up: ## Start Postgres and the service via compose
	docker compose up --build

.PHONY: down
down: ## Stop compose and remove volumes
	docker compose down -v

.PHONY: ci
ci: vet lint test ## Everything CI runs, locally
