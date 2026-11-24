---
name: post-journal
description: Create an automatic journal entry from another module's transaction (sale, goods receipt, payment, stock adjustment) with the correct invariants. Use whenever business code must post to the ledger.
---

# post-journal

Journals are created by calling the `finance` module's public interface **inside the caller's transaction**. Other modules never write to `finance` tables.

## Step 0: show the journal on paper first

Do not write code yet. Follow teaching mode in `AGENTS.md`. Show a worked example with real numbers: the business event, each journal line (account, debit, credit), and why it balances. Explain any account the developer has not seen. Wait for approval, then continue.

## Steps

1. Read `docs/modules/finance.md` and ADR-0004.
2. In the caller's `service_*.go`, after the document and stock movements are saved and within the same transaction, call `finance.PostJournal` with:
   - a source reference (document type + ID),
   - `store_id` and `company_id`,
   - lines with account, debit or credit `*_amount`, and a description.
3. The journal must balance (debit = credit). `finance` rejects an unbalanced journal with `JOURNAL_UNBALANCED`; do not work around it.
4. Never `UPDATE` or `DELETE` a journal. To correct or void, call `finance.ReverseJournal` for the original, then post the correct one.
5. Use `decimal` for all amounts. Round only at the documented points (see `docs/modules/finance.md`), never inside the journal lines.
6. Write an integration test that proves: the journal balances, it is created in the same transaction as the document (force a failure after posting and confirm both roll back), and a void creates a reversing journal.

## Typical postings (confirm accounts in `finance` seed)

- Sale: cash/bank debit, sales revenue credit, VAT payable credit (if taxable); cost of goods sold debit, inventory credit (using batch cost).
- Goods receipt: inventory debit, accounts payable credit (VAT input if applicable).
- Supplier payment: accounts payable debit, cash/bank credit.
- Stock adjustment: inventory ↔ adjustment expense/income.

## Never

- Post from outside a transaction.
- Edit journal rows.
- Read `finance` tables directly.
