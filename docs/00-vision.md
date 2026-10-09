# 00 · Product vision and end state

Read this first. It describes the **finished product after all four phases**, so that every
Phase 1 decision leaves room for what comes later. Phase 1 scope itself is in
`docs/02-phase1-scope.md`.

## 1. What we are building

A web application for an automotive electronics manufacturer (PCBA, IATF 16949) to create,
check and monitor three linked quality documents per product:

- **PFD**: Process Flow Diagram (no special format requirement).
- **PFMEA**: Process FMEA, in two methodologies:
  - **AIAG FMEA 4th edition (2008)**: S, O, D and RPN = S × O × D.
  - **AIAG-VDA FMEA Handbook 1st edition (2019)**: 7 steps, structure tree with 4M work
    elements, and Action Priority (AP) from an S-O-D table.
- **Control Plan (CP)**, in two formats:
  - **Template A**: APQP 2nd edition (2008) Control Plan form. Default partner of PFMEA 4th.
  - **Template B**: AIAG Control Plan 1st edition (CP-1, 2024). Adds the Safe Launch phase,
    a reaction-plan owner and volume-based frequencies. Default partner of AIAG-VDA.

The core idea: **the three documents are three views of one process dataset**. A process step,
characteristic, special-characteristic symbol or control is entered once and appears read-only
everywhere else. What still can go wrong is caught automatically by a rule engine (42 rules),
and a dashboard shows the state of every package.

## 2. Problems it solves

| Today (Excel) | With the system |
| --- | --- |
| PFD, PFMEA and CP are separate files that drift apart. | One dataset; documents cannot disagree on shared data. |
| Inconsistencies are found by hand before audits. | 42 rules run on save, on demand and nightly. Errors block approval. |
| Common processes (receiving, IQC, storage, packing, shipping) are copy-pasted into every model and edited separately. | **Template General**: edit once, release, and every model package is updated automatically. |
| No overview of overdue actions or document health. | Dashboard per customer, model, methodology and PIC. |
| Two methodology versions are mixed by hand. | The methodology is a property of each document; both coexist. |

## 3. End-state capabilities (after Phase 4)

1. **Packages.** One package per part/model holds its PFD, PFMEA and CP. Two package kinds:
   *general* (Template General) and *model*. Document numbers follow a configurable pattern
   (default `SH-` PFD, `SF-` PFMEA, `SC-` CP, then customer code and a sequence).
2. **Template General and Template by Model.** General processes are mandatory or optional.
   Model packages hold linked copies of general rows. Releasing a new template revision updates
   all linked packages: unchanged rows follow the template, local overrides are kept and
   flagged when they conflict, detached rows are left alone. Every sync is logged per cell and
   can be reversed by restoring an earlier template revision.
3. **Editors.** PFD table plus auto-drawn flow diagram; PFMEA worksheet (flat AIAG 4th grid, or
   AIAG-VDA structure tree plus worksheet); CP editor with columns that come from PFD/PFMEA shown
   read-only.
4. **Consistency engine.** 42 rules in groups K (cross-document), S (special characteristics),
   R (risk and controls), F (FMEA content), C (Control Plan), W (workflow and rework) and
   T (Template General). Levels: Error blocks approval, Warning needs a justification,
   Info is advice. Customers can override levels. Catalog: `docs/06-rules.md`.
5. **Workflow.** Draft → In Review → Approved → Released → Superseded. Release is per package,
   so the three documents always move together. Each release stores a frozen snapshot plus the
   Excel/PDF files. When a customer requires approval of changes, a package waits for it.
6. **Actions.** Recommended actions with PIC, target date, status, evidence and re-rating;
   overdue tracking.
7. **Monitoring.** Status cards, package health ranking, risk profile (AP for AIAG-VDA,
   S 9–10 and top RPN for 4th), rule Pareto, overdue actions, template sync state, trends,
   activity feed.
8. **Reports and exchange.** Excel A3 in the company's own layouts, PDF, import of legacy Excel,
   Markdown/JSON export for discussion with AI, traceability matrix, revision diff.
9. **Integration.** Read API for MES/ERP and NG data from production lines (Phase 4).

## 4. Roadmap and gates

