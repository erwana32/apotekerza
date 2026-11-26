# Documentation

Docs, rules, and ADRs are written in **English**. User-facing text (UI, receipts, drug labels, purchase orders, reports) is **Indonesian**, and Indonesian pharmacy terms are kept in their original form where English has no clean equivalent. `glossary.md` maps each term to its English identifier.

## Map

| Path | Contents |
|---|---|
| `architecture.md` | System overview, module list, dependencies, build order, deferred items |
| `modules/<module>.md` | One file per module: responsibilities, status, domain rules, permissions, error codes, deferred items |
| `adr/NNNN-<title>.md` | One architecture decision per file (`0000-template.md` is the template) |
| `conventions/naming.md` | Naming standard for files, tables, API, routes, permissions |
| `glossary.md` | Indonesian pharmacy term ↔ English identifier in code |
| `branding.md` | Company logo source (`source/`), usage rules, brand colours |
| `prd/` | One product requirements document per phase, written by the developer before the phase starts. Overrides provisional assumptions in the module docs |
| `runbooks/` | Operational procedures (backup, restore, release) |

## Rules

1. **Docs and code change together.** A behaviour change must update `modules/<module>.md` in the same change. A change without the doc update does not pass review.
2. **PRD before each phase.** The developer writes `docs/prd/phase-<N>-<name>.md` before a phase starts. Where a PRD and a module doc disagree, the PRD wins for what it covers; the AI then updates the module doc to match in the same change and reports the difference. A PRD never overrides a hard rule in `AGENTS.md` (change the rule first, with an ADR).
3. **One topic, one canonical document.** Never keep two competing versions. If a document and the code disagree, report it; do not silently pick one.
4. **ADRs** are required for architecture decisions (module boundaries, technology, core data model, cross-module policy). Use the format in `adr/0000-template.md`. An ADR always describes the current decision. When a decision changes, edit the ADR and add a dated line to its Change log (what changed and why); do not create a new ADR just to replace an old one. Numbers are sequential and never reused.
5. **Module status** uses: `Planned`, `In progress`, `Stable`. Update it when it changes.
6. **Regulatory assumptions** (narcotics, psychotropics, prescriptions, tax, retention) are written as assumptions and tagged `Needs APJ confirmation`. Do not state them as fact until confirmed.
7. **No personal data.** Examples of patients, doctors, and staff use invented data.
8. A document opens with one sentence of purpose, then content. Avoid generic prose that is not specific to this project.
9. Cross-references use paths relative to the repo root (e.g. `docs/modules/inventory.md`).
10. Keep Indonesian terms in their original form and italicise on first use in a document (*resep*, *racikan*), with the English identifier in parentheses.
