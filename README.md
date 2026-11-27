# apotekerza

A staged ERP for an Indonesian pharmacy (apotek). Version 1 serves one pharmacy: Finance, Inventory, Purchasing, POS, and the pharmacy service (prescriptions).

**Status:** Phase 0 (repository bootstrap) is done: the Go backend skeleton, the React web skeleton, the OpenAPI contract with generated clients, Docker Compose with Caddy, and the quality gates are in place. Phase 1 (foundation) is in progress: the shared platform packages and the identity backend (login, sessions, users, access groups, audit trail, permission guard) are done; the web screens and `masterdata` come next. No other business module exists yet. Start with `AGENTS.md` and `docs/architecture.md`; build progress is tracked in `source/start-setup.md`.

## Prerequisites

- Go 1.26+
- Node.js and pnpm (`corepack enable`)
- Docker (PostgreSQL and integration tests)
- GNU Make
- OpenCode (the AI tool this project is set up for)

## Commands

`make help` lists all of them. The gate before merging is `make check`.

## Map

| Path | Contents |
|---|---|
| `AGENTS.md` | Rules for AI (root), plus nested `AGENTS.md` per folder |
| `opencode.json`, `.opencode/` | OpenCode configuration: permissions, `plan` and `reviewer` agents (`.opencode/agents/`), and the skills for repeatable procedures (`.opencode/skills/`) |
| `docs/` | Architecture, modules, ADRs, conventions, glossary. Documentation rules in `docs/README.md` |
| `api/openapi.yaml` | API contract, the source of truth for the generated Go server and TypeScript client |
| `apps/backend` | Go backend (`cmd/server`, `cmd/migrate`, `internal/platform/*`, `migrations/`) |
| `apps/web` | React + TypeScript web app (Vite, Tailwind, shadcn/ui, Vitest, Playwright) |
| `packages/api-client` | Generated TypeScript client (never edited by hand) |
| `deploy/` | Docker Compose (base + production override), `Caddyfile`, `Caddyfile.local`, Dockerfiles, `.env.example` |
| `scripts/` | Checker scripts |
| `source/` | Brand source files: company logo (`logo.svg`, `logo-mark.png`), see `docs/branding.md`. Also `source/start-setup.md` (phased copy-paste prompts for building the app step by step) and `source/learn.md` (a hands-on tour: install, run, test, Docker, call the API) |

## Language

Docs, rules, code, and commits are in English. Everything the end user sees or prints is in Indonesian, and Indonesian pharmacy terms are kept where English has no clean equivalent (see `docs/glossary.md`).
