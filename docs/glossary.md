# Glossary

Indonesian pharmacy and accounting terms mapped to English identifiers in code. Always look a term up here before naming anything. The Indonesian term stays in the UI, printed documents, and reports. Definitions marked *assumption* come from general knowledge of Indonesian pharmacy practice and `Need APJ confirmation`.

## Organisation and roles

| Indonesian | English / identifier | Meaning |
|---|---|---|
| Apotek | `store` (table `stores`, column `store_id`) | One pharmacy (one licence, one APJ) |
| Perusahaan | `company` (`company_id`) | Legal entity that owns the store(s) |
| APJ (Apoteker Penanggung Jawab) | `pharmacist_in_charge` | Pharmacist legally responsible for the store |
| TTK (Tenaga Teknis Kefarmasian) | `pharmacy_technician` | Pharmacy technician (assistant) |
| Kasir | `cashier` | Person operating the POS |
| Pemilik | `owner` | Business owner |
| Administrator | `administrator` | System administrator, full access in version 1 |
| Grup akses | `access_group` | A named set of permissions assigned to users |
| PBF (Pedagang Besar Farmasi) | `supplier` (type `pbf`) | Pharmaceutical wholesaler |
| Pasien | `patient` | Owned by module `pharmacy` |
| Dokter | `doctor` | Prescriber, owned by module `pharmacy` |

## Products and stock

| Indonesian | English / identifier | Meaning |
|---|---|---|
| Produk / Obat | `product` | Any sellable item (drug, medical device, general goods) |
| Master produk | `product` in `masterdata` | Product definition (name, units, tax, min/max stock) |
| Daftar Produk | page `ProductListPage` (optional view `v_inventory_product_stock`) | Product master joined with stock, different from the master data screen |
| Golongan obat | `drug_class` | `otc`, `limited_otc`, `prescription`, `oot`, `psychotropic`, `narcotic`, `precursor` (*assumption*) |
| Alat kesehatan (Alkes) | `medical_device` | Product type |
| Barang umum (Umum) | `general_goods` | Non-drug goods such as food and skincare |
| Obat bebas (Biasa) | `otc` | Over-the-counter drug |
| OOT (Obat-Obat Tertentu) | `oot` | Certain regulated drugs (*assumption*) |
| Psikotropika | `psychotropic` | Regulated class |
| Narkotika | `narcotic` | Regulated class |
| Prekursor | `precursor` | Regulated class |
| Batch | `batch` | A received lot of one product |
| Nomor batch | `batch_number` | Lot identifier |
| ED / Kedaluwarsa | `expiry_date` | Expiry date of a batch |
| FEFO | `fefo` | First Expired, First Out: allocation rule for leaving stock |
| Biaya batch | `batch_unit_cost` | Unit cost of a batch, used for cost of goods sold |
| HPP | `cost_of_goods_sold` | Cost of goods sold |
| HNA | `purchase_price` | Net purchase price from the supplier |
| Harga jual | `selling_price` | Selling price per unit |
| Satuan | `unit` | Unit of measure |
| Satuan bertingkat | `unit_conversion` | Box → strip → tablet with conversion factors |
| Stok minimum / maksimum | `min_stock` / `max_stock` | Reorder baseline on the product: below/at min means reorder, above max means do not |
| Defecta (buku/daftar defecta) | `defecta` | List of products whose stock is low or empty and need re-ordering; shows min/max stock, last supplier, on order, in cart; owned by `purchase` |
| Sedang dipesan | `on_order_quantity` | Quantity ordered on SPs that is not received yet |
| Keranjang (defecta) | `defecta_cart` | Per-user list of products chosen to buy, with supplier and quantity, before saving as a plan |
| Rencana pembelian | `purchase_plan` | Draft created when the cart is saved; later converted into SP(s) |
| Rak | `rack` | Shelf location of a product, informational only in version 1 |
| Kartu stok | `stock_card` | Per product/batch movement history |
| Stok opname | `stock_take` | Physical stock count and adjustment |
| Pergerakan stok | `stock_movement` | One append-only stock book entry |

## Purchasing

| Indonesian | English / identifier | Meaning |
|---|---|---|
| SP (Surat Pesanan) | `purchase_order` | Purchase order sent to a supplier |
| Jenis SP | `order_type` | `regular` (Biasa: OTC), `general` (Umum: F&B, skincare, etc.), `precursor`, `oot`, `psychotropic`, `narcotic`, `medical_device` (Alkes) |
| Penerimaan barang | `goods_receipt` | Receipt of goods; batch and expiry are mandatory |
| Faktur PBF | `supplier_invoice` | Invoice from the supplier |
| Hutang dagang | `accounts_payable` | Amount owed to suppliers |
| Retur | `return` | Return of goods to the supplier or from a customer |

## Sales and pharmacy service

| Indonesian | English / identifier | Meaning |
|---|---|---|
| Penjualan | `sale` | A POS transaction |
| Sesi kasir | `cashier_session` | One shift of one cashier, opened and closed with cash reconciliation |
| Struk | `receipt` | Printed sale receipt |
| Resep | `prescription` | Doctor's prescription |
| Racikan | `compounding` | Compounded preparation made from several ingredients |
| Embalase | `embalase_fee` | Packaging fee on a sale |
| Jasa / tuslah | `service_fee` | Service (e.g. compounding) fee |
| Etiket | `label` | Drug label printed for the patient |
| Penyerahan obat | `dispensing` | Handing drugs to the patient |
| Buku catat narkotika/psikotropika | `controlled_substance_log` | Log of narcotics/psychotropics movements |
| Void | `void` | Cancellation done as a reversing document, never a deletion |
| Status dokumen | `status` | `draft`, `confirmed`, `posted`, `cancelled` (ADR-0010) |
| Nomor dokumen | `document_number` | Number assigned by the server when a document leaves `draft` |
| Riwayat | `history` | Per-document log of who did what and when, built on the audit trail |
| Kanal pembayaran | `payment_channel` | Cash, transfer, QRIS, debit, ... (configurable) |
| Pembulatan | `rounding_amount` | Rounding adjustment on a sale |

## Accounting

| Indonesian | English / identifier | Meaning |
|---|---|---|
| Bagan akun | `chart_of_accounts` | List of accounts |
| Jurnal | `journal_entry` | Balanced set of lines (debit = credit) |
| Buku besar | `general_ledger` | All posted journal lines per account |
| Neraca saldo | `trial_balance` | Balance per account |
| Laba rugi | `income_statement` | Income statement |
| Neraca | `balance_sheet` | Balance sheet |
| Laporan keuangan | `financial_statements` | Set of financial reports |
| PPN | `vat` | Value-added tax, configured per product |
| Jurnal pembalik | `reversing_journal` | Journal that reverses an earlier one |