| Phase | Name | Scope | Gate to pass |
| --- | --- | --- | --- |
| **1** | Foundation + AIAG 4th | Master data, packages, Template General (release and sync), PFD table + diagram, PFMEA 4th with RPN, CP Template A, 32 core rules, Excel A3 export, basic dashboard, bulk paste from Excel | One real package re-created in the system; its Excel export matches the old Excel |
| 2 | AIAG-VDA + CP-1 | 4M structure, AP, CP Template B, Safe Launch, structured reaction plan with owner, all 42 rules, review and approval, revisions and propagation to released packages, PDF, legacy Excel import | One AIAG-VDA package passes QA review with zero Errors |
| 3 | Full monitoring | Action tracking with evidence, e-mail notifications, KPIs and monthly trends, traceability matrix, revision diff, out-of-sync triggers from PFD/WI changes | All active packages monitored on the dashboard |
| 4 | Scale and reuse | Templates per variant/family, 4th → AIAG-VDA conversion assistant, Markdown/JSON export for AI, NG data integration, drag-and-drop PFD editor | Decided after Phase 3 |

Phase durations are not fixed; they depend on team size and QA test time.

## 5. Users and roles (end state)

| Role | Typical person | Can do |
| --- | --- | --- |
| Admin | IT / system owner | Master data, users, rule settings, everything else |
| Author | Process engineer | Create and edit packages they own or are members of |
| Reviewer | QE, production, maintenance | Review and comment (Phase 2 workflow) |
| Approver | QA Manager | Approve and release packages, release Template General, waive Warnings |
| Viewer | Operator, auditor | Read, export |

## 6. Non-functional requirements (all phases)

- **Hosting:** on-premises factory server (Linux, Docker Compose). No data leaves the site.
  PFMEA and CP content is usually confidential to the customer.
- **Performance targets (server time):** save one cell < 50 ms p95; open a 3,000-row worksheet
  < 300 ms; full check of one package < 2 s; Template General release synced to 24 packages
  < 10 s; Excel export < 3 s; dashboard < 100 ms. Details: `docs/03-architecture.md`.
- **Concurrency:** 10–50 simultaneous users; optimistic locking per row; live updates.
- **Auditability:** every change to content is logged (who, when, old → new) by database
  triggers that application code cannot bypass.
- **Language:** the whole user interface is in English (labels, messages, errors, exports, file
  names). AIAG column headers are exactly as on the AIAG forms; FMEA content stays in the
  language the company writes it in (usually English). Source-code comments are written in
  Bahasa Indonesia so the plant's team can maintain the code.
- **Time zone:** plant time zone setting (default `Asia/Jakarta`, WIB) for dates and "overdue".
- **Browsers:** current Chrome and Edge on factory PCs, which may be low-spec; keep the frontend
  light (no heavy UI kits).
- **Licensed content:** S/O/D criteria text and the AIAG-VDA AP table come from manuals the
  company has licensed. The application stores them as master data entered by the company and
  never ships that text.
- **Backups:** daily dump plus WAL archiving; restore tested before go-live.

## 7. Forward compatibility rules for Phase 1

Phase 1 must not build Phase 2+ features, but it must not block them either:

1. Keep methodology as data (`packages.pfmea_method`, `packages.cp_format`,
   `documents.methodology`). Never assume "4th" in table design; branch in code by methodology.
2. Keep `failure_modes` / `failure_effects` / `failure_causes` / `failure_chains` separate even
   though the 4th grid shows one flat row per chain. AIAG-VDA needs the network.
3. `failure_chains.ap`, `cp_phase = 'safe_launch'`, structured reaction plan columns and the
   workflow statuses already exist in the schema. Phase 1 code must not write them.
4. Revision snapshots (`package_revisions` + `package_snapshot()`) are used in Phase 1 for
   Template General only; Phase 2 reuses them for package releases.
5. The sync engine has a hook for released packages (`onReleasedPackageSynced`) that is a no-op
   in Phase 1 because model packages stay in Draft.
6. The API is versioned (`/api/v1`) and contract-first so a read API for MES/ERP can be added
   without touching the UI contract.

## 8. Never in scope

DFMEA, MSA/SPC calculation tools, document control of work instructions (only references are
stored), customer portals, multi-site tenancy.
