---
name: milestone
description: Run one Phase 1 milestone (M0–M13) from docs/10-milestones.md end to end - test cases first, implement, verify, review, tick, commit and push to main.
argument-hint: M<n>
disable-model-invocation: true
---

# Run milestone $ARGUMENTS

1. `git status -sb` must be clean and `main` up to date (`git pull --rebase --autostash`). The
   only exception is unfinished, uncommitted work of **$ARGUMENTS** left by an interrupted
   session: continue from it. Open `docs/10-milestones.md` and find **$ARGUMENTS**. If it does
   not exist, or an earlier milestone is not ticked, stop and tell the user. If `origin` is not
   set up or `docs/build-guide.md` is missing, do Langkah 1–3 of
   `docs/prompts/start-project.md` first.
2. Read every document in its "Read first" list and the parts of `api/openapi.yaml` and
   `db/migrations/00001_init.sql` it touches. Re-read "Language policy" and "Rules you must
   follow" in `CLAUDE.md`.
3. Plan before editing (plan mode): the **test cases** (ID `TC-M<nn>-<nnn>`, level, short
   scenario, expected result, requirement or user story they prove), files to create or change,
   endpoints (by `operationId`), and anything the spec leaves open. Ask the user only about real
   gaps or contradictions.
4. **Test cases first**: write `docs/test-cases/M<nn>-<slug>.md` in Indonesian using
   `docs/test-cases/_template.md`, commit `test(M<n>): define test cases` and push to `main`.
5. Write the automated tests for every case (names contain the case ID, Indonesian comment above
   each test). Run them and confirm they fail for the expected reason. Do not commit them on
   their own; they are committed together with the code that makes them pass.
6. Implement in small steps until the tests pass. Comments in Indonesian (file header, every
   function, type and test, non-obvious steps); every visible text in
   English (`web/src/lib/i18n/en.ts`, `backend/internal/i18n/en.go`). Commit and push each step
   that leaves `make check` green. Never implement a later milestone's feature; use an
   interface with a no-op and `TODO(M<k>)` when a dependency is needed.
7. Verify: every command under "Done when", then `make check`. Fix until green; never weaken or
   delete tests. Update the status column in the test-case file.
8. Run the `spec-reviewer` agent on the milestone's changes (it diffs from the
   `test(M<n>): define test cases` commit, so pushed work is included). Fix every blocker and
   major finding; re-run the checks.
9. Tick **$ARGUMENTS** in `docs/10-milestones.md`, update the stage table in
   `docs/build-guide.md` (and its installation section when a setup command changed), note
   any spec decision in the relevant doc, commit
   `feat($ARGUMENTS): <short summary>` and push to `main`. Wait for the GitHub Actions run of
   that commit:
   - id: `gh run list --commit "$(git rev-parse HEAD)" --limit 1 --json databaseId,status,conclusion`
     (retry a few seconds later until it appears);
   - wait: `gh run watch <id> --exit-status`, with a 10-minute timeout or in the background;
   - if it fails: read `gh run view <id> --log-failed`, fix, commit and push again.
10. Report to the user in Bahasa Indonesia: what was built, the test cases and their results,
    the CI run, how to try it locally, open questions, and what the next milestone needs from
    the company (`docs/02-phase1-scope.md` §5). If the session had to stop early, push what is
    green, leave the unfinished work uncommitted and list it in the report.
