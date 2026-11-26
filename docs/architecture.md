# Architecture

This document describes the shape of apotekerza, the boundaries between modules, and the order in which it is built. Decisions made here are recorded as ADRs in `docs/adr/`.

## 1. Goal and version 1 scope

A staged ERP for an Indonesian pharmacy (apotek). Version 1 serves **one pharmacy** (one `store`) per installation: Finance, Inventory, Purchasing, POS, and the pharmacy service (prescriptions). The schema already carries `store_id` and `company_id` so that multiple pharmacies and multiple companies can be added later without a large migration.

## 2. Overview

```
Browser (React SPA) --HTTPS--> Caddy --/api/*--> backend (Go, one binary) --> PostgreSQL
                                  `-- static web files
```

- **Modular monolith:** one Go binary, one PostgreSQL database, **one Postgres schema per module**. ADR-0001.
- **Spec-first API contract:** `api/openapi.yaml` generates the Go server and the TypeScript client. ADR-0002.
- **Data access:** `pgx` + `sqlc`, SQL migrations with `goose`, no ORM. ADR-0003.
- **Deployment:** one VPS, Docker Compose (`caddy`, `backend`, `postgres`). Caddy handles TLS, reverse proxy, and static files.
- **Phase 0 tooling:** stdlib `net/http`, `slog`, `shopspring/decimal`, ESLint + TypeScript strict + Vitest + Playwright, module path `github.com/erwana32/apotekerza`. ADR-0010.

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
2. **Never read another module's tables or schema.** Composite views use the public interfaces of both modules or a **read-only SQL view prefixed `v_`** published by the owning module as a contract (e.g. `inventory.v_product_stock`).
3. `masterdata` is the only shared kernel every module may import.
4. No circular dependencies. The direction is the table above. `pos` knows nothing about prescriptions: it only stores a neutral reference (`prescription_ref`). `pharmacy` creates transactions through the `pos` interface.
5. Enforcement: import-rule lint + `make check`. A violation is an architecture bug.

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

Permissions are `<module>.<resource>.<action>`. **Access groups** are created by the Administrator; a user may hold several groups and permissions are the union. Version 1 ships one built-in group, "Administrator" (all permissions, still recorded in the audit trail). Dedicated permissions guard sensitive data: `inventory.cost.read`, `pharmacy.patient.read`, `pos.sale.override_batch`. The backend checks permissions on every endpoint; the frontend hides menus accordingly. ADR-0007.

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

OpenCode v2. `AGENTS.md` is canonical (root and nested), skills in `.opencode/skills/`, configuration and permissions in `opencode.json`, read-only `plan` and `reviewer` agents. ADR-0008, ADR-0009.
