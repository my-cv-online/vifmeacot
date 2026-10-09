# 09 · Excel export (A3)

Gate 1 depends on this: the exported PFD, PFMEA and Control Plan must contain the same content
as the company's current Excel files and look acceptable to QA. Library:
`github.com/xuri/excelize/v2`. Code: `backend/internal/export`.

## 1. What is exported

| Document | Rows | File name |
| --- | --- | --- |
| PFD | one per step (`seq` order) | `SH-CB-001 PS-07 PFD 2026-10-08.xlsx` |
| PFMEA AIAG 4th | one per failure chain, expanded to one row per action when a chain has several actions | `SF-CB-001 PS-07 PFMEA 2026-10-08.xlsx` |
| Control Plan Template A | one per CP line of the chosen phase | `SC-CB-001 PS-07 CP Production 2026-10-08.xlsx` |

The date in the name is the export date in plant time. Phase 2 adds AIAG-VDA, CP-1 and PDF.

## 2. Company layouts or the default layout

Layouts live in `EXPORT_TEMPLATE_DIR` as pairs:

```text
PFD.xlsx          + PFD.json
PFMEA-AIAG4.xlsx  + PFMEA-AIAG4.json
CP-APQP2.xlsx     + CP-APQP2.json
```

The `.xlsx` is the company's blank form (logo, header labels, borders, column widths, print
settings). The `.json` says where the data goes:

```json
{
  "sheet": "PFMEA",
  "header": { "item": "C3", "modelYearProgram": "C4", "coreTeam": "C5", "processResponsibility": "J3",
              "keyDate": "J4", "fmeaNo": "Q2", "preparedBy": "Q4", "dateOrig": "Q5", "dateRev": "S5" },
  "table": {
    "firstRow": 10,
    "styleRow": 10,
    "columns": { "stepFunction": "A", "requirements": "B", "failureMode": "C", "effects": "D", "s": "E",
                 "class": "F", "cause": "G", "prevention": "H", "o": "I", "detection": "J", "d": "K",
                 "rpn": "L", "actions": "M", "responsibilityTarget": "N", "actionTakenCompletion": "O",
                 "newS": "P", "newO": "Q", "newD": "R", "newRpn": "S" }
  },
  "merge": ["stepFunction", "requirements", "failureMode", "effects", "s", "class"],
  "printTitleRows": "1:9"
}
```

- When `fmeaNo` or `cpNo` is missing from the header, `doc_no` is printed instead
  (`docs/07-template-general.md` §2); the PS-07 golden files rely on this.
- `header` keys are the document header keys (`docs/04-data-model.md` §5) plus `docNo`,
  `revision`, `customer`, `partNo`, `partName`, `changeLevel`, `phaseMark.prototype` /
  `phaseMark.preLaunch` / `phaseMark.production` (an "X" in the ticked box).
- `table.columns` keys are fixed per document (§3). A column missing from the mapping is not
  exported.
- Every written row copies the cell styles of `styleRow` (borders, font, wrap, alignment).
- The loader validates the JSON at startup and logs which layouts are active; a broken layout
  disables itself and the default layout is used (with a warning in the log and in
  `GET /exports/{id}` as `error` text only when generation fails).

Until the company provides its forms (needed by milestone M10), the **default layout** built in
code is used: same column keys, Arial 9 pt, thin borders, header block in rows 1–7, column
headers in rows 8–9 with the AIAG wording, widths chosen so the sheet fits A3 landscape.

## 3. Column keys and sources

**PFMEA (AIAG 4th)** — sources in `docs/01-domain-glossary.md` §3:
`stepFunction`, `requirements`, `failureMode`, `effects`, `s`, `class`, `cause`, `prevention`,
`o`, `detection`, `d`, `rpn`, `actions`, `responsibilityTarget`, `actionTakenCompletion`,
`newS`, `newO`, `newD`, `newRpn`.

