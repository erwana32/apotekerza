.PHONY: help gen gen-check lint test test-integration test-e2e check-naming check dev-up dev-down

BACKEND_DIR := apps/backend
WEB_DIR := apps/web
COMPOSE := deploy/docker-compose.yml

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
	@cd $(BACKEND_DIR) && go test -tags=integration ./...

test-e2e: ## Playwright smoke tests (needs browsers: pnpm exec playwright install chromium)
	@pnpm --filter ./$(WEB_DIR) run test:e2e

check-naming: ## Check file naming rules
	@bash scripts/check-naming.sh

check: check-naming gen-check lint test ## Gate before merging

dev-up: ## Start the local dev environment
	@docker compose -f $(COMPOSE) up -d --build

dev-down: ## Stop the local dev environment (data volumes are kept)
	@docker compose -f $(COMPOSE) down
