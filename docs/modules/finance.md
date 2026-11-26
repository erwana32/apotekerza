# Module: finance

Chart of accounts, double-entry journals, general ledger, accounts payable, cash/bank, and financial reports.

- **Status:** Planned
- **Depends on:** `masterdata`
- **Table prefix:** `finance_` (one schema, ADR-0001)

## Responsibilities

- Chart of accounts seeded with a standard Indonesian set, editable.
- Journals: balanced (debit = credit), append-only. Corrections use `reversing_journal`.
- Automatic journals from other modules through the `finance` interface, inside the caller's transaction (sale, goods receipt, supplier payment, stock adjustment).
- General ledger, trial balance, income statement, balance sheet, financial statements.
- Accounts payable per supplier (PBF): invoices, payments, ageing.
- Cash and bank accounts; cashier session cash reconciliation differences.

## Rules and invariants

- Every journal balances. A journal is posted atomically with the document that caused it.
- Currency is IDR only in version 1; a `currency` column exists.
- Money is `numeric(18,2)`. **Rounding** happens only at documented points: per line total and at the sale `rounding_amount` (cash rounding). Document the exact rule here when implemented.
- Reports read only the ledger; there are no summary tables.
- VAT is posted according to the product's tax setting. Which items are taxable `Needs APJ confirmation` (and an accountant).

## Public interface

`PostJournal`, `ReverseJournal`, `GetAccountBalance`, `RegisterPayable`, `RecordPayment`.

## Permissions

`finance.journal.read|post`, `finance.account.read|write`, `finance.payable.read|write`, `finance.report.read`.

## Error codes

`JOURNAL_UNBALANCED`, `ACCOUNT_NOT_FOUND`, `PERIOD_CLOSED` (period closing deferred but reserved).

## Deferred

Customer receivables and credit limits, multi-currency, e-Faktur, period closing, consolidation across companies.
