# Module: inventory

Batches, expiry, stock movements, FEFO allocation, batch cost, stock take, and stock reports.

- **Status:** Planned
- **Depends on:** `masterdata`, `finance`
- **Table prefix:** `inventory_` (one schema, ADR-0001)

## Responsibilities

- Stock book: append-only `stock_movements` per product and batch (`quantity` signed, in the smallest unit, `batch_unit_cost`, reason, source document reference).
- Batches: batch number + `expiry_date` per received lot. Products with `tracks_batch = false` use one implicit batch.
- **FEFO allocation:** leaving stock takes the earliest `expiry_date` first. Automatic by default; manual batch choice needs `pos.sale.override_batch` and records a reason. Expired batches cannot be sold. A near-expiry warning threshold is configurable.
- **Costing:** cost per batch. COGS = sum(quantity × batch cost). Inventory value = sum(remaining batch quantity × batch cost). Extra purchase costs are allocated to batches at receipt. Price corrections after receipt need a correction document.
- Stock take: physical count creates adjustment movements and a journal through `finance`.
- Opening stock: a "opening balance" goods receipt that creates its own journal (no shortcut around the ledger).

## Reports (computed from the stock book)

Stock/Inventory report (stock card per product and per batch), **Product List** (stock joined with master data on screen), **Expiry report**. The **Defecta** is owned by `purchase` (it ends in a purchase plan); `inventory` owns the stock tables that the Defecta joins.

### Stock figures used by the Defecta

Per product: stock on hand (sum of batches that are not expired, in the smallest unit) and the nearest expiry. `purchase` may join the inventory tables directly (ADR-0001); a shared query or view `v_inventory_product_stock` is optional. The exact definition of "available stock" (whether near-expiry batches count) is an assumption: see `docs/modules/purchase.md`, Defecta.

## Rules and invariants

- Stock never goes negative by default; selling more than available is rejected with `INSUFFICIENT_STOCK` (a setting for backorder is deferred).
- Every movement is posted together with its journal in the same transaction (see `post-journal` and `stock-movement` skills).
- Corrections are reversing movements. No `UPDATE`/`DELETE` of movements.

## Public interface

`ReserveAndIssue` (FEFO), `Receive`, `Adjust`, `GetStock`, `GetBatchesForProduct`.

## Permissions

`inventory.stock.read`, `inventory.stock.adjust`, `inventory.stocktake.write`, `inventory.cost.read` (batch cost, sensitive), `inventory.report.read`.

## Error codes

`INSUFFICIENT_STOCK`, `BATCH_EXPIRED`, `BATCH_REQUIRED`, `EXPIRY_REQUIRED`.

## Deferred

Racks as stock locations, multi-warehouse transfers, consignment, FIFO as a per-product costing policy.