**Control Plan (Template A)** — sources in `docs/01-domain-glossary.md` §4:
`processNo`, `processName`, `machines`, `charNo`, `productChar`, `processChar`, `class`, `spec`,
`evalTechnique`, `sampleSize`, `sampleFreq`, `controlMethod`, `reactionPlan`.

**PFD** — `opNo`, `symbolOperation`, `symbolInspection`, `symbolTransport`, `symbolStorage`,
`symbolDelay`, `symbolDecision` (each gets "●" when the step has that symbol;
`operation_inspection` marks both operation and inspection), `processName`, `function`,
`machines`, `inputs`, `outputs`, `productChars`, `processChars` (one per line, `char_no name`),
`class`, `department`, `wiRef`, `flows` (one per line: "NG → 75 Rework (touch-up): rework"),
`notes` ("optional", "rework").

Formatting rules:

- Multi-value cells are joined with line breaks; wrap text on.
- `class` uses the customer's printed symbol (`customer_sc_symbols.customer_symbol`); without a
  mapping the internal code is printed (S02 already reports it).
- S, O, D, RPN are numbers, centred; empty when NULL.
- Dates print as `dd/mm/yyyy`.
- When a chain has no action but a justification: `actions` = "Justification: <text>".
- Responsible: user display name, or `responsible_text`; then the target date on a new line.

## 4. Merging

- PFMEA: `stepFunction` merges over all rows of the step; `requirements`, `failureMode`,
  `effects`, `s`, `class` over all rows of the failure mode; chain columns (`cause`,
  `prevention`, `o`, `detection`, `d`, `rpn`) over the rows of a chain that has several
  actions. Only columns listed in `merge` (company layout) or the default list are merged.
- CP: `processNo`, `processName` merge over consecutive lines of the same step; `charNo`,
  `productChar`, `processChar`, `class`, `spec` over consecutive lines of the same
  characteristic.
- PFD: no merges.

## 5. Page setup

A3 landscape, fit to 1 page wide (as many pages tall as needed), narrow margins (1 cm, header
and footer 0.5 cm), repeat the title rows (`printTitleRows`, default rows 1–9) on every page,
footer left `<docNo> rev <revision>`, centre the document title, right "Page &P / &N". Freeze
panes below the column headers.

## 6. Generation, cache and delivery

1. `POST /packages/{id}/exports {docType, phase}`: look up `export_files` for (package,
   docType, phase, format, current `content_version`). A `done` row whose file exists → 200.
   A `queued`/`running` row → 202 with it. Otherwise insert a `queued` row and a River job
   `export_file` in the same transaction → 202.
2. The job reads the view data (same SQL as the screens, `REPEATABLE READ`), writes the file to
   `EXPORT_DIR/<packageId>/<exportId>.xlsx.tmp`, fsyncs, renames to `.xlsx`, sets `done`,
   `file_path`, `file_size`, and notifies `export.ready` to the requesting user. Errors set
   `failed` with the message.
3. `GET /exports/{id}/file` streams the file with `Content-Disposition: attachment;
   filename*=UTF-8''<name>`.
4. The cache is invalidated by content changes (`content_version`) and cleared explicitly when
   an admin changes customers, symbol tables, classes or user names (rows deleted, files
   removed). The daily cleanup job deletes files older than 30 days except the newest per
   document.

Target: < 3 s for the PFMEA of the 3,000-chain performance package. Use the normal excelize API
first; if `make perf` misses the target, write the table body with `StreamWriter` (re-writing
the header rows from the layout) — measure before changing.

## 7. Gate 1 comparison

`pfmea xlsx-compare --old <old.xlsx> --new <new.xlsx> --sheet <name> --from-row <n>` (tool in
`backend/cmd/pfmea`, milestone M10) compares the text of the table area cell by cell after
normalising whitespace and line breaks, ignoring formatting, and prints the differences. QA
reviews the remaining differences and the printed layout. The comparison report is attached to
the Gate 1 checklist (`docs/11-testing.md` §6).
