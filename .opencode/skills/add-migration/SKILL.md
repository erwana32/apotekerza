---
name: add-migration
description: Change the database schema with an add-only goose migration and matching sqlc queries. Use for any table, column, index, or view change.
---

# add-migration

Migrations are **add-only**. Never edit a migration that already exists.

## Step 0: propose the table design and wait for approval

Do not write SQL yet. Follow teaching mode in `AGENTS.md`. Present, for each new or changed table:

- purpose in one sentence and which module owns it;
- a column table: name, type, nullable, default, and **why** for every non-obvious column (e.g. why `numeric(18,4)` for quantity, why `company_id`);
- primary key, foreign keys, unique constraints, check constraints, and the indexes with the query each one serves;
- whether it is append-only (and the trigger that enforces it);
- how it relates to other tables, as a small text diagram;
- alternatives considered (e.g. one table versus two, JSON column versus real columns) with your recommendation.

Start from what `docs/modules/<module>.md` already says about the entity. Wait for the developer's approval or changes. Only then continue with the steps below, one table per step.

## Steps

1. Read `apps/backend/AGENTS.md` and the module doc.
2. Create `apps/backend/migrations/<timestamp>_<module>_<description>.sql` (e.g. `20261124120000_inventory_create_stock_batches.sql`). Use goose `-- +goose Up` / `-- +goose Down` sections.
3. Rules for the SQL:
   - All objects live in the single `public` schema. Table names are `<module>_<plural_name>` (`inventory_stock_batches`), the prefix being the owning module (ADR-0001). Foreign keys may point to tables of other modules (`ON DELETE RESTRICT` by default).
   - Tables plural `snake_case`; primary key `id` (UUIDv7); foreign keys `<entity>_id`; scope column is `company_id` only: documents and their lines, operational data, users, and the audit log carry it; access groups, link tables, session tokens, and (version 1) products and contacts do not. No `store_id`, no `branch_id`.
   - `created_at`, `updated_at` as `timestamptz`.
   - Money `numeric(18,2)` named `*_amount`; quantities `numeric(18,4)`. Never `float`.
   - Ledger and stock tables are append-only: do not add update/delete paths. Corrections are reversing rows.
   - Views are optional, for reports only, named `v_<module>_<name>`. Other modules may read tables directly.
   - Add indexes for the queries you write and for every foreign key you filter on.
4. Write or update the sqlc queries in `store/queries/query_<entity>.sql` (`GetXByID`, `ListXs`, `InsertX`, `UpdateX`). Run `make gen`.
5. Add an integration test that exercises the new schema against a real PostgreSQL.
6. Update `docs/modules/<module>.md` if behaviour or invariants changed.
7. Run `make check`.

## Never

- Modify an existing migration (the permission config asks for confirmation on `apps/backend/migrations/*` for this reason).
- Backfill data inside a schema migration without a separate, reviewed step.
- Edit `store/gen/`.
