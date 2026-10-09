# 10 · Milestones for Phase 1

Fourteen milestones, each small enough for one Claude Code session. Do them in order; each
one ends with green checks, commits pushed to `main` on GitHub and a green CI run. Tick the
boxes in this file as part of the milestone.

## How to run a milestone

1. Start a fresh session. Run `/milestone M<n>` (skill in `.claude/skills/milestone`).
   `git pull --rebase --autostash` first; the working tree must be clean unless it holds
   unfinished work of the same milestone from an interrupted session.
2. Read `CLAUDE.md` and the "Read first" documents of the milestone. Nothing else is required.
3. Plan in plan mode: the **test cases** of the milestone (IDs, level, expected result, which
   requirement or user story they prove), files to create or change, open questions. Ask the
   human only when the spec is silent or contradictory; otherwise follow the spec.
4. **Test cases first**: write `docs/test-cases/M<nn>-<slug>.md` (format and language in
   `docs/test-cases/README.md`), commit `test(M<n>): define test cases` and push.
5. Write the automated tests for those cases; run them and see them fail for the right reason.
   Do not commit them on their own: they are committed and pushed together with the code that
   makes them pass (`main` on GitHub is never red).
6. Implement in small steps until the tests pass. Comments in Indonesian, user-facing text in
   English (`CLAUDE.md`). Commit and push each step that leaves `make check` green.
7. Run the "Done when" commands. All must pass, plus `make check`; update the status column of
   the test-case file.
8. Run the `spec-reviewer` agent on the milestone's changes; fix every blocker and major
   finding.
9. Tick the checklist below, update the stage table in `docs/build-guide.md` (and its
   installation section when the milestone adds or changes a setup command), commit
   `feat(M<n>): <summary>`, push to `main`, and wait for the CI run to be green.

Never start a later milestone's feature early. If something from a later milestone is needed
as a dependency, add a minimal interface with a no-op implementation and a `TODO(M<k>)`
comment.

## Checklist

- [x] M0 Scaffold
- [ ] M1 Database, seed, sqlc, test harness
- [ ] M2 API contract, auth, users
- [ ] M3 Master data
- [ ] M4 Packages, app shell, create from template
- [ ] M5 PFD editor, changes engine, realtime
- [ ] M6 PFMEA AIAG 4th worksheet
- [ ] M7 Control Plan Template A
- [ ] M8 Rule engine and findings
- [ ] M9 Template General release and sync
- [ ] M10 Excel export
- [ ] M11 Dashboard, activity, search
- [ ] M12 Bulk paste from Excel
- [ ] M13 Hardening, performance, deployment, LDAP, Gate 1

---

## M0 · Scaffold

**Goal:** an empty but runnable repository with pinned tools and one command that checks
everything.

**Read first:** `docs/03-architecture.md` §2–3, §8 (configuration).

**Deliver**

- `go.mod` (module `pfmea`, `go 1.27`), `backend/cmd/pfmea/main.go` with `serve`
  (`/healthz` returns 200 while the process runs; `/readyz` runs the registered readiness checks
  and returns 200 when all pass, 503 otherwise; M0 registers none, M1 adds "database reachable
  and migrations current"), `backend/internal/config` (env parsing, defaults, validation errors
  listed together).
- `Makefile` with every target of `docs/03-architecture.md` §3.3. Working in M0: `tools` (Go
  tools and `npm ci` in `web/`), `db`, `dev`, `test`, `lint`, `build`, `check`. The others print
  "available from M<n>" and exit 0 until their milestone: `migrate`, `migrate-down`, `seed`,
  `test-rules` and `gen` (M1; `gen` is complete from M2), `e2e` (M2), `perf` (M13).
- `docker-compose.yml` (dev `db` service `postgres:18` with healthcheck and volume),
  `.env.example`, `.gitignore` (ignores the copied web build with
  `backend/internal/webui/dist/*` and keeps `!backend/internal/webui/dist/.keep`), `.golangci.yml`.
- `.github/workflows/ci.yml`: on every push to `main` (and on pull requests) run `make tools` and
  `make check` on `ubuntu-latest` with Go from `go.mod` and Node 24; from M2 a second job runs
  `make e2e` with Playwright Chromium. Testcontainers uses the runner's Docker.
