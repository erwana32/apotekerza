---
name: stock-movement
description: Receive, issue, or adjust stock per batch with FEFO allocation and per-batch cost. Use whenever code changes stock quantity or batch data.
---

# stock-movement

Stock changes only through the `inventory` module's public interface, as append-only stock movements, in the caller's transaction.

## Rules (see `docs/modules/inventory.md`, ADR-0005)

- Quantity is stored in the **smallest unit**. Convert tiered units (box → strip → tablet) before calling.
- Every movement has: product, batch, signed `quantity`, `batch_unit_cost`, reason, and a source document reference. Movements are never updated or deleted; a correction is a reversing movement.
- **Receiving:** batch number and expiry date are mandatory for products with `tracks_batch = true` (`BATCH_REQUIRED`, `EXPIRY_REQUIRED`). Products without batch tracking use the implicit batch. Allocate extra purchase costs to batches at receipt.
- **Issuing (selling/dispensing):** allocate **FEFO** automatically: earliest `expiry_date` first, skipping expired batches (`BATCH_EXPIRED`), and split across batches when one batch is not enough. Manual batch choice is allowed only with the `pos.sale.override_batch` permission and a recorded reason.
- **Insufficient stock** is rejected with `INSUFFICIENT_STOCK`; stock never goes negative by default.
- **Cost:** COGS uses the cost of the batch(es) actually issued. Do not average across batches.
- Each movement is posted together with its journal (skill `post-journal`) in the same transaction.

## Step 0: show the allocation on paper first

Do not write code yet. Follow teaching mode in `AGENTS.md`. Show a worked FEFO example with numbers: the batches (quantity, expiry, cost), the sale or receipt, which batches are used, the resulting movements, and the cost. Include one edge case (split across batches, or an expired batch). Wait for approval, then continue.

## Steps

1. Read `docs/modules/inventory.md`.
2. Call `inventory.ReserveAndIssue` / `Receive` / `Adjust` from your `service_*.go`, inside the transaction.
3. Write unit tests (table-driven) for the FEFO allocation: ties on expiry, split across batches, expired batches skipped, quantity conversion.
4. Write an integration test with a real PostgreSQL: concurrent issues must not oversell the same batch (use row locking or a constraint, and test it).
5. Update the module doc if rules changed. Run `make check`.

## Never

- Update the stock book.
- Compute available stock by reading another module's tables.
- Sell an expired batch.
