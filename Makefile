SHELL := /bin/bash
.DEFAULT_GOAL := help

-include .env
export

DATABASE_URL ?= postgres://listou:listou@localhost:5432/listou?sslmode=disable
API_DIR := apps/api

.PHONY: help
help: ## Show available commands
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

.PHONY: install
install: ## Install JS and Go dependencies
	pnpm install
	cd $(API_DIR) && go mod download

.PHONY: up
up: ## Start Postgres in Docker
	docker compose up -d postgres

.PHONY: down
down: ## Stop Docker services
	docker compose down

.PHONY: dev
dev: ## Run API and web locally with hot reload (needs Postgres: make up)
	@$(MAKE) migrate
	@trap 'kill 0' EXIT; \
	  (cd $(API_DIR) && go run ./cmd/api) & \
	  pnpm --filter @listou/web dev & \
	  wait

.PHONY: dev-docker
dev-docker: ## Run the full stack in Docker (web, api, postgres)
	docker compose up --build

.PHONY: migrate
migrate: ## Apply all pending migrations
	cd $(API_DIR) && go run ./cmd/migrate up

.PHONY: migrate-down
migrate-down: ## Roll back the latest migration
	cd $(API_DIR) && go run ./cmd/migrate down

.PHONY: migrate-status
migrate-status: ## Show migration status
	cd $(API_DIR) && go run ./cmd/migrate status

.PHONY: seed
seed: ## Load development seed data (idempotent)
	cd $(API_DIR) && go run ./cmd/migrate seed

.PHONY: db-reset
db-reset: ## Drop all migrations, re-apply and seed
	cd $(API_DIR) && go run ./cmd/migrate reset && go run ./cmd/migrate up && go run ./cmd/migrate seed

.PHONY: test
test: test-api test-web ## Run all tests

.PHONY: test-api
test-api: ## Run Go tests (integration tests use TEST_DATABASE_URL when set)
	cd $(API_DIR) && go test -race ./...

.PHONY: test-web
test-web: ## Run frontend unit tests
	pnpm -r test

.PHONY: e2e
e2e: ## Run Playwright tests against a running stack
	pnpm --filter @listou/web e2e

.PHONY: lint
lint: ## Lint Go and TypeScript
	cd $(API_DIR) && go vet ./... && test -z "$$(gofmt -l .)"
	pnpm -r lint
	pnpm -r typecheck

.PHONY: format
format: ## Format Go and TypeScript
	cd $(API_DIR) && gofmt -w .
	pnpm format

.PHONY: build
build: ## Build API binary and web app
	cd $(API_DIR) && CGO_ENABLED=0 go build -o bin/api ./cmd/api && CGO_ENABLED=0 go build -o bin/migrate ./cmd/migrate
	pnpm --filter @listou/web build