- `web/`: SvelteKit 3 + Svelte 5 + TypeScript 6 + Vite 8, `adapter-static` with
  `fallback: 'index.html'`, `ssr = false`; ESLint, Vitest, Playwright config; `@fontsource`
  IBM Plex Sans/Mono; `src/lib/i18n/en.ts` (all UI strings, English); the app shell layout of `docs/08-screens.md` §1
  with placeholder pages; Vite dev proxy `/api` → `:8080` (WebSocket included).
- `backend/internal/webui` embedding `dist/` (committed placeholder `dist/.keep`; when empty the
  server answers "UI not built yet") and serving `index.html` for unknown non-API paths;
  hashed assets get `Cache-Control: public, max-age=31536000, immutable`.
- `docs/build-guide.md` (Indonesian): written in the first session by Langkah 3 of
  `docs/prompts/start-project.md`; if it is missing, write it as described there.

**Done when**

- `make check` passes locally and the first CI run on GitHub is green.
- `make dev` → http://localhost:5173 shows the shell; `curl -i localhost:8080/healthz` → 200.
- `make build && ./bin/pfmea serve` serves the SPA on :8080.
- A fresh clone from GitHub passes the installation steps of `docs/build-guide.md` that exist
  in M0, up to a green `make check`.
- Every hand-written file has its comments in Indonesian; every visible text is English.

**Not in M0:** database access, API endpoints.

**Decisions recorded in M0** (2026-10-09)

- `make tools` also installs air (live reload for `make dev`); tool versions are pinned in the
  Makefile and built with the Go version of `go.mod`. `make lint` runs `redocly lint` from M0.
- `docker-compose.yml` publishes Postgres on `127.0.0.1:${PG_PORT:-5432}`; the Vite proxy follows
  the port of `HTTP_ADDR`. Response bodies of `/healthz` and `/readyz`: `docs/03-architecture.md`
  §8. `APP_BASE_URL` must be an origin (scheme and host only) because M2 compares it with
  `Origin`.
- Repository-level tests live in `backend/internal/repotest`.
- `pfmea serve` force-closes connections still open 8 s after a shutdown signal; per-request read
  deadlines come with the M2 middleware chain (a global `ReadTimeout` would cut WebSockets).
- Decided after M0 (2026-10-09): `Me.devMode` in the contract tells the SPA to show the
  "DEMO DATA" badge; the CSP allows SvelteKit's inline start-up script by its SHA-256 hash
  computed by the server (`docs/03-architecture.md` §8). Both are built in M2.

## M1 · Database, seed, sqlc, test harness

**Goal:** the schema and demo data work end to end, with fast database tests.

**Read first:** `docs/04-data-model.md`, `docs/06-rules.md` §1, §5, §7, `docs/11-testing.md` §2.

**Deliver**

- `db/embed.go`; `pfmea migrate up|down|status` (goose, embedded files); `pfmea seed-demo`
  (`DEV_MODE` only). `/readyz` registers the check "database reachable and migrations current"
  (`docs/03-architecture.md` §8).
- `backend/internal/store`: pgx pool, `WithTx(ctx, actor, fn)` with `set_config` of
  `app.user_id`, `app.request_id`, `app.source`, retries on 40001 and 40P01, Postgres error
  mapping that also covers errors raised at `COMMIT`, and `LockPackage(ctx, tx, packageID)`
  (`docs/03-architecture.md` §3.1).
- `sqlc.yaml` + first queries (users, packages) + generated code.
- `backend/internal/testdb`: PostgreSQL 18 via testcontainers (or `TEST_DATABASE_URL`);
  migrate + seed once into a template database; each test gets `CREATE DATABASE … TEMPLATE`.
- Trigger tests: chain `s` follows effect S (insert, update, delete, move effect); `rpn` and
  `new_rpn` generated; version bumps only on user-meaningful change (not on `s`/`rpn`, not on
  no-op updates); `content_version` +1 per statement per touched package; audit rows store only
  changed columns and carry `app.*` settings; delete of a step with failure modes fails
  (RESTRICT); cross-package references fail (composite FKs); `snake_to_camel`.
