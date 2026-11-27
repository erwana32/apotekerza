# Naming conventions

The naming standard for files, tables, API, routes, and permissions. `make check-naming` (`scripts/check-naming.sh`) enforces the file rules. The goal: the role of a file is visible from its name alone, and a test file sits right next to the file it tests.

## 1. Backend (Go): `apps/backend/internal/<module>/`

Files are `snake_case.go`, **prefixed by role**, then the entity or use case.

| Prefix | Role | Folder | Example |
|---|---|---|---|
| `model_` | Domain entity and pure rules (no database or HTTP) | `domain/` | `model_product.go`, `model_stock_batch.go` |
| `service_` | One use case per file, owns the transaction | `app/` | `service_receive_goods.go`, `service_allocate_fefo.go` |
| `handler_` | HTTP handler, one file per resource, thin | `httpapi/` | `handler_product.go` |
| `mapper_` | Convert domain types to/from API types | `httpapi/` | `mapper_product.go` |
| `repo_` | Wrapper over sqlc code | `store/` | `repo_stock_batch.go` |
| `query_` | sqlc SQL file, one per entity | `store/queries/` | `query_stock_batch.sql` |
| `job_` | Background job (later) | `app/` | `job_expiry_alert.go` |

- Only `module.go` (public interface), `errors.go`, and `doc.go` have no prefix. They are the only files allowed at the module root.
- **Unit tests:** `<same name>_test.go`, e.g. `service_receive_goods_test.go`. The suffix must be lowercase `_test.go`, otherwise Go does not run it.
- **Integration tests:** `<same name>_integration_test.go` with `//go:build integration`, run with `make test-integration`.
- Generated files end in `.gen.go` or live in `store/gen/` and are never edited.
- Identifiers: entity `PascalCase` singular without a `Model` suffix (`Product`, `StockBatch`); service functions start with a verb (`ReceiveGoods`, `PostSale`, `VoidSale`); domain errors `ErrXxx` (`ErrInsufficientStock`); test names `Test<Unit>_<Behaviour>` (`TestReceiveGoods_RejectsMissingBatch`).
- sqlc query names: `GetXByID :one`, `ListXs :many`, `InsertX :one`, `UpdateX`.

## 2. Database

- **Schema name: `public`, the only schema.** There is no `CREATE SCHEMA <module>`; the module is told apart by the table prefix, not by a schema (ADR-0001). The goose version table `goose_db_version` also lives in `public`; it is the one table without a module prefix, because it belongs to the migration tool, not to a module. Never add a schema, and never rely on `search_path`: write plain table names.
- **Database name (`POSTGRES_DB`): `db_apotekerza_dev` (local dev), `db_apotekerza_stg` (staging), `db_apotekerza_prod` (production)**; the role `POSTGRES_USER` is `apotekerza` in all of them. One installation is one database (version 1: one pharmacy), so the name is the product plus the environment, not the pharmacy and not a module (decided by the developer, 2026-11-27: the environment is in the name so a database is identifiable in a client or a backup file). Underscores only, so no name needs double quotes in SQL. It is set per environment in the `.env` file listed in `deploy/AGENTS.md` (templates `deploy/.env.*example`); the production override refuses to start without it. `POSTGRES_DB` is read only when the Postgres volume is first created: changing it on an existing volume does not rename the database (rename it with `ALTER DATABASE`, or start from a new volume). Tests clone a throwaway database from `apotekerza_template`; those names are an implementation detail of `platform/testutil` and never appear in configuration.
- Every table is named `<module>_<plural_name>` so the owner is visible: `inventory_stock_batches`, `finance_journal_entries`, `masterdata_products` (ADR-0001).
- Tables: module prefix + plural `snake_case`. Columns: `snake_case`.
- Primary key `id` (UUIDv7). Foreign key `<entity>_id`. Scope column is `company_id` only (ADR-0001): required on documents and their lines, operational data, users, and the audit log; **not** on access groups, link tables, session tokens, or (version 1) products and contacts. There is no `store_id` or `branch_id`.
- Timestamps: `created_at`, `updated_at` (`timestamptz`). Client-origin time: `client_created_at`.
- Money: `*_amount` (`numeric(18,2)`). Quantities: `quantity` or `*_quantity` (`numeric(18,4)`).
- Optional read-only views for reports: `v_<module>_<name>` (e.g. `v_inventory_product_stock`). Not a required contract.
- Link (many-to-many) tables: `<module>_<left_plural>_<right_plural>`-style names that read as the relationship, e.g. `identity_user_groups`, `identity_group_permissions`. The primary key is the pair of foreign keys; no surrogate `id`.
- Booleans: `is_<state>` or `has_<thing>`, `NOT NULL` with an explicit default (`is_active`, `is_builtin`, `tracks_batch`).
- Status and other closed value sets: a `text` column with a `CHECK`, not a PostgreSQL `ENUM`, so adding a value is an add-only migration without `ALTER TYPE`. Values are lowercase `snake_case` and match the shared vocabulary (ADR-0010: `draft`, `confirmed`, `posted`, `cancelled`). The column is called `status` (or `<thing>_status` when a table has two).
- Migrations: `<timestamp>_<module>_<description>.sql`, e.g. `20261124120000_inventory_create_stock_batches.sql`. Add-only.

