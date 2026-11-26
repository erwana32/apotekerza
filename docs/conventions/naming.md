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

- One schema (`public`). Every table is named `<module>_<plural_name>` so the owner is visible: `inventory_stock_batches`, `finance_journal_entries`, `masterdata_products` (ADR-0001).
- Tables: module prefix + plural `snake_case`. Columns: `snake_case`.
- Primary key `id` (UUIDv7). Foreign key `<entity>_id`. Always `store_id` and `company_id` on business tables.
- Timestamps: `created_at`, `updated_at` (`timestamptz`). Client-origin time: `client_created_at`.
- Money: `*_amount` (`numeric(18,2)`). Quantities: `quantity` or `*_quantity` (`numeric(18,4)`).
- Optional read-only views for reports: `v_<module>_<name>` (e.g. `v_inventory_product_stock`). Not a required contract.
- Migrations: `<timestamp>_<module>_<description>.sql`, e.g. `20261124120000_inventory_create_stock_batches.sql`. Add-only.

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
