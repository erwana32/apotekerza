---
name: add-endpoint
description: Add or change an API endpoint spec-first: edit api/openapi.yaml, generate code, then write handler, service, tests, and frontend client usage. Use for any new or changed endpoint.
---

# add-endpoint

The contract comes first. Never write a handler without a spec entry.

## Step 0: propose the endpoint and wait for approval

Do not edit the spec yet. Follow teaching mode in `AGENTS.md`. Present: path and method, `operationId`, request and response fields (with types and decimal strings for money), the permission, the error codes, and how it maps to the use case. Give your recommendation if there are alternatives (e.g. sub-resource versus action). Wait for approval, then continue step by step.

## Steps

1. Read `api/AGENTS.md` and `docs/modules/<module>.md`.
2. Edit `api/openapi.yaml`:
   - Path `/api/v1/<module>/<resource-plural-kebab>`; non-CRUD actions are noun sub-resources.
   - `operationId` `VerbNoun`, unique across the spec.
   - Schemas PascalCase, properties `camelCase`. Money and quantities are decimal **strings**.
   - `x-permission: <module>.<resource>.<action>` on the operation.
   - Operations that create a transaction document take `clientRequestId`.
   - Error responses use the shared `Error` shape with an UPPER_SNAKE `code`.
3. Run `make gen`. Never edit the generated files.
4. Implement:
   - `service_<use_case>.go` in `app/` (owns the transaction, enforces rules).
   - `handler_<resource>.go` in `httpapi/` (thin: read input, call service, map result and errors). The method name equals the `operationId`.
   - `mapper_<resource>.go` for domain ↔ API type conversion.
   - Queries in `store/queries/query_<entity>.sql` (skill `add-migration` if the schema changes).
5. Write tests: unit tests for pure logic, an integration test (real PostgreSQL) for any posting, stock, or journal behaviour.
6. Frontend: use the generated client in `packages/api-client`, route and permission per `apps/web/AGENTS.md`.
7. Update `docs/modules/<module>.md` (permissions, error codes, behaviour) in the same change.
8. Run `make check`.

## Checklist

- Does the handler contain business logic? Move it to the service.
- Is the permission declared and checked?
- Is the error `code` listed in the module doc and mapped to Indonesian text in the frontend?
