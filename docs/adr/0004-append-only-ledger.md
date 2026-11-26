# 0004. Double-entry, append-only ledger and stock book

- **Status:** Accepted
- **Date:** 2026-11-24

## Context

Sales, goods receipts, and stock movements must produce journals automatically, and the pharmacy needs an auditable history. Cash-only records would have to be rewritten later.

## Decision

Finance uses double-entry bookkeeping from the start. Journal lines and stock movements are append-only: a correction or void is a reversing journal or reversing stock movement, never an `UPDATE` or `DELETE`. Cross-module posting (sale, stock, journal) happens synchronously in one database transaction. Sales carry an idempotency key. A balance identity for sales is a property test. The append-only rule is enforced by database triggers on the journal and stock movement tables, so it holds whichever module writes (ADR-0001).

## Consequences

- Reports are computed from the ledger and stock book with no summary tables, so they always match the transactions.
- History is complete and auditable.
- Voids create documents instead of removing rows, so reports must handle reversals.
- Other modules may write these tables directly (ADR-0001), so checks in application code (journal balance, FEFO, `INSUFFICIENT_STOCK`) can be skipped by a direct `INSERT`. Recommended mitigations, to be decided in the Phase 2 PRD: a deferred constraint trigger that rejects an unbalanced journal, and the convention (not enforced) that code posts through `finance.PostJournal` and `inventory.Receive`/`ReserveAndIssue`.

## Alternatives considered

- **Simple cash in/out:** easier at first, but automatic journals from POS and stock would require rewriting the data later.
- **Asynchronous posting (outbox) for journals:** risks a sale existing without its journal. Outbox stays only for non-critical side effects.

## Change log

- 2026-11-26: added that triggers enforce append-only whichever module writes, and the mitigations for direct writes, after ADR-0001 moved to one schema with free SQL access.
