---
name: add-module
description: Create a new backend module that respects the module boundaries. Use only after the developer approved the module and it is listed in docs/architecture.md.
---

# add-module

Adding a module is an architecture change. Do not start without the developer's approval.

## Before you start

1. Confirm the module is in the table in `docs/architecture.md`. If not, stop and ask. Add it there first, with an ADR (skill `write-adr`) when it changes boundaries.
2. Read `docs/modules/<module>.md` (create it from the existing ones if it does not exist), `docs/conventions/naming.md`, and `apps/backend/AGENTS.md`.
3. Decide the dependencies. They must not create a cycle (see the table in `docs/architecture.md`).

## Step 0: propose the module shape and wait for approval

Follow teaching mode in `AGENTS.md`. Present the module's responsibilities, its public interface (`module.go` functions with inputs and outputs), what it depends on, what depends on it, and the tables it will own at a high level. Wait for approval before creating folders.

## Steps

1. Create the folders:
   ```
   apps/backend/internal/<module>/{domain,app,httpapi,store/queries}
   ```
2. Create the root files: `module.go` (public interface: only the functions other modules need), `errors.go` (domain errors).
3. Create a migration that starts the module's schema (skill `add-migration`): `CREATE SCHEMA <module>;`.
4. Register the module in the application wiring (dependency injection in `internal/platform`). Wire only through the public interfaces of other modules.
5. Add the module's permissions to the permission registry (every `x-permission` in `api/openapi.yaml`).
6. Create `docs/modules/<module>.md` (responsibilities, depends on, permissions, error codes, deferred) and set its status in `docs/architecture.md`.
7. Run `make check-naming` and `make check`.

## Rules

- File prefixes and folders follow `docs/conventions/naming.md` (`model_` in `domain/`, `service_` in `app/`, `handler_`/`mapper_` in `httpapi/`, `repo_` in `store/`).
- Never read another module's tables. Use its `module.go` or a `v_` view it publishes.
- Cross-module calls happen inside the caller's transaction.
