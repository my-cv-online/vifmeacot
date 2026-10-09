---
name: add-rule
description: Add or change a consistency rule (SQL file + fixture) of the rule engine, keeping the demo baseline green. Use whenever a rule's logic, message, level or catalog entry changes.
argument-hint: <RULE CODE> <what should change>
---

# Add or change rule $ARGUMENTS

Contract: `docs/06-rules.md` §1. Steps:

1. Decide the code and level. New catalog entries need a new migration that inserts into
   `rules` (never edit `00001_init.sql`); Phase 2 codes stay unimplemented in Phase 1.
2. **Test case first**: add the case to `docs/test-cases/rules.md` (ID `TC-RULE-<CODE>-<n>`,
   Indonesian), commit and push it. Then write or extend
   `backend/internal/rules/testdata/<CODE>.sql` with `-- fixture:`, `-- scenario:`
   (Indonesian) and `-- expect:` (the complete output of the rule on PS-07 after the fixture,
   baseline included). Run `make test-rules` and see it fail.
3. Write `backend/internal/rules/sql/<CODE>.sql`: header lines, `@package_id` filter, columns
   `object_type, object_id, field, key, params` in order, camelCase `field` and `params` keys,
   indexed joins, no reads of `findings`. List every result-changing table in `-- reads:`.
4. Message: one English line in Go `text/template` (users read it), only variables that exist
   in `params`. Other comments in the file are Indonesian.
5. Run `make test-rules` until green. Check the demo baseline (`docs/06-rules.md` §5): if the
   change alters it on purpose, update §5 and the baseline test in the same change.
6. Update the rule table in `docs/06-rules.md` §4 (what it checks, reads, target) and the
   deep-link mapping if the target is new.
7. Run `make check`, commit (`feat(rules): …` or `fix(rules): …`) and push to `main`; wait for
   the CI run to be green.
