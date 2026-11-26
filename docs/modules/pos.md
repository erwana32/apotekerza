# Module: pos

Cashier sessions, sales, payments per channel, voids, and receipts.

- **Status:** Planned
- **Depends on:** `masterdata`, `inventory`, `finance`
- **Table prefix:** `pos_` (one schema, ADR-0001)

## Responsibilities

- **Cashier session:** open and close a shift with a cash reconciliation difference. Its lifecycle is `open` → `closed` (not the document statuses of ADR-0010).
- **Sale:** lines (product, unit, quantity, price, discount), document-level charges (`embalase_fee`, `service_fee`, shipping, additional cost, `rounding_amount`), discount, tax. Stock leaves by FEFO automatically (a setting; manual batch choice needs `pos.sale.override_batch`).
- **Payments** split across configurable channels (cash, transfer, QRIS, debit, ...). The channel list is data, extensible later (credit is deferred).
- **Void:** a reversing document with reversing stock movements and journal. A sale is never deleted.
- **Receipt:** printed from the browser (58/80 mm print CSS). Indonesian text.
- **Idempotency:** each sale carries `client_request_id`; client and server timestamps are stored; the invoice number is assigned by the server at posting.
- Sales referencing a prescription only store a neutral `prescription_ref`; `pos` knows nothing about prescriptions otherwise.

## Rules and invariants

For every non-voided sale (property test):

```
total = sum(line_subtotal) - discount_total + service_fee + embalase_fee
        + additional_cost + tax_total + rounding_amount
sum(payments) >= total   (change = sum(payments) - total for cash)
```

How a prescription's own total enters this identity is decided with the `pharmacy` module and recorded here.

- Selling a controlled drug (`prescription`, `oot`, `psychotropic`, `narcotic`, `precursor` classes) requires a verified `prescription_ref`. Which classes and rules `Need APJ confirmation`.
- Sale, stock movements, and journal are one transaction.

## Reports

Sales report (by period, session, product, payment channel), gross profit per sale (sale amount minus batch cost).

## Public interface

`CreateSaleFromPrescription` (used by `pharmacy`), `GetSale`.

## Permissions

`pos.session.open|close`, `pos.sale.create|void|read`, `pos.sale.override_batch`, `pos.report.read`.

## Error codes

`SESSION_NOT_OPEN`, `INSUFFICIENT_STOCK`, `BATCH_EXPIRED`, `PRESCRIPTION_REQUIRED`, `PAYMENT_INSUFFICIENT`, `DUPLICATE_REQUEST` (idempotent replay returns the original result).

## Deferred

Offline mode, customer receivables, membership/points, price tiers, bundles, WhatsApp receipts.