- `make test-rules`: the fixture test of `docs/06-rules.md` §7 (rule SQL only; messages are
  checked from M8) plus the baseline counts (PS-07 = 14, GENERAL = 0).

**Done when:** `make db migrate seed` works; `make test` and `make test-rules` pass;
`pfmea migrate down` (from M8: River's migrations first, then goose to version 0) leaves no
application tables, functions or types; only goose's version table and the `pg_trgm`
extension remain.

**Not in M1:** HTTP endpoints, rule engine.

## M2 · API contract, auth, users

**Goal:** the generated API skeleton, login and user administration.

**Read first:** `docs/05-api.md` §1–3, §11; `api/openapi.yaml` (auth, users, Problem);
`docs/08-screens.md` §1–2.

**Deliver**

- `make gen`: oapi-codegen (`api/oapi-codegen.yaml`) → `backend/internal/httpapi/gen`;
  openapi-typescript → `web/src/lib/api/schema.d.ts`; sqlc.
- `httpapi`: server struct implementing `gen.StrictServerInterface`; every operation not yet
  built returns a 501 Problem (code `internal`, title "Not implemented yet") from `unimplemented.go`, and methods
  move out of that file as milestones implement them. Middleware chain of
  `docs/03-architecture.md` §5; strict-server error handlers mapping errors to Problem Details.
- `auth`: argon2id (PHC), sessions with cache and sliding expiry, login rate limiter,
  permission helpers (`canEditPackage`, `canRelease`, `canWaive`, `isAdmin`).
- Endpoints: `login`, `logout`, `getMe`, `changeOwnPassword`, `listUsers`, `createUser`,
  `updateUser`. `pfmea init --admin <name>` (admin + empty Template General).
- Web: API client (`openapi-fetch`, credentials, `X-Request-Id`, Problem → typed error),
  login page, auth guard, user menu, `/master/users`, "DEMO DATA" badge from `Me.devMode`.
- Security headers of `docs/03-architecture.md` §8, with `script-src` hashes of the inline
  scripts of the embedded `index.html` computed at start-up.

**Done when:** API tests for 401, 403 (role and Origin), 415, 429; the demo hash verifies
`pfmea-dev-2026`; Playwright: log in as `apratama`, see the shell, log out; `make lint`
includes `redocly lint`.

## M3 · Master data

**Goal:** admins maintain customers, classes, parts, control library, criteria, banned terms,
rules and settings.

**Read first:** `docs/05-api.md` §11 (master rows), `docs/08-screens.md` §12,
`docs/01-domain-glossary.md` §2.

**Deliver:** the master-data endpoints of `api/openapi.yaml` (tag `master`), validation,
`delete_blocked` for used classes, the `/master/*` screens. Changes that affect rules call a
`checks.Enqueuer` interface (no-op until M8, `TODO(M8)`); changes that affect exports clear
the export cache (no-op until M10).

**Done when:** API tests for CRUD, 409 `version_conflict`, 403 for non-admins, audit rows
present; Playwright: admin edits Customer B's symbol table.

## M4 · Packages, app shell, create from template

**Goal:** list, create and open packages; a model package is created from Template General.

**Read first:** `docs/07-template-general.md` §1–2, §3.1; `docs/02-phase1-scope.md` §3
(permissions); `docs/05-api.md` §1–3, §7 (conventions, errors, permissions, `MutationResult`);
`docs/08-screens.md` §4–5.

**Deliver:** `listPackages`, `createPackage` (create-from-template algorithm, document
numbers, header defaults), `getPackage` (permissions), `updatePackage`, `deletePackage`,
`replacePackageMembers`, `markPackageReviewed`, `getTemplateGeneral` (complete:
`hasUnreleasedChanges` compares `content_version` with the latest revision's, `lastSyncRun` is
null until M9 creates runs). Screens: packages list, create dialog,
package overview (cards for documents, members, settings; findings and template cards come in
M8/M9), sidebar package groups, breadcrumbs. After creation call `checks.Enqueuer` (no-op).

**Done when:** US-01 passes as an API integration test and in Playwright, except its last
clause ("a check run has executed"), which M8 adds; document numbers increase per customer; an
author cannot change the owner; a viewer cannot create.

## M5 · PFD editor, changes engine, realtime

**Goal:** the PFD can be edited by several users at once; the generic editing engine exists.

**Read first:** `docs/05-api.md` §4–7, §10; `docs/03-architecture.md` §4–6;
`docs/08-screens.md` §1 (common behaviours), §6.

**Deliver**

- `domain/entities.go` (all entities, fields, columns, types, limits, sync columns flag).
- `changes` engine for all entities (whitelist, validation, grouping, versions, template locks
  and overrides, derived rows, events). Virtual fields may wait for M6.
- `createStep`, `deleteStep`, `createStepFlow`, `deleteStepFlow`, `createCharacteristic`,
  `deleteCharacteristic`, `moveRow`, `getPfdView`; delete `dryRun` and blockers.
- `realtime`: publisher, listener, hub, presence, `/api/v1/ws`.
- Web: reusable grid component (Tabulator wrapper with lock states, conflict popover, finding
  markers placeholder), package store (normalized, applies events, ignores own `requestId`),
  PFD screen with header form, characteristics and flows editors, diagram (Svelte Flow + ELK
  worker), presence avatars.

**Done when:** US-02, US-03 and US-07 (on PFD fields) pass; tests for `locked_by_template`
and `override_reason_required` using PS-07's linked rows; Playwright with two browser
contexts sees the other user's change within 1 s; local benchmark of a one-cell save < 50 ms.

## M6 · PFMEA AIAG 4th worksheet

**Read first:** `docs/01-domain-glossary.md` §3; `docs/05-api.md` §4, §6–7; `docs/08-screens.md`
§7; `docs/11-testing.md` §5 (performance data set).

**Deliver:** `getPfmeaView`, `createPfmeaRow`, `createFailureEffect`, `deleteFailureEffect`,
`createControl`, `deleteControl`, `createAction`, `deleteAction`, `deleteFailureMode`,
`deleteFailureChain`; virtual fields; `completedOn` default; the PFMEA screen (grouping, merged
display, sorting, highlights, effect/control/action editors, library picker, criteria tooltips,
Action Results toggle, keyboard, clipboard into existing rows); `pfmea perf-gen`
(`docs/11-testing.md` §5), used from now on for scale tests.

**Done when:** US-04 passes; US-07 also passes on `failureChain.d` and on the virtual
`failureChain.dcText` (with `targetVersion`); tests for `not_unique_target` and the last-chain
delete rule; `PERF-3000` from `pfmea perf-gen` loads its PFMEA view in < 300 ms server time
and scrolls smoothly in Playwright (time to first row recorded).

## M7 · Control Plan Template A

**Read first:** `docs/01-domain-glossary.md` §4, `docs/05-api.md` §4, §7, `docs/08-screens.md` §8.

**Deliver:** `getControlPlanView`, `createCpLine`, `deleteCpLine`, `createCpLinesFromPfmea`,
reaction plan editing, the CP screen (header form, phase tabs, read-only PFD columns, link
column, "+ Lines from PFMEA controls" dialog).

**Done when:** US-06 passes except its R01 clause (checked in M8): two lines created with step,
characteristic, class, spec, control and failure mode links; the covered control is disabled in
the dialog and returned in `skipped` by the API.

## M8 · Rule engine and findings

**Read first:** `docs/06-rules.md` (all), `docs/08-screens.md` §9.

**Deliver:** rule loader, executor (snapshot parallelism), fingerprints, findings upsert,
`check_runs`, `package_stats`, `findings.changed` events, the in-process scheduler, River
(client, `pfmea migrate` runs River migrations, workers `check_packages`, periodic nightly
job, startup catch-up), `runCheck`, `listPackageFindings`, `listFindings`, `waiveFinding`,
`unwaiveFinding` (both refresh `package_stats`), `lastCheckRun` in findings lists; replace the
M3/M4 no-op enqueuer; `pfmea seed-demo` runs a full check after loading; findings screens, cell
markers, deep links, quick fixes for R01, R02, R03, W04; findings card on the package overview.

**Done when:** `make test-rules` also asserts the rendered messages of §5; US-01 (check clause),
US-06 (R01 fixed), US-08 and US-09 pass; package counters change right after a waive;
waived-to-error transition tested; incremental check after a save < 300 ms locally.

## M9 · Template General release and sync

**Read first:** `docs/07-template-general.md` (all), `docs/05-api.md` §5, §10,
`docs/08-screens.md` §5, §10.

**Deliver:** `template/columns.go`, `PlanSync` (pure), apply + `template_sync` River job,
`releaseTemplate`, `restoreTemplateRevision`, `getTemplateDiff`, `getTemplateImpact`,
`listTemplateRevisions`, `updateOverridePolicy`, `listSyncRuns`, `getSyncRun`, `retrySyncRun`,
`getTemplateLink`, `addGeneralSteps`, `detachLinkedRow`, `revertLinkedRow`,
`resolveSyncConflict`; the hook interface; Template General screen, package template card,
T03/T04 quick fixes, `sync.progress` and `package.reload` UX.

**Done when:** planner unit tests cover every case of §4.1 and §8; US-10 and US-11 pass as
integration tests and in Playwright; on the perf data set (24 linked packages) a release is
synced to all of them in < 10 s locally.

## M10 · Excel export

**Read first:** `docs/09-excel-export.md`, `docs/01-domain-glossary.md` §3–5,
`docs/07-template-general.md` §2 (document header defaults).

**Deliver:** layout loader and default layouts, writers for PFD, PFMEA, CP, `export_file` job,
`createExport`, `getExport`, `downloadExport`, cache and cleanup, cache clearing on master-data
changes (replace the M3 no-op), export buttons, `pfmea xlsx-compare`.

**Done when:** US-12 passes; golden tests open each exported file of PS-07 with excelize and
compare header cells and table cell text with `backend/internal/export/testdata/*.golden.json`;
page setup is A3 landscape with print titles.

## M11 · Dashboard, activity, search

**Read first:** `docs/08-screens.md` §1, §3; `api/openapi.yaml` (tag `monitoring`).

**Deliver:** `getDashboard` (from `package_stats` and indexed queries), `listActivity`
(audit rows grouped by request id, English summaries), `search`; dashboard screen, activity
lists, top-bar search.

**Done when:** US-13 passes (numbers equal the findings and actions lists for the same
filters); dashboard < 100 ms locally on the perf data set (24 linked packages).

## M12 · Bulk paste from Excel

**Read first:** `docs/05-api.md` §8, `docs/08-screens.md` §11.

**Deliver:** `pasteRows` (TSV parsing done in the browser; server-side mapping, fill-down,
matching, validation, creation with `pgx.Batch`), the paste dialog on PFD, PFMEA and CP.

**Done when:** US-05 passes with two fixtures that M12 writes in
`backend/internal/worksheet/testdata/`:
- `pfmea-40-rows.tsv`: a header row plus 40 rows for new steps 30 "Solder paste printing" and
  40 "SPI" of PS-07 (2 steps × 2 characteristics × 2 failure modes × 5 causes; repeated
  cells left empty as Excel copies merged cells). Pasting it creates exactly 2 steps,
  4 characteristics, 8 failure modes, 8 effects, 40 causes, 40 chains and 80 controls;
- `pfmea-40-rows-broken.tsv`: the same rows with D = 11 in row 7 and O = "x" in row 23 (rows
  counted from the first data row; D and O are never filled down). It creates nothing and
  reports exactly those 2 row errors.
2,000 rows paste in < 3 s.

## M13 · Hardening, performance, deployment, LDAP, Gate 1

**Read first:** `docs/03-architecture.md` §8–9, `docs/11-testing.md` §5–6.

**Deliver:** `pfmea perf` with pass/fail thresholds (`make perf`) and a `perf` compose profile
that runs a second app container against the `pfmea_perf` database;
`deploy/` (Dockerfile, compose `prod` profile, Caddyfile, `postgresql.conf`, `backup.sh`,
`restore.sh`); LDAP login; `/metrics`; security headers review; README operations section in
Indonesian; Gate 1 checklist ready for QA.

**Done when:** `make perf` meets every target on the reference machine; a backup is restored
on a second machine (documented); LDAP login works against the company directory or a test
container; every item of the Gate 1 checklist has an owner and a way to verify it.
