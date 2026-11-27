# Learn apotekerza, step by step

A hands-on tour of this repository: install the tools, run the quality gate, start the stack, look inside the database, and call the API. Every step says what to type, what you should see, and why it matters. Do the steps in order; each one builds on the previous one.

This guide is for **learning and operating** the project. The prompts that **build** the application (phase by phase, for the AI) are in `source/start-setup.md`. Rules for AI and for humans are in `AGENTS.md` and `docs/`.

Every command and every expected output below was run on macOS (Apple silicon) on 2026-11-27 against the state of `main` at that time. Lines starting with `$` are what you type. Lines without `$` are what you should see (shortened where it says `...`).

## Contents

1. [Install the tools](#1-install-the-tools)
2. [Get the code and look around](#2-get-the-code-and-look-around)
3. [Install dependencies](#3-install-dependencies)
4. [Run the quality gate](#4-run-the-quality-gate)
5. [Code generation (spec first)](#5-code-generation-spec-first)
6. [Run the tests](#6-run-the-tests)
7. [Start the dev stack](#7-start-the-dev-stack)
8. [Call the health API](#8-call-the-health-api)
9. [Look inside the database](#9-look-inside-the-database)
10. [Create the first Administrator and log in](#10-create-the-first-administrator-and-log-in)
11. [Staging and production environments](#11-staging-and-production-environments)
12. [Your daily workflow](#12-your-daily-workflow)
13. [Troubleshooting](#13-troubleshooting)
14. [Command cheat sheet](#14-command-cheat-sheet)

---

## 1. Install the tools

You need five things on your machine.

| Tool | Version | Why |
|---|---|---|
| Go | 1.26 or newer | The backend. `apps/backend/go.mod` says `go 1.26.0` |
| Node.js | a current LTS or newer (the web image uses Node 25) | The web app and the generated TypeScript client |
| pnpm | 10.34.9 (pinned in `package.json`) | Package manager for `apps/web` and `packages/api-client` |
| Docker Desktop | current | PostgreSQL, the dev stack, and the integration tests |
| GNU Make | any | Every project command is a `make` target |

Optional: **OpenCode**, the AI tool the repository is configured for (`opencode.json`, `.opencode/`).

Check what you have:

```sh
$ go version
go version go1.27.4 darwin/arm64
$ node -v
v25.9.0
$ corepack enable
$ pnpm -v
10.34.9
$ docker version --format '{{.Server.Version}}'
$ make --version | head -1
```

`corepack enable` makes the `pnpm` that is pinned in `package.json` available, so everyone uses the same pnpm version. If `pnpm -v` prints a different number, run `corepack prepare pnpm@10.34.9 --activate`.

Docker must be **running** (the whale icon is up) before you reach step 6.

## 2. Get the code and look around

```sh
$ git clone https://github.com/erwana32/apotekerza.git
$ cd apotekerza
$ ls
```

The map of the repository (full table in `README.md`):

```text
apotekerza/
├── AGENTS.md              rules for AI (and a good first read for you)
├── api/openapi.yaml       the API contract: the source of truth
├── apps/
│   ├── backend/           Go: cmd/ (server, migrate, createadmin), internal/<module>/, migrations/
│   └── web/               React + TypeScript (Vite, Tailwind)
├── packages/api-client/   generated TypeScript client (never edited by hand)
├── deploy/                Docker Compose, Caddyfiles, Dockerfiles, .env templates
├── docs/                  architecture, ADRs, module docs, conventions, glossary
├── scripts/               checker scripts (naming rules)
└── source/                logo files, start-setup.md (build prompts), learn.md (this guide)
```

Read these four files first, in this order. They take about twenty minutes and answer most "why is it like this?" questions:

1. `AGENTS.md`: the hard rules (module boundaries, money as decimals, one schema, documents lifecycle).
2. `docs/architecture.md`: the modules and how they depend on each other.
3. `docs/adr/0001-modular-monolith.md`: why one database, one schema, and tables prefixed by module.
4. `docs/conventions/naming.md`: how files, tables, routes, and permissions are named.

The backend module layout you will see everywhere:

```text
apps/backend/internal/<module>/
├── module.go     the only file other modules may import (public interface)
├── domain/       model_*.go: entities and pure rules (no database, no HTTP)
├── app/          service_*.go: one use case per file, owns the transaction
├── httpapi/      handler_*.go, mapper_*.go: thin HTTP layer
└── store/        repo_*.go + queries/query_*.sql: database access (sqlc)
```

The file prefix tells you the role of a file before you open it. `make check-naming` enforces this.

## 3. Install dependencies

```sh
$ pnpm install
$ cd apps/backend && go mod download && cd ../..
```

The Go tools (`sqlc`, `oapi-codegen`, `goose`, `golangci-lint`) are declared in `apps/backend/go.mod` and run as `go tool <name>`, so there is nothing else to install globally. The first run of each tool compiles it, which takes a minute.

## 4. Run the quality gate

```sh
$ make help
```

lists every command. The one to know first:

```sh
$ make check
```

This is the gate every change must pass before merging. It runs four checks in order:

| Part | What it checks |
|---|---|
| `check-naming` | File names follow the prefix rules (`model_`, `service_`, `handler_`, ...) |
| `gen-check` | Generated code is in sync with `api/openapi.yaml` and the SQL queries |
| `lint` | `go vet`, `golangci-lint` (including the module-boundary rules), and ESLint for the web |
| `test` | Go unit tests and the web unit tests (Vitest) |

Expected: it ends without `Error` and the exit code is `0` (`echo $?`).

You can run any part alone: `make check-naming`, `make lint`, `make test`.

> `gen-check` compares generated files with what is committed in git. On a clean checkout it passes. If you have uncommitted changes to generated files it reports "out of sync" until you commit them. That is expected, not a bug.

Try it: break a rule on purpose and watch the gate catch it.

```sh
$ touch apps/backend/internal/identity/domain/user.go
$ make check-naming
NAMING: apps/backend/internal/identity/domain/user.go: files in domain/ must start with model_
Naming check failed. See docs/conventions/naming.md
```

Delete the file again: `rm apps/backend/internal/identity/domain/user.go`.

## 5. Code generation (spec first)

This project is **spec first** (ADR-0002): the API is defined in `api/openapi.yaml` and the code is generated from it.

```sh
$ make gen
```

runs three generators:

| Generator | Input | Output |
|---|---|---|
| `oapi-codegen` | `api/openapi.yaml` | Go server interface in `apps/backend/internal/platform/http/gen/` |
| `sqlc` | `apps/backend/internal/<module>/store/queries/*.sql` | typed Go code in `store/gen/` |
| `openapi-typescript` | `api/openapi.yaml` | `packages/api-client/src/generated/api.d.ts` |

Never edit anything in a `gen/` folder or in `generated/`. Change the spec or the SQL, then run `make gen`.

Look at one operation in `api/openapi.yaml`:

```yaml
/api/v1/health:
  get:
    operationId: GetHealth
    x-permission: platform.health.read
    x-public: true
```

- `operationId` becomes the Go method name (`GetHealth`).
- `x-permission` is the permission a user needs. A guard reads it from the spec for **every** operation, so an endpoint cannot be added without a permission.
- `x-public: true` marks the few endpoints that need no login (health and login).

## 6. Run the tests

Three levels, from fast to slow:

```sh
$ make test               # unit tests: Go + web, no Docker needed
$ make test-integration   # Go tests against a real PostgreSQL (needs Docker running)
$ pnpm --filter ./apps/web exec playwright install chromium   # one time
$ make test-e2e           # browser smoke tests
```

How the integration tests work: `internal/platform/testutil` starts **one** PostgreSQL container for each test package, migrates a template database once, and gives every test its own copy of it. Tests never share data. They run with `-p 1` (one package at a time) so containers do not race. The first run downloads `postgres:17-alpine`.

Expected: a list of `ok  github.com/erwana32/apotekerza/...` lines and no `FAIL`.

Test files sit next to the file they test: `service_authenticate.go` and `service_authenticate_test.go`; integration tests end in `_integration_test.go` and start with `//go:build integration`.

## 7. Start the dev stack

The stack has four services (`deploy/docker-compose.yml`):

```text
postgres  ──healthy──▶  migrate (runs once, exits)  ──succeeded──▶  backend  ──healthy──▶  caddy
```

- `postgres`: PostgreSQL 17, database `db_apotekerza_dev`, user `apotekerza`.
- `migrate`: applies the goose migrations, then exits.
- `backend`: the Go server on port 8080 inside Docker (not published to your machine).
- `caddy`: serves the web build and forwards `/api/*` to the backend. It publishes ports **80** and **443**.

Before you start, make sure ports 80 and 443 are free:

```sh
$ lsof -nP -iTCP:80 -sTCP:LISTEN
```

No output means free. Then:

```sh
$ make dev-up
...
 Container deploy-postgres-1 Healthy
 Container deploy-migrate-1 Started
 Container deploy-backend-1 Healthy
 Container deploy-caddy-1 Started
```

The first build takes a few minutes (Go modules and pnpm install). Later builds use the cache. No `.env` file is needed: the dev defaults are in the compose file.

Check that everything is up:

```sh
$ make dc CMD="ps"
NAME                STATUS
deploy-backend-1    Up ... (healthy)
deploy-caddy-1      Up ...
deploy-postgres-1   Up ... (healthy)
```

`migrate` is not listed because it already finished. Show it with `make dc CMD="ps -a"`, and read what it did:

```sh
$ make dc CMD="logs migrate"
migrate-1  | ... OK   20261126120000_identity_create_tables.sql
migrate-1  | ... OK   20261126121000_identity_seed_administrator.sql
migrate-1  | ... OK   20261127100000_identity_seed_endpoint_permissions.sql
migrate-1  | ... goose: successfully migrated database to version: 20261127100000
```

### `make dc`: one entry point for docker compose

```sh
$ make dc CMD="logs -f backend"             # dev (default); Ctrl+C stops following
$ make dc ENV=stg CMD="ps"                  # staging
$ make dc ENV=prod CMD="ps"                 # production, on the server
```

`ENV` is `dev` (default), `stg`, or `prod`. It picks the right Compose project name, env file, and compose files, so you never type `-p`, `--env-file`, or `-f` yourself.

To stop the stack (your data is kept in a Docker volume):

```sh
$ make dev-down
```

> `make dc CMD="down -v"` also **deletes the database volume**. Use it only when you really want an empty database.

## 8. Call the health API

There are two health endpoints. They are different on purpose.

```sh
$ curl -i http://localhost/api/v1/health
HTTP/1.1 200 OK
Content-Type: application/json
Via: 1.1 Caddy
X-Request-Id: ...

{"status":"ok"}
```

`/api/v1/health` is part of the API contract (`api/openapi.yaml`). It is public, goes through Caddy, and every response carries an `X-Request-Id` that also appears in the backend logs.

```sh
$ make dc CMD='exec backend wget -qO- http://localhost:8080/healthz'
ok
```

`/healthz` is an operational probe outside the contract, used by Docker's own healthcheck. Caddy sends every path that is not `/api/*` to the web app, so `http://localhost/healthz` would return the web page, not `ok`. That is why this one is called from inside the container.

Open `http://localhost/` in a browser to see the web app (currently the Phase 0 skeleton).

## 9. Look inside the database

Open an interactive SQL session in the running Postgres container:

```sh
$ make dc CMD='exec postgres psql -U apotekerza -d db_apotekerza_dev'
db_apotekerza_dev=# \dt
```

(`\q` leaves.) Or run one statement at a time without the prompt:

```sh
$ make dc CMD='exec postgres psql -U apotekerza -d db_apotekerza_dev -c "\dt"'
 Schema |            Name            | Type  |   Owner
--------+----------------------------+-------+------------
 public | goose_db_version           | table | apotekerza
 public | identity_access_groups     | table | apotekerza
 public | identity_audit_log         | table | apotekerza
 public | identity_group_permissions | table | apotekerza
 public | identity_sessions          | table | apotekerza
 public | identity_user_groups       | table | apotekerza
 public | identity_users             | table | apotekerza
```

Things to notice (they come straight from `docs/conventions/naming.md`):

- **One schema, `public`.** The owner of a table is visible from its prefix: `identity_...`. Other modules will add `inventory_...`, `finance_...`, and so on in the same schema.
- `goose_db_version` is the migration tool's own table, the only one without a module prefix.

See the built-in group that a migration seeded:

```sh
$ make dc CMD='exec postgres psql -U apotekerza -d db_apotekerza_dev -c "select id, name, is_builtin from identity_access_groups"'
                  id                  |     name      | is_builtin
--------------------------------------+---------------+------------
 0193b7a0-0000-7000-8000-000000000001 | Administrator | t
```

and the migrations that were applied:

```sh
$ make dc CMD='exec postgres psql -U apotekerza -d db_apotekerza_dev -c "select version_id, is_applied from goose_db_version order by id"'
   version_id   | is_applied
----------------+------------
              0 | t
 20261126120000 | t
 20261126121000 | t
 20261127100000 | t
```

Why there is **no user yet**: the first Administrator needs a real password, so it is created by a command, never seeded with a default password. That is the next step.

Try the append-only rule (the audit log refuses changes, enforced by a database trigger). Run this only after step 10 has produced an audit row:

```sh
$ make dc CMD='exec postgres psql -U apotekerza -d db_apotekerza_dev -c "delete from identity_audit_log"'
ERROR:  identity_audit_log is append-only: DELETE is not allowed
```

## 10. Create the first Administrator and log in

`cmd/createadmin` creates the first user and puts it in the built-in `Administrator` group. The password is read from an environment variable so it never appears in your shell history or the process list.

```sh
$ make dc CMD="exec -e CREATEADMIN_PASSWORD='choose-a-long-password' backend /app/createadmin -company $(uuidgen | tr A-Z a-z) -username admin"
created administrator "admin" in group "Administrator"
```

- `-company` is a UUID. The `masterdata` module does not exist yet (Phase 1.4), so any UUID works for now. When `masterdata` arrives, it creates the default company and adds the database link; until then use one fixed UUID per database.
- Usernames are lowercase (`a-z`, `0-9`, `.`, `_`, `-`, 3 to 64 characters) and unique for the whole installation.
- Passwords must be at least 8 characters.

Log in. The server answers with a **session cookie**; `curl -c` saves it in a file and `-b` sends it back:

```sh
$ curl -s -i -c cookies.txt -H 'Content-Type: application/json' \
    -d '{"username":"admin","password":"choose-a-long-password"}' \
    http://localhost/api/v1/identity/login
HTTP/1.1 200 OK
Set-Cookie: apotekerza_session=...; Path=/; Expires=...; HttpOnly; Secure; SameSite=Lax

{"permissions":["identity.audit.read","identity.group.read", ...],"user":{...}}
```

The cookie is `HttpOnly` (JavaScript cannot read it) and `Secure`. curl and Chrome send a `Secure` cookie to `http://localhost`; **Safari does not**, so use Chrome for browser tests on localhost.

Now use the session:

```sh
$ curl -s -b cookies.txt http://localhost/api/v1/identity/me
{"permissions":[...],"user":{"displayName":"Administrator","username":"admin",...}}

$ curl -s -b cookies.txt http://localhost/api/v1/identity/permissions
{"items":["identity.audit.read","identity.group.read","identity.group.write","identity.session.write","identity.user.read","identity.user.write","platform.health.read"]}
```

The permission list is not typed by hand: it is read from every `x-permission` in `api/openapi.yaml`.

Create an access group and a user in it:

```sh
$ curl -s -b cookies.txt -H 'Content-Type: application/json' \
    -d '{"name":"Kasir","permissions":["identity.user.read"]}' \
    http://localhost/api/v1/identity/access-groups
{"id":"<group-id>","isBuiltin":false,"name":"Kasir","permissions":["identity.user.read"],...}

$ curl -s -b cookies.txt -H 'Content-Type: application/json' \
    -d '{"username":"kasir.satu","displayName":"Kasir Satu","password":"another-password","groupIds":["<group-id>"]}' \
    http://localhost/api/v1/identity/users
```

(Use the `id` from the first response in place of `<group-id>`.)

Four behaviours worth seeing with your own eyes:

```sh
# 1. Pagination is keyset-based: ask for one item, get a cursor for the next page.
$ curl -s -b cookies.txt 'http://localhost/api/v1/identity/users?limit=1'
{"items":[{"username":"admin",...}],"nextCursor":"..."}

# 2. The last Administrator cannot be deactivated (the id comes from the "me" response).
$ curl -s -X PATCH -b cookies.txt -H 'Content-Type: application/json' \
    -d '{"displayName":"Administrator","isActive":false}' \
    http://localhost/api/v1/identity/users/<your-user-id>
{"code":"LAST_ADMINISTRATOR","message":"at least one active Administrator must remain"}

# 3. Permissions are enforced: log in as kasir.satu, then ask for the access groups.
$ curl -s -c kasir.txt -H 'Content-Type: application/json' \
    -d '{"username":"kasir.satu","password":"another-password"}' http://localhost/api/v1/identity/login
$ curl -s -b kasir.txt http://localhost/api/v1/identity/access-groups
{"code":"PERMISSION_DENIED","message":"permission denied"}

# 4. Logout revokes the session at once.
$ curl -s -o /dev/null -w '%{http_code}\n' -X POST -b cookies.txt -c cookies.txt http://localhost/api/v1/identity/logout
204
$ curl -s -b cookies.txt http://localhost/api/v1/identity/me
{"code":"SESSION_EXPIRED","message":"session is missing or expired"}
```

Finally, read who did what. Every create, update, login, and logout is recorded:

```sh
$ curl -s -b cookies.txt 'http://localhost/api/v1/identity/audit-events?limit=5'
```

> The audit log is append-only, so these test users and their history stay in your dev database. If you want a clean slate, `make dev-down` and then remove the volume (`make dc CMD="down -v"`).

Delete `cookies.txt` and `kasir.txt` when you finish; they hold live session tokens.

## 11. Staging and production environments

One compose file, three environments. Each environment has its own Compose **project name**, so its containers and its database volume are separate.

| Environment | Command | Compose project | Env file | Compose files | Database |
|---|---|---|---|---|---|
| dev (your machine) | `make dev-up` | `deploy` | none (defaults) | base | `db_apotekerza_dev` |
| staging (your machine) | `make stg-up` | `apotekerza-stg` | `deploy/.env.stg` | base | `db_apotekerza_stg` |
| production (the server) | `make prod-up` | `apotekerza-prod` | `deploy/.env` on the server | base + prod | `db_apotekerza_prod` |

Set up **staging** on your machine:

```sh
$ cp deploy/.env.stg.example deploy/.env.stg
$ openssl rand -hex 32          # paste the output as POSTGRES_PASSWORD in deploy/.env.stg
$ make stg-up
$ curl -i http://localhost/api/v1/health
$ make stg-down                 # when you are done
```

Staging also listens on ports 80 and 443, so stop the dev stack (`make dev-down`) before `make stg-up`.

Set up **production** on the server (never on a laptop):

```sh
$ cp deploy/.env.prod.example deploy/.env
$ nano deploy/.env              # set APP_DOMAIN (DNS must already point here) and the passwords
$ make prod-up
```

The production override refuses to start if `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, or `APP_DOMAIN` is missing. Caddy then obtains a TLS certificate for `APP_DOMAIN` automatically, so ports 80 and 443 must be open.

Rules that keep you out of trouble:

- **Real `.env*` files are never committed.** `.gitignore` ignores them, and `.dockerignore` keeps them out of image builds. Only the `*.example` templates are in git.
- **A different password for every environment.** Use letters and digits only (`openssl rand -hex 32`): the password is placed in a connection URL, where characters like `/ + = :` would break it.
- **`POSTGRES_DB` is read only when the Postgres volume is first created.** Changing it later does not rename an existing database. Rename with `ALTER DATABASE`, or start from a new volume.
- **Never copy production data into staging or dev** without anonymising it: it contains patient data.

Create the first Administrator in an environment by adding `ENV=` to the same command:

```sh
$ make dc ENV=stg CMD="exec -e CREATEADMIN_PASSWORD='...' backend /app/createadmin -company <uuid> -username admin"
```

## 12. Your daily workflow

```text
git checkout -b slice/<short-description>   one branch per vertical slice
  ... change code and docs ...
make check                                   must pass
git commit                                   Conventional Commits, in English
git push, open a pull request                you push; the AI never pushes
```

Commit messages look like `feat(inventory): add FEFO allocation` or `fix(identity): protect last administrator`.

**Add an endpoint** (the order matters, because the spec comes first):

1. Edit `api/openapi.yaml`: path, `operationId`, schemas, and `x-permission`.
2. `make gen`.
3. Implement the new method of the generated interface: handler (thin), then the use case in `app/`.
4. If it introduces a new permission, add a migration that grants it to the `Administrator` group (permissions are data, not synced at startup).
5. Write the tests, then `make check`.

**Change the database** with an add-only goose migration: `apps/backend/migrations/<timestamp>_<module>_<description>.sql`. Never edit a migration that already exists. Names of tables, columns, constraints, and indexes follow `docs/conventions/naming.md` (section 2).

**Work with the AI.** The repository is set up for OpenCode:

- `AGENTS.md` is the canonical rule file; `docs/` holds the decisions.
- `.opencode/skills/` has procedures for repeatable tasks (`add-module`, `add-endpoint`, `add-migration`, `post-journal`, `stock-movement`, `write-adr`, `review-change`).
- **Teaching mode:** the AI explains its plan in small steps and waits for your OK before it writes code or tables.
- `source/start-setup.md` holds the phased prompts, from Phase 0 to Phase 8, with a progress table at the top.
- The AI never runs `git push` and asks before committing.

When a decision changes, update the ADR in `docs/adr/` (they are living documents with a change log), not just the code.

## 13. Troubleshooting

| Symptom | Cause and fix |
|---|---|
| `bind: address already in use` on `make dev-up` | Port 80 or 443 is taken (often by another stack). Find it with `lsof -nP -iTCP:80 -sTCP:LISTEN` and stop it. |
| `migrate` exits with `database "db_apotekerza_dev" does not exist` | The Postgres volume was created earlier under another database name. Rename it: `make dc CMD='exec postgres psql -U apotekerza -d postgres -c "ALTER DATABASE <old> RENAME TO db_apotekerza_dev"'` (only if the old database is empty or you are sure), or start from a new volume. |
| `backend` is not `healthy` | `make dc CMD="logs backend"`. A wrong `DATABASE_URL` or a failed migration shows up there. |
| `make dc` says `Missing deploy/.env.stg` | Copy `deploy/.env.stg.example` to `deploy/.env.stg` and fill in the password. |
| `Generated code is out of sync` from `make check` | Run `make gen` and commit the result. |
| `make test-integration` cannot start a container | Docker Desktop is not running. |
| `Error: ... unexpected container status "created"` | Run the tests through `make test-integration`; it passes `-p 1` so packages do not race for containers. |
| `make check-naming` rejects a file | Add the role prefix (`model_`, `service_`, `handler_`, `mapper_`, `repo_`, `query_`, `job_`) or fix the test suffix (`_test.go`). The message names the rule. |
| `lint` reports an import from another module's `domain/`, `app/`, `httpapi/`, or `store/` | Modules may import each other only through `module.go` (ADR-0001). Call the public interface instead. |
| Login works in Chrome but not in Safari on `localhost` | Safari does not send `Secure` cookies over plain `http`. Use Chrome. |
| `SESSION_EXPIRED` right after login with curl | You forgot `-b cookies.txt` (send) or `-c cookies.txt` (save) on the matching command. |
| The first build is very slow | Go modules and pnpm packages are downloaded once; later builds use the Docker cache. |

## 14. Command cheat sheet

```sh
# Quality
make check                 # the gate: naming + generated code + lint + unit tests
make check-naming          # file naming rules only
make lint                  # go vet, golangci-lint, ESLint
make test                  # unit tests (Go + web)
make test-integration      # Go tests on a real PostgreSQL (Docker)
make test-e2e              # Playwright smoke tests
make gen                   # regenerate code from the spec and SQL
make gen-check             # fail if generated code is out of sync with git

# Docker, development
make dev-up                # build and start the dev stack
make dev-down              # stop it (data kept)
make dc CMD="ps"           # any docker compose command; ENV=dev|stg|prod
make dc CMD="logs -f backend"
make dc CMD='exec postgres psql -U apotekerza -d db_apotekerza_dev'

# Docker, other environments
make stg-up / make stg-down        # staging on this machine (deploy/.env.stg)
make prod-up / make prod-down      # production, on the server (deploy/.env)

# API
curl -i http://localhost/api/v1/health
curl -s -c cookies.txt -H 'Content-Type: application/json' -d '{"username":"...","password":"..."}' http://localhost/api/v1/identity/login
curl -s -b cookies.txt http://localhost/api/v1/identity/me
```

## Where to go next

- Build the next phase: open `source/start-setup.md`, read the progress table, and copy the next prompt (Phase 1.3: the web app shell and identity screens).
- Understand a decision: `docs/adr/`.
- Learn a module: `docs/modules/<module>.md`.
- Look up an Indonesian pharmacy term: `docs/glossary.md`.
