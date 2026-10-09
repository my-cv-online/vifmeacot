# 07 · Template General: create, release, sync, overrides

Template General holds the processes every product goes through (receiving, IQC, storage,
marking, packing, shipping …) with their PFD, PFMEA and Control Plan content. Model packages
contain **linked copies** of those rows. When the approver releases a new template revision,
every linked model package is updated automatically, except where the model deliberately
differs. This is the most delicate part of Phase 1; implement it exactly as written and test
the planner as a pure function.

## 1. Concepts

| Term | Meaning |
| --- | --- |
| General package | The single package with `kind = 'general'` (code `GENERAL`). Its content rows are the **working copy**; they are edited with the normal editors and have `origin = 'local'`. |
| Revision | A frozen `package_snapshot()` of the general package in `package_revisions` (`rev_no` 1, 2, …). `packages.revision` of the general package = latest released revision. |
| Mandatory / optional | `process_steps.general_mode` in the general package. Mandatory steps are copied into every model package (T01 checks); optional ones only when chosen. |
| Linked row | A model row with `origin = 'general'`, `source_id` = id of the template row, `source_rev` = template revision of the last change applied to it. |
| `sync_status` | `in_sync` (equals the template), `override` (some fields differ on purpose, see `overrides`), `detached` (no longer follows the template, `detach_reason` required, T04), `review` (the template deleted the source row but local rows still depend on it, T03). |
| Local child | A local row below a linked row, e.g. a model-specific characteristic 20-03 on general step 20. Allowed. |
| Override policy | `packages.override_policy` of the general package: which columns a model may override, e.g. `{"failure_effects": ["s"], "cp_lines": ["sample_size", "sample_freq"], "reaction_plans": ["text"]}`. Only sync columns that are editable in Phase 1 (`docs/05-api.md` §4) are accepted. Live (not versioned). |
| Package link | `package_links`: which revision a model follows (`synced_rev`) and the sync state (`in_sync`, `pending`, `syncing`, `failed`). |

### Sync columns

Only these columns are copied and synchronised (`*` = reference, remapped from template ids to
model ids through `source_id`). Everything else (`id`, `package_id`, `seq`, link columns,
timestamps, `version`, `general_mode`, the derived `failure_chains.s` / `rpn`, `ap`) never syncs.
`documents` and `actions` are never copied or synced.

| Table | Sync columns |
| --- | --- |
| `process_steps` | `op_no`, `name`, `function`, `symbol`, `kind`, `is_optional`, `machines`, `inputs`, `outputs`, `department`, `wi_ref`, `not_analyzed`, `not_analyzed_reason` |
| `step_flows` | `from_step_id`*, `to_step_id`*, `kind`, `disposition`, `label` |
| `characteristics` | `step_id`*, `char_no`, `kind`, `name`, `spec`, `lsl`, `usl`, `unit`, `sc_symbol_id` |
| `failure_modes` | `step_id`*, `characteristic_id`*, `text` |
| `failure_effects` | `failure_mode_id`*, `level`, `text`, `s` |
| `failure_causes` | `failure_mode_id`*, `text` |
| `failure_chains` | `failure_mode_id`*, `failure_cause_id`*, `o`, `d`, `justification`, `o_evidence` |
| `controls` | `failure_chain_id`*, `kind`, `text`, `control_library_id`, `is_system_control`, `system_control_reason` |
| `cp_lines` | `phase`, `step_id`*, `characteristic_id`*, `control_id`*, `failure_mode_id`*, `machines`, `eval_technique`, `gauge`, `sample_size`, `sample_freq`, `freq_basis`, `control_method`, `is_error_proofing`, `ep_verify_freq` |
| `reaction_plans` | `cp_line_id`*, `text`, `isolation`, `stop_process`, `recovery`, `owner_user_id`, `owner_text`, `instruction_ref` |

Dependency order (parents first): `process_steps`, `characteristics`, `step_flows`,
`failure_modes`, `failure_effects`, `failure_causes`, `failure_chains`, `controls`, `cp_lines`,
`reaction_plans`. Deletes run in reverse order.

