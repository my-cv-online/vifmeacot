# 08 · Screens (Phase 1)

The static mockups in `docs/mockups/` show the **end state** (they include AIAG-VDA, CP-1,
review workflow and other later-phase elements) and were drawn with Indonesian labels. Use them
for layout, density and colours only. This document is authoritative for Phase 1: what to build,
what to hide, how each screen behaves and the **English** text it shows. All UI text is English
and lives in `web/src/lib/i18n/en.ts`; AIAG form column headers are written exactly as on the
forms. §14 maps the main mockup labels to their English labels.

## 1. App shell

- **Top bar** (`#14181D`): logo + "PFD · PFMEA · CP System" (link to `/`), global search
  "Search packages, steps, characteristics" (`GET /search`, results grouped Packages / Steps /
  Characteristics, Enter opens the first hit), badge "DEMO DATA" only when `DEV_MODE` (the UI reads `devMode` from `getMe`), user menu
  (name, role, "Change password", "Log out"). The notification bell is Phase 3: do not render it.
- **Sidebar**: MONITORING (Dashboard `/`, Consistency check `/findings`), PACKAGES (All packages
  `/packages`), TEMPLATE (Template General `/template-general`), then one group per recently
  opened package (max 2, from local storage) titled `PACKAGE <code> · AIAG 4TH` with PFD, PFMEA,
  Control Plan, Findings. Admins also see MASTER DATA (Customers, Parts, Characteristic classes,
  Control library, S/O/D criteria, Banned terms, Rules, Users, Settings).
- **Breadcrumb** on every page (`Dashboard / Package PS-07 Power Supply Board / PFMEA`).
- **Connection banner** when the WebSocket is down: "Live updates disconnected. Reconnecting…".
  Saving still works over HTTP.
- Unsaved state never exists: every edit is sent immediately (one request per gesture).
- Dates are shown as `08 Oct 2026` (and `08 Oct 2026, 07:45` with time) in the plant time zone.

### Design tokens (from the mockups)

| Token | Value |
| --- | --- |
| Page background / surface / subtle surface | `#EEF0F3` / `#FFFFFF` / `#F7F8FA` |
| Borders: row / control / strong | `#E6E9EE` / `#D5DAE1` / `#C3CAD3` |
| Text: primary / secondary / muted | `#16191D` / `#4A5361` / `#7D8794` |
| Accent (links, primary buttons) / hover | `#0A57A8` / `#073F7C` |
| Selected navigation | background `#E6EEF8`, text `#0A3F7A` |
| Error chip / bar | `#FCE8E6` + `#8F1D14` / `#B42318` |
| Warning chip / bar | `#FDF0D9` + `#7A4100` / `#C26A00` |
| Info chip | `#E6EEF8` + `#0A3F7A` |
| OK chip | `#F2F9F3` + `#1D4620` |
| GENERAL chip (linked template row) | `#DDF1EE` + `#0B5A50` |
| Read-only (from another document) cell | background `#F7F8FA`, text `#4A5361` |
| Fonts | IBM Plex Sans (UI), IBM Plex Mono (numbers, codes, op no), self-hosted via `@fontsource` |
| Sizes | body 14 px, tables 13.5 px, touch targets ≥ 44 px, focus outline 2 px accent |

### Common behaviours

| Situation | UI |
| --- | --- |
| 409 `version_conflict` | Automatic rebase when the user's field was not changed by others (`docs/05-api.md` §6.3); otherwise the cell shows a conflict popover: "Changed by <name> to <value>." with "Keep my value" / "Use server value". |
| 409 `locked_by_template` | Toast "This field follows Template General. Detach the row from the template or change it in Template General." |
| 422 `override_reason_required` | Dialog "Override reason" (text area, at least 3 characters), then re-send with `overrideReason`. |
| 422 validation | The cell stays in edit mode with the message under it. |
| 409 `delete_blocked` | Dialog "Cannot delete" listing the blocking rows with links. |
| Delete | Confirmation dialog built from `dryRun=true` ("This will delete: 3 causes, 3 chains, 6 controls, 1 action"). |
| Other users' edits | Cells flash briefly; presence avatars at the top of the grid; a cell focused by someone else gets a coloured outline with their initials. |
| `package.reload` | Reload the view, keep scroll position and selection, toast "Template General rev N applied" (reason `template_sync`). |
| Findings on cells | Small coloured corner (error/warning/info) with the message as tooltip; "Open cell" deep links scroll to and focus the cell (`?focus=<entity>:<id>:<field>`). |
| Template lock state | Locked: grey background + lock icon. Overridable: normal cell, editing asks for a reason. Overridden: small "O" badge; tooltip shows reason, user, date and template value; context menu "Revert to template". Rows show a GENERAL chip on their first column. |
| Read-only users (viewer, reviewer) | Same screens, no edit affordances, no create/delete buttons. |

