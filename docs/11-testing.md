# 11 · Testing and acceptance

**Test cases come first.** Every change (new feature, update, bug fix) starts by writing or
extending its test cases in `docs/test-cases/` (format: `docs/test-cases/README.md`). The test
cases are committed and pushed before the feature code; then the automated tests are written
and seen failing; then the code is written until they pass. Failing tests are never committed
on their own; they are committed and pushed together with the code that makes them pass, so
`main` on GitHub stays green. Every automated test carries the ID
of the test case it implements in its name, and every test case lists the automated test that
covers it.

## 1. Layers

| Layer | Tool | What | Command |
| --- | --- | --- | --- |
| Go unit | `testing` | Pure functions: validation, entity map, document numbers, fingerprints, message rendering, `PlanSync`, paste mapping, diff | `make test` |
| Go integration | `testing` + real PostgreSQL 18 | Store, triggers, services, handlers through `httptest` with a cookie jar | `make test` |
| Rule fixtures | Go test over `backend/internal/rules/{sql,testdata}` | Every rule, baseline of `docs/06-rules.md` §5 | `make test-rules` |
| Contract | Go test | Views of the demo packages decode into generated types with `DisallowUnknownFields`; no operation returns 501 at the end of Phase 1 | `make test` |
| Frontend unit | Vitest | Store joins (view → grid rows), lock-state derivation, conflict rebase, TSV parsing and fill-down, formatting | `make test` |
| End to end | Playwright (Chromium) | User stories US-01 … US-13 | `make e2e` |
| Performance | `pfmea perf` | Targets of `docs/03-architecture.md` §9 | `make perf` |
| Test-case documents | Markdown in `docs/test-cases/` | The list of cases each milestone or change must satisfy, written before the code | review |

`make check` = `make gen` (must not change tracked files) + `make lint` + `make test` +
`make test-rules`. It must pass before every commit and push. The GitHub Actions workflow runs
it again on every push to `main`; from M2 it also runs `make e2e`.

## 2. Database in tests

- `backend/internal/testdb` starts `postgres:18` (with `-c jit=off`) once per `go test`
  process with testcontainers-go, or uses `TEST_DATABASE_URL` when set (CI without Docker).
- It creates a template database once: migrate up, load `db/seed/demo.sql`. Each test calls
  `testdb.New(t)` which runs `CREATE DATABASE t_<random> TEMPLATE pfmea_tpl` (tens of
  milliseconds) and drops it in `t.Cleanup`. Tests can run in parallel.
- Time is injected: services take a `clock.Clock`; tests use 2026-10-08 in `Asia/Jakarta`, the
  date the demo data and fixtures assume. The server honours `DEV_FAKE_TODAY=2026-10-08` only
  when `DEV_MODE=true` (used by E2E).
- Demo ids are deterministic (uuid5). Prefer looking rows up by business keys (`code`,
  `op_no`, `char_no`) in tests for readability.

## 3. Rule fixtures

Algorithm (Go port of the harness used while writing the spec):

```text
for each testdata/<CODE>.sql:
  expect = JSON after "-- expect:"
  BEGIN
    exec fixture SQL
    rows = run sql/<CODE>.sql with @package_id = PS-07, @today = 2026-10-08
  ROLLBACK
  assert len(rows) == len(expect)
  every expected item matches a distinct row: same object_type and field, params ⊇ expected params
```

Plus: all rules on PS-07 give exactly the 14 findings of `docs/06-rules.md` §5 (with rendered
messages from M8) and on GENERAL give none.

## 4. End-to-end scenarios

Playwright runs with one worker against a server built with `make build`, started with
`DEV_MODE=true`, `DEV_FAKE_TODAY=2026-10-08` and a dedicated database. Before each spec file
the helper `resetDb()` stops the server, runs `pfmea seed-demo --reset` (drop schema, full
migrate including River, seed, and from M8 a full check of every package) and starts the
server again (about 2 s); the restart avoids stale prepared statements, cached sessions and
River state. Specs live in `web/tests/e2e/` and use the demo users (password
`pfmea-dev-2026`). Each test title starts with its test-case ID, for example
`test('TC-M04-001 create MB-01 with the optional laser marking step', …)`.

