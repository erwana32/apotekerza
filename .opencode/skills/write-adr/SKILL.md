---
name: write-adr
description: Record an architecture decision as an ADR in docs/adr. Use when a decision affects module boundaries, technology, the core data model, or a cross-module policy.
---

# write-adr

## When an ADR is required

Module boundaries, technology choices, core data model rules (ledger, stock, ids, money), cross-module policies, and anything that makes a hard rule in `AGENTS.md` change. Small implementation choices do not need one.

## Step 0: discuss the decision first

Follow teaching mode in `AGENTS.md`. State the decision, the context, and the alternatives in a few lines and wait for the developer to confirm before writing the file.

## Steps

1. Find the next number: list `docs/adr/` and take the highest `NNNN` + 1. Numbers are never reused.
2. Copy `docs/adr/0000-template.md` to `docs/adr/NNNN-<kebab-title>.md`.
3. Fill it: Status (`Proposed` until the developer accepts, then `Accepted`), Date, Context (facts about this project), Decision (active voice, "We use ..."), Consequences (easier, harder, accepted risks), Alternatives considered (each with the main reason it was not chosen).
4. Reference the ADR from the relevant section of `docs/architecture.md` and from the affected `docs/modules/<module>.md`.
5. If it replaces an earlier decision: set the old ADR's status line to `Superseded by NNNN`. This status line is the **only** allowed edit to an accepted ADR. The new ADR states which one it supersedes.
6. If the decision changes a hard rule, update `AGENTS.md` (root or nested) in the same change.

## Never

- Edit the body of an Accepted ADR.
- Write an ADR that states regulatory facts as certain; mark assumptions `Needs APJ confirmation`.
