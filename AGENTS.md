# Project instructions: apotekerza

Applies to every AI tool (OpenCode). This file is deliberately short. Detailed rules live in the nested `AGENTS.md` of each folder and in `docs/`.

This project stands alone. Do not reference, copy from, or take assumptions from other projects on this machine.

## Goal

A staged ERP for an Indonesian **pharmacy (apotek)**, starting with **one pharmacy** (one `company`) per installation. Version 1 scope: Finance, Inventory, Purchasing, POS, and the pharmacy service (prescriptions). The application code is built incrementally in vertical slices (see `docs/architecture.md`, "Build order").

Stack: Go, PostgreSQL, React (TypeScript) + Tailwind, OpenAPI, Docker Compose, Caddy.

## Language

- **English:** docs, rules, skills, ADRs, code, identifiers, table and endpoint names, commit messages, code comments.
- **Indonesian:** everything the end user sees or prints: UI labels, validation and error messages shown to users, receipts (struk), drug labels (etiket), purchase orders (SP), reports.
- **Indonesian pharmacy terms are kept as-is** when they have no clean English equivalent (resep, racikan, embalase, defecta, PBF, APJ, TTK, SP, OOT, ED). `docs/glossary.md` maps each term to its English identifier. Always look a term up there before naming anything.
- Reply to the developer in the language they write in.

## Repo map

| Path | Contents | Detailed rules |
|---|---|---|
| `apps/backend/` | Go backend, modular monolith | `apps/backend/AGENTS.md` |
| `apps/web/` | React frontend (Vite SPA) | `apps/web/AGENTS.md` |
| `packages/api-client/` | Generated TypeScript client. **Never edited by hand** | `api/AGENTS.md` |
| `api/openapi.yaml` | API contract, the source of truth | `api/AGENTS.md` |
| `deploy/` | Docker Compose, Caddyfile, `.env.example` | `deploy/AGENTS.md` |
| `docs/` | Architecture, modules, ADRs, conventions, glossary | `docs/README.md` |
| `.opencode/skills/` | Skills for repeatable procedures (`<name>/SKILL.md`) | table below |
| `scripts/` | Checker scripts (`check-naming.sh`) | |
| `source/` | `logo.svg` and `logo-mark.png`: **the only source of the company logo, never edited**. Also `start-setup.md` (the phased build guide) and `learn.md` (the hands-on learning guide for people) | `docs/branding.md` |

## Commands

`make help` lists everything. The important ones:

- `make check`: the gate before merging (naming, generated code in sync, lint, tests).
- `make gen`: generate code from `api/openapi.yaml` and SQL queries.
- `make test` / `make test-integration`: unit tests / integration tests (real PostgreSQL).
- `make test-e2e`: Playwright smoke tests (needs `pnpm exec playwright install chromium` once).
- `make dev-up` / `make dev-down`: the local Docker Compose stack (volumes are kept).

## Hard rules

1. **Module boundaries (ADR-0001).** In Go, a module calls another module only through that module's `module.go`, and never imports another module's `domain/`, `app/`, `httpapi/`, or `store/`. In SQL there is **one schema** and tables are prefixed by module (`inventory_stock_batches`); queries may read, join, and write tables of other modules, and foreign keys may point to any table. Prefer posting through `finance.PostJournal` and the inventory functions so their checks run (convention, not enforced).
2. **The ledger and the stock book are append-only.** Corrections and voids are reversing journals or reversing stock movements. Never `UPDATE` or `DELETE` journal or stock movement rows; database triggers enforce it whichever module writes (ADR-0004, ADR-0001).
3. **Money is never `float`.** Database `numeric`, Go `decimal`. Item quantities are `numeric` too.
4. **Spec-first.** Change `api/openapi.yaml` first, run `make gen`, then write handlers and the client.
5. **Never hand-edit** generated code (`**/gen/**`, `*.gen.go`, `packages/api-client/src/generated/**`) or existing migrations. A schema change is a new migration.
6. **Thin handlers.** A handler reads input, calls a service, and maps the result. Business logic and transaction boundaries live in services.
7. **Cross-module transactions are synchronous**, in one database transaction (sale + stock + journal succeed or fail together).
8. **Tests are mandatory** for finance and stock logic. Integration tests use a real PostgreSQL, never a mocked database.
9. **Naming** follows `docs/conventions/naming.md`. `make check-naming` enforces it.
10. **Patient data is sensitive.** Never put personal data in logs, tests, examples, or docs. Use invented data.
11. **Do not add modules, large dependencies, or features outside the list** in `docs/architecture.md` without the developer's approval. Deferred features are listed there.
12. **Do not invent pharmacy regulations** (narcotics, psychotropics, prescription retention, tax). Write them as assumptions and mark them `Needs APJ confirmation` in the module doc.
13. **Scope is `company` only (Odoo style, ADR-0001).** One pharmacy (apotek) is one `company` (table `masterdata_companies`, column `company_id`); a branch (cabang) is a child company (`parent_id`), and a legal entity owning several pharmacies is their parent company. There is no `store`, `store_id`, or `branch_id`. (`store/` is only the name of the Go data-access folder.)
14. **Documents follow ADR-0010.** Status from `draft`, `confirmed`, `posted`, `cancelled`; a draft never creates journal lines or stock movements; confirmed and posted documents are immutable and cancelled by a reversing document; the server assigns the number when the document leaves `draft`.

