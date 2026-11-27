# Start setup: phased prompts

Copy-paste prompts for building apotekerza with an AI coding tool, one phase and one vertical slice at a time. The code is written in this repository, which also holds the rules the AI must follow (`AGENTS.md`, `docs/`, `.opencode/skills/`). The AI tool is OpenCode (ADR-0008) and its working directory must be this repository. The app should work like Odoo (`docs/architecture.md`, section 1).

## How to use

1. Work **one prompt at a time**, in order. Each prompt is one reviewable slice.
2. Start every new chat session with the **Session preamble** below, then paste the prompt.
3. Make the AI **plan first and work step by step** (teaching mode, see `AGENTS.md`). It announces each step, proposes designs (tables, endpoints, rules) before writing them, waits for your OK, and explains afterwards. Say **"batch mode"** to let it do several steps at once, and **"teaching mode"** to go back.
4. One branch per slice: `slice/<short-description>`. Merge only when `make check` is green.
5. After each slice run the **Review prompt**, fix findings, then commit yourself. The AI never pushes and always asks before committing.
6. If the AI breaks a hard rule from `AGENTS.md`, stop it and paste the **Drift prompt**.
7. **Before each phase, write its PRD** in `docs/prd/phase-<N>-<name>.md` (goal, scope, screens, rules, acceptance criteria). The project is built step by step, so detail each phase when you reach it. The AI reads the PRD first; where it disagrees with a module doc, the PRD wins and the AI updates the module doc.
8. **ADRs describe the current decision.** When you change a decision, the AI edits that ADR and adds a dated line to its Change log (no chain of superseded ADRs). It creates a new ADR only for a new topic.
9. Replace `<...>` placeholders before pasting. Items marked **Decision** are yours to make first.

OpenCode reads `AGENTS.md` automatically; the preamble below makes it read the rest.

## Recurring prompts

### Session preamble (paste at the start of every session)

```text
You are working in the apotekerza repository (a staged ERP for an Indonesian pharmacy).
Before doing anything, read in this order: AGENTS.md, the PRD of the current phase in docs/prd/ if it exists (it overrides provisional assumptions in the module docs; report any difference you find), the nested AGENTS.md of every folder you will touch, docs/architecture.md, docs/conventions/naming.md, docs/glossary.md, and docs/modules/<module>.md for every module involved. Follow every hard rule in AGENTS.md. The app works like Odoo: one schema with module-prefixed tables and free SQL across modules (ADR-0001), documents with draft/confirmed/posted/cancelled statuses, server-assigned numbers, list and form views, and per-document history (ADR-0010).
Rules of engagement:
- Teaching mode (see AGENTS.md): plan first, then work one small step at a time. Before each step say what you will build, why, and the alternatives with your recommendation, and wait for my OK. For any table, endpoint, or posting rule, propose the design first (columns and types, constraints, indexes, relations; or path, request, response, permission, errors; or a worked example) and wait for approval. After each step explain what you built, how to verify it, and what comes next. Never run several steps on one OK.
- Work in one vertical slice. Do not touch modules outside the slice.
- Spec-first: change api/openapi.yaml, run make gen, then write code. Never edit generated code or existing migrations.
- Update docs/modules/<module>.md in the same change. For an architecture decision use the write-adr skill: edit the existing ADR if one covers the topic (add a Change log line), otherwise create the next ADR.
- Run make check at the end and report the real output. Do not git push. Ask before git commit.
- If something in the docs is ambiguous or an assumption needs APJ confirmation, ask me instead of guessing.
```

### Explain prompt (any time you want to learn more)

```text
Explain what you just built in plain words: the purpose, how the pieces connect (module, layer, files), why you chose this design over the alternatives, and what could go wrong. Then show me the command to run and what output to expect. Do not change anything.
```

### Review prompt (after every slice)

```text
Use the reviewer agent (skill review-change) to review the current diff against AGENTS.md, the nested AGENTS.md files, and docs/modules/<module>.md. Report findings most severe first with path:line, the rule broken, and the consequence. Separate hard-rule violations, confident bugs, and suggestions. Do not edit anything.
```

### Fix review findings

```text
Fix the findings below one at a time. For each: write or adjust a failing test first when it is a bug, make the minimal change, and re-run make check. Do not widen scope.
<paste findings>
```

### Drift prompt (when the AI breaks a rule)

```text
Stop. Your last change broke this rule: <quote the rule from AGENTS.md or the docs>. Revert that part, redo it correctly, and tell me what you changed. Then run make check-naming and make check.
```

