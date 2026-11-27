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
3. Decide the module's table prefix (`<module>_`). There is no per-module schema (ADR-0001); the first migration creates the module's first tables (skill `add-migration`).
4. Add the module to `apps/backend/.golangci.yml`: its own depguard module rule (the imports allowed by the dependency table in `docs/architecture.md`) and its name in the four `layer-*` rules. depguard matches package paths by prefix and does not expand globs, so every module is listed explicitly; a module missing from the file is not checked.
5. Register the module in the application wiring (dependency injection in `internal/platform`). Wire only through the public interfaces of other modules.
6. Add the module's permissions to the permission registry (every `x-permission` in `api/openapi.yaml`).
7. Create `docs/modules/<module>.md` (responsibilities, depends on, permissions, error codes, deferred) and set its status in `docs/architecture.md`.
8. Run `make check-naming` and `make check`.

## Rules

- File prefixes and folders follow `docs/conventions/naming.md` (`model_` in `domain/`, `service_` in `app/`, `handler_`/`mapper_` in `httpapi/`, `repo_` in `store/`).
- In Go, call another module only through its `module.go`. In SQL, other modules' tables may be read and joined (ADR-0001).
- Cross-module calls happen inside the caller's transaction.