## Read first before touching

| When working on | Read |
|---|---|
| Cross-module change, new module | `docs/architecture.md`, `docs/adr/` |
| The current phase | `docs/prd/phase-<N>-<name>.md` if it exists (it overrides provisional assumptions in module docs) |
| A module | `docs/modules/<module>.md` |
| File, table, route, or API names | `docs/conventions/naming.md` |
| Pharmacy terms | `docs/glossary.md` |
| Logo, colours, any branded UI or printed document | `docs/branding.md` |

## Skills

| Skill | Use for |
|---|---|
| `add-module` | Creating a new module that respects module boundaries |
| `add-endpoint` | A new endpoint: OpenAPI → generate → handler → tests |
| `add-migration` | Schema migrations and sqlc queries |
| `post-journal` | Creating automatic journals with the right invariants |
| `stock-movement` | Stock movements per batch, FEFO, batch cost |
| `write-adr` | Recording an architecture decision |
| `review-change` | Review checklist (used by the `reviewer` agent) |

## Documentation travels with code

A behaviour change must update `docs/modules/<module>.md` in the same change. A new architecture decision needs an ADR in `docs/adr/` (skill `write-adr`). ADRs describe the **current** decision. When a decision changes, edit the ADR so it states the new decision and add a dated line to its Change log; do not leave a chain of superseded ADRs.

## Teaching mode (default: on)

The developer is learning while building, not only using. Work in small steps and get confirmation before each one. One step = one concept (one table, one endpoint, one service, one screen).

- **Before each step:** say "Step N of M", what you will build, why, where it fits (module and layer), the alternatives you considered with your recommendation, and the files you will touch. Then wait for the developer's OK. Never run several steps on one OK.
- **Design first when there is a choice.** For a table, propose the design before any SQL: columns with types and why, constraints, indexes, relations, whether it is append-only, and whether it carries `company_id` (see `docs/conventions/naming.md`). For an endpoint, propose path, `operationId`, request, response, permission, and error codes. For a posting or stock rule, show a worked example with numbers. Wait for approval or changes.
- **After each step:** explain what was built in plain words, point to the files, give the command to run and verify it, and name what comes next.
- **Explain new concepts once, briefly,** when first used (e.g. FEFO, double-entry, idempotency key, append-only). Use the Indonesian term with the English identifier from `docs/glossary.md`.
- The developer can say "batch mode" to allow several steps at once for the rest of the session, and "teaching mode" to return. Without that, stay in teaching mode.

## Workflow and agent authority

- One branch per vertical slice. Commit messages in English, Conventional Commits style (`feat(inventory): ...`).
- Merge to `main` only after `make check` is green.
- Agents **never** `git push`. `git commit` always asks first. Destructive commands (`git reset --hard`, `rm -rf`, `docker compose down -v`) are denied by `opencode.json`.
- The `plan` and `reviewer` agents are read-only.
- If these rules conflict with a request, state the conflict and ask. Never silently break a rule.
