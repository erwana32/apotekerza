# 0005. Stock per batch, FEFO allocation, cost per batch

- **Status:** Accepted
- **Date:** 2026-11-24

## Context

A pharmacy must know the expiry of what it sells and must sell the earliest-expiring stock first. FEFO is an allocation rule, not a valuation method, so costing must be decided separately.

## Decision

Stock is tracked per batch (batch number + expiry date). Leaving stock is allocated automatically by FEFO (earliest expiry first); a cashier cannot choose a batch unless they hold `pos.sale.override_batch`, and the override reason is recorded. Expired batches cannot be sold, and a near-expiry warning threshold is configurable. Cost is taken **per batch** (the batch unit cost from receipt), so cost of goods sold is the sum of quantity × batch cost, and inventory value is the sum of remaining batch quantity × batch cost. `tracks_batch` is a product attribute; a product without it uses one implicit batch. Stock is stored in the smallest unit.

## Consequences

- No averaging calculation to maintain. Purchase price differences are preserved per batch.
- Corrections of a purchase price after receipt need a correction document. Extra purchase costs are allocated to batches at receipt.
- Reports (stock card, expiry, defecta) read the stock book directly.

## Alternatives considered

- **Moving average per product:** simpler to describe, but loses the actual cost of each batch and adds an averaging step.
- **Stock per product only:** cannot support expiry handling.
