# Module: identity

Users, sessions, access groups, permissions, and the audit trail.

- **Status:** In progress (Phase 1.2, backend)
- **Depends on:** none
- **Table prefix:** `identity_` (one schema, ADR-0001)

## Responsibilities

- Authentication: password login, cookie sessions (`HttpOnly`, `Secure`, `SameSite`), argon2id password hashing, logout, session expiry.
- Authorisation: permissions `<module>.<resource>.<action>`; access groups = named sets of permissions; a user can hold several groups, permissions are the union.
- Audit trail: append-only record of who did what and when for patient-data access, stock movements and corrections, journal posting, and Administrator actions. Never stores personal data contents, only references. It also feeds the per-document history shown on every document page (ADR-0010).

## Records are master records, not documents

Users, access groups, and sessions are **not** documents. ADR-0010's lifecycle (`draft` → `confirmed`/`posted` → `cancelled`) and server-assigned document numbers apply to transactional documents only. Identity records are edited in place while active and **deactivated** (`is_active = false`) instead of deleted, so history and foreign keys stay intact. See the Change log of ADR-0010.

## Tables

All tables live in the single `public` schema with the `identity_` prefix (ADR-0001). Migration: `20261126120000_identity_create_tables.sql`.

| Table | Purpose | Notes |
|---|---|---|
| `identity_users` | Login accounts | Carries `company_id` (the user's company). `username` is unique across the installation (login is typed before any company is known, as in Odoo), stored lowercase (`CHECK` + unique index). `password_hash` holds a self-describing argon2id string. `is_active` replaces delete. |
| `identity_access_groups` | Named sets of permissions | **Global**: no `company_id`, `name` unique (as Odoo's groups). `is_builtin` marks the seeded `Administrator` group, which the UI must not delete. |
| `identity_group_permissions` | Permissions granted to a group | PK `(access_group_id, permission)`. One row per permission string from the registry. |
| `identity_user_groups` | Group membership | PK `(user_id, access_group_id)`; a user's permissions are the union over their groups. |
| `identity_sessions` | Login sessions | Stores only the SHA-256 of the cookie token (`token_hash`, 32 bytes), so a database read cannot be replayed. `expires_at` and `revoked_at`. |
| `identity_audit_log` | Audit trail and per-document history | **Append-only**, enforced by a trigger rejecting `UPDATE`/`DELETE` (SQLSTATE `23001`), the same rule as the ledger (ADR-0004). |

### `identity_audit_log` event shape

| Column | Meaning |
|---|---|
| `actor_user_id` | Who acted; `NULL` for system actions such as migrations. |
| `action` | Stable verb: `create`, `update`, `deactivate`, `login`, `logout`, and later `confirm`, `post`, `cancel`. |
| `entity` | Type of record, e.g. `user`, `access_group`, `goods_receipt`, `sale`. |
| `entity_id` | The record or document this event belongs to; the history page reads by `(entity, entity_id)`. |
| `changes` | `jsonb` of field → `{from, to}` **references only**; never personal-data contents (hard rule 10). |

The audit log carries `company_id` (the company whose record the event concerns), because history is read per company. Link tables (`identity_group_permissions`, `identity_user_groups`) and `identity_sessions` carry no scope column; they inherit it through their foreign keys.

## Seeding

- The built-in **Administrator** group is global (no company), so its seed needs no company id. It is seeded by migration `20261126121000_identity_seed_administrator.sql` with `is_builtin = true`.
- Its permissions are seeded per migration, one row per `x-permission` in `api/openapi.yaml`. Permissions are data, so they are recorded by migration rather than synced at startup: every migration that adds an endpoint with a new permission also grants it to Administrator.
- **No user is seeded.** The first Administrator requires a real password, so it is created by a command (`cmd/createadmin`), never by a default credential.

## Version 1 behaviour

- One built-in group "Administrator" with all permissions (seeded). The Administrator creates further groups and users in the UI. Roles beyond Administrator are created by the Administrator later, so permissions must be listed in a registry each module contributes to.
- The permission registry is the list of all `x-permission` values in `api/openapi.yaml`.
- The company is single in version 1: no per-company scoping of roles.

## Public interface (`module.go`)

`Authenticate`, `CurrentUser(ctx)`, `HasPermission(ctx, permission)`, `RecordAudit(ctx, event)`.

## Permissions

`identity.user.read|write`, `identity.group.read|write`, `identity.audit.read`.

## Error codes

`INVALID_CREDENTIALS`, `SESSION_EXPIRED`, `PERMISSION_DENIED`.

## Deferred

Per-company role assignment (a user's allowed companies), external identity providers, two-factor authentication.