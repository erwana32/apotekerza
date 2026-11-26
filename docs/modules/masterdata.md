# Module: masterdata

Shared kernel: contacts, products, tiered units, tax, and the `company` (the pharmacy). The only module every other module may import.

- **Status:** Planned
- **Depends on:** `identity`
- **Table prefix:** `masterdata_` (one schema, ADR-0001)

## Responsibilities

- **Company (apotek):** `masterdata_companies`, one row in version 1 with `parent_id` empty, holding the licence number (SIA) and the APJ name. A branch (cabang) is a child row (`parent_id`); a legal entity owning several pharmacies is their parent row. `company_id` is the only scope column (ADR-0001). A fixed default company id is a documented constant; this module's migration creates that row, and tables created earlier (e.g. `identity_users`) get their foreign key to it here.
- **Contacts:** suppliers (including PBF) and customers. Patients and doctors are **not** here: they belong to `pharmacy` (sensitive data).
- **Products:** one definition per sellable item: name, `product_type` (`drug`, `medical_device`, `general_goods`), `drug_class`, `tracks_batch`, default `rack` (shelf location, informational), `min_stock`, `max_stock`, VAT setting, barcode(s).
- **Units:** tiered units with conversion factors (box → strip → tablet). Stock is stored in the smallest unit; selling price is defined per unit.
- **Tax:** VAT is configuration per product (taxable or not, rate editable). Rates are not hard-coded. `Needs APJ confirmation` / accountant confirmation for which products are taxable.

## Rules

- `min_stock` and `max_stock` live on the product master and are the baseline for the Defecta (owned by `purchase`): at or below `min_stock` means reorder, above `max_stock` means do not.
- `min_stock` and `max_stock` are **entered on the product master form** (masterdata), not on the Defecta. Decided by the developer: both must be entered and stored in the **smallest unit** of the product (the form shows the unit name next to the field and accepts no other unit). Also decided (2026-11-25, developer accepted the recommendation): both are optional (empty means the product is not part of the Defecta); `max_stock` must be greater than or equal to `min_stock`; version 1 has one value per product (a per-store value comes with multi-pharmacy).
- Product master is distinct from the **Product List** screen, which is a read-only join of master data and stock (a query joining `masterdata_products` with the stock tables of `inventory`).
- Deactivating a product never deletes it (history must stay readable).

## Public interface

`GetProduct`, `ListProducts`, `GetUnitConversion`, `GetContact`, `GetStore`, `GetTaxRule`.

## Permissions

`masterdata.product.read|write`, `masterdata.contact.read|write`, `masterdata.company.read|write`, `masterdata.tax.read|write`.

## Deferred

Multi-company management UI (company picker, branch tree, consolidated reports), price tiers (membership, minimum quantity), bundles.