### Constraints, indexes, triggers, and functions

Always give database objects an explicit name. Never rely on the name PostgreSQL generates: Go code matches some of these names to turn a database error into a domain error (for example `identity_users_username_key` becomes `ErrDuplicateUsername`), so a changed name silently breaks error handling. Every name starts with the owning table, so it stays unique inside the single schema.

| Object | Pattern | Example |
|---|---|---|
| Primary key | default `<table>_pkey` (do not rename) | `identity_users_pkey` |
| Foreign key | default `<table>_<column>_fkey` (do not rename) | `identity_user_groups_user_id_fkey` |
| Unique constraint or unique index | `<table>_<columns>_key` | `identity_users_username_key` |
| Plain index | `<table>_<columns>_idx` | `identity_sessions_user_id_idx` |
| Check constraint | `<table>_<column>_<rule>` | `identity_users_username_lowercase`, `identity_sessions_token_hash_sha256` |
| Trigger | `<table>_<behaviour>` | `identity_audit_log_no_update_delete` |
| Trigger function | `<table>_<behaviour>` (one function may serve several triggers) | `identity_audit_log_append_only` |

For an index on several columns, list the columns in index order, joined by `_` (`identity_audit_log_entity_id_created_at_idx`). If a name would exceed PostgreSQL's 63-byte limit, shorten the column part, never the table prefix. A Go constant that holds a constraint name lives in the same package as the repository that matches it and is covered by an integration test.

## 3. API (OpenAPI)

- Path: `/api/v1/<module>/<resource-plural-kebab>` (e.g. `/api/v1/inventory/stock-batches`).
- `operationId`: `VerbNoun` PascalCase, unique across the spec (`ListProducts`, `ReceiveGoods`). The Go handler method has the same name.
- Schemas PascalCase (`Product`, `CreateProductRequest`, `ProductList`). JSON properties `camelCase`.
- Error codes: `UPPER_SNAKE` (`INSUFFICIENT_STOCK`).
- Permission extension on every operation: `x-permission: <module>.<resource>.<action>`.
- Exception: `GET /api/v1/health` is a cross-cutting infrastructure operation with `x-permission: platform.health.read` and `x-public: true`. `platform` is not a business module; `x-public` marks it for unauthenticated access when the auth middleware is implemented.

## 4. Frontend (React): `apps/web/src/features/<module>/<resource>/`

| Role | Pattern | Example |
|---|---|---|
| Page (view) | `<Resource><Action>Page.tsx` in `pages/` | `ProductListPage.tsx`, `ProductDetailPage.tsx` |
| Component | `<Name>.tsx` PascalCase in `components/` | `StockBatchTable.tsx`, `ProductForm.tsx`, `ProductFormModal.tsx` |
| Data hook | `use<Thing>.ts` in `hooks/`, wraps TanStack Query | `useProductList.ts` |
| Form schema | `<thing>FormSchema.ts` (zod) in `schemas/` | `productFormSchema.ts` |
| Test | `<same name>.test.ts(x)` next to the file | `ProductListPage.test.tsx` |
| Non-component file | `camelCase.ts` | `formatQuantity.ts` |

- **Routes:** `/<module>/<resource-plural-kebab>`; list `/inventory/products`, new `/inventory/products/new`, detail `/inventory/products/:productId`, edit `/inventory/products/:productId/edit`. Always `new` (never `create`/`add`) and `:<entity>Id` (never `:id`).
- **Paths** come from one nested object: `PATHS.inventory.products.detail(id)`. No string literals in components.
- One `routes.tsx` per module. Each route declares its permission (`permission: 'inventory.product.read'`).
- The generated client in `packages/api-client` is never edited.

## 5. Permissions

`<module>.<resource>.<action>`, e.g. `inventory.product.read`, `pos.sale.void`. Sensitive data has its own permission (`inventory.cost.read`, `pharmacy.patient.read`). The cross-cutting `platform.health.read` is the documented exception (see §3).

## 6. Documents and commits

- Docs: `docs/modules/<module>.md`, `docs/adr/NNNN-<kebab-title>.md` (zero-padded sequence, never reused).
- Commits: English, Conventional Commits: `feat(inventory): add FEFO allocation`.
- Branches: one per vertical slice, `slice/<short-description>`.

## 7. Language of names

Identifiers are English. Where an Indonesian pharmacy term has no clean English equivalent, `docs/glossary.md` fixes the identifier (e.g. `defecta`, `embalase_fee`). Look the term up there before naming. User-facing text (UI, receipts, labels, reports) is Indonesian and never hard-coded in the backend.
