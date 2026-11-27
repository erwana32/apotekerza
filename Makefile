.PHONY: help gen gen-check lint test test-integration test-e2e check-naming check dev-up dev-down dc stg-up stg-down prod-up prod-down

BACKEND_DIR := apps/backend
WEB_DIR := apps/web
COMPOSE := deploy/docker-compose.yml
COMPOSE_PROD := deploy/docker-compose.prod.yml
STG_ENV := deploy/.env.stg
PROD_ENV := deploy/.env

# docker compose per environment: its project name (separate containers and database
# volume), env file, and compose files. `make dc` and the stg-/prod- targets share these.
DC_dev  := docker compose -f $(COMPOSE)
DC_stg  := docker compose -p apotekerza-stg --env-file $(STG_ENV) -f $(COMPOSE)
DC_prod := docker compose -p apotekerza-prod --env-file $(PROD_ENV) -f $(COMPOSE) -f $(COMPOSE_PROD)
ENV ?= dev

help: ## List commands
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-18s %s\n", $$1, $$2}'

gen: ## Generate code from api/openapi.yaml and SQL queries
	@cd $(BACKEND_DIR) && go generate ./...
	@pnpm --filter ./packages/api-client run generate

gen-check: gen ## Fail if generated code is out of sync with what is committed
	@git diff --exit-code -- ':(glob)**/gen/**' packages/api-client/src/generated || \
		{ echo "Generated code is out of sync. Run make gen and commit the result."; exit 1; }

lint: ## Lint Go and web
	@cd $(BACKEND_DIR) && go vet ./... && go tool golangci-lint run ./...
	@pnpm --filter ./$(WEB_DIR) run lint

test: ## Unit tests for Go and web (no integration tests)
	@cd $(BACKEND_DIR) && go test ./...
	@pnpm --filter ./$(WEB_DIR) run test

test-integration: ## Go integration tests (real PostgreSQL via testcontainers, needs Docker)
	@cd $(BACKEND_DIR) && go test -tags=integration -p 1 ./...

test-e2e: ## Playwright smoke tests (needs browsers: pnpm exec playwright install chromium)
	@pnpm --filter ./$(WEB_DIR) run test:e2e

check-naming: ## Check file naming rules
	@bash scripts/check-naming.sh

check: check-naming gen-check lint test ## Gate before merging

dev-up: ## Start the local dev environment
	@docker compose -f $(COMPOSE) up -d --build

dev-down: ## Stop the local dev environment (data volumes are kept)
	@docker compose -f $(COMPOSE) down

dc: ## docker compose for any environment: make dc [ENV=dev|stg|prod] CMD="logs -f backend"
	$(if $(strip $(CMD)),,$(error Usage: make dc [ENV=dev|stg|prod] CMD="ps"   e.g. CMD="logs -f backend"))
	$(if $(DC_$(ENV)),,$(error Unknown ENV '$(ENV)': use dev, stg, or prod))
	$(if $(filter stg,$(ENV)),@test -f $(STG_ENV) || { echo "Missing $(STG_ENV): copy deploy/.env.stg.example and fill it in."; exit 1; })
	$(if $(filter prod,$(ENV)),@test -f $(PROD_ENV) || { echo "Missing $(PROD_ENV): copy deploy/.env.prod.example and fill it in."; exit 1; })
	@$(DC_$(ENV)) $(CMD)

# Each environment has its own Compose project name, so its containers and its database
# volume are separate (a shared project name would share the Postgres volume).
stg-up: ## Start staging on this machine (needs deploy/.env.stg)
	@$(MAKE) --no-print-directory dc ENV=stg CMD="up -d --build"

stg-down: ## Stop staging (data volume is kept)
	@$(MAKE) --no-print-directory dc ENV=stg CMD="down"

prod-up: ## Start production, on the server only (needs deploy/.env)
	@$(MAKE) --no-print-directory dc ENV=prod CMD="up -d --build"

prod-down: ## Stop production (data volume is kept)
	@$(MAKE) --no-print-directory dc ENV=prod CMD="down"
