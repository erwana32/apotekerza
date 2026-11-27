---
name: review-change
description: Review checklist for a change in apotekerza. Used by the read-only reviewer agent; use it to review a diff against the project's hard rules.
---

# review-change

Read-only. Use `git diff`, `git log`, `git show`. Do not edit files.

## Steps

1. Read the root `AGENTS.md`, the nested `AGENTS.md` of every folder in the diff, and `docs/modules/<module>.md` of every module touched.
2. Walk the checklist below. For each finding give `path/file:line`, the rule it breaks, and the consequence.
3. Report most severe first, and separate: hard-rule violations, confident bugs, suggestions. Do not report what you cannot point to in the code.

## Checklist

**Boundaries**
- Does any module import another module's `domain/`, `app/`, `httpapi/`, or `store/`? Does any code skip an owner's checks by writing its tables directly (journals without `finance.PostJournal`, stock without FEFO)? Reading other modules' tables is allowed (ADR-0001).
- Is there a dependency cycle against the table in `docs/architecture.md`?

**Ledger and stock**
- Any `UPDATE`/`DELETE` on journal or stock movement rows? Corrections must be reversing entries.
- Is a journal posted in the same transaction as its document and stock movements?
- FEFO respected? Expired batches blocked? Batch and expiry mandatory on receipt when tracked?
- Cost per batch, no averaging?

**Documents (ADR-0010)**
- Statuses only from `draft`, `confirmed`, `posted`, `cancelled`? Does a draft ever create journal lines or stock movements? Can a confirmed or posted document be edited or deleted (it must be cancelled by a reversing document)?
- Is the document number assigned by the server when the document leaves `draft`? Does the document page show its history?

**Money and numbers**
- Any `float` for money or quantity? Decimal strings in the API?
- Rounding only at documented points?

**API and generated code**
- Spec changed first? `operationId` unique and `VerbNoun`? `x-permission` set?
- Any hand edit of generated code or an existing migration?
- Database naming (`docs/conventions/naming.md`, section 2): table `<module>_<plural>`, explicit constraint/index/trigger names starting with the table (`_key`, `_idx`), booleans `is_`/`has_`, status as `text` + `CHECK` (no `ENUM`), no new schema, no `search_path`. Does a Go constant that matches a constraint name still match the migration?
- Handler thin; logic in the service?

**Security and data**
- Permission checked on the backend for every endpoint?
- Patient data or personal data in logs, tests, examples, or docs?
- Secrets or `.env` values committed?

**Tests and docs**
- Tests for finance/stock logic? Integration tests on a real PostgreSQL, no mocked database?
- `docs/modules/<module>.md` updated in the same change? ADR for an architecture decision?
- Regulatory claims written as assumptions with `Needs APJ confirmation`?

**Naming and language**
- File prefixes and test suffixes per `docs/conventions/naming.md` (`make check-naming` passes)?
- English in code, docs, and commits; Indonesian only for user-facing text; term names per `docs/glossary.md`?
