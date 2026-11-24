# 0007. Built-in authentication and permission-based access groups

- **Status:** Accepted
- **Date:** 2026-11-24

## Context

Version 1 needs login and role-based access, not federation. Roles must be configurable by the Administrator, and one user may hold several roles.

## Decision

Authentication is built into the `identity` module (passwords hashed with argon2id, cookie sessions with `HttpOnly`, `Secure`, `SameSite`) behind an interface so it can be replaced later. Authorisation uses permissions named `<module>.<resource>.<action>`. The Administrator creates **access groups** (sets of permissions); a user can belong to several groups and their permissions are the union. Version 1 ships one built-in group, "Administrator", with all permissions. The backend checks permissions on every endpoint (from `x-permission` in the spec); the frontend hides menus and buttons accordingly but is never the enforcement. Sensitive data has dedicated permissions (`inventory.cost.read`, `pharmacy.patient.read`, `pos.sale.override_batch`). Access to patient data, stock corrections, and Administrator actions is recorded in an audit trail. Per-store role scoping is deferred with multi-pharmacy.

## Consequences

- No per-role route files in the frontend: one route tree per module, each route declares its permission.
- Granularity stops at screen and action plus sensitive-data permissions, not per column everywhere.
- The Administrator has full access, so the audit trail is the control.

## Alternatives considered

- **External identity provider (Keycloak, Zitadel):** one more service to operate for a need that is only login and RBAC.
