# Deploy (`deploy/`)

One VPS, Docker Compose. General rules are in the root `AGENTS.md`.

## Planned contents

- `docker-compose.yml`: `caddy`, `backend`, `postgres`. The static web build (`pnpm build`) is served by Caddy.
- `Caddyfile`: automatic TLS, reverse proxy `/api/*` to `backend`, everything else static files with SPA fallback.
- `.env.example`: variable names **without secret values**.

## Rules

- **Secrets never enter git.** Real values live only in `.env` on the server. `.env.example` holds placeholders. Agents do not read `.env`.
- User passwords are hashed with argon2id. Session cookies are `HttpOnly`, `Secure`, `SameSite`.
- The database volume is named and never removed automatically. `docker compose down -v` is denied for agents.
- Environments: local dev and production. Staging is added before a real go-live.
- **Backups:** daily encrypted `pg_dump` to storage outside the server, retention 30 daily + 12 monthly, restore tested monthly. The procedure is written in `docs/runbooks/`. Prescription data retention is not decided yet: confirm with the APJ.
- Application logs are JSON to stdout with a request ID; Docker log rotation is enabled.
- Migrations run as a separate step before the new `backend` starts.
