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

1. First check whether an existing ADR already covers the topic. If it does, **edit that ADR** (step 5) instead of creating a new one. Otherwise find the next number: list `docs/adr/` and take the highest `NNNN` + 1. Numbers are never reused.
2. Copy `docs/adr/0000-template.md` to `docs/adr/NNNN-<kebab-title>.md`.
3. Fill it: Status (`Proposed` until the developer accepts, then `Accepted`), Date (of the decision), Context (facts about this project), Decision (active voice, "We use ..."), Consequences (easier, harder, accepted risks), Alternatives considered (each with the main reason it was not chosen).
4. Reference the ADR from the relevant section of `docs/architecture.md` and from the affected `docs/modules/<module>.md`.
5. If the decision changes an existing ADR: rewrite its Context, Decision, Consequences, and Alternatives so they describe the **new** decision, keep the old option as an alternative with the reason it was dropped, and add a dated line to its Change log. Update every doc that cites it.
6. If the decision changes a hard rule, update `AGENTS.md` (root or nested) in the same change.

## Never

- Leave an ADR describing a decision that is no longer in force, or create a superseding ADR just to replace an old one.
- Write an ADR that states regulatory facts as certain; mark assumptions `Needs APJ confirmation`.
