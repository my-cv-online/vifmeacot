# Test cases

Every change starts here. Before any feature code is written or changed (new feature, update or
bug fix), its test cases are written in this folder, committed and pushed to `main`. Then the
automated tests are written and seen failing (they are committed and pushed together with the
code that makes them pass, never on their own), then the code. This rule comes from the project
owner and is repeated in `CLAUDE.md` and `docs/11-testing.md`.

## Files and IDs

| Work | File | ID pattern |
| --- | --- | --- |
| A milestone | `M<nn>-<slug>.md`, e.g. `M05-pfd-editor.md` | `TC-M05-001`, `TC-M05-002`, … |
| A later change to a feature | the file of the milestone that owns the feature (continue its numbering) | `TC-M05-013`, … |
| A bug fix | `bugs.md` | `TC-BUG-001`, … |
| A consistency rule | `rules.md` (one or more cases per rule; the fixture in `backend/internal/rules/testdata/` is the automated test) | `TC-RULE-K01-1`, … |

IDs are never reused or renumbered. A case that no longer applies is marked "dihapus" with the
reason; it is not deleted.

## Language

The content is written in **Bahasa Indonesia** so the team can review it. IDs, file names,
code identifiers and the English UI labels quoted in the steps stay as they are.

## Content of a test case

Use `_template.md`. Each case has:

- **ID and title** (one line, what is proven);
- **Level**: unit, integrasi (database or API), E2E (browser), performa, or manual (only for
  checks that cannot be automated, e.g. the printed A3 layout at Gate 1);
- **Requirement**: user story (`US-xx`), capability (`P1-xx`), rule code or doc section;
- **Preconditions**: data (usually `db/seed/demo.sql`), user and role, starting screen;
- **Steps**: numbered, concrete (values to type, buttons with their English labels);
- **Expected result**: observable and checkable (values, status codes, messages, counts,
  timing limits);
- **Automated test**: file and test name, which contains the ID
  (`TestCreatePackage_TC_M04_001`, `test('TC-M04-001 …')`);
- **Status**: `belum dibuat` → `gagal` (test written, code not yet) → `lulus`.

Every case must have an automated test except level manual. Every user story of
`docs/02-phase1-scope.md` §4 is covered by at least one E2E case in its milestone.
