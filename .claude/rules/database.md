---
paths:
  - "db/**"
  - "sqlc.yaml"
  - "backend/internal/store/**"
---
# Database rules

- `db/migrations/00001_init.sql` is final once M1 is committed. Never edit an applied
  migration: add `0000N_<what>.sql` with `-- +goose Up` / `-- +goose Down` sections (wrap
  functions and DO blocks in `-- +goose StatementBegin/End`), and test `make migrate-down`
  then `make migrate` round trips.
- Every content table has `package_id`, composite foreign keys `(package_id, x_id)`, the
  template link columns (`origin`, `source_id`, `source_rev`, `sync_status`, `overrides`,
  `detach_reason`) except `actions`, `version` + `updated_at` with `trg_touch_row`, and
  `attach_audit(...)`. Pass derived columns as trigger arguments so they never bump `version`.
- Derived values stay in the database (triggers, generated columns). Do not duplicate them in
  Go.
- Add an index for every new lookup by `package_id` + parent id used by views or rules.
- snake_case in SQL; API names are camelCase (`snake_to_camel()` exists for data-driven names).
- sqlc queries live in `backend/internal/store/queries/*.sql`, one file per area, named
  `-- name: VerbNoun :one|:many|:exec|:batchexec`. Run `make gen` after changes.
- Update `docs/04-data-model.md` in the same change as any schema change, and keep
  `db/seed/demo.sql` loading cleanly.
- Keep `jit = off` (server config and test sessions); it slows the view and rule queries.
- SQL comments are written in Bahasa Indonesia; text values that users see (rule titles, seed
  data shown in the UI) are English.
