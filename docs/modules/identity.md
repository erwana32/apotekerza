# Module: identity

Users, sessions, access groups, permissions, and the audit trail.

- **Status:** Backend done (Phase 1.2); web screens come in Phase 1.3
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
- **No user is seeded.** The first Administrator requires a real password, so it is created by a command (`cmd/createadmin`), never by a default credential. The password is read from `CREATEADMIN_PASSWORD` so it never appears in the process list or in shell history:

  ```sh
  CREATEADMIN_PASSWORD='...' go run ./cmd/createadmin -company <company-uuid> -username admin
  ```

  In the Docker stack the same binary is in the backend image (`Dockerfile.backend`), and `DATABASE_URL` is already set in that container:

  ```sh
  # production, on the server
  docker compose -p apotekerza-prod --env-file deploy/.env \
    -f deploy/docker-compose.yml -f deploy/docker-compose.prod.yml \
    exec -e CREATEADMIN_PASSWORD='...' backend /app/createadmin -company <company-uuid> -username admin

  # staging, on a laptop
  docker compose -p apotekerza-stg --env-file deploy/.env.stg -f deploy/docker-compose.yml \
    exec -e CREATEADMIN_PASSWORD='...' backend /app/createadmin -company <company-uuid> -username admin
  ```

  It joins the builtin `Administrator` group, which already holds every permission from migration `20261127100000_identity_seed_endpoint_permissions.sql`.

## Endpoints

Every operation declares `x-permission` in `api/openapi.yaml`; `Login` and the health probe are the only `x-public: true` ones. The guard (`httpapi/handler_guard.go`) is a generated strict middleware, so it sees each `operationId` and runs before the handler: an endpoint added to the spec cannot be served unguarded, and an operation that declares no permission is denied rather than served open.

| Operation | Route | Permission |
|---|---|---|
| `Login` | `POST /api/v1/identity/login` | public, sets the session cookie |
| `Logout` | `POST /api/v1/identity/logout` | `identity.session.write` |
| `GetMe` | `GET /api/v1/identity/me` | `identity.user.read` |
| `ListUsers`, `GetUser` | `GET /api/v1/identity/users[/{userId}]` | `identity.user.read` |
| `CreateUser` | `POST /api/v1/identity/users` | `identity.user.write` |
| `UpdateUser` | `PATCH /api/v1/identity/users/{userId}` | `identity.user.write` |
| `ListAccessGroups`, `GetAccessGroup` | `GET /api/v1/identity/access-groups[/{accessGroupId}]` | `identity.group.read` |
| `CreateAccessGroup` | `POST /api/v1/identity/access-groups` | `identity.group.write` |
| `UpdateAccessGroup` | `PATCH /api/v1/identity/access-groups/{accessGroupId}` | `identity.group.write` |
| `ListPermissions` | `GET /api/v1/identity/permissions` | `identity.group.read` |
| `ListAuditEvents` | `GET /api/v1/identity/audit-events` | `identity.audit.read` |

List endpoints are keyset-paginated: the response carries an opaque base64url `nextCursor` and the request takes `limit` plus `cursor` (see `api/AGENTS.md`). Users and groups are ordered and paged by `username`/`name`; audit events by `id DESC`, because timestamps can tie inside one transaction and UUIDv7 ids are time-ordered.

**The last Administrator is protected.** An `UpdateUser` that would leave no *active* user in an *active* builtin group (deactivating the only Administrator, or removing the group from them) is refused with `LAST_ADMINISTRATOR` and rolled back, because nobody could repair the installation through the UI. The check locks the builtin group rows for the transaction, so two concurrent updates cannot each remove a different Administrator. When no Administrator exists yet (before `createadmin` has run) there is nothing to protect and updates are not blocked.

**Login timing.** An unknown username runs the same argon2id work as a known one (`domain.BurnPasswordCheck` against a throwaway hash made with the current parameters), so response time does not reveal which usernames exist. Together with the single `INVALID_CREDENTIALS` code, the endpoint does not enumerate users. `CurrentSession` separates an unknown, revoked, or expired token (`SESSION_EXPIRED`, 401) from an infrastructure failure (`INTERNAL_ERROR`, 500), so a database outage is not shown to every user as "logged out".

