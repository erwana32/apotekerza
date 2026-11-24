# 0009. Skills live in `.opencode/skills/`

- **Status:** Accepted
- **Date:** 2026-11-25

## Context

ADR-0008 placed skills in `.agents/skills/`, which OpenCode scans directly. The developer wants the skills to be visible and checkable next to the other OpenCode configuration in `.opencode/`. Placing skills under `.opencode/agents/` is not possible: OpenCode registers every Markdown file there as an agent (verified: a file in `.opencode/agents/skill/` appeared as the primary agent `skill/add-endpoint`).

## Decision

Skills live in `.opencode/skills/<name>/SKILL.md`, which is OpenCode's own recommended location. This supersedes the skill location in ADR-0008; the rest of ADR-0008 (OpenCode as the tool, `AGENTS.md` canonical, permissions, `plan` and `reviewer` agents, language rules) remains in force. Agents stay in `.opencode/agents/`. Nothing is kept in `.agents/`.

## Consequences

- OpenCode configuration, agents, and skills are all under `.opencode/`.
- Other AI tools (Claude Code, Codex) do not read `.opencode/skills/`. This project uses OpenCode only, so this is accepted. If another tool is adopted, add a link or copy for it and record that in a new ADR.
- Skill names stay `kebab-case` and equal to the folder name.

## Alternatives considered

- **Keep `.agents/skills/` plus an index document:** tool-neutral, but the skills would not sit with the rest of the OpenCode setup.
- **`.opencode/agents/skill/`:** rejected, files there are loaded as agents.
