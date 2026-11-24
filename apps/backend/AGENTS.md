# Backend (`apps/backend`)

Go, modular monolith, PostgreSQL. General rules are in the root `AGENTS.md`. This file covers backend specifics only.

## Module layout

```
internal/<module>/
  module.go        public interface of the module (the only thing other modules may import)
  errors.go        domain errors (ErrXxx)
  domain/          model_*.go   entities and pure rules, no database or HTTP
  app/             service_*.go one use case per file, owns the transaction
  httpapi/         handler_*.go, mapper_*.go
  store/           repo_*.go, queries/query_*.sql, gen/ (sqlc output, never edited)
internal/platform/ shared infrastructure (config, db, log, http server, auth middleware, money, ids)
migrations/        goose migrations, add-only
```

Import direction: `httpapi → app → domain` and `app → store`. `domain` imports nothing from this project except shared types (`platform/money`, `platform/ids`). Other modules are imported only through the package `internal/<module>` (which holds `module.go`).

## Naming

See `docs/conventions/naming.md`. In short: Go files carry a role prefix (`model_`, `service_`, `handler_`, `mapper_`, `repo_`, `job_`), unit tests are `<name>_test.go`, integration tests are `<name>_integration_test.go` with `//go:build integration`.

## Implementation rules

- **One use case = one `service_*.go` file** with one exported function. A service opens the transaction via `platform/db`, calls `store`, and calls other modules through their interfaces inside the same transaction.
- **Handlers** are generated from OpenAPI `operationId` (`oapi-codegen` strict server). A handler checks the permission (from the spec), calls the service, and maps domain errors to API codes (`INSUFFICIENT_STOCK`, ...).
- **IDs** are UUIDv7. Sales documents carry a `client_request_id` (idempotency key) and store both `client_created_at` and `created_at` (server time). The server assigns the invoice number at posting time.
- **Time** is `timestamptz` (UTC). Time zone conversion happens only at display.
- **Money** is `numeric(18,2)` + `decimal.Decimal`. **Item quantities** are `numeric(18,4)`. Rounding happens only at the points documented in `docs/modules/finance.md`.
- **Audit trail** records access to patient data, stock movements, corrections, and Administrator actions. Never log personal data contents.
- **Errors** are wrapped with `%w`. Never swallow an error. Never `panic` on a request path.
- **Queries** live in `store/queries/query_<entity>.sql` (sqlc), named `GetXByID`, `ListXs`, `InsertX`, `UpdateX`. No ORM.
- **Migrations** are `migrations/<timestamp>_<module>_<description>.sql`, one Postgres schema per module (`CREATE SCHEMA <module>`). Never modify an existing migration (skill `add-migration`).
- **User-facing messages** are not composed in the backend: the API returns a stable error `code`; the frontend maps it to Indonesian text.

## Tests

- Unit tests: table-driven for pure logic (`domain/`, pricing rules, FEFO, rounding).
- Integration tests: real PostgreSQL via testcontainers, no mocked database. Mandatory for journal posting, stock movements, goods receipt, POS sales, and voids.
- Name pattern: `Test<Unit>_<Behaviour>`, e.g. `TestReceiveGoods_RejectsMissingBatch`.
- Commands: `make test`, `make test-integration`.

## Do not

- Read tables of another schema (`inventory.*` from module `pos`, etc.).
- Put business logic in a handler or in `.sql`.
- Edit `store/gen/`, `*.gen.go`, or old migrations.
- Add a new module without approval (see skill `add-module`).
