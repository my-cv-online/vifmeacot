---
name: fmea-domain
description: Domain knowledge for PFD, PFMEA (AIAG FMEA 4th edition) and Control Plan (APQP 2nd ed Template A) in this project. Use when working on worksheet logic, rules, Excel mapping, user-facing wording or when a quality-engineering term is unclear.
---

# FMEA domain quick reference

Full primer: `docs/01-domain-glossary.md`. The essentials:

- **Linkage**: PFD steps and characteristics are the single source; PFMEA and Control Plan
  reference them by id and show them read-only. The special-characteristic class lives on the
  characteristic and is printed with the customer's own symbol.
- **AIAG 4th row** = one failure chain = one failure mode + one cause. Effects belong to the
  failure mode; chain S = highest effect S (trigger). RPN = S × O × D computed by the database;
  NULL while a factor is missing. S/O/D are integers 1–10; S 9–10 needs an action or a
  justification (F07). AIAG discourages RPN thresholds; F08 only runs when a customer sets one.
- **Controls**: prevention lowers O, detection sets D. A detection method cannot justify a D
  better than the library's `d_min` (R04). Detection controls should appear in the CP on the
  same step and characteristic (R01); prevention controls too unless they are system controls
  with a reason (R02).
- **Actions**: PIC, target date, status; when done, re-rate new S/O/D (F09); lowering S needs a
  design change note (F10); overdue is computed in the plant time zone (F11).
- **Control Plan Template A columns**: Part/Process No., Process Name, Machine/Device/Jig/Tools,
  Characteristic No./Product/Process, Special Char. Class, Specification/Tolerance,
  Evaluation/Measurement Technique, Sample Size, Sample Freq., Control Method, Reaction Plan.
  Phases: Prototype, Pre-launch, Production (Safe Launch is CP-1, Phase 2).
- **IATF 16949**: rework/repair steps must be analysed in the PFMEA (W01); inspection steps
  need an NG disposition (W02); special characteristics must cascade to PFMEA and CP (S01).
- **Template General**: common processes maintained once; linked rows in model packages follow
  the template except recorded overrides (policy) and detached rows (`docs/07-template-general.md`).
- **Language**: FMEA content is usually English and is stored as typed; the UI is English;
  code comments are Indonesian.
- **Licensed text**: never reproduce AIAG/VDA rating tables or AP tables; the company enters
  them in master data.

When in doubt about a domain behaviour, prefer the documents in `docs/`; if they are silent,
ask the user (a quality engineer) rather than inventing rules.
