# 0010. Odoo-style document states, screens, and record history

- **Status:** Accepted
- **Date:** 2026-11-26

## Context

The developer's goal is that the application works like Odoo (see `docs/architecture.md`, section 1). ADR-0001 already adopted Odoo's data layout. Odoo also has a recognisable working model for documents, screens, and per-record history. These behaviours were accepted before Phase 1 continued, so that every module is built the same way.

## Decision

1. **Document states.** Every transactional document (purchase order/SP, goods receipt, supplier invoice, sale, journal entry, stock take, purchase plan; for a purchase plan `confirmed` means converted into SP(s)) has a status from one shared vocabulary: `draft`, `confirmed`, `posted`, `cancelled`. Each document type's module doc lists which of these it uses and the allowed transitions. A **cashier session** is not a document in this sense: it has its own lifecycle `open` → `closed`, defined in `docs/modules/pos.md`. A `draft` can be edited and deleted; once `confirmed` or `posted` the document is immutable. **Cancelling is a reversing document** (reversing journal, reversing stock movement), never a deletion or an edit, consistent with ADR-0004. A draft never creates journal lines or stock movements.
2. **Sequence numbers.** The server assigns a document number when the document leaves `draft` (confirmed or posted), from a sequence per document type per company. SP sequences are per `order_type`. Number formats are set in each phase's PRD.
3. **Screens.** Every resource has a **list view** (search, filters, grouping, CSV export, print) and a **form view** for the record, as in Odoo. There is no kanban view.
4. **Record history.** Each document page shows its history (who did what, when, and what changed), built on the audit trail of ADR-0007. This is the equivalent of Odoo's chatter without messages or attachments. A simple version arrives in Phase 2.
5. **No installable modules.** Unlike Odoo, all modules are always present in the single binary. Features that need switching are settings, added only when needed.

## Consequences

- One consistent lifecycle and screen pattern across all modules makes PRDs, tests, and AI-assisted work uniform.
- Immutable confirmed documents plus reversal match the append-only ledger and stock book.
- Draft rows must be clearly separated from posted effects, because a draft never creates journal lines or stock movements.
- Every list screen has the same minimum work (filters, export), which adds effort per resource but removes per-screen design decisions.

## Alternatives considered

- **Free-form per-module statuses:** more flexible, but each module would reinvent transitions and reports would not share filters.
- **Editing posted documents with a log:** simpler for users, but it breaks the append-only ledger and the audit story.
- **Installable modules:** closer to Odoo, but adds a plugin system with no benefit for one pharmacy.

## Change log

- 2026-11-26: decided; cashier session excluded from the shared document statuses.
- 2026-11-26: clarified scope after Phase 1.2 — the lifecycle and server-assigned numbers apply to transactional documents only. Identity records (`users`, `access_groups`, `sessions`) are master records: edited in place while active and deactivated with `is_active`, never deleted.
- 2026-11-27: document number sequences are per company instead of per store, after ADR-0001 moved to `company_id` only.
