---
paths:
  - "backend/internal/rules/**"
---
# Rule engine rules

- Each rule = `sql/<CODE>.sql` + `testdata/<CODE>.sql`. Header lines `-- rule:`, `-- reads:`,
  `-- message:` (Go text/template, English because users read it), `-- returns:`; other comment
  lines and the fixture `-- scenario:` are Indonesian; parameters `@package_id`,
  `@today`; columns `object_type, object_id, field, key, params` in that order
  (`docs/06-rules.md` §1).
- Write the test case (`docs/test-cases/rules.md`, ID `TC-RULE-<CODE>-<n>`) and the fixture
  first; the fixture's `expect` is the complete output of the rule on PS-07 after the fixture.
  Run `make test-rules`.
- `field` must be the API field name (camelCase); `params` keys camelCase; filter every query
  by `@package_id`; never read `findings`; keep joins indexed.
- `reads` must list every table whose change can change the result; it drives incremental
  checks.
- The demo baseline (PS-07 = 14 findings with the messages in `docs/06-rules.md` §5, GENERAL =
  0) must stay true; if a rule change alters it on purpose, update §5 in the same change.
- Phase 2 rules (K04, F04–F06, C02–C06, W03) are not implemented in Phase 1.
