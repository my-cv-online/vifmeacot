---
paths:
  - "backend/**/*.go"
  - "db/**/*.go"
---
# Go backend rules

- Handlers in `backend/internal/httpapi` are thin: generated decode → one service call → map the
  result. Validation, permissions and business rules live in the service packages.
- Pass `context.Context` as the first argument everywhere; no package-level mutable state
  except the parsed config and the embedded files.
- Errors: wrap with `fmt.Errorf("…: %w", err)`; return typed application errors
  (`apperr.Problem` with a `ProblemCode` from `api/openapi.yaml`) for anything the client must
  see; map PostgreSQL errors centrally in `store` (`docs/05-api.md` §2). Never leak SQL or stack
  traces to clients.
- Writes only through `store.WithTx(ctx, actor, fn)`; a transaction that writes package content
  calls `store.LockPackage` first (lock order: `package_links` → package row → content rows).
  Reads that need a consistent view of several tables use a `REPEATABLE READ READ ONLY`
  transaction.
- User-facing texts produced by the server are English and come from `backend/internal/i18n/en.go`.
- Comments are written in Bahasa Indonesia: a short comment at the top of every file (the
  `// Package x ...` doc comment in one file per package), a doc comment on every function,
  type and constant (exported or not) and on non-obvious steps. Doc comments start with the
  identifier name: `// LockPackage mengunci baris paket sebelum konten ditulis.` Log messages
  stay English.
- Time from the injected `clock.Clock` (plant time zone); ids from `uuid.NewV7()` when Go must
  know them before insert (template copy and sync), otherwise the database default.
- Logging with `log/slog` (JSON); include `request_id`, `user_id`, `package_id` attributes.
- Tests: table-driven; database tests use `testdb.New(t)` (real PostgreSQL 18, demo seed);
  run with `-race` in `make test`. Test names contain the test-case ID from `docs/test-cases/`
  (e.g. `TestCreatePackage_TC_M04_001`), with an Indonesian comment describing the scenario;
  subtest names and failure messages (`t.Errorf`, `t.Fatalf`) are English.
  A bug fix starts with a new test case and a failing test.
- Keep packages within the dependency direction of `docs/03-architecture.md` §3.1; services
  talk to realtime and jobs through small interfaces so they stay testable.
- Run `gofmt`, `go vet` and `golangci-lint` (via `make lint`) on `./backend/... ./db/...`.
