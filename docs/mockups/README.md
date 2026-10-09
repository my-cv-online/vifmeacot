# Mockups

Static HTML mockups (open the `.html` files in a browser; links between them work) and full-page
screenshots (`.png`, 1440 px wide). They show the **end state after all phases** with sample
data, so several elements are not Phase 1. `docs/08-screens.md` is authoritative for Phase 1;
use the mockups for layout, density and colours.

**Language:** the mockups were drawn with Indonesian labels before the project chose an English
user interface. The application shows English text only; `docs/08-screens.md` gives every label
and §14 there maps the mockup labels to the English ones. Do not copy Indonesian text from the
mockups into the application.

The mockups load IBM Plex from Google Fonts for convenience. The application must self-host the
fonts (`@fontsource`), because the factory network may have no internet access.

| File | Screen | Phase 1 use | Ignore in Phase 1 |
| --- | --- | --- | --- |
| `Main` | Dashboard | Layout of filters, KPI cards, package health, rule Pareto, overdue actions, activity | AIAG-VDA packages, "Profil Action Priority" (replace with the 4th risk profile), statuses other than Draft, "Out of sync" (PFD/WI) card, Model/Metodologi/Periode filters, "Ekspor ringkasan", "Jalankan cek sekarang", notification bell |
| `Pfmea4th` | PFMEA AIAG 4th worksheet | Header block, toolbar, grouped grid, column set, RPN note | "Konversi ke AIAG-VDA", "Ajukan review", Rev badge other than the document revision |
| `PfmeaVda` | PFMEA AIAG-VDA | Not in Phase 1 (reference for Phase 2 only) | Everything |
| `ControlPlan` | Control Plan | Header block, grid with grey read-only PFD columns, "+ Baris dari kontrol PFMEA", phase tabs | Template B (CP-1), Safe Launch tab and banner, C04 badge, Owner column, In Review status |
| `Pfd` | PFD editor | Step table with GENERAL/opsional chips, characteristics, diagram with NG branches and legend, delete-protection note | "Impor dari Excel" (legacy file import; Phase 1 has "Tempel dari Excel"), In Review status |
| `TemplateGeneral` | Template General | Tabs, general process list (Wajib/Opsional/Berubah), per-process change table, impact cards and conflict table, "Rilis dan sinkronkan" | Decision buttons in the impact table (decisions happen in each model package after release), "Released: revisi naik otomatis" card, approval chip wording |
| `Findings` | Consistency check results | Count cards, filter chips, findings table, detail panel with quick fixes | "Approval terkunci" wording, AIAG-VDA rules (F04–F06) and CP-1 rules (C02–C06) in the sample list |

Sample data in the mockups (package names, counts, revision numbers) is illustrative and does
not match `db/seed/demo.sql`; tests use the seed, never the mockups.
