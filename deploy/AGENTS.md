# Deploy (`deploy/`)

One VPS, Docker Compose. General rules are in the root `AGENTS.md`.

## Contents

- `docker-compose.yml` (base, local dev): `postgres`, `migrate` (runs once before the backend), `backend`, and `caddy` (built from `Dockerfile.web`, serves the static web build). `make dev-up` / `make dev-down` use it. Its dev-only database password fallback must never be used in production.
- `docker-compose.prod.yml`: production override. It makes `POSTGRES_*` and `APP_DOMAIN` mandatory (the stack refuses to start without them). Use it together with the base file.
- `Caddyfile` (production, automatic TLS for `APP_DOMAIN`) and `Caddyfile.local` (plain HTTP on localhost for dev): both reverse-proxy `/api/*` to `backend:8080` and serve static files with SPA fallback. `CADDYFILE` in `.env` selects which one is mounted.
- `Dockerfile.backend` (builds `server` and `migrate`) and `Dockerfile.web`.
- `.env.example`: variable names **without secret values**.

## Rules

- **Secrets never enter git.** Real values live only in `.env` on the server. `.env.example` holds placeholders. Agents do not read `.env`.
- User passwords are hashed with argon2id. Session cookies are `HttpOnly`, `Secure`, `SameSite`.
- The database volume is named and never removed automatically. `docker compose down -v` is denied for agents.
- Environments: local dev and production. Staging is added before a real go-live.
- **Backups:** daily encrypted `pg_dump` to storage outside the server, retention 30 daily + 12 monthly, restore tested monthly. The procedure is written in `docs/runbooks/`. Prescription data retention is not decided yet: confirm with the APJ.
- Application logs are JSON to stdout with a request ID; Docker log rotation is enabled.
- Migrations run as a separate step before the new `backend` starts.
