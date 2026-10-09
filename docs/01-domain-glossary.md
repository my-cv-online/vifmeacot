# 01 · Domain primer and glossary

Read this before touching PFMEA or Control Plan code. It explains the quality documents in the
terms the code uses. The company is an automotive electronics (PCBA) plant certified to
IATF 16949. Engineers write FMEA content mostly in English, and the UI is in English too.

## 1. The three documents and how they link

Automotive quality planning (APQP) requires, for every product, three documents that must
agree with each other ("linkage"):

```text
PFD  (what happens, in which order)       step 30 Solder paste printing → characteristic 30-01 Paste volume (SC)
  ↓ every step is analysed
PFMEA (what can go wrong, how bad, how likely, how detected)
       failure mode "Insufficient paste" → effects (S) → causes (O) → controls (prevention / detection, D)
  ↓ every control and special characteristic is put into production
Control Plan (how it is controlled on the line)
       30-01 Paste volume · SPI 3D 100% · every board · reaction plan "stop line, …"
```

In Excel the three files drift apart. In this system they are **three views of one dataset**:
a step, characteristic or special-characteristic class is entered once (in the PFD) and shown
read-only in the PFMEA and the Control Plan. Rules (`docs/06-rules.md`) catch what can still be
inconsistent.

## 2. Glossary

| Term | Meaning | In the data model |
| --- | --- | --- |
| APQP | Advanced Product Quality Planning: the automotive project framework that produces PFD, PFMEA, CP | — |
| IATF 16949 | Automotive quality management standard; requires the linkage, rework analysis (8.7.1.4), special-characteristic control | rules W01, S01–S04 |
| CSR | Customer-specific requirements (e.g. symbol for critical characteristics, RPN threshold) | `customers`, `customer_sc_symbols`, `rule_overrides` |
| Package | One product's PFD + PFMEA + CP, released together | `packages`, `documents` |
| PFD | Process Flow Diagram: ordered process steps with symbols, machines, characteristics, NG/rework flows | `process_steps`, `step_flows`, `characteristics` |
| Process step / operation | One operation, numbered (`op_no` 10, 20, 25 …) | `process_steps` |
| Step symbol | Operation ○, inspection □, combined ⊡, transport ⇨, storage ▽, delay D, decision ◇ | `process_steps.symbol` |
| Rework / repair | A step that corrects nonconforming product; must be analysed in the PFMEA | `process_steps.kind = 'rework'`, W01 |
| NG flow, disposition | Where nonconforming parts go (rework, scrap, hold, return) | `step_flows` (`kind`, `disposition`), W02 |
| Characteristic | Product characteristic (feature of the part, e.g. solder joint) or process characteristic (parameter, e.g. reflow profile) | `characteristics.kind` |
| Special characteristic | Characteristic with safety/regulatory (critical, CC) or fit/function (significant, SC) importance | `characteristics.sc_symbol_id` → `sc_symbols` |
| Symbol conversion | Each customer prints its own symbol for the company's classes (◆, Ⓢ …) | `customer_sc_symbols`, S02 |
| PFMEA | Process Failure Mode and Effects Analysis | `failure_*`, `controls`, `actions` |
| Function / requirement | What the step must achieve; in the 4th form the requirement is the characteristic | `process_steps.function`, `failure_modes.characteristic_id` |
| Failure mode (FM) | How the step can fail to meet the requirement ("Insufficient paste") | `failure_modes` |
| Effect (FE) | Consequence for your plant, the ship-to plant or the end user | `failure_effects` (`level`, `s`) |
| Cause (FC) | Why the failure mode happens ("Stencil clogged") | `failure_causes` |
| Failure chain | One FM + one cause = one row of the AIAG 4th worksheet | `failure_chains` |
| Severity (S) | 1–10, seriousness of the worst effect; 9–10 = safety or regulatory | `failure_effects.s`; chain `s` = max (trigger) |
| Occurrence (O) | 1–10, likelihood of the cause given prevention controls | `failure_chains.o` |
| Detection (D) | 1–10, ability of detection controls to find it before it leaves (1 = certain) | `failure_chains.d` |
| RPN | Risk Priority Number = S × O × D (AIAG 4th). AIAG discourages fixed thresholds; some customers still require one | `failure_chains.rpn` (generated), F08 |
| AP | Action Priority H/M/L from the AIAG-VDA S-O-D table (Phase 2) | `failure_chains.ap` |
| Prevention control | Stops the cause or reduces its occurrence (feeder barcode interlock, PM) | `controls.kind = 'prevention'` |
| Detection control | Finds the failure mode or cause (AOI, ICT, visual) | `controls.kind = 'detection'` |
| System control | A prevention handled by a system (ERP/MES/WMS) and not by a CP line | `controls.is_system_control`, R02 |
| Error-proofing (poka-yoke) | Control that makes the error impossible or certainly detected | `control_library.is_error_proofing`, `cp_lines.is_error_proofing`, R06 |
| Recommended action | Improvement with PIC, target date, status and re-rating (new S/O/D) | `actions`, F07–F11 |
| Justification | Why no (further) action is needed for a high-risk row | `failure_chains.justification` |
| Control Plan (CP) | How each characteristic is controlled in production | `cp_lines`, `reaction_plans` |
| CP phase | Prototype, Pre-launch, Production (Safe Launch in CP-1, Phase 2) | `cp_lines.phase` |
| Evaluation / measurement technique | How it is measured (SPI 3D, caliper) | `cp_lines.eval_technique`, `gauge` |
| Sample size / frequency | How many, how often (5 pcs / every lot; 100% / every board) | `cp_lines.sample_size`, `sample_freq`, `freq_basis` |
| Control method | How the process is controlled (SPC chart, check sheet, interlock) | `cp_lines.control_method` |
| Reaction plan | What the operator does when out of control (stop, contain, correct) | `reaction_plans.text` (Template A) |
| Change level | Engineering revision of the part (A, B, C …) shown in document headers | `parts.change_level`, document `header.changeLevel`, K06 |
| WI | Work instruction; only its number and revision are stored | `process_steps.wi_ref` |
| Template General | Processes common to all products, maintained once and synced into every model package | `docs/07-template-general.md` |
| PCBA terms | SMT (surface mount), pick and place (mounter, feeders, nozzles), reflow oven, SPI (solder paste inspection), AOI (automated optical inspection), ICT (in-circuit test), MSD (moisture-sensitive device, floor life, MSL) | demo data |