Keep this table in Go as data (`backend/internal/template/columns.go`); the planner, the copy
code, the policy validation and the diff all use it.

## 2. Create a model package (`POST /packages`)

In one transaction:

1. Require a released template (`latestRev ≥ 1`), otherwise 422 ("Template General has not
   been released yet"). Phase 1 accepts only `AIAG4` + `APQP2`.
2. Lock the customer row (`SELECT … FOR UPDATE`) and take `doc_seq = max(doc_seq) + 1` for
   that customer.
3. Insert the package (`status = draft`, `revision = 0`, `last_reviewed_at = now()`), the
   owner/members and three documents. Document numbers come from `doc_number_pattern`:
   `{customer}` = customer code, `{seq:03}` = zero-padded `doc_seq` (e.g. `SF-CA-001`). The
   general package uses `<prefix>-<generalSuffix>` (`SH-GEN`, `SF-GEN`, `SC-GEN`). Header
   defaults (keys in `docs/04-data-model.md` §5):
   - PFD: `partNo`, `partName`, `changeLevel`, `customer` (customer name), `preparedBy`
     (owner's name), `date` (today);
   - PFMEA: `item` (part name), `partNo`, `changeLevel`, `fmeaNo` = `doc_no`, `coreTeam`
     (package core team), `preparedBy`, `dateOrig` (today);
   - CP: `phase = production`, `cpNo` = `doc_no`, `partNo`, `partName`, `changeLevel`,
     `supplierPlant` (package plant), `coreTeam`, `dateOrig` (today).
   When `fmeaNo` or `cpNo` is missing (as in the demo data), screens and exports show `doc_no`.
4. Load the latest revision snapshot. Selected steps = all `mandatory` steps + the
   `optionalStepIds` (422 if an id is not an optional step of that revision).
5. Copy, in dependency order, every snapshot row that belongs to the selected steps: new ids
   (UUIDv7 generated in Go), references remapped, `origin = 'general'`, `source_id` = template
   row id, `source_rev` = revision, `sync_status = 'in_sync'`, `overrides = {}`, `seq` copied.
   A flow whose `to_step_id` points to a step that is not selected is not copied.
6. Insert `package_links (synced_rev = revision, state = 'in_sync')`.
7. Commit, then run a full check (`trigger = sync`).

`POST /packages/{id}/general-steps` reuses step 5 for steps added later (optional ones, or a
mandatory one reported by T01), copying from the **synced** revision. The new steps get a `seq`
between their neighbours by numeric `op_no` (midpoint of the gap; renumber the package's steps
by 10 only when there is no gap). 422 `duplicate` when a local step already uses that `op_no`.

## 3. Release (`POST /templates/general/release`, approver or admin)

1. **Before** opening the release transaction, run a full check of the general package and
   wait for it. (The check writes findings on other connections; running it inside the
   release transaction while it holds the package lock would wait forever.)
2. Begin the transaction:
   `SELECT content_version, revision FROM packages WHERE id = $general FOR NO KEY UPDATE`.
3. 409 `version_conflict` if `contentVersion` in the request differs from `content_version`
   (someone edited after the approver looked at the diff). Because `content_version` only
   grows, an equal value also proves that the check of step 1 saw exactly this content.
4. 409 `release_blocked` (`reasons[]` says why) when the general package has open Errors in
   `findings`, when it has no process step, or when `latestRev ≥ 1` and the working copy
   equals the latest revision (§3.1).
5. `rev = revision + 1`; insert `package_revisions (rev_no = rev, snapshot = package_snapshot(id), content_version, change_note, created_by)`;
   update `packages.revision = rev`.
6. Insert `sync_runs (from_rev = rev - 1, to_rev = rev, status = queued, packages_total = number of linked packages)`.
   With zero linked packages the run is created as `done`.
7. For every linked package: set `package_links.state = 'pending'` and insert a River job
   `template_sync {syncRunId, packageId}` with `InsertTx` (same transaction).
8. `pg_notify` `sync.progress`; commit; respond 202 `{revNo, syncRunId}`.

### 3.1 What counts as a change

The diff (`GET /templates/general/diff`, the release guard and the "Berubah" chips) compares
two snapshots table by table by row id, over the sync columns of §1 plus
`process_steps.general_mode` and `seq` of every table; `documents` and `actions` are ignored.
A change of only `general_mode` or `seq` is releasable: it matters for new packages even though
the sync ignores it. Each diff row carries the `stepId` it belongs to so the UI can group
changes per process. `hasUnreleasedChanges` in `GET /templates/general` is a cheap
approximation: `packages.content_version >` the latest revision's `content_version`, or `true`
when no revision exists yet. It can be `true` while the diff is empty (an edit that was undone,
a header-only edit, a restore of the latest revision); the release then answers
`release_blocked` and the UI shows the reason.

## 4. Sync one model package (River job `template_sync`)

The job is idempotent and safe to retry (River retries up to 3 times).

```text
BEGIN
  SET LOCAL app.source = 'sync'; app.user_id = <run.created_by>; app.request_id = 'sync:<runId>'
  link = SELECT … FROM package_links WHERE model_package_id = $pkg FOR UPDATE
  if link.synced_rev >= run.to_rev: mark this package done in the run; COMMIT; return   -- newer run already applied
  SELECT 1 FROM packages WHERE id = $pkg FOR NO KEY UPDATE                            -- same lock order as API writes
  set link.state = 'syncing'
  old = snapshot(rev = link.synced_rev); new = snapshot(rev = run.to_rev)             -- may skip revisions
  model = all rows of the package (linked and local) for the 10 content tables
  plan = PlanSync(old, new, model, policy)                                            -- pure Go, no I/O
  apply(plan): inserts (parents first) → updates → review marks → deletes (children first)
  insert sync_changes rows for every planned item
  close older unresolved conflicts on a (row, column) that gets a new conflict, or whose
    template value is back at overrides[c].templateValue                              -- resolved_at = now(), resolution NULL
  link.synced_rev = run.to_rev; link.state = 'in_sync'; link.last_sync_run_id = run.id
  hooks.OnReleasedPackageSynced(ctx, tx, pkg, run)                                    -- Phase 1: no-op
  UPDATE sync_runs SET packages_done = packages_done + 1, … (finish the run when done + failed = total)
  pg_notify package.reload (reason template_sync) and sync.progress
COMMIT
then run a full check of the package (trigger = sync)
```

On the final failed attempt: in a new transaction set `link.state = 'failed'`, increment
`packages_failed`, finish the run if complete, notify. T02 then shows the package as behind.
`POST /sync-runs/{id}/retry` re-queues the failed packages of that run.

Sync run lifecycle:

| Event | Change to `sync_runs` |
| --- | --- |
| Release | inserted as `queued` (as `done` with `finished_at` when no package is linked) |
| First job of the run starts | `running`, `started_at = now()` (only while still `queued`) |
| A package is applied | `packages_done + 1` |
| A package fails its last attempt | `packages_failed + 1` |
| `packages_done + packages_failed = packages_total` | `done` when `packages_failed = 0`, otherwise `failed`; `finished_at = now()` |
| `POST /sync-runs/{id}/retry` | per failed package: `packages_failed − 1`, link `state = pending`, new job; the run returns to `running` with `finished_at = NULL` |

In `sync_changes`, `row_id` is the model row for `update`, `conflict`, `delete` and skipped
updates; only a skipped insert (no model row exists) stores the template row id.

### 4.1 Planner rules (`PlanSync`)

Notation: template row `S` (by id) in `old` and/or `new`; model copy `M` (the model row with
`source_id = S.id`); for a column `c`: `O = old[S][c]`, `N = new[S][c]`, `L = M[c]`, all compared
as JSON after remapping references to model ids.

**Updated row** (`S` in old and new, `O ≠ N` for some sync column):

| Situation | Result | `sync_changes.action` |
| --- | --- | --- |
| No `M` (not selected optional step, or removed) | nothing | — |
| `M` is `detached` or `review` | nothing | — |
| `c` not overridden in `M` | `M[c] = N` | `update` (old `L`, new `N`) |
| `c` overridden and `L = N` | template caught up: `M[c]` unchanged, remove the override key | `update` |
| `c` overridden and `overrides[c].templateValue = N` | nothing (already acknowledged with "Keep local") | `skip_override` |
| `c` overridden otherwise | keep `L` | `conflict` (old `L`, new `N`) → T03 |
| `N` is a reference whose target has no copy in the model | column not changed | `skip_override` |
| `N` would break a unique key in the model (`op_no`, `char_no` already used by another row) | column not changed | `conflict` |

After the columns: if anything changed set `M.source_rev = new rev`; `M.sync_status` =
`override` when `overrides` is non-empty, else `in_sync`.

**New row** (`S` only in new):

- `process_steps`: insert only when `general_mode = 'mandatory'`. A new optional step appears
  in `GET /packages/{id}/template-link` → `availableGeneralSteps`.
- Other tables: insert when every required parent has a copy in the model (or is inserted by
  the same plan) and that parent is neither `detached` nor `review`. The optional references
  `cp_lines.control_id` and `cp_lines.failure_mode_id` become NULL when their target has no
  copy; a flow whose `to_step_id` target has no copy is skipped.
- Skipped inserts are logged as `skip_override` with `row_id` = template row id and
  `new_value = {"label": …}`; an `op_no` / `char_no` collision with a local row is also skipped
  (T01 then reports a missing mandatory step).
- Inserted rows: values from `new`, `origin = general`, `source_rev = new rev`,
  `sync_status = in_sync`; `seq` = after the last sibling (steps: by numeric `op_no`, as in §2).
  Logged as `insert` with `new_value = {"label": …}`.

**Deleted row** (`S` only in old) with copy `M`:

- `M` is `detached` → nothing.
- `M` has overrides, or rows outside the deletion set reference it (local children, local CP
  lines, detached rows) → `M.sync_status = 'review'`; logged as `conflict` with
  `column_name = NULL`, `old_value = {"label": …}`, `new_value = NULL` → T03 row conflict.
- otherwise → delete `M`; logged as `delete` with `old_value = {"label": …}`.
- The deletion set is computed to a fixpoint: a row is deletable when all rows referencing it
  are deletable.

Labels are human-readable: step `"20 IQC"`, characteristic `"20-02 Part number vs AVL"`, text
rows by their text (max 80 characters), CP line `"<char_no> <phase>"`.

## 5. Working with linked rows in a model package

| Action | Endpoint | Effect |
| --- | --- | --- |
| Override a field | `/changes` with `overrideReason` | Allowed for policy fields only (`docs/05-api.md` §5). |
| Revert overrides | `/linked-rows/revert` `{fields?}` | Restore values from the **synced** revision snapshot, drop the override keys, `sync_status` → `in_sync` when none remain. |
| Detach | `/linked-rows/detach` `{reason}` | `sync_status = detached`, row becomes fully editable, future syncs ignore it. Children stay linked. |
| Re-attach | `/linked-rows/revert` on a detached row | Copy all sync columns from the synced revision, `sync_status = in_sync`, clear reason and overrides. 422 when the template row no longer exists. |
| Resolve a column conflict | `/sync-conflicts/{id}/resolve` | `follow_template`: set the column to the new template value, remove the override key. `keep_local`: keep the value and set `overrides[c].templateValue` to the new template value (acknowledged). Both mark the conflict resolved; afterwards `sync_status` is `override` while overrides remain, otherwise `in_sync`. |
| Resolve a row conflict (`review`) | same | `keep_local`: detach with reason "Deleted from Template General rev N; kept locally" (or the given reason). `follow_template`: delete the row and everything that depends on it; `dryRun` lists them first. |
| Remove an optional general step | `DELETE …/steps/{id}` | Allowed for optional general steps without local rows below them. |
| Add general steps | `/general-steps` | §2. |

All of these write audit rows, publish realtime events and trigger an incremental check (T03 and
T04 update immediately).

## 6. Restore an older revision (rollback)

`POST /templates/general/restore {revNo, contentVersion}` (approver/admin) makes the working
copy equal to revision `revNo` by applying the difference, in one transaction:

1. Lock the general package row (`FOR NO KEY UPDATE`); 409 when `contentVersion` is stale.
2. For the 10 sync tables only (`documents` and `actions` of the general package are not
   touched), compare working rows and snapshot rows by id:
   - rows not in the snapshot → delete (children first);
   - rows only in the snapshot → insert with the same id (parents first; effects before chains
     so the S trigger is right);
   - rows in both → update the columns that differ.
   Restored columns: the sync columns, `seq` and `general_mode`. Inserts take `id` and
   `package_id` from the snapshot; updates never change them. `version`, `created_at`,
   `updated_at`, `s`, `rpn` and `new_rpn` are never written: triggers and defaults keep them
   right, so versions only grow and editors holding old versions get 409.
   The unique checks on `op_no`, `char_no` and step `seq` are deferred to commit, so values may
   swap inside the transaction.
3. Publish `package.reload` (reason `restore`), commit, then run a check.

The working copy now equals the old revision; the approver reviews the diff and releases it as
a **new** revision. Model packages then sync from their revision to the new one, which undoes
the later changes. Nothing is ever deleted from `package_revisions`.

## 7. Phase 2 hook

`OnReleasedPackageSynced(ctx, tx, pkg, run)` is called after a package's sync is applied.
Phase 1: no-op (model packages stay `draft`). Phase 2: for released packages it creates the
next draft revision, and when the customer requires change approval it sets
`package_links.state = 'awaiting_customer'`. Keep the signature; do not implement Phase 2.

## 8. Edge cases (each needs a test)

1. Release twice quickly (rev 2 and 3): the rev 3 job may diff 1 → 3; the rev 2 job then finds
   `synced_rev ≥ 2` and only counts as done.
2. Two jobs for the same package run concurrently: the `FOR UPDATE` on `package_links`
   serialises them.
3. A user edits a linked row while the sync runs: normal optimistic locking; one of them gets
   409 and retries.
4. Template changes an overridden column (conflict, T03 open) and the next revision changes it
   back to the override's `templateValue`: the open conflict is closed and no new one is
   created.
5. Template deletes step 20 while a model package has a local characteristic 20-03 on it
   (built in the test; the seed has none): step 20 goes to `review`, its linked children that
   the template also deleted are removed, 20-03 stays.
6. Template renumbers a step to an `op_no` that a local step uses: column conflict.
7. Package deleted while its job is queued: the job finds no link and finishes as done.
8. Restore rev 1 after rev 2 and 3, release as rev 4: a model on rev 3 ends with rev 1 content
   except its overrides and detached rows.

## 9. Tests

- **Planner unit tests** (table-driven, no database): one case per row of the tables in §4.1
  plus the edge cases in §8, using small hand-built snapshots.
- **Integration test with the seed** (US-10, US-11): on the demo data, change in GENERAL the
  step 20 detection control "Visual check of label vs PO" to "Barcode scan of reel label vs PO
  and AVL" with library item "Barcode scan vs reference", its chain's D from 6 to 4, and CP line
  20-02 `sample_freq` to `Every reel`; release rev 2; run the job; assert that PS-07 has the new
  control text, library item and D (and no new R04), keeps `Every 3 reels`, has exactly one open
  T03 (`cpLine.sampleFreq`), and that `sync_runs` shows 1 of 1 done. Then resolve with
  `follow_template` and assert that T03 is fixed, the override is gone and the line is
  `in_sync`.
- **Create-from-template test** (US-01): package for part MB-01 with the optional step 65 has
  linked steps 10, 20, 25, 65, 100, 110 and documents `SH-CA-001`, `SF-CA-001`, `SC-CA-001`.
