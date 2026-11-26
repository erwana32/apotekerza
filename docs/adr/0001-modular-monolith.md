# 0001. Modular monolith, one database, one schema with module-prefixed tables (Odoo style)

- **Status:** Accepted
- **Date:** 2026-11-24

## Context

The product grows in stages from Finance, Inventory, and POS toward a pharmacy ERP, and it should work like Odoo (see `docs/architecture.md`, section 1). A single developer works with AI tools. Strong consistency across modules (sale + stock + journal) is required.

## Decision

- **One Go binary and one PostgreSQL database**, organised as modules (`identity`, `masterdata`, `finance`, `inventory`, `purchase`, `pos`, `pharmacy`). `masterdata` is the shared kernel every module may import in Go.
- **One schema.** All tables live in the default `public` schema. There is no `CREATE SCHEMA <module>`.
- **Table prefix = owning module:** `<module>_<plural_name>`, e.g. `inventory_stock_batches`, `finance_journal_entries`, `masterdata_products`, `pos_sales`, `purchase_orders`, `pharmacy_patients`, `identity_users`. This is Odoo's layout, where `account_move` and `stock_move` sit side by side and any module can join any table.
- **Free cross-module access in SQL.** Any module may read, join, and write the tables of other modules, and foreign keys may point to any table (`ON DELETE RESTRICT` by default).
- **Go structure keeps the boundary.** Modules call each other only through `module.go`, and import lint forbids importing another module's `domain/`, `app/`, `httpapi/`, or `store/`. Only SQL is free; Go package boundaries are not.
- **Invariants live in the database.** The append-only rule of ADR-0004 is enforced by triggers on journal and stock movement tables, so it holds no matter which module writes.
- **Scope is `company_id` only, as in Odoo.** A pharmacy is a company; a branch (cabang) is a child company (`parent_id`), and a legal entity owning several pharmacies is their parent. There is no `store_id` or `branch_id`. `company_id` is on documents and lines, operational data, users, and the audit log; it is not on access groups (global), link tables, or session tokens, and not on products and contacts in version 1 (shared; an optional `company_id` is added when multi-company is built). A warehouse or location dimension is added only when one pharmacy needs several.
- **Cross-module transactions** are one database transaction (ADR-0004).
- Read-only views are optional, for reports, named `v_<module>_<name>`.

## Consequences

- Reports such as the Product List and the Defecta are plain joins over inventory, masterdata, and purchase tables.
- Coupling between modules is higher, and extracting a module into its own service later is harder, because other modules' SQL references its tables.
- Because other modules may write ledger and stock tables directly, checks that live in application code (journal balance, FEFO allocation, `INSUFFICIENT_STOCK`) can be bypassed by a direct `INSERT`. Mitigations are in ADR-0004.
- Reviews must watch for direct writes that skip the owner's checks (skill `review-change`).
- Strong cross-module consistency stays a plain database transaction.

## Alternatives considered

- **Microservices from day one:** distributed transactions and operations cost far exceed the need for one pharmacy.
- **One Postgres schema per module with no cross-module table access:** clearer ownership and a stronger boundary, but more ceremony for reports and joins, and not the Odoo layout the developer wants. It was the first design; the developer replaced it on 2026-11-26.
- **Single schema, reads free but writes only by the owning module:** the safest variant and the one recommended in discussion, because it protects ledger and stock invariants. The developer chose free writes as in Odoo.

## Change log

- 2026-11-26: changed from one schema per module (no cross-module table reads) to one schema with module-prefixed tables and free SQL access, to follow Odoo.
- 2026-11-27: removed the separate `store` concept; scope is `company_id` only with branches as child companies, following Odoo's schema (no `branch_id` column anywhere in Odoo; branches are companies with a parent).
