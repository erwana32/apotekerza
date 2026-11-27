# Deploy (`deploy/`)

One VPS, Docker Compose. General rules are in the root `AGENTS.md`.

## Contents

- `docker-compose.yml` (base, local dev): `postgres`, `migrate` (runs once before the backend), `backend`, and `caddy` (built from `Dockerfile.web`, serves the static web build). `make dev-up` / `make dev-down` use it. Its dev-only database password fallback must never be used in production.
- `docker-compose.prod.yml`: production override. It makes `POSTGRES_*` and `APP_DOMAIN` mandatory (the stack refuses to start without them). Use it together with the base file.
- `Caddyfile` (production, automatic TLS for `APP_DOMAIN`) and `Caddyfile.local` (plain HTTP on localhost for dev): both reverse-proxy `/api/*` to `backend:8080` and serve static files with SPA fallback. `CADDYFILE` in `.env` selects which one is mounted.
- `Dockerfile.backend` (builds `server`, `migrate`, and `createadmin`; the first Administrator is created inside the running stack, see `docs/modules/identity.md`, "Seeding") and `Dockerfile.web`.
- `.dockerignore` (repo root, because the build context is the repo root): keeps `.git`, every `.env*` file, `node_modules`, `dist`, docs, and Markdown out of the build context. A real env file must never reach a build, so do not remove the `**/.env*` lines.
- `.env.example`: every variable the stack reads, **without secret values**, and a table of which template to copy for which environment. `.env.stg.example` and `.env.prod.example` are the per-environment templates (`.gitignore` allows `.env.example` and `.env.*.example`, and ignores every real `.env*`).

## Rules

- **Secrets never enter git.** Real values live only in `.env` on the server. `.env.example` holds placeholders. Agents do not read `.env`.
- User passwords are hashed with argon2id. Session cookies are `HttpOnly`, `Secure`, `SameSite`.
- The database volume is named and never removed automatically. `docker compose down -v` is denied for agents.
- **Environments** (each has its own Compose project name, so containers and the Postgres volume are separate; a shared name would share the volume):

  | Environment | Command | Compose project | `.env` file | Compose files | Database |
  |---|---|---|---|---|---|
  | local dev | `make dev-up` | `deploy` | none (defaults) | base | `db_apotekerza_dev` |
  | staging (laptop) | `make stg-up` | `apotekerza-stg` | `deploy/.env.stg` | base | `db_apotekerza_stg` |
  | production (server) | `make prod-up` | `apotekerza-prod` | `deploy/.env` on the server | base + prod | `db_apotekerza_prod` |

  `make dc` runs any `docker compose` command for an environment with the right project name, env file, and compose files: `make dc CMD="ps"` (dev), `make dc ENV=stg CMD="logs -f backend"`, `make dc ENV=prod CMD="exec backend sh"`. `stg-up`, `stg-down`, `prod-up`, and `prod-down` are shortcuts over it.

  `stg-down` and `prod-down` stop the stack and keep the data volume. Staging on a server is not decided yet. Passwords differ per environment (generate with `openssl rand -hex 32`: letters and digits only, because the password sits in a connection URL). Never copy production data into staging or local without anonymising it (patient data, hard rule 10).
- **Backups:** daily encrypted `pg_dump` to storage outside the server, retention 30 daily + 12 monthly, restore tested monthly. The procedure is written in `docs/runbooks/`. Prescription data retention is not decided yet: confirm with the APJ.
- Application logs are JSON to stdout with a request ID; Docker log rotation is enabled.
- Migrations run as a separate step before the new `backend` starts.
