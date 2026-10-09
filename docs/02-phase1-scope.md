# 02 · Phase 1 scope: Foundation + AIAG 4th

**Goal:** the team can build and maintain packages that use PFMEA AIAG 4th edition and Control
Plan Template A (APQP 2nd ed), with Template General synced into every model package, 32 core
consistency rules, Excel A3 export and a basic dashboard.

**Gate 1 (Phase 1 is done when all are true):**

1. One real package chosen by QA is re-created in the system (bulk paste allowed).
2. Its Excel exports (PFD, PFMEA, CP) contain the same content as the old Excel files and QA
   accepts the layout.
3. All 32 Phase 1 rules give correct results on that package (QA reviews every finding).
4. Template General with the company's real general processes is released and synced.
5. Performance targets in `docs/03-architecture.md` §9 are met on the factory server with a
   generated 3,000-chain package.
6. A backup has been restored successfully on a second machine.

## 1. In scope

Each capability has an ID used by milestones (`docs/10-milestones.md`) and tests.

| ID | Capability | Acceptance criteria (summary) |
| --- | --- | --- |
| P1-01 | Login, sessions, roles | Local accounts with argon2id passwords; sessions in PostgreSQL; roles admin, approver, reviewer, author, viewer; permission matrix §3 enforced in the API. LDAP/AD login behind a config flag (milestone M13). |
| P1-02 | Master data | CRUD for customers (code, CSR settings: RPN threshold, CC minimum S, symbol conversion table), special-characteristic classes, parts, control library (kind, method category, D range, error-proofing, strong), S/O/D criteria text for AIAG 4th, banned terms, rule enable/level per customer, users, settings. Admin only. |
| P1-03 | Packages | List with filters; create a model package from the released Template General (all mandatory processes, chosen optional ones); automatic document numbers; document headers; owner and members. |
| P1-04 | PFD editor | Step table (op no, symbol, name, function, machines, inputs/outputs, department, WI ref, optional, rework), characteristics per step (product/process, spec, special class), NG/rework/scrap/return flows, auto-drawn diagram, reorder, delete protection with list of blocking rows. |
| P1-05 | PFMEA AIAG 4th worksheet | One row per failure chain grouped by step; effects with level and S; S = max effect S; RPN = S × O × D computed by the server; prevention/detection controls (optionally linked to the control library); recommended actions with PIC, target, status and re-rating; justification; sort by step / S / S×O / RPN; S 9–10 highlighted; RPN threshold shown only when the customer sets one. |
| P1-06 | Control Plan Template A | Lines per phase (Prototype, Pre-launch, Production); step, characteristic, special class and spec shown read-only from PFD; evaluation technique, gauge, sample size, frequency (+ basis), control method, error-proofing (+ verification frequency), reaction plan text; "+ Lines from PFMEA controls" creates lines from selected controls. |
| P1-07 | Concurrent editing | Every row has a version; stale writes return 409 with the current value; other users' changes appear live; presence shows who is in the package. |
| P1-08 | Consistency engine | 32 rules (`docs/06-rules.md`); run after every save (only affected rules), on "Run check", after a sync and nightly; findings with level, message, deep link to the cell; Open → Fixed automatically; Warning/Info can be waived by approver/admin with a reason; Errors cannot be waived. |
| P1-09 | Template General | Edit the working copy with the same editors; mark processes mandatory/optional; override policy; diff working copy vs last release; impact preview; release with change note; background sync to all model packages with progress; overrides, detach and conflict resolution in model packages; restore an older revision; revision history; list of linked packages. |
| P1-10 | Excel A3 export | PFD, PFMEA 4th and CP Template A (per phase) as .xlsx in the company's layout templates; A3 landscape; merged cells per step / failure mode; cached per content version. |
| P1-11 | Dashboard and search | KPI cards, package health, 4th risk profile (S 9–10, top RPN), rule Pareto, overdue actions, template sync state, activity feed; global search for packages, steps and characteristics. |
| P1-12 | Bulk paste from Excel | Paste a block copied from Excel (tab-separated) into a preview dialog for PFD steps, PFMEA rows or CP lines; map columns; validate; create rows in one transaction. |
| P1-13 | Operations | Docker Compose deployment, configuration by env vars, health and metrics endpoints, backup and restore scripts, performance test suite. |

## 2. Out of scope (and where it goes)

| Not in Phase 1 | Phase |
| --- | --- |
| AIAG-VDA PFMEA (structure tree, 4M work elements, AP, rules F04–F06) | 2 |
| CP Template B (CP-1), Safe Launch, structured reaction plan with owner (C02–C06) | 2 |
| Review/approval workflow, package release, package revisions, revision bump and customer-approval wait after template sync | 2 |
| PDF export (Gotenberg), import of legacy Excel files (K04) | 2 |
| Action evidence files, e-mail notifications, KPI trends, traceability matrix, revision diff for model packages, PFD/WI change triggers (W03) | 3 |
| Templates per family/variant, 4th → AIAG-VDA conversion, AI export, NG data, drag-and-drop PFD | 4 |

Hide or disable UI elements of later phases that appear in the mockups (for example
"Convert to AIAG-VDA", "Submit for review", Template B, Safe Launch). Never stub them with fake data.

## 3. Roles and permissions in Phase 1

