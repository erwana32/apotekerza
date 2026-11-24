.PHONY: help gen gen-check lint test test-integration check-naming check dev-up dev-down

BACKEND_DIR := apps/backend
WEB_DIR := apps/web
COMPOSE := deploy/docker-compose.yml

help: ## List commands
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-18s %s\n", $$1, $$2}'

gen: ## Generate code from api/openapi.yaml and SQL queries
	@if [ -f $(BACKEND_DIR)/go.mod ]; then (cd $(BACKEND_DIR) && go generate ./...); else echo "skip: apps/backend not created yet"; fi
	@if [ -f packages/api-client/package.json ]; then pnpm --filter ./packages/api-client run generate; else echo "skip: packages/api-client not created yet"; fi

gen-check: gen ## Fail if generated code is out of sync with what is committed
	@if git rev-parse --git-dir >/dev/null 2>&1; then \
		git diff --exit-code -- ':(glob)**/gen/**' packages/api-client/src/generated || \
		{ echo "Generated code is out of sync. Run make gen and commit the result."; exit 1; }; \
	else echo "skip: not a git repo"; fi

lint: ## Lint Go and web
	@if [ -f $(BACKEND_DIR)/go.mod ]; then (cd $(BACKEND_DIR) && go vet ./... && if command -v golangci-lint >/dev/null 2>&1; then golangci-lint run; else echo "golangci-lint not installed, skipped"; fi); else echo "skip: apps/backend not created yet"; fi
	@if [ -f $(WEB_DIR)/package.json ]; then pnpm --filter ./$(WEB_DIR) run lint; else echo "skip: apps/web not created yet"; fi

test: ## Unit tests for Go and web (no integration tests)
	@if [ -f $(BACKEND_DIR)/go.mod ]; then (cd $(BACKEND_DIR) && go test ./...); else echo "skip: apps/backend not created yet"; fi
	@if [ -f $(WEB_DIR)/package.json ]; then pnpm --filter ./$(WEB_DIR) run test; else echo "skip: apps/web not created yet"; fi

test-integration: ## Go integration tests (real PostgreSQL via testcontainers, needs Docker)
	@if [ -f $(BACKEND_DIR)/go.mod ]; then (cd $(BACKEND_DIR) && go test -tags=integration ./...); else echo "skip: apps/backend not created yet"; fi

check-naming: ## Check file naming rules
	@bash scripts/check-naming.sh

check: check-naming gen-check lint test ## Gate before merging

dev-up: ## Start the local dev environment
	@if [ -f $(COMPOSE) ]; then docker compose -f $(COMPOSE) up -d; else echo "skip: $(COMPOSE) not created yet"; fi

dev-down: ## Stop the local dev environment (data volumes are kept)
	@if [ -f $(COMPOSE) ]; then docker compose -f $(COMPOSE) down; else echo "skip: $(COMPOSE) not created yet"; fi