## 2. Sign in `/login`

Username, password, "Sign in". Errors: "Wrong username or password." (401), "Too many attempts.
Try again in N minutes." (429). After sign-in go to the page the user wanted, else `/`.

## 3. Dashboard `/` (mockup `Main`)

Data: `GET /dashboard?customerId=&ownerId=` (one request, < 100 ms).

- Filters: Customer, Owner. Hide Model/part, Methodology and Period (Phase 3).
- KPI cards: "Active packages"; "Packages without errors" (percent and "x of y packages"); "Open
  errors" ("in n packages"); "Open warnings" ("in n packages"); "Overdue actions" ("of n open
  actions"); "Template sync" ("in sync/total · Rev N"; failed count in red). Do not show "Out of
  sync PFD/WI" (Phase 3).
- "Package health" table (worst first): Package (link), Customer, Errors, Warnings, Overdue
  actions, "PFD steps covered by CP" (bar `stepsInCp / stepsTotal`), Last check.
- "Risk profile (AIAG 4th)" replaces "Action Priority profile": chains with S 9–10 and the top 5
  RPN rows (package · step · failure mode · S·O·D · RPN, link to the chain).
- "Most violated rules": top 5 rules, bar coloured by level, link to `/findings?rule=`.
- "Overdue actions": action, package · step, PIC, target date, days late (link to the chain).
- "Recent activity": the latest 15 activity items.
- Hide "Export summary" and "Run checks now" (no cross-package check button in Phase 1).

## 4. Packages `/packages`

Table (`GET /packages`, keyset pagination): Code, Name, Customer, Part / change level,
Methodology ("AIAG 4th · CP A"), Status (Draft), Template ("Rev 1 · in sync" / "pending" /
"failed"), Errors, Warnings, Overdue actions, Updated. Filters: customer, owner, text, health.

"+ New package" (author, approver, admin) opens a dialog:

1. Code, Name, Customer, Part (filtered by customer; shows its change level).
2. Methodology: "PFMEA AIAG 4th + Control Plan Template A (APQP 2nd ed)". AIAG-VDA and CP-1 are
   visible but disabled with the label "Phase 2".
3. "Optional general processes": checkboxes from the latest Template General revision; the
   mandatory ones are listed as fixed.
4. Owner (default: me), members, plant, core team.
5. "Create package" → `POST /packages` → navigate to `/packages/:id/pfd`.

## 5. Package overview `/packages/:id`

Header: code, name, status chip, customer, part / change level, owner. Cards:

- **Documents**: PFD, PFMEA, Control Plan with document numbers; "Open" and "Export Excel A3"
  (CP asks for the phase).
- **Findings**: counts by level, link to `/packages/:id/findings`, "Run check".
- **Template General**: "Follows Rev N" vs the latest revision, state; open conflicts with
  "Follow template" / "Keep local" (column conflicts) or "Detach from template" / "Follow
  template (delete)" (row conflicts, with dry-run confirmation); rows in review; detached rows
  with reasons; overridden rows; "Add general process" listing `availableGeneralSteps`.
- **Members** (owner, approver, admin can edit), **Settings** (name, plant, core team, owner),
  "Mark as reviewed" (clears W04), "Delete package" (admin, type the code to confirm).

## 6. PFD `/packages/:id/pfd` (mockup `Pfd`)

Data: `GET /views/pfd`. Layout: step table on the left (≈ 60 %), diagram on the right.

- Header block: document number, part, change level, prepared by, date (editable header keys
  via `/changes` entity `document`).
- Buttons: "+ Add step" (dialog: op no, name, symbol, after which step), "Paste from Excel"
  (bulk paste, §11). The mockup's file import ("Import from Excel", legacy files) is Phase 2.
- Step table (Tabulator): No., Symbol (select with icons), Process (name; chips GENERAL /
  optional / rework), Function, Machines / tools (tag editor), Characteristics (summary
  "product · process", click to expand), Class, Department, WI. Toggle "Show optional
  processes". Row menu: Move up / Move down, Mark as not analysed (reason), Delete, Detach from
  template.
- Expanded row: characteristics sub-table (No., Type, Name, Specification, LSL, USL, Unit,
  Class) with "+ Characteristic"; NG flows of the step (Type NG / Rework / Scrap / Return, To
  step or "Leaves the flow", Disposition, Label) with "+ NG flow".