A user is **deactivated**, never deleted. Deactivating a user also revokes every live session, so an old cookie stops working immediately. The built-in `Administrator` group cannot be renamed, deactivated, or stripped of permissions (`BUILTIN_GROUP_IMMUTABLE`): renaming it silently would weaken the only account that can repair the installation.

## Version 1 behaviour

- One built-in group "Administrator" with all permissions (seeded). The Administrator creates further groups and users in the UI. Roles beyond Administrator are created by the Administrator later, so permissions must be listed in a registry each module contributes to.
- The permission registry is the list of all `x-permission` values in `api/openapi.yaml`, read from the embedded spec (`embedded-spec: true` in `oapi-codegen.yaml`) at startup. The same data drives `ListPermissions` and the guard, so the registry and the checks cannot drift.
- A group may only be granted a permission that exists in the registry (`UNKNOWN_PERMISSION`), so a typo cannot create a dead permission that no endpoint checks.
- The company is single in version 1: no per-company scoping of roles.

## Public interface (`module.go`)

`Authenticate(ctx, token)`, `CurrentUser(session)`, `HasPermission(session, permission)`, `RecordAudit(ctx, event)`. Another module posts its history with `RecordAudit`; when it already holds a transaction, it uses the service's `RecordAuditTx` so the change and its history commit together.

## Permissions

`identity.user.read|write`, `identity.group.read|write`, `identity.audit.read`.

## Error codes

| Code | Status | Meaning |
|---|---|---|
| `INVALID_CREDENTIALS` | 401 | Wrong username **or** wrong password; the same code for both, so the response cannot probe which usernames exist. |
| `SESSION_EXPIRED` | 401 | The cookie is missing, unknown, revoked, or past `expires_at`. |
| `PERMISSION_DENIED` | 403 | Signed in, but the user's groups lack the operation's `x-permission`. |
| `USER_INACTIVE` | 403 | A deactivated user tried to sign in. |
| `USER_NOT_FOUND` | 404 | No such user. |
| `ACCESS_GROUP_NOT_FOUND` | 404 | No such access group. |
| `USERNAME_TAKEN` | 409 | The username already exists (unique across the installation). |
| `ACCESS_GROUP_NAME_TAKEN` | 409 | The access group name already exists. |
| `BUILTIN_GROUP_IMMUTABLE` | 409 | An attempt to weaken the builtin Administrator group. |
| `LAST_ADMINISTRATOR` | 409 | An update would leave no active Administrator. |
| `UNKNOWN_PERMISSION` | 422 | A granted permission is not in the registry. |
| `VALIDATION_FAILED` | 422 | A field rule failed; `details` names the field. |

## Deferred

Per-company role assignment (a user's allowed companies), external identity providers, two-factor authentication.

## Known limitations (decided, not yet built)

Recorded after the Phase 1.2 review so they are decisions, not surprises:

- **No login rate limit or lockout.** Brute force is limited only by argon2id cost. Add throttling (per username and per address) in the hardening phase, or earlier if the installation is exposed to the internet.
- **Cookie `Secure` is always on.** Chrome accepts it on `http://localhost`, Safari does not; develop in Chrome (the local Caddy serves plain `http://localhost`). A config switch is added only if it becomes a real problem.
- **Read endpoints are not filtered by `company_id` yet.** `ListUsers`, `GetUser`, and `ListAuditEvents` return the whole installation, which is correct while there is one company. Add the `company_id` filter when the multi-company UI is built (hard rule 13).
- **The use-case layer reads the OpenAPI spec through `platform/http/gen`** to build the permission registry. Allowed by the lint rules (`platform` is shared), but if it becomes awkward, move spec reading into `platform` and pass the registry in.
- **Session expiry uses two clocks.** `expires_at` is computed from the injected clock, but the live-session filter uses the database's `now()`. Tests that fake the clock cannot expire a session; set `expires_at` directly in SQL for those tests.
