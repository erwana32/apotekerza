# apotekerza

A staged ERP for an Indonesian pharmacy (apotek). Version 1 serves one pharmacy: Finance, Inventory, Purchasing, POS, and the pharmacy service (prescriptions).

**Status:** skeleton of docs and rules. There is no application code yet. Start with `AGENTS.md` and `docs/architecture.md`.

## Prerequisites

- Go 1.25+
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
| `api/openapi.yaml` | API contract (created with the first endpoint) |
| `apps/backend`, `apps/web`, `packages/api-client` | Application code (created incrementally) |
| `deploy/` | Docker Compose, Caddyfile, `.env.example` |
| `scripts/` | Checker scripts |
| `source/` | Brand source files: company logo (`logo.svg`, `logo-mark.png`), see `docs/branding.md`. Also `source/start-setup.md`: phased copy-paste prompts for building the app step by step |

## Language

Docs, rules, code, and commits are in English. Everything the end user sees or prints is in Indonesian, and Indonesian pharmacy terms are kept where English has no clean equivalent (see `docs/glossary.md`).