- Diagram (Svelte Flow, read-only, auto layout by ELK `layered`, direction down, in a Web
  Worker, re-layout debounced 300 ms): one node per step in `seq` order with its symbol, op no
  and name; "G" badge for general rows; dashed side boxes for NG/rework/scrap flows with their
  label; rework steps offset to the side; a red dot on nodes with open errors. Clicking a node
  selects the table row. Legend: Operation, Inspection, Storage, Transport, Rework (optional),
  General process.
- Footer note: "Steps still used by the PFMEA or the Control Plan cannot be deleted; the
  affected rows are listed."

## 7. PFMEA AIAG 4th `/packages/:id/pfmea` (mockup `Pfmea4th`)

Data: `GET /views/pfmea` (normalized; the store joins it into one grid row per chain).

- Header block: Item, Model Year / Program, Core Team, Process Responsibility, Key Date, FMEA
  Number, Prepared By, FMEA Date (Orig. / Rev.) — editable fields of `documents.header`.
- Title chips: "AIAG FMEA 4th ed", "Rev N", "Draft". Buttons: "Export Excel A3", "Run check".
  Hide "Convert to AIAG-VDA" (Phase 4) and "Submit for review" (Phase 2).
- Toolbar: "Sort by" (Process number · Highest S · S×O · RPN), RPN threshold chip ("RPN
  threshold: not used" or "Customer RPN threshold: 120"), toggle "Action Results columns",
  counters ("n rows · S 9–10: n · highest RPN n · n overdue actions"), "+ Row".
- Grid: grouped by step (group header "50 · Pick and place — Function: …"). Columns as the
  AIAG 4th form (`docs/01-domain-glossary.md` §3):
  - Requirements (characteristic, read-only text; click to choose another characteristic of
    the same step),
  - Potential Failure Mode (`failureMode.text`),
  - Potential Effect(s) of Failure (`failureMode.effectText`; with several effects the cell
    shows a list and opens the effect editor: level, text, S),
  - Sev (`failureMode.effectS` when there is one effect, otherwise read-only = chain S),
  - Class (read-only symbol from the characteristic),
  - Potential Cause(s) of Failure (`failureCause.text`),
  - Current Process Controls Prevention (`failureChain.pcText`, list editor when there are
    several; library picker),
  - Occ (`failureChain.o`),
  - Current Process Controls Detection (`failureChain.dcText`, list editor, library picker),
  - Det (`failureChain.d`),
  - RPN (read-only, server value),
  - Recommended Action(s), Responsibility & Target Completion Date, Action Taken & Completion
    Date, S/O/D/RPN results (action editor: one line per action; status select; PIC user
    picker or free text), plus Justification (`failureChain.justification`) in the action
    editor.
  Cells of the failure mode (Requirements, Failure Mode, Effects, Sev, Class) are shown only on
  the first chain of each failure mode, like merged cells in Excel.
- Highlights: S 9–10 cells warning-tinted; RPN ≥ customer threshold bold warning; S/O/D cells
  show the criteria text of `rating_criteria` as tooltip when the company has entered it.
- Context menu: Add cause to this failure mode, Add failure mode to this step, Manage effects,
  Manage controls, Manage actions, Move up / Move down, Delete row, Detach from template,
  Revert to template.
- "+ Row" dialog (`POST /pfmea-rows`): step, characteristic, failure mode (new text or
  existing), effect + S, cause, O, prevention, detection (+ library), D.
- Keyboard: arrows, Tab / Shift+Tab, Enter or F2 to edit, Esc to cancel, Ctrl+C copies cells,
  Ctrl+V into existing rows sends one `/changes` batch; Ctrl+V of new rows suggests the paste
  dialog. No undo in Phase 1 (the audit log keeps the history).
- Footer: "RPN = Sev × Occ × Det, computed by the server. The Excel export merges cells per
  process as on the AIAG 4th form."

## 8. Control Plan Template A `/packages/:id/control-plan?phase=production` (mockup `ControlPlan`)

Data: `GET /views/control-plan?phase=`.

- Header block (APQP form header, `docs/04-data-model.md` §5) including the phase tick boxes
  and approvals.
- Template selector shows only "A · APQP 2nd ed" (Template B is Phase 2). Phase tabs:
  Prototype, Pre-launch, Production with line counts; hide Safe Launch and its progress banner.
- Buttons: "Run check", "Export Excel A3", "+ Lines from PFMEA controls" (dialog listing all
  PFMEA controls grouped by step with checkboxes; controls that already have a line in this
  phase are disabled), "+ Line" (step, characteristic).
- Grid columns as the form (`docs/01-domain-glossary.md` §4). Grey read-only columns from the
  PFD: Part/Process Number, Process Name/Operation Description, Characteristics No. / Product /
  Process, Special Char. Class, Product/Process Specification/Tolerance (note: "Grey columns
  come from the PFD or the PFMEA; change them there."). Editable: Machine, Device, Jig, Tools
  for Mfg. (tags, suggestions from the step's machines), Evaluation/Measurement Technique,
  Gauge, Sample Size, Sample Freq. (+ basis), Control Method, Error-proofing (checkbox +
  verification frequency), Reaction Plan, and a link column "PFMEA" (linked control / failure
  mode, picker) used to fix R03.
- Hide the CP-1 Owner column and the C04 badges (Phase 2).
- Error/Warning counters in the grid header link to the package findings filtered to CP.

## 9. Findings `/findings` and `/packages/:id/findings` (mockup `Findings`)

- Header: "Consistency check results", last run from `lastCheckRun` ("32 rules run · 08 Oct
  2026, 07:45 · manual by R. Saputri"), package selector (cross-package view: all packages),
  "Run again" (package view; `POST /checks`).
- Count cards: Error, Warning, Info, Waived. Banner (Phase 1 text): "Errors must be fixed and
  cannot be waived. Warnings and infos can be waived by an approver with a reason." Hide the
  "Approval locked" wording (Phase 2).
- Filter chips (All, Error, Warning, Info, Waived), rule select ("Rule"), table: Level, Code,
  Finding (message), Location ("PFMEA · 60" / "CP · 50-01" / "PFD · 25"), Actions ("Open
  cell", "Waive" for approver/admin on Warning/Info, "View" for waived ones). The cross-package
  view adds a Package column.
- Detail panel: level chip, code, rule title, location, message, "What is checked" (rule
  description), history (first seen, last seen, status), waiver (who, when, reason) and quick
  fixes:
  - R01/R02: "Create CP line from this control" (`/cp-lines/from-pfmea`), "Link to an existing
    CP line" (picker → `/changes` `cpLine.controlId`);
  - R02: "Mark as system control" (reason → `control.isSystemControl`);
  - R03: "Link to PFMEA";
  - T03: "Follow template" / "Keep local" (or the row-conflict pair);
  - T04: "Enter reason";
  - W04: "Mark as reviewed".
  Footer: "Errors cannot be waived. Once the data is fixed, the status changes to Fixed
  automatically."
- Live: `findings.changed` updates counts and rows without a reload.

## 10. Template General `/template-general` (mockup `TemplateGeneral`)

Data: `GET /templates/general`, `/diff`, `/impact`, `/revisions`, `/sync-runs/{id}`.

- Header: "Template General", chip "Rev N" (latest), chip "Unreleased changes" when
  `hasUnreleasedChanges` is true, "Active release: Rev N · date", owner, "Used by n model
  packages". Buttons: "Compare with Rev N" (diff dialog grouped by document: Document, Row,
  Field, Rev N, Working copy) and "Release and sync" (approver/admin; disabled while
  `hasUnreleasedChanges` is false; a refused release shows its `reasons`).
- Tabs:
  - **General processes**: list of general steps with "Mandatory" / "Optional" chips (toggle
    for editors) and "Changed" when changed since the last release; selecting one shows its
    changes and links "Edit in PFD / PFMEA / Control Plan", which open the editors of the
    general package (`/packages/<generalId>/pfd` …). "+ Add general process" creates a step in
    the general package.
  - **Revision history**: revision, change note, by, date, sync result; "Restore to working
    copy" (approver/admin; the confirmation explains that a new release is needed).
  - **Linked packages (n)**: linked packages, synced revision, state, open conflicts; "Retry"
    for failed ones.
- "Impact of releasing Rev N+1" (`GET /impact`, recomputed after each save, debounced): cards
  "linked model packages", "will be updated", "conflicting overrides (T03)", "rows to review";
  table "Overrides that will conflict" (Package, Row · field, Local value, Rev N+1, Override
  reason). Read-only in Phase 1: the decision ("Follow template" / "Keep local") is taken in
  each model package after the release (T03). Hide the "Released packages get a new revision"
  card (Phase 2).
- Release dialog: change note (required), impact summary, `release_blocked` reasons when
  refused. After 202: progress panel driven by `sync.progress` ("Synced 18 of 24 packages"),
  failures listed with "Retry".
- "Override policy" (approver/admin): per entity, checkboxes of the fields that models may
  override.

## 11. Bulk paste dialog (PFD, PFMEA, CP)

1. The user copies cells in Excel and presses Ctrl+V in the dialog's paste area (TSV).
2. Column mapping row above the preview: each column gets a key (`PasteColumn`), guessed from a
   header row when present (`Op`, `Process`, `Failure Mode`, `Sev` …); "Ignore" is allowed.
   Checkbox "Fill empty cells from the row above (Excel merged cells)" (on by default).
3. "Check" → `dryRun: true`; the preview shows per-row errors (red) and warnings (amber) and
   the totals ("Will create: 4 steps, 12 characteristics, 30 failure modes …").
4. "Save" (enabled when there are no errors) → `dryRun: false`; then the view reloads.

Allowed columns per target, matching and validation rules: `docs/05-api.md` §8. The dialog
shows only the columns allowed for the screen it was opened from.

## 12. Master data (admin) `/master/*`

Simple tables with a side form; deactivate instead of delete (except banned terms and unused
classes).

| Route | Content |
| --- | --- |
| `/master/customers` | Code, name, RPN threshold (empty = off), CC minimum S (empty = off), notes, active; symbol conversion table editor (class → customer symbol, label). |
| `/master/parts` | Customer, part no, name, change level, family, active. |
| `/master/sc-symbols` | Code, name, critical (S04), description, order. |
| `/master/control-library` | Name, kind, method category, D min–max (detection only), error-proofing, strong, active. |
| `/master/rating-criteria` | Methodology (AIAG 4th only in Phase 1), tabs S / O / D, ten text areas (rating 10 → 1). Note: "Enter the text from the company's licensed AIAG manual." |
| `/master/banned-terms` | Term, applies to, reason, suggestion, active. |
| `/master/rules` | Code, title, level, enabled, phase ("Phase 2" rows disabled), per-customer overrides. |
| `/master/users` | Username, name, department, role, local/LDAP, active, reset password. |
| `/master/settings` | Document number pattern, review interval (months), plant time zone, nightly check time. |

## 13. Phase 2+ elements to hide (never fake)

"Convert to AIAG-VDA", "Submit for review", review/approval statuses other than Draft, CP
Template B and Safe Launch (phase tab and banner), CP Owner column, AP profile and AP columns,
structure tree / 4M work elements, "Import from Excel" (legacy file import), PDF export,
notifications, KPI trends and period filter, "Export summary", PFD/WI out-of-sync KPI,
customer-approval waiting states.

## 14. Mockup labels → English labels

| Mockup (Indonesian) | App (English) |
| --- | --- |
| Sistem PFD · PFMEA · CP | PFD · PFMEA · CP System |
| Cari paket, step, karakteristik | Search packages, steps, characteristics |
| Data contoh | Demo data |
| Cek konsistensi / Hasil cek konsistensi | Consistency check / Consistency check results |
| Paket aktif, Paket bebas Error, Error terbuka, Aksi lewat tenggat, Sinkron template | Active packages, Packages without errors, Open errors, Overdue actions, Template sync |
| Kesehatan paket, Step PFD tercakup CP | Package health, PFD steps covered by CP |
| Aturan paling sering dilanggar, Aktivitas terbaru | Most violated rules, Recent activity |
| Langkah proses, Tampilkan proses opsional, + Tambah step | Process steps, Show optional processes, + Add step |
| Diagram alur, Operasi, Inspeksi, Simpan, Proses general | Flow diagram, Operation, Inspection, Storage, General process |
| Urutkan, Nomor proses, Ambang RPN: tidak dipakai, Kolom Action Results | Sort by, Process number, RPN threshold: not used, Action Results columns |
| Ekspor Excel A3, Cek sekarang, Jalankan ulang | Export Excel A3, Run check, Run again |
| Baris Control Plan, + Baris dari kontrol PFMEA, Fase | Control Plan lines, + Lines from PFMEA controls, Phase |
| Temuan, Kode, Lokasi, Tindakan, Buka sel, Lihat | Findings, Code, Location, Actions, Open cell, View |
| Buat baris CP dari kontrol ini, Tautkan ke baris CP yang sudah ada | Create CP line from this control, Link to an existing CP line |
| Proses general, Riwayat revisi, Paket pengguna, Wajib, Opsional, Berubah | General processes, Revision history, Linked packages, Mandatory, Optional, Changed |
| Bandingkan Rev N, Rilis dan sinkronkan, Dampak jika Rev N dirilis | Compare with Rev N, Release and sync, Impact of releasing Rev N |
| Override yang perlu diputuskan, Ikut template, Tetap lokal | Overrides that will conflict, Follow template, Keep local |
| Lepas dari template, Kembalikan ke template | Detach from template, Revert to template |
| Tahap 2 | Phase 2 |