### Docs sync prompt

```text
Compare the code in <module> with docs/modules/<module>.md. List every difference (behaviour, permissions, error codes, invariants, deferred items). Do not guess which side is right: report differences and propose the doc edit or code fix for each, then wait for my decision.
```

### ADR prompt

```text
We decided: <decision>. Use the write-adr skill to create the next ADR in docs/adr/. Reference it from docs/architecture.md and the affected module doc. If it changes a hard rule, update AGENTS.md in the same change.
```

### Bug prompt

```text
Bug: <what happens, what should happen>. First write a failing test that reproduces it (integration test with a real PostgreSQL if it involves stock, journals, or sales). Then fix it with the smallest change, re-run the test and make check. Do not refactor unrelated code.
```

---

## Progress

| Phase | Status |
|---|---|
| 0. Repository bootstrap (0.1–0.6) | Done. Merged on `main`. Technology choices are in ADR-0009 |
| 1. Foundation | In progress. **1.1** shared building blocks: done (PR #2). **1.2** identity backend: done (PR #3 data layer; PR #5 `app/` and `httpapi/` layers, argon2id, cookie sessions, 13 endpoints, permission guard from the spec, `cmd/createadmin`; hardening in `fix/identity-hardening`: createadmin in the image, last-Administrator protection, login timing, session error handling). Open decisions are in `docs/modules/identity.md` ("Known limitations"). **1.3–1.5** not started |
| 2–8 | Not started |

Update this table when a phase or prompt is finished. Current decisions live in `docs/adr/` (0001 modular monolith with one schema, Odoo style; 0004 append-only ledger; 0005 stock per batch and FEFO; 0008 OpenCode and rules; 0009 Phase 0 technology; 0010 Odoo-style documents, screens, history).

## Where the code lives

**Decided: the code is written in this repository** (`apotekerza`), alongside the rule kit (`AGENTS.md`, `docs/`, `.opencode/`, `opencode.json`, `Makefile`, `scripts/`, `source/`). The AI reads its rules from the working directory, so run every prompt below with this repository as the working directory. Docs and code then change together in one place.

---

## Phase 0: Repository bootstrap

**Status: done.** The prompts below are kept for reference, for example when starting another project from this kit.

**Goal:** an empty but fully wired skeleton where `make check` passes and the stack runs locally. No business logic.

**Decisions first:** Go module path (e.g. `github.com/<account>/apotekerza`), git hosting, HTTP router (stdlib `net/http` or a small router), logger (`slog`), decimal library for Go, which JS test and lint tools. Record non-obvious choices in an ADR.

**Prompt 0.1: repo and workspaces**

```text
Initialise the repository structure described in AGENTS.md and docs/architecture.md; do not create business modules yet.
- git init (do not commit), keep the existing files.
- apps/backend: go.mod with module path <GO_MODULE_PATH>, the Go version you pin (1.26 now; keep deploy/Dockerfile.backend on the same version), internal/platform/ placeholder packages.
- pnpm workspace at the root: apps/web (Vite + React + TypeScript + Tailwind + shadcn/ui) and packages/api-client (generated client, with a generate script).
- Keep the skeleton minimal. The root Makefile expects these names: package scripts `lint` and `test` in apps/web, a `generate` script in packages/api-client, and `go generate ./...` directives in apps/backend. Make the code match the Makefile, or fix the Makefile if it is wrong, and make sure make lint and make test work (they currently skip missing parts).
Plan first, then wait for my OK.
```

**Prompt 0.2: backend platform skeleton**

```text
Create apps/backend/internal/platform with: config from environment variables, structured JSON logging (slog) with a request ID middleware, a pgx connection pool, an HTTP server with graceful shutdown, a transaction helper (WithTx) used by services, and goose migration running at startup or via a command (as a separate step per deploy/AGENTS.md). Add a health endpoint. Follow the naming rules (role prefixes) and apps/backend/AGENTS.md. Add unit tests and one integration test using testcontainers PostgreSQL.
```

**Prompt 0.3: OpenAPI skeleton and code generation**

```text
Create api/openapi.yaml with the conventions from api/AGENTS.md: the shared Error shape, paging parameters (limit, cursor), decimal-string types for money and quantity, the x-permission extension, and one operation GetHealth. Wire oapi-codegen (strict server) into apps/backend and openapi-typescript + openapi-fetch into packages/api-client so make gen works and make gen-check fails when generated code is stale. Generated files must live in store/gen or **/gen/** and packages/api-client/src/generated so opencode.json denies hand edits.
```

**Prompt 0.4: Docker and Caddy**

```text
Create deploy/docker-compose.yml (caddy, backend, postgres with a named volume and healthchecks), deploy/Caddyfile (automatic TLS for the real domain, plain HTTP on localhost for local dev, reverse proxy /api/* to backend, static web build with SPA fallback), Dockerfiles for backend and web, and a local dev setup so make dev-up and make dev-down work. Secrets only via deploy/.env (never committed). Follow deploy/AGENTS.md. Do not remove volumes in any script.
```

**Prompt 0.5: quality gates**

```text
Make make check meaningful on the empty skeleton:
- golangci-lint config including import rules (depguard or equivalent) that enforce the module boundaries from docs/architecture.md (no imports into another module's domain/app/httpapi/store).
- ESLint, TypeScript strict, Vitest, and Playwright set up for apps/web.
- A testcontainers helper in internal/platform/testutil for integration tests.
- scripts/check-naming.sh stays as is; make check must call it.
Report the real output of make check. If it is not green, fix before finishing.
```

**Prompt 0.6: ADR for technology choices**

```text
We chose these technologies in Phase 0: <HTTP router>, <logger>, <decimal library>, <JS test and lint tools>, <Go module path>. Use the write-adr skill to record them as the next ADR in docs/adr/, with the alternatives considered. Reference it from docs/architecture.md.
```

**Exit criteria:** `make check` green; `make dev-up` starts caddy, backend, postgres; `GET /api/v1/health` works through Caddy; generated code is in sync.

---

## Phase 1: Foundation (platform shared code, identity, masterdata)

**Goal:** login, access groups, and the product/contact/unit/tax master data.

**Prompt 1.1: shared building blocks**

```text
In internal/platform add: money (decimal wrapper, JSON as decimal string, rounding helpers that make the rounding point explicit), ids (UUIDv7), clock (injectable time), pagination (limit/cursor helpers), and a domain-error to API-error mapper. Every file follows the naming rules. Table-driven unit tests for money (no float anywhere) and pagination. Document nothing in modules yet; update docs/architecture.md only if a decision changed.
```

**Prompt 1.2: identity backend**

```text
Implement module identity per docs/modules/identity.md and docs/adr/0007-access-groups-permissions.md.
- Migration for the identity tables (prefix identity_, one shared schema, ADR-0001). Scope is company_id only: identity_users and identity_audit_log carry company_id; identity_access_groups is global (no company_id, unique name); identity_group_permissions, identity_user_groups, and identity_sessions carry no scope column; username is unique across the installation; the Administrator seed needs no company id: users, access_groups, group_permissions, user_groups, sessions, and an append-only audit_log (block UPDATE/DELETE with a trigger); the audit_log is also the source of the per-document history (ADR-0010).
- argon2id password hashing, cookie sessions (HttpOnly, Secure, SameSite).
- Permission registry built from every x-permission in api/openapi.yaml; seed the built-in group "Administrator" with all permissions.
- Endpoints (spec first): login, logout, me, users CRUD, access groups CRUD, list permissions, read audit log.
- Middleware that enforces x-permission on every operation; module.go exposes Authenticate, CurrentUser, HasPermission, RecordAudit.
- Unit and integration tests, including: a user with two groups gets the union of permissions; an endpoint without permission returns PERMISSION_DENIED; audit rows cannot be updated.
Update docs/modules/identity.md.
```

**Prompt 1.3: identity frontend**

```text
In apps/web build the app shell and identity screens: BaseLayout (with the company logo from source/logo.svg per docs/branding.md: copy it at build time into apps/web/public/brand, never redraw or recolour it; define Tailwind tokens brand.blue, brand.pink, brand.green from the logo and a darker blue that passes WCAG AA for text and buttons), login page, route guard that reads each route's permission, the nested PATHS object, menu filtered by permission, and pages to manage users and access groups, each as a list view (search, filters, CSV export) and a form view per ADR-0010 (the Administrator edits permissions of a group per module/resource/action). All visible text in Indonesian; backend error codes mapped to Indonesian in shared/errorMessages. Follow apps/web/AGENTS.md and the naming rules. Vitest for logic and components, one Playwright test for login and permission-hidden menu.
```

**Prompt 1.4: masterdata backend**

```text
Implement module masterdata per docs/modules/masterdata.md.
- Tables masterdata_*: companies (the pharmacy; one seeded row with a documented fixed default id and empty parent_id, optional parent_id for branches later; ask me for the name, licence number, and APJ name, and use invented placeholders in seeds and tests, never real data; add the foreign key from identity_users.company_id to it), contacts (supplier incl. PBF, customer), units and tiered unit conversions (box -> strip -> tablet), tax rules (configurable rate, taxable flag per product), products (name, product_type drug/medical_device/general_goods, drug_class, tracks_batch, default rack, min_stock, max_stock, barcode(s), tax setting).
- Patients and doctors are NOT here.
- Endpoints spec-first with permissions masterdata.*.
- module.go: GetProduct, ListProducts, GetUnitConversion, GetContact, GetStore, GetTaxRule.
- Tests: unit conversion maths (always smallest unit), deactivating a product never deletes it.
Update docs/modules/masterdata.md.
```

**Prompt 1.5: masterdata frontend**

```text
Build the masterdata screens, each resource with a list view (search, filters, CSV export) and a form view per ADR-0010: product master list and form (including min_stock and max_stock, entered only in the product's smallest unit with the unit name shown next to the field, tracks_batch, rack, tiered units, tax setting), contacts, units, tax rules. This is the product master screen, not the Product List (stock joined with master) which comes in Phase 3. Indonesian UI text, id-ID number formatting, permissions on routes, tests per apps/web/AGENTS.md.
```

**Exit criteria:** log in as Administrator, create a group and a user, create products with tiered units; audit log records the actions; `make check` green.

---

## Phase 2: Walking skeleton (first vertical slice, thin across modules)

**Goal:** login -> create product -> receive goods with batch and expiry -> sell at POS with FEFO -> automatic balanced journal. Everything is minimal on purpose; it proves every layer works together. Documents already follow ADR-0010 (draft / confirmed / posted / cancelled, immutable once confirmed or posted, server-assigned numbers, list and form views).

**Prompt 2.1: finance minimal**

```text
Implement the minimum of module finance per docs/modules/finance.md and skill post-journal: tables finance_*, a seeded minimal chart of accounts (cash, inventory, accounts payable, VAT output, sales revenue, cost of goods sold), journal entries and lines (append-only via trigger), PostJournal (rejects unbalanced with JOURNAL_UNBALANCED; propose whether to also add a deferred constraint trigger that rejects an unbalanced journal even when another module writes the tables directly, ADR-0004, and wait for my decision) and ReverseJournal exposed in module.go. Integration tests: balance check, append-only, reversal. No reports yet.
```

**Prompt 2.2: inventory minimal**

```text
Implement the minimum of module inventory per docs/modules/inventory.md, ADR-0005, and skill stock-movement: tables inventory_*, batches (batch number, expiry date; implicit batch when tracks_batch is false), append-only stock_movements (trigger), Receive and ReserveAndIssue (FEFO, split across batches, skip expired, INSUFFICIENT_STOCK; the convention is that other modules call these instead of writing the tables directly, ADR-0004), cost per batch, each movement posts its journal through finance in the same transaction. Table-driven unit tests for FEFO allocation and an integration test proving concurrent issues cannot oversell a batch.
```

**Prompt 2.3: purchase goods receipt (minimal)**

```text
Implement only the goods receipt of module purchase: an endpoint and a page to receive goods against a supplier with mandatory batch number and expiry date for batch-tracked products (BATCH_REQUIRED / EXPIRY_REQUIRED). It calls inventory.Receive and creates the payable journal in one transaction. SP, supplier invoices, returns, and payments are later phases.
```

**Prompt 2.4: POS sale (minimal)**

```text
Implement the minimum of module pos per docs/modules/pos.md: a sale with lines, FEFO issue through inventory, single cash payment, journal (cash, revenue, VAT if taxable, COGS and inventory using batch cost), client_request_id idempotency (a replay returns the original result), server-assigned invoice number, and a minimal POS screen in PosLayout. Add the sale balance identity as a property test. No sessions, split payments, void, or receipts yet.
```

**Prompt 2.5: end-to-end test**

```text
Write one Playwright test for the walking skeleton: log in, create a product, receive goods with two batches of different expiry, sell more than one batch's quantity, then assert stock per batch (earliest expiry consumed first) and that the journal balances. Run it against the docker compose stack. Fix anything the test exposes.
```

**Prompt 2.6: document history**

```text
Add the simple version of document history (ADR-0010): every document page shows who did what and when and what changed (status changes, edits to a draft, posting, cancelling), read from the audit trail of module identity. Write it once as a reusable component and endpoint pattern, then use it on the goods receipt and the sale. No messages or attachments.
```

**Exit criteria:** the end-to-end test passes; document history works on the goods receipt and the sale; `docs/modules/*.md` for the touched modules match the code (run the Docs sync prompt); `make check` green. Run the Review prompt on the whole phase.

---

## Phase 3: Inventory (full)

**Prompt 3.1: stock card and Product List**

```text
Add the stock card per product and per batch (movement history with running balance) and the Product List: build the ProductListPage as a read query that joins the inventory_* stock tables with masterdata_products (cross-module reads are allowed, ADR-0001). Both support filters, CSV export, and browser print. Update docs/modules/inventory.md.
```

**Prompt 3.2: expiry report and stock figures**

```text
Add the Expiry report (expired and near-expiry, configurable threshold) with the warning shown on the stock screen and POS, with CSV export and print. Make sure the stock figures are available as one reusable query (or optional view v_inventory_product_stock), per product: the stock on hand (only batches that are not expired, in the smallest unit) and the nearest expiry, because the Defecta in module purchase will read it. The Defecta itself is built in Phase 5, not here. Update docs/modules/inventory.md.
```

**Prompt 3.3: stock take and opening balance**

```text
Add stock take (physical count produces adjustment movements and a journal through finance, permission inventory.stocktake.write) and the opening-balance goods receipt (its own journal, no shortcut around the ledger). Corrections are reversing movements only. Tests include a stock take that changes batch quantities and keeps the book append-only.
```

**Prompt 3.4: FEFO setting and override**

```text
Make FEFO automatic by default as a setting, and add the permission pos.sale.override_batch with a recorded reason for manual batch choice. Hide batch choice for cashiers without the permission. Cover with unit and integration tests and update docs/modules/inventory.md and docs/modules/pos.md.
```

---

## Phase 4: Finance (full)

**Prompt 4.1: chart of accounts and journals UI**

```text
Complete the chart of accounts (standard Indonesian seed, editable), journal list and detail, and manual journals with the same balance rules. Manual journals are still append-only; corrections use reversal. Document the rounding rules (per-line and sale rounding_amount) in docs/modules/finance.md.
```

**Prompt 4.2: ledger reports**

```text
Build General Ledger, Trial Balance, Income Statement, Balance Sheet, and Financial Statements. They read only the ledger (no summary tables), with period filters, CSV export, and browser print. Add integration tests that cross-check report totals against posted journals.
```

**Prompt 4.3: accounts payable and cash**

```text
Build accounts payable per supplier (PBF) with ageing, supplier payments, and cash/bank accounts. Payments post journals in the same transaction. Report: Accounts Payable per PBF with CSV and print. Customer receivables stay deferred.
```

---

## Phase 5: Purchase (full)

**Prompt 5.1: purchase orders (SP)**

```text
Build the purchase order (SP) per docs/modules/purchase.md: order_type regular, general, precursor, oot, psychotropic, narcotic, medical_device; a separate number sequence per type per company; printable (browser print, Indonesian text, company and APJ details); a status field that leaves room for a future approval step. Do NOT build approval now. Link goods receipt to an SP. Use the ADR-0010 statuses for the SP (draft, confirmed, cancelled); how much of a confirmed SP is received comes from the goods receipts, not from extra statuses, so quantity on order can be computed later.
```

**Prompt 5.2: Defecta, cart, and purchase plan**

```text
Build the Defecta in module purchase exactly as described in docs/modules/purchase.md (section "Defecta, cart, and purchase plan"). Show per product: min_stock, max_stock, stock (joined from the inventory_* tables), last supplier, quantity on order, quantity in the cart, and a Buy button. Buy opens a choice of supplier and quantity and adds the row to the cart; saving the cart creates draft purchase plans automatically (one per supplier). Add the conversion of a plan into SP(s), split by supplier and order_type. Read stock and min/max with joins on the inventory_* and masterdata_products tables (cross-module reads are allowed, ADR-0001). The assumptions listed in that section are not approved yet: propose the design for each (which rows show, suggested quantity, last supplier, on order, cart scope, plan to SP split) and wait for my decision before building. CSV export and print for the list. Update docs/modules/purchase.md and docs/glossary.md.
```

**Prompt 5.3: supplier invoices, returns, payments**

```text
Add supplier invoices (faktur PBF) matched to goods receipts, returns to the supplier (reversing stock movements and journals), and the payment flow into finance. Add the Purchase report (period, supplier, product, order type) with CSV export and print.
```

---

## Phase 6: POS (full)

**Prompt 6.1: cashier sessions**

```text
Add cashier sessions: open and close a shift with cash reconciliation and the difference posted through finance. A sale requires an open session (SESSION_NOT_OPEN).
```

**Prompt 6.2: payments and document charges**

```text
Add split payments across configurable channels (cash, transfer, QRIS, debit), document-level discount, embalase_fee, service_fee, shipping, additional cost, and rounding_amount, with tax per product. Keep the sale balance identity property test green and extended. Credit and receivables are deferred.
```

**Prompt 6.3: void**

```text
Add void: a reversing document with reversing stock movements and a reversing journal, never a delete. Permission pos.sale.void, audit entry, tests for stock and ledger after a void.
```

**Prompt 6.4: receipt and POS ergonomics**

```text
Add receipt printing from the browser (58/80 mm print CSS, Indonesian text), barcode scanner input (keyboard type), and a full-screen touch-friendly PosLayout. Add the Sales report (by period, session, product, payment channel) and gross profit per sale (sale minus batch cost) with CSV export and print.
```

**Prompt 6.5: prescription hook**

```text
Add the neutral prescription_ref field on sales and the rule that controlled drug classes require a verified prescription_ref (PRESCRIPTION_REQUIRED). pos must know nothing else about prescriptions. Mark which classes and rules need APJ confirmation in docs/modules/pos.md and ask me before hard-coding any class list.
```

---

## Phase 7: Pharmacy

**Prompt 7.1: patients and doctors**

```text
Build module pharmacy per docs/modules/pharmacy.md: patients and doctors owned only here (other modules store IDs), permission pharmacy.patient.read with audit on every read, and no personal data in logs, tests, or examples (use invented data).
```

**Prompt 7.2: prescription workflow and dispensing**

```text
Build the prescription workflow received -> verified (APJ) -> prepared -> dispensed. Dispensing creates a sale through the pos interface with prescription_ref and issues stock FEFO inside one transaction. Add simple compounding (ingredients, quantities, compounding fee). Mark every regulatory assumption with "Needs APJ confirmation" and ask me before choosing retention periods or APJ-only rules.
```

**Prompt 7.3: labels and controlled-substance log**

```text
Add drug label (etiket) printing from the browser (label print CSS, Indonesian text) and the controlled-substance log for narcotics and psychotropics recorded on every relevant movement. Export of that log stays deferred, but keep the data it will need.
```

---

## Phase 8: Hardening and go-live

**Prompt 8.1: initial data import**

```text
Add CSV import with templates for products (units, drug class, min/max stock), suppliers, opening stock per batch with expiry and batch cost, and opening account balances. Stock and balances must go through the opening-balance receipt/journal path (no shortcut around the ledger). Validate rows, report errors per row, and make the import idempotent.
```

**Prompt 8.2: backups and restore**

```text
Create the backup runbook and scripts in deploy/ and docs/runbooks/: daily encrypted pg_dump to storage outside the server, retention 30 daily plus 12 monthly, and a restore procedure that I can test monthly. Never use docker compose down -v. Ask me for the storage target; do not invent credentials.
```

**Prompt 8.3: security and audit review**

```text
Review the whole repository against AGENTS.md and docs for: permission checks on every endpoint, personal data in logs, secrets in git, session cookie flags, argon2id parameters, audit trail coverage (patient data, stock corrections, Administrator actions), and SQL injection or unsafe string building. Report findings with path:line and severity; do not fix until I decide.
```

**Prompt 8.4: staging and release**

```text
Add a staging environment (copy of compose with its own database and secrets), a release runbook (migrate, deploy, verify, rollback), and a smoke test script run after each deploy. Confirm with me the VPS details and domain; do not invent them.
```

**Prompt 8.5: APJ confirmation checklist**

```text
List every item in docs/ tagged "Needs APJ confirmation" (controlled-drug rules, prescription retention, APJ-only actions, SP types per approver, taxable products, Defecta formula). Produce one checklist document I can take to my APJ and accountant, with the exact current assumption for each item and the module doc it lives in.
```

**Exit criteria:** import tested on a copy of real data, backup and restore proven, security review findings resolved, APJ checklist answered and the docs updated, `make check` green, staging verified.

---

## After go-live

Pick one deferred item from `docs/architecture.md` section 9 (customer receivables, membership/points, tiered pricing, SP approval, offline POS, multi-pharmacy, SATUSEHAT/BPJS, controlled-drug report export), write an ADR for it, and run it as a new phase with the same pattern: preamble, plan, one slice per prompt, review, docs sync.
