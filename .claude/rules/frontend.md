---
paths:
  - "web/**"
---
# Frontend rules (SvelteKit 3 SPA, Svelte 5)

- SPA only: `adapter-static` with `fallback: 'index.html'`, `export const ssr = false`. No server
  routes in SvelteKit; all data comes from `/api/v1` through `src/lib/api/client.ts`
  (openapi-fetch + generated `schema.d.ts`). Never call `fetch` directly from components.
- Svelte 5 runes (`$state`, `$derived`, `$effect`, `$props`). Shared state lives in
  `src/lib/stores/*.svelte.ts` classes; the package store keeps the normalized view, applies
  realtime events and ignores events carrying its own `requestId`.
- Every user-visible string is English and comes from `src/lib/i18n/en.ts`; AIAG form column
  headers are written exactly as on the forms. Component, file and route names are English.
- Comments (TypeScript, Svelte, CSS) are written in Bahasa Indonesia: a short comment at the
  top of every file or component, on every function, store class and test, and on
  non-obvious steps.
- Grids use the shared Tabulator wrapper (`src/lib/components/grid`). Update rows in place
  (`updateData`, `addData`, `deleteRow`); never rebuild the table on every change. Virtual DOM
  rendering on; no per-cell Svelte components inside Tabulator.
- The PFD diagram uses Svelte Flow; ELK layout runs in `elk.worker.ts`, debounced 300 ms.
- No UI component kits or CSS frameworks; use the design tokens of `docs/08-screens.md` §1 as
  CSS custom properties in `src/app.css`. Fonts come from `@fontsource` (no CDN).
- Accessibility: touch targets ≥ 44 px, visible focus, keyboard operation of grids and dialogs,
  `aria-*` on icon buttons.
- Phase 2+ controls are not rendered (`docs/08-screens.md` §13). Never show fake data.
- Tests: Vitest for stores and utilities; Playwright specs per user story in `tests/e2e`; every
  test title starts with its test-case ID (`TC-M05-002 …`).
