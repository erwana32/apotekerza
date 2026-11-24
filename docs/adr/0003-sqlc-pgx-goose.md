# 0003. pgx + sqlc + goose, no ORM

- **Status:** Accepted
- **Date:** 2026-11-24

## Context

Finance and stock code needs explicit, reviewable SQL, exact transactions, and exact numeric handling.

## Decision

Database access uses `pgx`. Queries are written as SQL and turned into Go by `sqlc`. Migrations are plain SQL files managed by `goose`, add-only. One Postgres schema per module. Money is `numeric(18,2)` and quantities `numeric(18,4)`, mapped to a decimal type in Go, never `float`.

## Consequences

- SQL is visible in review and testable against a real database.
- No hidden queries or lazy loading.
- More SQL to write than with an ORM, which is acceptable for a ledger-centred system.

## Alternatives considered

- **GORM or another ORM:** hides transaction and query behaviour that matters for ledger correctness.
- **Hand-written `database/sql`:** no type generation, more boilerplate.
