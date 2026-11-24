# 0002. OpenAPI is written first and generates both sides

- **Status:** Accepted
- **Date:** 2026-11-24

## Context

The backend (Go) and frontend (TypeScript) must agree on the contract, and AI tools need one place to read and change it.

## Decision

`api/openapi.yaml` is the source of truth. The Go server is generated with `oapi-codegen` (strict server) and the TypeScript client with `openapi-typescript` + `openapi-fetch` into `packages/api-client`. CI fails when generated code is out of sync (`make gen-check`). Money and quantities travel as decimal strings. Every operation declares `x-permission`.

## Consequences

- An agent changes the contract first, then the implementation.
- Generated code is never edited by hand (denied in `opencode.json`).
- A little ceremony per endpoint in exchange for a contract that cannot drift.

## Alternatives considered

- **Code first, derive the spec:** the contract would drift and be harder for AI to reason about.
- **Hand-written client:** duplicates types and invites mismatch.
