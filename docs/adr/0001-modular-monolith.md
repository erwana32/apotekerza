# 0001. Modular monolith with one Postgres schema per module

- **Status:** Accepted
- **Date:** 2026-11-24

## Context

The product grows in stages from Finance, Inventory, and POS toward a pharmacy ERP, with a module reference similar to Odoo. A single developer works with AI tools. Strong consistency across modules (sale + stock + journal) is required.

## Decision

One Go binary and one PostgreSQL database, organised as modules (`identity`, `masterdata`, `finance`, `inventory`, `purchase`, `pos`, `pharmacy`). Each module owns one Postgres schema. A module is used by others only through its public `module.go`. Another module's tables are never read. Composite reads use public interfaces or read-only `v_` views published by the owner. `masterdata` is the only shared kernel.

## Consequences

- Cross-module atomicity is a plain database transaction (see ADR-0004).
- Boundaries are enforced by import lint and by schema ownership, not by the network.
- A module can be extracted into a service later because it already only talks through an interface.
- Discipline is required: a boundary violation is an architecture bug.

## Alternatives considered

- **Microservices from day one:** distributed transactions and operations cost far exceed the need for one pharmacy.
- **Single unstructured monolith:** fastest at first, but AI-assisted changes would erode boundaries quickly.