## 3. AIAG FMEA 4th edition worksheet ↔ data model

One worksheet row per failure chain, grouped by step and failure mode (merged cells in Excel).

| # | Form column | Source |
| --- | --- | --- |
| 1 | Process Step / Function | `process_steps.op_no` + `name` (+ `function` on a second line) |
| 2 | Requirements | characteristic of the failure mode: `char_no` + `name` (+ `spec`) |
| 3 | Potential Failure Mode | `failure_modes.text` |
| 4 | Potential Effect(s) of Failure | all `failure_effects.text` of the failure mode, one per line (level prefix when not `unspecified`) |
| 5 | Sev | `failure_chains.s` |
| 6 | Class | characteristic's class printed with the customer's symbol |
| 7 | Potential Cause(s) of Failure | `failure_causes.text` |
| 8 | Current Process Controls – Prevention | `controls` (prevention) text, one per line |
| 9 | Occur | `failure_chains.o` |
| 10 | Current Process Controls – Detection | `controls` (detection) text, one per line |
| 11 | Detec | `failure_chains.d` |
| 12 | RPN | `failure_chains.rpn` |
| 13 | Recommended Action(s) | `actions.text`; when there is no action, `justification` prefixed with "Justification:" |
| 14 | Responsibility & Target Completion Date | responsible user's display name or `responsible_text`, `target_date` |
| 15 | Action Results: Action Taken & Completion Date | `action_taken`, `completed_on` |
| 16–19 | Action Results: S, O, D, RPN | `new_s`, `new_o`, `new_d`, `new_rpn` |

