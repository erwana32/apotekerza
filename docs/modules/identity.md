# Module: identity

Users, sessions, access groups, permissions, and the audit trail.

- **Status:** Planned
- **Depends on:** none
- **Schema:** `identity`

## Responsibilities

- Authentication: password login, cookie sessions (`HttpOnly`, `Secure`, `SameSite`), argon2id password hashing, logout, session expiry.
- Authorisation: permissions `<module>.<resource>.<action>`; access groups = named sets of permissions; a user can hold several groups, permissions are the union.
- Audit trail: append-only record of who did what and when for patient-data access, stock movements and corrections, journal posting, and Administrator actions. Never stores personal data contents, only references.

## Version 1 behaviour

- One built-in group "Administrator" with all permissions (seeded). The Administrator creates further groups and users in the UI. Roles beyond Administrator are created by the Administrator later, so permissions must be listed in a registry each module contributes to.
- The permission registry is the list of all `x-permission` values in `api/openapi.yaml`.
- The `store` is single in version 1: no store scoping of roles.

## Public interface (`module.go`)

`Authenticate`, `CurrentUser(ctx)`, `HasPermission(ctx, permission)`, `RecordAudit(ctx, event)`.

## Permissions

`identity.user.read|write`, `identity.group.read|write`, `identity.audit.read`.

## Error codes

`INVALID_CREDENTIALS`, `SESSION_EXPIRED`, `PERMISSION_DENIED`.

## Deferred

Per-store role assignment, external identity providers, two-factor authentication.
