# Module: pharmacy

The pharmacy service: patients, doctors, prescriptions, compounding, drug labels, and the narcotics/psychotropics log.

- **Status:** Planned
- **Depends on:** `masterdata`, `inventory`, `pos`
- **Schema:** `pharmacy`

All regulatory points below are **assumptions** until confirmed: `Needs APJ confirmation`.

## Responsibilities

- **Patients and doctors:** owned only by this module. Other modules store only an ID. Patient data is sensitive: access requires `pharmacy.patient.read` and is audited.
- **Prescription (resep)** workflow: `received → verified (by the APJ) → prepared → dispensed`. Dispensing creates a sale through the `pos` interface with `prescription_ref`.
- **Compounding (racikan):** simple composition (ingredients and quantities) plus a compounding service fee.
- **Drug label (etiket):** printed from the browser (label sticker print CSS), Indonesian text.
- **Controlled-substance log:** narcotics and psychotropics movements are recorded from day one. Export for reporting is deferred, but the data model keeps what that export will need.

## Rules

- Dispensing consumes stock through `inventory` (FEFO) inside the sale transaction.
- APJ-only actions (assumption): verifying a prescription, dispensing narcotics/psychotropics, stock corrections.
- Prescription retention period is not decided: `Needs APJ confirmation`. Do not hard-code a number.
- No personal data in logs, tests, or examples.

## Public interface

`GetPrescription`, `VerifyPrescription`, `DispensePrescription`.

## Permissions

`pharmacy.patient.read|write`, `pharmacy.doctor.read|write`, `pharmacy.prescription.read|write|verify|dispense`, `pharmacy.controlled_log.read`.

## Error codes

`PRESCRIPTION_NOT_VERIFIED`, `PATIENT_REQUIRED`, `CONTROLLED_SUBSTANCE_RULE`.

## Deferred

Electronic prescription / SATUSEHAT, BPJS/PRB, telepharmacy, counselling documentation, controlled-substance report export.