Header (`documents.header` of the PFMEA): Item, Model Year(s)/Program(s), Core Team, Process
Responsibility, Key Date, FMEA Number, Page, Prepared By, FMEA Date (Orig.), (Rev.).

## 4. Control Plan Template A (APQP 2nd edition form) ↔ data model

| # | Form column | Source |
| --- | --- | --- |
| 1 | Part/Process Number | `process_steps.op_no` |
| 2 | Process Name / Operation Description | `process_steps.name` |
| 3 | Machine, Device, Jig, Tools for Mfg. | `cp_lines.machines` (joined with ", ") |
| 4 | Characteristics – No. | `characteristics.char_no` |
| 5 | Characteristics – Product | `characteristics.name` when `kind = product` |
| 6 | Characteristics – Process | `characteristics.name` when `kind = process` |
| 7 | Special Char. Class | characteristic's class printed with the customer's symbol |
| 8 | Product/Process Specification/Tolerance | `characteristics.spec` (or `lsl`–`usl` `unit` when `spec` is empty) |
| 9 | Evaluation/Measurement Technique | `cp_lines.eval_technique` (+ gauge in parentheses) |
| 10 | Sample – Size | `cp_lines.sample_size` |
| 11 | Sample – Freq. | `cp_lines.sample_freq` |
| 12 | Control Method | `cp_lines.control_method` (+ "Error-proofing, verifikasi: …" when `is_error_proofing`) |
| 13 | Reaction Plan | `reaction_plans.text` |

Header: Prototype / Pre-Launch / Production tick boxes, Control Plan Number, Key Contact/Phone,
Date (Orig.), Date (Rev.), Part Number/Latest Change Level, Core Team, Part Name/Description,
Supplier/Plant, Supplier Code, approvals (Supplier/Plant, Customer Engineering, Customer
Quality, Other) with dates. Keys in `docs/04-data-model.md` §5.

Template B (AIAG CP-1, 2024) adds Safe Launch, a structured reaction plan with an owner and
volume-based frequencies. It is Phase 2; the columns already exist in `reaction_plans`.

## 5. PFD ↔ data model

The PFD has no mandated form. Columns used by the editor and the default export: Op No,
symbol, process name, function/description, machines/equipment, inputs, outputs, product and
process characteristics (with class), department/PIC, WI reference, optional/rework marks, and
NG/rework/scrap/return flows with disposition. The diagram is drawn automatically from `seq`
order plus `step_flows`.

## 6. Domain invariants the code must keep

1. Steps, characteristics and classes are defined only in the PFD; PFMEA and CP reference
   them by id and show them read-only.
2. A failure chain belongs to exactly one failure mode and one cause; its S is the highest S of
   the failure mode's effects (database trigger); RPN is computed by the database.
3. The special-characteristic class lives only on the characteristic; the PFMEA "Class" and the
   CP "Special Char. Class" display it; printing uses the customer's symbol.
4. CP lines belong to a step and one of its characteristics; spec and class come from the
   characteristic and are not editable in the CP.
5. S, O, D are integers 1–10. S 9–10 requires an action or a justification (F07).
6. A completed action must be re-rated (F09); lowering S needs a design change (F10).
7. Rework steps are analysed in the PFMEA (W01); inspection steps have an NG disposition (W02).
8. Template-controlled cells of linked rows change only through the template or a recorded
   override.
9. Methodology is data, never an assumption: code branches on `packages.pfmea_method` and
   `packages.cp_format`.

## 7. Licensed content

S/O/D rating tables and AIAG-VDA AP tables are copyrighted (AIAG / VDA). The application ships
empty criteria; the company enters the text from its licensed manuals in master data
(`rating_criteria`). Never add manual text to code, seeds, fixtures or tests.