| Capability | viewer | reviewer | author | approver | admin |
| --- | :-: | :-: | :-: | :-: | :-: |
| Read packages, worksheets, findings, dashboard; export Excel | ✓ | ✓ | ✓ | ✓ | ✓ |
| Create a model package | | | ✓ | ✓ | ✓ |
| Edit model package content | | | owner or member with edit | ✓ | ✓ |
| Override / detach general rows, resolve sync conflicts in a package | | | owner or member with edit | ✓ | ✓ |
| Edit Template General working copy | | | owner or member with edit of the general package | ✓ | ✓ |
| Release Template General, restore an older revision | | | | ✓ | ✓ |
| Waive Warning/Info findings | | | | ✓ | ✓ |
| Master data, rules, users, settings | | | | | ✓ |

Reviewer behaves like viewer in Phase 1; the review workflow arrives in Phase 2.

## 4. User stories

Format: Given / When / Then. The demo data in `db/seed/demo.sql` is the starting point.

**US-01 Create a model package.** Given Template General rev 1 is released, when an author
creates package `MB-01` for Customer A with part `MB-01` and selects the optional process
"Laser marking + traceability", then the package contains steps 10, 20, 25, 65, 100 and 110 as
linked copies (label GENERAL), documents `SH-CA-001`, `SF-CA-001`, `SC-CA-001` exist, the
package link shows rev 1 in sync, and a check run has executed.

**US-02 Edit the PFD.** When an author adds step 30 "Solder paste printing" between 25 and 50
and adds characteristic 30-01 "Solder paste volume" (SC), then the diagram redraws without a
page reload and other users viewing the package see the new step within 1 s.

**US-03 Delete protection.** When an author deletes step 50 that has PFMEA rows, then the API
returns 409 `delete_blocked` listing the failure modes and CP lines, and nothing is deleted.

**US-04 PFMEA row.** When an author adds a failure chain with effect S 7, O 4, D 4, then the
grid shows S 7 and RPN 112 computed by the server, and changing the effect S to 8 updates S and
RPN on every chain of that failure mode.

**US-05 Bulk paste.** When an author pastes 40 rows copied from the old Excel PFMEA into the
paste dialog and confirms, then steps, characteristics, failure modes, effects, causes, chains
and controls are created in one transaction, or nothing is created and the dialog lists the
invalid rows.

**US-06 CP from PFMEA.** Given PS-07, when an author opens "+ Lines from PFMEA controls" for
phase Production, then controls that already have a line in that phase (e.g. "ICT 100%" on
60-02) are shown disabled; after selecting "AOI 100%" (step 60, 60-03) and the second "Daily
golden sample check" (step 70, 70-01) and confirming, two CP lines are created with step,
characteristic, special class and specification from the PFD and links to the control and
failure mode, and the R01 finding for "AOI 100%" becomes Fixed. Through the API, a control that
already has a line is returned in `skipped` and creates nothing.

**US-07 Concurrent edit.** Given two users open the same chain (version 3), when user A saves
D = 5 and user B then saves D = 6, then B receives 409 `version_conflict` with A's value and
the grid shows a conflict prompt; no value is silently overwritten.

**US-08 Fix a finding.** Given finding R04 on step 90, when the author changes D from 3 to 6,
then after the save the finding is Fixed without pressing "Run check".

**US-09 Waive.** When an approver waives Warning R02 with the reason "Oven alarm connected to
the andon", then the finding shows Waived with name and date; an author cannot waive; an Error
cannot be waived by anyone.

**US-10 Release Template General.** Given PS-07 follows rev 1 and overrides the sample frequency
of CP line 20-02 ("Every 3 reels"), when the approver changes in GENERAL the step 20 detection
control "Visual check of label vs PO" to "Barcode scan of reel label vs PO and AVL" with library
method "Barcode scan vs reference", its D from 6 to 4, and the sample frequency of CP line 20-02
from "Every lot" to "Every reel", and releases rev 2, then PS-07 receives the new control text,
library method and D, keeps "Every 3 reels", shows one T03 conflict for that cell, and the sync
run shows 1 of 1 package done.

**US-11 Resolve a conflict.** When the author chooses "Follow template" on that T03 conflict, the
cell takes the template value, the override is removed and T03 becomes Fixed. "Keep local"
keeps the local value and records the new template value as acknowledged.

**US-12 Export.** When a user exports the PFMEA of PS-07, an .xlsx in the company's A3 layout
downloads in under 3 s; exporting again without changes returns the cached file instantly.

**US-13 Dashboard.** The dashboard numbers equal the open findings and overdue actions shown on
the findings and package screens for the same filters.

## 5. Inputs the company must provide

| Input | Needed by milestone |
| --- | --- |
| Customer list, symbol conversion tables, CSR thresholds (RPN, CC minimum S) | M3 |
| S/O/D criteria text for AIAG 4th from the licensed manual | M3 |
| Control library with D ranges per detection method | M8 |
| List of general processes (mandatory/optional) and their PFD/PFMEA/CP content | M9 |
| Current Excel forms for PFD (SH), PFMEA (SF) and CP (SC) | M10 |
| One real package (Excel) for Gate 1 | M12 |
| Server, network, TLS certificate, Active Directory details if used | M13 |

Until real inputs arrive, build and test with `db/seed/demo.sql` and the default export layout.

Decision (2026-10-09): the project uses **fictitious dummy data** instead of these inputs. Each
milestone that needs an input designs its dummy version (in its own words, never copied from the
licensed AIAG manuals) and documents it; the repository is public and must never receive real
company data. Gate 1 item 1 uses a dummy package until the owner names a real one.

## 6. Open decisions (owner: project team)

- [ ] Use Active Directory login in Phase 1 or local accounts only.
- [ ] Document number pattern (default `SF-{customer}-{seq:03}`).
- [ ] Owner of Template General and who approves template releases.
- [ ] Which real package is used for Gate 1.
- [ ] Server specification and backup target.
