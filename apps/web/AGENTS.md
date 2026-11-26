# Frontend (`apps/web`)

React + TypeScript, Vite SPA, Tailwind + shadcn/ui. General rules are in the root `AGENTS.md`. This file covers frontend specifics only.

## Layout

```
src/
  app/                  bootstrap, providers, root router, layouts (BaseLayout, PosLayout)
  features/<module>/<resource>/
    pages/              <Resource><Action>Page.tsx
    components/         <Name>.tsx (Table, Form, FormModal, ...)
    hooks/              use<Thing>.ts  (wraps TanStack Query)
    schemas/            <thing>FormSchema.ts (zod)
  features/<module>/routes.tsx   one route tree per module
  shared/               cross-module components and utilities
  paths.ts              nested PATHS object
```

## Rules

- **The API is reached only through `packages/api-client`** (generated, never edited). Do not hand-write `fetch` calls to endpoints.
- **Routes:** `/<module>/<resource-plural-kebab>`, add `/new`, detail `/:<entity>Id`, edit `/:<entity>Id/edit`. Never `create`/`add`, never `:id`. Paths come only from `PATHS` (e.g. `PATHS.inventory.products.detail(id)`), never as string literals in components.
- **Permission per route:** every route declares `permission: '<module>.<resource>.<action>'`. The route guard and the menu read it. Hiding a button in the UI is not security: the backend always checks too.
- **One `routes.tsx` per module.** There are no per-role route files: roles are access groups the Administrator can change.
- **Screen pattern (ADR-0010):** every resource has a list view (search, filters, grouping, CSV export, print) and a form view; documents also show their history. No kanban views.
- **Forms:** `react-hook-form` + `zod`. Schemas live in `schemas/`. Validation messages are in Indonesian.
- **Server data:** TanStack Query. Query keys come from one function per resource (`productKeys.list(params)`), never scattered literals.
- **UI language:** all user-visible text is Indonesian. Numbers and money are formatted `id-ID` through `shared/format`; never format by hand. Backend error codes are mapped to Indonesian messages in one place (`shared/errorMessages`).
- **Styling:** Tailwind + shadcn/ui.
- **POS** (`/pos`) uses the full-screen touch-friendly `PosLayout` and a keyboard-type barcode scanner. Receipts and drug labels are printed with browser print CSS (58/80 mm paper, label stickers), not a local print agent.
- **Tests:** Vitest for logic and components (`<name>.test.tsx` next to the file), Playwright for critical paths (login, POS sale, journal posting).

## Naming

See `docs/conventions/naming.md`. Pages end with `Page.tsx`, hooks start with `use`, tests are `<name>.test.ts(x)`. `make check-naming` enforces it.

## Do not

- Edit `packages/api-client/src/generated/**`.
- Write route paths as string literals in components.
- Compute business values (tax, totals, stock) in the frontend as the source of truth: the server computes, the frontend displays (a preview is fine, the final value comes from the server).
- Store patient data in `localStorage` or logs.
