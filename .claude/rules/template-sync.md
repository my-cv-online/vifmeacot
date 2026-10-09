---
paths:
  - "backend/internal/template/**"
---
# Template General rules

- Follow `docs/07-template-general.md` exactly. The sync columns per table live in
  `columns.go` and are used by copy, planner, diff and policy validation.
- `PlanSync(old, new, model, policy)` is pure (no database, no clock); every case of §4.1 and
  every edge case of §8 has a table-driven test with hand-built snapshots.
- The sync job is idempotent: lock `package_links` `FOR UPDATE`, skip when `synced_rev >=
  to_rev`, apply inserts → updates → review marks → deletes, log every item in
  `sync_changes`, set `app.source = 'sync'`, and finish with a full check after commit.
- Never sync `seq`, link columns, `general_mode`, derived columns, `documents` or `actions`.
- `OnReleasedPackageSynced` stays a no-op in Phase 1.
