# Module: purchase

Purchase orders (SP, per type), goods receipt, supplier invoices, returns, and supplier payments.

- **Status:** Planned
- **Depends on:** `masterdata`, `inventory`, `finance`
- **Table prefix:** `purchase_` (one schema, ADR-0001)

## Responsibilities

- **Purchase order (SP)** with `order_type`: `regular` (Biasa: OTC), `general` (Umum: F&B, skincare, etc.), `precursor`, `oot`, `psychotropic`, `narcotic`, `medical_device` (Alkes). Numbering is a separate sequence per `order_type` per store. Printable (browser print) with store and APJ details.
- **Goods receipt:** batch number and expiry date are **mandatory** for batch-tracked products; a receipt without them is rejected. Creates stock movements (`inventory`) and a payable (`finance`) in one transaction.
- **Supplier invoice (faktur PBF)** matched to the receipt; payable in `finance`.
- **Returns** to the supplier create reversing stock movements.
- **Payments** to suppliers through `finance`.

## SP approval (not in version 1)

SP approval by role (e.g. only the APJ for narcotics/psychotropics/precursors/OOT) is **not developed yet**. The design keeps `order_type` on every SP and a `status` field (shared vocabulary of ADR-0010: `draft`, `confirmed`, `cancelled`) with room for an approval step, so the workflow can be added without reshaping data. Which types require which approver `Needs APJ confirmation`.

## Defecta, cart, and purchase plan

The **Defecta** (buku/daftar defecta) is the list of drugs, medical devices, and other goods whose stock is low or empty, so they can be re-ordered. Defined by the developer (2026-11-25): it shows each product with `min_stock`, `max_stock`, last supplier, quantity on order (*sedang dipesan*), and quantity in the cart (*dalam keranjang*), plus a **Buy** button. Buying a row lets the user choose the supplier and the quantity and adds it to the cart; saving the cart automatically creates **draft purchase plans** (*rencana pembelian*).

Ownership: the Defecta, cart, and purchase plan belong to `purchase`. Stock comes from a join on the `inventory_*` stock tables, `min_stock`/`max_stock` from `masterdata_products`, last supplier and quantity on order from `purchase`'s own data. A purchase plan is not an SP: it is a draft that a person later turns into SP(s).

The first two items below (which rows show, suggested quantity) are **decided** because they follow from `min_stock`/`max_stock`. The rest are **accepted provisionally** (the developer chose to follow the recommendations for now, 2026-11-25) and will be detailed in the Phase 5 PRD (`docs/prd/`). Until then, build as written:

- **Which rows show (decided):** products whose stock is at or below `min_stock`. Rows already covered by quantity on order (stock + on order above `min_stock`) are hidden by default and can be shown with a filter.
- **Suggested quantity (decided)** = `max_stock` − (stock + on order + in cart), never below zero, computed in the smallest unit and rounded up to the product's default purchase unit.
- **Stock** counts only batches that are not expired.
- **Last supplier** = supplier of the most recent goods receipt of the product.
- **Quantity on order** = ordered minus received, on `confirmed` SPs that are not fully received (how much is received comes from the goods receipts, not from an SP status). `draft` SPs and draft purchase plans do not count as ordered.
- **Cart**: one cart per user. Saving it creates one draft plan per chosen supplier. Draft plans are editable until converted.
- **Plan → SP conversion**: splits by supplier and by `order_type` (a plan line of a controlled class goes on the matching SP type). Which classes must be on separate SPs `Needs APJ confirmation`.

## Reports

Purchase report (by period, supplier, product, order type).

## Public interface

`CreatePurchaseOrder`, `ReceiveGoods`, `RegisterSupplierInvoice`, `ReturnToSupplier`, `ListDefecta`, `AddToDefectaCart`, `SaveDefectaCart`, `ConvertPlanToOrders`.

## Permissions

`purchase.order.read|write`, `purchase.receipt.read|write`, `purchase.invoice.read|write`, `purchase.report.read`, `purchase.defecta.read`, `purchase.plan.read|write`.

## Error codes

`BATCH_REQUIRED`, `EXPIRY_REQUIRED`, `ORDER_TYPE_INVALID`, `SUPPLIER_REQUIRED`, `PLAN_ALREADY_CONVERTED`.

## Deferred

SP approval workflow, consignment, fully automatic ordering (the Defecta only suggests and creates drafts; a person converts a plan into SP(s)).
