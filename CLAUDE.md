# PFMEA · Control Plan · PFD system

Web application for an automotive electronics (PCBA, IATF 16949) plant to create, check and
monitor three linked documents per product: Process Flow Diagram, PFMEA and Control Plan.
**Current scope: Phase 1** — AIAG FMEA 4th edition, Control Plan Template A (APQP 2nd ed),
Template General synced into model packages, 32 consistency rules, Excel A3 export, dashboard,
bulk paste. The end state (4 phases) is described so you never block later phases.

## Where things are specified

Read this file and the "Read first" list of the current milestone. Open other docs only when
needed (they are not imported on purpose).

| Doc | Content |
| --- | --- |
| `docs/00-vision.md` | End state after Phase 4, roadmap and gates, non-functional requirements, forward-compatibility rules |
| `docs/01-domain-glossary.md` | Domain primer, glossary, AIAG form columns ↔ data model |
| `docs/02-phase1-scope.md` | Capabilities P1-01…13, out of scope, permission matrix, user stories US-01…13, Gate 1 |
| `docs/03-architecture.md` | Stack and versions, repo layout, read/write path, realtime, jobs, config, performance targets, decision log |
| `docs/04-data-model.md` | Intent of the schema (`db/migrations/00001_init.sql` is the truth) |
| `docs/05-api.md` | API rules (contract: `api/openapi.yaml`), editing semantics, locks, WebSocket protocol |
| `docs/06-rules.md` | Rule engine and the 32 Phase 1 rules (`backend/internal/rules/sql`, fixtures in `testdata`) |
| `docs/07-template-general.md` | Create from template, release, sync planner, overrides, conflicts |
| `docs/08-screens.md` | Screens, UI behaviour, English UI text, what to hide |
| `docs/09-excel-export.md` | Excel export layouts and mapping |
| `docs/10-milestones.md` | M0…M13 with deliverables and done criteria; tick boxes there |
| `docs/11-testing.md` | Test-first rule, test layers, database harness, E2E, performance, Gate 1 |
| `docs/test-cases/` | Test cases written before the code (format in its README; content in Indonesian) |
| `docs/build-guide.md` | Build stages and installation guide for the team (Indonesian; written in the first session) |
| `docs/prompts/start-project.md` | Prompt of the first session: GitHub setup, build guide, M0 |
| `docs/mockups/` | Static HTML mockups of the end state with Indonesian labels (layout reference only) |

Already written and executed while preparing this spec (on PostgreSQL 16 with a `uuidv7()`
shim; M1 re-runs everything on PostgreSQL 18): the schema (`db/migrations/00001_init.sql`),
demo data (`db/seed/demo.sql`), 32 rule queries with fixtures. The OpenAPI file passes
`redocly lint` and generates TypeScript types; the Go generator runs for the first time in M2.

## Stack

Go 1.27 (`net/http`, pgx v5, sqlc, goose, River, coder/websocket, excelize), PostgreSQL 18,
Svelte 5 + SvelteKit 3 static SPA embedded in the Go binary (Tabulator grid, Svelte Flow +
ELK diagram), OpenAPI 3.0.3 → oapi-codegen v2 strict server + openapi-typescript/openapi-fetch.
Pinned versions: `docs/03-architecture.md` §2.

## Layout

```text
api/            openapi.yaml (contract), oapi-codegen.yaml, redocly.yaml
db/             migrations/ (goose), seed/demo.sql, embed.go
backend/cmd/pfmea/            single binary: serve, migrate, init, seed-demo, perf-gen, perf, xlsx-compare
backend/internal/<module>/    config, apperr, clock, i18n, testdb, httpapi(+gen), auth, store(+queries, sqlc),
                              domain, worksheet, changes, rules(sql, testdata), template, export, realtime, jobs, webui
web/            SvelteKit app: src/routes, src/lib/{api,components,stores,i18n}, tests/e2e
deploy/         Dockerfile, Caddyfile, postgresql.conf, backup/restore scripts
.github/        workflows/ci.yml (M0)
```

`go.mod` is at the repository root (module `pfmea`). Run Go tools with explicit patterns
(`./backend/... ./db/...`) so `web/node_modules` is never scanned.

## Commands

| Command | Does |
| --- | --- |
| `make tools` | Install pinned code generators and linters, and the web dependencies (`npm ci`) |
| `make db` / `make migrate` / `make seed` | Start PostgreSQL 18 in Docker / migrate / load demo data |
| `make gen` | oapi-codegen + sqlc + openapi-typescript (never edit their output) |
| `make dev` | Go server (:8080) + Vite (:5173) with proxy |
| `make test` · `make test-rules` · `make e2e` | Go + Vitest · rule fixtures · Playwright |
| `make lint` · `make perf` · `make build` | Linters incl. redocly · performance targets · production binary |
| `make check` | gen + lint + test + test-rules; must be green before every commit and push |

Demo users (development only, never production): `admin`, `rsaputri` (approver),
`apratama` and `dhidayat` (authors), `swulandari` (reviewer), `operator1` (viewer); password
`pfmea-dev-2026`. Demo "today" is 2026-10-08.

## Language policy