| Spec | Story | Milestone |
| --- | --- | --- |
| `us01-create-package.spec.ts` | Create MB-01 with optional step 65 | M4 (check clause: M8) |
| `us02-edit-pfd.spec.ts` | Add step 30 and characteristic 30-01; second browser sees it < 1 s | M5 |
| `us03-delete-protection.spec.ts` | Delete step 50 → blockers dialog | M5 |
| `us04-pfmea-row.spec.ts` | S/RPN computed; effect S change updates chains | M6 |
| `us05-bulk-paste.spec.ts` | 40 rows pasted; broken input creates nothing | M12 |
| `us06-cp-from-pfmea.spec.ts` | Two CP lines from uncovered controls; R01 fixed | M7, M8 |
| `us07-concurrent-edit.spec.ts` | Two users, same cell, conflict prompt | M5 (PFD), M6 (chain and virtual fields) |
| `us08-fix-finding.spec.ts` | Change D on step 90 → R04 fixed | M8 |
| `us09-waive.spec.ts` | Approver waives R02; author cannot; Error cannot | M8 |
| `us10-release-template.spec.ts` | Release rev 2, PS-07 receives changes, T03 on 20-02 | M9 |
| `us11-resolve-conflict.spec.ts` | "Follow template" and "Keep local" | M9 |
| `us12-export.spec.ts` | PFMEA export < 3 s; second export cached | M10 |
| `us13-dashboard.spec.ts` | Dashboard numbers equal the findings and actions lists | M11 |

## 5. Performance

`pfmea perf-gen` (milestone M6, idempotent) works only on a database whose name ends in `_perf`
(for example `pfmea_perf`) and refuses any other. On an empty database it migrates and loads
the demo seed (Template General rev 1, PS-07 and the demo users), immediately replaces every
user's password with `PERF_PASSWORD` (required environment variable, so no publicly known
password exists there), then creates model packages from Template General with the normal
create-from-template service until 24 packages are linked: PS-07, `PERF-3000` and 22 small
ones. `PERF-3000` has 40 steps,
200 characteristics, 1,000 failure modes × 3 causes = 3,000 chains, 6,000 controls, 500
actions, 2,000 CP lines across three phases. Text values are realistic lengths (30–120
characters).

`pfmea perf` (milestone M13) runs the measurements of `docs/03-architecture.md` §9 against a
running server (`--base-url`), logging in as `apratama` and `rsaputri` with `PERF_PASSWORD`;
those credentials exist only in a perf database, so it cannot run against production by
accident. It prints a table with p50/p95/max and target, writes `perf-report.json`, and exits
non-zero when a target is missed. Its release measurement publishes a new revision of the perf
database's Template General, never a real one. Before Gate 1, run it on the factory server
with a second app container (`docker compose --profile perf`, bound to `127.0.0.1` and
stopped after the run) pointing at `pfmea_perf`; record the machine (CPU, RAM, disk) in the
report.

## 6. Gate 1 checklist (Phase 1 acceptance)

| # | Criterion | How it is verified | Owner |
| --- | --- | --- | --- |
| 1 | A real package re-created in the system | QA picks the package; engineers rebuild it with bulk paste and editors | QA + process engineering |
| 2 | Excel exports match the old files | `pfmea xlsx-compare` report with no unexplained differences; QA signs the printed A3 layout | QA |
| 3 | All 32 rules correct on that package | QA reviews every finding (true/false positive) and missed issues; fixes become rule fixtures | QA |
| 4 | Real Template General released and synced | Template contains the company's general processes; release shows all packages `done` | QA Manager (approver) |
| 5 | Performance targets met | `perf-report.json` from the factory server (perf profile, `pfmea_perf` database), all targets green | IT |
| 6 | Backup restored on a second machine | `deploy/restore.sh` log and a login on the restored copy | IT |

## 7. Definition of done (every change)

- Test cases written in `docs/test-cases/` and pushed before the code; every case has an
  automated test whose name contains the case ID; status column updated.
- `make check` green locally, pushed to `main`, CI run green; E2E of the touched user story
  green.
- Comments in Bahasa Indonesia in every hand-written file (file header, every function, type,
  constant and test, non-obvious steps); identifiers, file names and commit messages in
  English.
- No new user-facing string outside `web/src/lib/i18n/en.ts` (UI) or
  `backend/internal/i18n/en.go` (server); rule messages stay in the rule SQL headers; all of
  them in English.
- API changes made in `api/openapi.yaml` first; generated code regenerated, not edited.
- Schema changes as a new migration file; docs (`docs/04-data-model.md`) updated in the same
  change.
- No feature of a later phase; no `TODO` without a milestone tag.
