# API contract (`api/`)

`api/openapi.yaml` is the **source of truth** for the contract. The Go server (`oapi-codegen`, strict server) and the TypeScript client (`openapi-typescript` + `openapi-fetch` in `packages/api-client`) are generated from it. CI fails the build when generated code is out of sync (`make gen-check`).

## Workflow for changing the API

1. Edit `api/openapi.yaml` (skill `add-endpoint`).
2. `make gen`.
3. Implement the handler and service in `apps/backend`, then the page in `apps/web`.
4. Write tests and update `docs/modules/<module>.md`.

## Spec rules

- Path: `/api/v1/<module>/<resource-plural-kebab>`, e.g. `/api/v1/inventory/stock-batches`. Non-CRUD actions are noun sub-resources (`POST /api/v1/purchase/goods-receipts`), not verbs in the path.
- `operationId`: `VerbNoun` PascalCase, **unique across the whole spec** (`ListProducts`, `ReceiveGoods`). Standard verbs: `List`, `Get`, `Create`, `Update`, `Delete`; otherwise a domain verb (`Receive`, `Post`, `Void`).
- Schemas: PascalCase (`Product`, `CreateProductRequest`, `ProductList`). JSON properties are `camelCase`.
- Every operation declares its permission with the extension `x-permission: <module>.<resource>.<action>`. An operation that needs no session also sets `x-public: true` (e.g. `GetHealth` uses `x-permission: platform.health.read` and `x-public: true`).
- Reuse the shared components: `Error`, `Money`, `Quantity`, and the `Limit` and `Cursor` parameters. Do not redefine money or quantity types per endpoint.
- `/healthz` is an operational probe used by Docker healthchecks and lives outside this contract; `/api/v1/health` is the in-contract liveness endpoint.
- Operations that create a transaction document accept `clientRequestId` (idempotency).
- Errors use a single shape (`Error`: `code` UPPER_SNAKE, `message` for developers in English, `details`). Error codes are listed in `docs/modules/<module>.md`. User-facing Indonesian text is produced by the frontend from `code`.
- Paged lists use the same `limit` and `cursor` on every resource.
- Money and quantities are sent as **decimal strings** (e.g. `"12500.00"`), not JSON numbers, to avoid losing precision.
- The version is in the path (`/v1`). A breaking change means a new version or an ADR.

## Do not

- Edit generated code in `apps/backend/**/gen/` or `packages/api-client/src/generated/`.
- Add an endpoint in code without a spec.
