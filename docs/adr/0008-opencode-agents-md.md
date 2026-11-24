# 0008. OpenCode as the AI tool, `AGENTS.md` as the canonical rules

- **Status:** Superseded by 0009 (skill location only; see ADR-0009)
- **Date:** 2026-11-24

## Context

The project is built with AI assistance and the AI must follow the architecture and naming rules consistently.

## Decision

The AI tool is OpenCode (v2). Rules live in `AGENTS.md`: a short root file plus nested files in `apps/backend`, `apps/web`, `api`, and `deploy`. OpenCode loads nested files when an agent works in that area, and does not import `@file` references, so the one always-loaded extra file is `docs/conventions/naming.md` via `instructions` in `opencode.json`. Skills live in `.agents/skills/<name>/SKILL.md`, which OpenCode scans directly (no symlinks). Permissions in `opencode.json` deny `git push` and destructive commands, ask before `git commit`, and deny hand edits of generated code. Existing migrations require confirmation to edit. The `plan` agent is overridden to be read-only and a read-only `reviewer` agent is defined in `.opencode/agents/`. Docs and rules are in English; user-facing text and pharmacy terms stay Indonesian (see `docs/glossary.md`).

## Consequences

- One canonical set of rules for every tool that reads `AGENTS.md`.
- Hard rules are backed by permissions and by `make check`, not only by prose.
- The `opencode.json` permission and agent schema was written from the OpenCode 2 documentation and should be validated against the installed version when first used.

## Alternatives considered

- **Per-tool rule files (`CLAUDE.md`, `.cursor`):** duplicates rules and invites two competing versions.
