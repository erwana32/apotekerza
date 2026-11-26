# 0009. Phase 0 technology choices

- **Status:** Accepted
- **Date:** 2026-11-26

## Context

Phase 0 built an empty but fully wired skeleton in which `make check` passes and the Docker Compose stack runs locally. Before any code was written, the developer had to choose the HTTP router, logger, decimal library, JavaScript test and lint tools, and Go module path. Go 1.25 was pinned from the start. ADR-0002 and ADR-0003 already decide the API generators and the database stack, so this ADR records only the remaining choices. Every tool added later (oapi-codegen, testcontainers, golangci-lint) was pinned to a version compatible with Go 1.25.

## Decision

We use the standard library `net/http` for HTTP routing, relying on Go 1.22+ method and wildcard patterns. We use `log/slog` with a `JSONHandler` writing to stdout, with an `X-Request-ID` middleware that reuses an inbound ID or generates a random 128-bit one and attaches a request-scoped logger to the context.

We use `github.com/shopspring/decimal` for money and item quantities; the library is introduced with the shared `platform/money` package in Phase 1.1. Money stays `numeric(18,2)` in PostgreSQL and decimal strings in the API, and quantities `numeric(18,4)`.

For the web client we use ESLint with `typescript-eslint`, TypeScript strict mode, Vitest for logic and component tests, Playwright for critical end-to-end paths, and pnpm 10 for the workspace. The Go module path is `github.com/erwana32/apotekerza`, matching the GitHub repository. Code generation and database tooling remain governed by ADR-0002 and ADR-0003.

## Consequences

It becomes easier to keep the dependency tree small, to correlate JSON logs with a request, to avoid floating-point errors in prices and quantities, and to work in the web client with familiar tooling. The module path needs no vanity-domain hosting or redirect.

It becomes harder to use router conveniences such as a mature middleware ecosystem, and decimal arithmetic is more verbose than float arithmetic. Playwright adds browser binaries to developer machines, and pinning tools with `go tool` enlarges `go.sum`. We accept those costs in exchange for reproducibility and a small runtime image.

## Alternatives considered

- `chi` and `gin`/`echo`: more routing features and middleware, but an unnecessary dependency and framework lock-in; the generated strict server only needs an `http.Handler`.
- `zerolog` and `zap`: faster than `slog`, but an extra dependency with no benefit at this scale; `deploy/AGENTS.md` already requires JSON logs with a request ID, which `slog` provides.
- `cockroachdb/apd`: more precise control over decimal context, but a verbose API; `ericlagergren/decimal`: fast, but a smaller ecosystem than shopspring.
- Jest: needs extra configuration with Vite and ESM; Cypress: heavier than Playwright for the critical paths named in `apps/web/AGENTS.md`.
- A vanity-domain Go module path: needs hosting and a redirect with no benefit now; `example.com/...`: states the wrong owner.