| What | Language |
| --- | --- |
| Everything a user sees: UI labels, messages, errors, rule messages, Excel output, file names | English (`web/src/lib/i18n/en.ts`, `backend/internal/i18n/en.go`, rule SQL headers) |
| Comments in every hand-written file (Go, TypeScript, JavaScript, Svelte, HTML, CSS, SQL, Makefile, YAML, Dockerfile, shell) | Bahasa Indonesia |
| Identifiers, file and folder names, API fields, database names, log messages, commit messages | English |
| Specification docs for the AI (`CLAUDE.md`, `docs/0*.md`, `docs/1*.md`) | English |
| Docs for the team (`README.md`, `docs/build-guide.md`, `docs/test-cases/*.md`) | Bahasa Indonesia (file names English) |

Comments are there so the team can follow the code, so write them everywhere: a short header
comment in every hand-written file saying what it is for, a comment on every function, type,
constant and test, and a comment on every step inside a function that is not obvious. Explain
the why; do not restate the code line by line. Go doc comments start with the identifier name
followed by Indonesian text, e.g. `// LockPackage mengunci baris paket sebelum konten ditulis.`
Generated code (`gen/`, `sqlc/`, `schema.d.ts`) and JSON files (no comment syntax) are exempt;
the descriptions inside `api/openapi.yaml` stay English because they are the API documentation.

## Rules you must follow

- IMPORTANT: **Test cases first.** Before writing or changing feature code (new feature,
  update or bug fix), write the test cases in `docs/test-cases/` (format in its README),
  commit and push them; then write the automated tests and see them fail; then write the code.
  Every automated test name contains its test-case ID. Never weaken or delete a test to get
  green.
- IMPORTANT: **Everything goes to `main` on GitHub.** One branch, `main`, remote `origin`.
  Small Conventional Commits in English (`feat:`, `fix:`, `test:`, `docs:`, `chore:`,
  `refactor:`, `ci:`). Push after every step that leaves `make check` green. Failing tests are
  never committed on their own: commit them together with the code that makes them pass, once
  `make check` is green. Before the Makefile exists (start of M0), documentation-only commits
  are pushed without it. Never push a red `main`, never force-push, never rewrite pushed
  history, never commit secrets (`.env`) or build output.
  A session ends with a clean working tree, `main` equal to `origin/main` and a green CI run
  (`gh run list`, `gh run watch`). If a session must stop inside a milestone, push what is
  green, leave the unfinished work uncommitted and list it in the report; the next session
  continues from it.
- IMPORTANT: Build only Phase 1. AIAG-VDA, CP-1/Safe Launch, review/approval workflow, PDF,
  legacy Excel import, notifications and trends are later phases: hide their UI, never fake
  them, but keep the forward-compatibility rules of `docs/00-vision.md` §7.
- IMPORTANT: Integrity lives in PostgreSQL. Triggers maintain `failure_chains.s`, `rpn`,
  `new_rpn`, `version`, `updated_at`, `content_version` and `audit_log`. Code never writes
  those columns and never computes S or RPN for storage.
- Every write goes through `store.WithTx` (sets `app.user_id`, `app.request_id`,
  `app.source`), locks the package row first (`FOR NO KEY UPDATE`) when it writes package
  content, uses optimistic locking (`WHERE id = $1 AND version = $2`), and publishes realtime
  events with `pg_notify` inside the same transaction.
- No ORM. SQL lives in sqlc queries or in pgx calls with parameters; never concatenate values
  into SQL. Only whitelisted identifiers from `domain/entities.go` (and the validated snapshot
  id in the rule engine) may be interpolated.
- Contract first: change `api/openapi.yaml` (and `docs/05-api.md` when semantics change), run
  `make gen`, then implement. Never hand-edit `backend/internal/httpapi/gen`,
  `backend/internal/store/sqlc` or `web/src/lib/api/schema.d.ts`.
- `db/migrations/00001_init.sql` is final; after M1 is committed never edit a migration, add
  `0000N_<what>.sql` and update `docs/04-data-model.md`.
- Rules: one SQL file plus one fixture per rule; the demo baseline (PS-07 = 14 findings,
  GENERAL = 0) must stay green. Use the `/add-rule` skill.
- Template General behaviour follows `docs/07-template-general.md` exactly; `PlanSync` is a
  pure function with table-driven tests.
- Permissions are enforced in services (`docs/02-phase1-scope.md` §3), not only in the UI.
- Performance budgets (`docs/03-architecture.md` §9) are requirements: one SQL statement per
  view, no N+1 queries, batched writes, no work inside transactions that waits on I/O.
- Licensed AIAG/VDA text (S/O/D criteria, AP tables) never appears in code, seeds or tests.
- Time comes from the injected clock in the plant time zone; never call `time.Now()` in
  business logic.
- Enum values travel as their database values (`in_sync`, `pre_launch`).
- When the spec is silent or contradictory, ask instead of guessing, then write the decision
  into the relevant doc.

## Workflow

One milestone per session: `/milestone M<n>` → plan mode (test cases + files) → test-case file
committed and pushed → failing tests → code → `make check` and the milestone's E2E specs →
`spec-reviewer` agent → tick `docs/10-milestones.md` and the stage table in
`docs/build-guide.md` → commit, push to `main`, CI green → short report in Indonesian.
Definition of done: `docs/11-testing.md` §7.
