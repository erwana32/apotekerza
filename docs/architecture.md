# Architecture

This document describes the shape of apotekerza, the boundaries between modules, and the order in which it is built. Decisions made here are recorded as ADRs in `docs/adr/`.

## 1. Goal and version 1 scope

A staged ERP for an Indonesian pharmacy (apotek). Version 1 serves **one pharmacy** (one `store`) per installation: Finance, Inventory, Purchasing, POS, and the pharmacy service (prescriptions). The schema already carries `store_id` and `company_id` so that multiple pharmacies and multiple companies can be added later without a large migration.

**Working model: similar to Odoo** (developer's goal, 2026-11-26). The module set, the single schema with module-prefixed tables (ADR-0001), access groups, double-entry accounting, per-batch stock with FEFO, cashier sessions, and purchase → receipt → payable all follow how Odoo works. Documents follow one lifecycle (`draft → confirmed/posted → cancelled`, cancel = reversal, server-assigned sequence numbers), every resource has a list view and a form view, and each document shows its history (ADR-0010). Modules are not installable: all are always in the binary. Odoo behaviours beyond these are decided per phase in the PRDs; until then, do not assume them.

## 2. Overview

```
Browser (React SPA) --HTTPS--> Caddy --/api/*--> backend (Go, one binary) --> PostgreSQL
                                  `-- static web files
```

- **Modular monolith:** one Go binary, one PostgreSQL database, **one schema with tables prefixed by module (Odoo style)**. ADR-0001.
- **Spec-first API contract:** `api/openapi.yaml` generates the Go server and the TypeScript client. ADR-0002.
- **Data access:** `pgx` + `sqlc`, SQL migrations with `goose`, no ORM. ADR-0003.
- **Deployment:** one VPS, Docker Compose (`caddy`, `backend`, `postgres`). Caddy handles TLS, reverse proxy, and static files.
- **Phase 0 tooling:** stdlib `net/http`, `slog`, ESLint + TypeScript strict + Vitest + Playwright, module path `github.com/erwana32/apotekerza`. ADR-0009. The decimal library (`shopspring/decimal`) is decided in ADR-0009 and introduced with `platform/money` in Phase 1.1.

## 3. Modules and dependencies

| Module | Responsibility | Depends on | Status |
|---|---|---|---|
| `identity` | Users, sessions, access groups, permissions, audit trail | none | Planned |
| `masterdata` | Contacts (suppliers/customers; patients live in `pharmacy`), products, tiered units, tax, `store`. Shared kernel | `identity` | Planned |
| `inventory` | Batches and expiry, stock movements, FEFO, batch cost, min/max stock, stock take | `masterdata`, `finance` | Planned |
| `finance` | Chart of accounts, double-entry journals, general ledger, accounts payable, cash/bank, financial reports | `masterdata` | Planned |
| `purchase` | Defecta, cart, purchase plans (draft), purchase orders (SP, per type), goods receipt, supplier invoices, returns, payments | `masterdata`, `inventory`, `finance` | Planned |
| `pos` | Cashier sessions, sales, payments per channel, void, receipts | `masterdata`, `inventory`, `finance` | Planned |
| `pharmacy` | Patients, doctors, prescriptions, compounding, drug labels, narcotics/psychotropics log | `masterdata`, `inventory`, `pos` | Planned |

Details per module: `docs/modules/<module>.md`.

### Module boundary rules

1. A module imports only the public package of another module (`internal/<module>`, containing `module.go`). No imports into another module's `domain/`, `app/`, `httpapi/`, or `store/`.
2. **SQL is free across modules (ADR-0001).** One schema, tables prefixed by module (`inventory_stock_batches`). Queries may read, join, and write other modules' tables and foreign keys may cross modules, as in Odoo. Append-only ledger and stock tables are protected by triggers (ADR-0004), not by module boundaries. Posting through the owner's functions (`finance.PostJournal`, `inventory.Receive`) is the convention so their checks run.
3. `masterdata` is the only shared kernel every module may import in Go.
4. No circular dependencies. The direction is the table above. `pos` knows nothing about prescriptions: it only stores a neutral reference (`prescription_ref`). `pharmacy` creates transactions through the `pos` interface.
5. Enforcement: Go import-rule lint + `make check`. A Go import violation is an architecture bug. Direct writes that skip an owner's checks are caught in review (`review-change`).

### Cross-module transactions

Synchronous, in **one database transaction**. Example: a POS sale reduces stock (`inventory`), posts a journal (`finance`), and saves the document (`pos`) atomically. Asynchronous events (outbox) are only for non-critical side effects and are not needed in version 1. ADR-0004 and ADR-0005 describe the core data rules.

## 4. Core data rules

- **The ledger and the stock book are append-only.** A correction or void is a reversing journal or reversing stock movement. ADR-0004.
- **Money** is `numeric(18,2)` + `decimal`; item quantities are `numeric(18,4)`. Never `float`.
- **Stock is tracked per batch** (batch number + expiry date), leaves **FEFO** automatically, and is costed **per batch**. A product without batches uses one implicit batch (`tracks_batch = false`). ADR-0005.
- **Tiered units** (box → strip → tablet): stock is stored in the smallest unit.
- **Sale balance identity** is a property test (see `docs/modules/pos.md`).
- **IDs** are UUIDv7. Time is `timestamptz` UTC.

## 5. POS: online-first, open to offline

Version 1 is online-only. Accepted risk: the cashier cannot sell while the internet is down. So that offline can be added later without reworking the contract: draft IDs from the client (UUIDv7), `client_request_id` for idempotency, both client and server timestamps stored, invoice number assigned by the server at posting. Options if connectivity is unreliable: a local server at the pharmacy, or an offline POS. ADR-0006.

## 6. Access control

Permissions are `<module>.<resource>.<action>`. **Access groups** are created by the Administrator; a user may hold several groups and permissions are the union. Version 1 ships one built-in group, "Administrator" (all permissions, still recorded in the audit trail). Dedicated permissions guard sensitive data: `inventory.cost.read`, `pharmacy.patient.read`, `pos.sale.override_batch`. The backend checks permissions on every endpoint; the frontend hides menus accordingly. ADR-0007. Exception: `GET /api/v1/health` is a cross-cutting infrastructure operation with `platform.health.read` and `x-public: true`; `platform` is not a business module. `GET /healthz` is a separate operational liveness probe outside the API contract.

## 7. Version 1 reports

Stock/Inventory report (stock card per product and per batch), **Product List** (stock joined with the product master), **Defecta** (owned by `purchase`: shows min/max stock, last supplier, on order, in cart, and creates draft purchase plans), **Purchase report**, **Expiry report**, Sales report, General Ledger, Trial Balance, Income Statement, Balance Sheet, Accounts Payable per supplier (PBF), Financial Statements. Export to CSV and print from the browser. All are computed from the ledger and the stock book, with no separate summary tables.

## 8. Build order

1. Foundation: `identity`, `masterdata`.
2. `inventory` (batch, expiry, FEFO).
3. `finance` (journals, chart of accounts).
4. `purchase`.
5. `pos`.
6. `pharmacy`.

**First vertical slice** (walking skeleton): login → product master → receive goods with batch and expiry → sell at POS with FEFO → automatic journal, no prescription. This slice exercises every layer (OpenAPI, sqlc, cross-module transaction, integration tests, Docker, Caddy).

## 9. Deferred from version 1

Customer receivables and credit limits, membership/points, tiered pricing, bundles, WhatsApp sending, consignment, SATUSEHAT/BPJS, multi-pharmacy (a `store` picker in the UI), SP approval (the design keeps the SP type and a hook for it), racks as stock locations, offline POS, narcotics/psychotropics export (the data is still recorded), currencies other than IDR, e-Faktur.

## 10. Operations

See `deploy/AGENTS.md`. In short: local dev and production, secrets in the server `.env`, argon2id, daily encrypted `pg_dump` to storage outside the server, JSON logs.

## 11. AI tooling

OpenCode v2. `AGENTS.md` is canonical (root and nested), skills in `.opencode/skills/`, configuration and permissions in `opencode.json`, read-only `plan` and `reviewer` agents. ADR-0008, ADR-0008.
