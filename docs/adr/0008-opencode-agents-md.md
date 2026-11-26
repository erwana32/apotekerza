# 0008. OpenCode as the AI tool, `AGENTS.md` as the canonical rules, skills in `.opencode/skills/`

- **Status:** Accepted
- **Date:** 2026-11-24

## Context

The project is built with AI assistance and the AI must follow the architecture and naming rules consistently. The developer wants the OpenCode configuration, agents, and skills visible together in `.opencode/`.

## Decision

The AI tool is OpenCode (v2).

- **Rules** live in `AGENTS.md`: a short root file plus nested files in `apps/backend`, `apps/web`, `api`, and `deploy`. OpenCode loads nested files when an agent works in that area, and does not import `@file` references, so the one always-loaded extra file is `docs/conventions/naming.md` via `instructions` in `opencode.json`.
- **Skills** live in `.opencode/skills/<name>/SKILL.md`, OpenCode's own recommended location. Nothing is kept in `.agents/`. Skill names are kebab-case and equal to the folder name.
- **Agents** live in `.opencode/agents/`: a read-only `reviewer`; the built-in `plan` agent is overridden to be read-only in `opencode.json`. Only agent definitions go in that folder: OpenCode registers every Markdown file there as an agent (verified), so skills must not be placed in it.
- **Permissions** in `opencode.json` deny `git push` and destructive commands, ask before `git commit`, deny hand edits of generated code, and ask before editing existing migrations.
- **Language:** docs and rules are in English; user-facing text and pharmacy terms stay Indonesian (see `docs/glossary.md`).

## Consequences

- One canonical set of rules for every tool that reads `AGENTS.md`.
- Hard rules are backed by permissions and by `make check`, not only by prose.
- Other AI tools (Claude Code, Codex) do not read `.opencode/skills/`. This project uses OpenCode only, so this is accepted. If another tool is adopted, add a link or copy for it and record that here.
- Whether OpenCode discovers skills in `.opencode/skills/` and how the permission globs behave on real paths should be confirmed the first time they are used.

## Alternatives considered

- **Per-tool rule files (`CLAUDE.md`, `.cursor`):** duplicates rules and invites two competing versions.
- **Skills in `.agents/skills/`:** scanned by OpenCode and tool-neutral, but not next to the rest of the OpenCode setup.
- **Skills under `.opencode/agents/skill/`:** rejected, files there are loaded as agents.

## Change log

- 2026-11-25: moved skills from `.agents/skills/` to `.opencode/skills/`.
