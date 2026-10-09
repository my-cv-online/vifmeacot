# Aturan konsistensi — Test case

Kebutuhan: P1-08, `docs/06-rules.md` · Milestone: M1 (query dan fixture), M8 (mesin aturan dan pesan)

Setiap aturan Tahap 1 punya satu fixture di `backend/internal/rules/testdata/<CODE>.sql`.
Fixture adalah test otomatisnya: dijalankan di atas data demo dalam transaksi yang di-rollback.
Isi `-- expect:` adalah **seluruh** hasil aturan setelah fixture (termasuk temuan baseline).
Aturan baru atau perubahan aturan menambah case di file ini lebih dulu (lihat `/add-rule`).

## Ringkasan

| ID | Judul | Level | Status |
| --- | --- | --- | --- |
| TC-RULE-BASE-1 | Semua aturan pada PS-07 menghasilkan tepat 14 temuan baseline | integrasi | lulus (pesan mulai M8) |
| TC-RULE-BASE-2 | Semua aturan pada GENERAL tidak menghasilkan temuan | integrasi | lulus |
| TC-RULE-C01-1 | Sample size dan reaction plan kosong pada baris CP step 90. | integrasi | lulus |
| TC-RULE-F01-1 | Teks failure mode kosong dan D tidak diisi. | integrasi | lulus |
| TC-RULE-F02-1 | Dua failure mode diketik dalam satu sel. | integrasi | lulus |
| TC-RULE-F03-1 | Teks effect yang sama dinilai 6 sekali dan 7 dua kali. | integrasi | lulus |
| TC-RULE-F07-1 | S dinaikkan menjadi 9 pada chain tanpa aksi atau justifikasi. | integrasi | lulus |
| TC-RULE-F08-1 | Customer B menetapkan ambang RPN 100; chain dengan RPN >= 100 tanpa aksi dilaporkan. | integrasi | lulus |
| TC-RULE-F09-1 | Aksi selesai tanpa D baru. | integrasi | lulus |
| TC-RULE-F10-1 | S baru lebih rendah dari S saat ini tanpa catatan perubahan desain. | integrasi | lulus |
| TC-RULE-F11-1 | Aksi kedua yang lewat target. | integrasi | lulus |
| TC-RULE-F12-1 | Nomor dokumen diketik di sel kontrol. | integrasi | lulus |
| TC-RULE-K01-1 | Step baru tanpa failure mode dilaporkan. | integrasi | lulus |
| TC-RULE-K02-1 | Menghapus baris CP 50-03 menambah K02 kedua di samping baseline demo 60-03. | integrasi | lulus |
| TC-RULE-K03-1 | Failure mode step 50 yang menunjuk karakteristik milik step 60. | integrasi | lulus |
| TC-RULE-K05-1 | Mesin tak terdaftar kedua pada baris CP. | integrasi | lulus |
| TC-RULE-K06-1 | Header PFMEA kini juga berbeda dengan data part (part number). | integrasi | lulus |
| TC-RULE-R01-1 | Menghapus baris CP 60-01 membuat detection control-nya tidak tercakup. | integrasi | lulus |
| TC-RULE-R02-1 | Menandai zone temperature alarm sebagai kontrol sistem menghilangkan temuan demo. | integrasi | lulus |
| TC-RULE-R03-1 | Memutus tautan baris CP AOI dari PFMEA menambah R03 kedua. | integrasi | lulus |
| TC-RULE-R04-1 | D di step 90 diperbaiki; D terlalu optimis di step 50 (AOI hanya membenarkan D >= 3). | integrasi | lulus |
| TC-RULE-R05-1 | O = 3 dengan satu-satunya prevention control dihapus. | integrasi | lulus |
| TC-RULE-R06-1 | Menambah frekuensi verifikasi menghilangkan temuan demo. | integrasi | lulus |
| TC-RULE-S01-1 | Karakteristik SC baru yang belum ada di PFMEA maupun CP. | integrasi | lulus |
| TC-RULE-S02-1 | Karakteristik CC kedua di paket Customer B (Customer B tidak punya pemetaan CC). | integrasi | lulus |
| TC-RULE-S03-1 | Menandai 90-01 sebagai SC memperlihatkan kontrolnya yang lemah (visual 100%). | integrasi | lulus |
| TC-RULE-S04-1 | Menurunkan syarat CC customer menjadi 8 menghilangkan temuan demo. | integrasi | lulus |
| TC-RULE-T01-1 | Step Shipping tertaut (wajib di rev 1) dihapus dari paket model. | integrasi | lulus |
| TC-RULE-T02-1 | Template General merilis rev 2 tetapi paket model masih mengikuti rev 1. | integrasi | lulus |
| TC-RULE-T03-1 | Dua konflik sinkron yang belum diselesaikan: konflik override kolom pada baris CP 20-02 dan konflik baris pada step 20 (dihapus di template, masih dipakai baris lokal). | integrasi | lulus |
| TC-RULE-T04-1 | Baris tertaut dilepas dari template tanpa alasan. | integrasi | lulus |
| TC-RULE-W01-1 | Menganalisis step rework menghilangkan temuan demo. | integrasi | lulus |
| TC-RULE-W02-1 | Final inspection tanpa cabang NG. | integrasi | lulus |
| TC-RULE-W04-1 | Review terbaru menghilangkan temuan demo. | integrasi | lulus |

## TC-RULE-BASE-1 — Semua aturan pada PS-07 menghasilkan tepat 14 temuan baseline

- **Level:** integrasi (database) · **Kebutuhan:** `docs/06-rules.md` §5
- **Prasyarat:** data demo, `@today` = 2026-10-08.
- **Langkah:** jalankan ke-32 aturan Tahap 1 pada paket PS-07.
- **Hasil yang diharapkan:** tepat 14 hasil: 5 error (K02, K06, R01, S02, W01), 7 warning
  (F11, K05, R02, R03, R04, R06, S04), 2 info (F12, W04), dengan objek dan field seperti tabel
  di `docs/06-rules.md` §5. Mulai M8 pesan bahasa Inggris yang dirender juga harus sama persis.
- **Test otomatis:** `backend/internal/rules/rules_test.go` › `TestBaseline/TC-RULE-BASE-1`
- **Status:** lulus untuk aturan, objek, field dan jumlah per level (M1); pesan yang dirender
  dibandingkan mulai M8

## TC-RULE-BASE-2 — Semua aturan pada GENERAL tidak menghasilkan temuan

- **Level:** integrasi (database) · **Kebutuhan:** `docs/06-rules.md` §5
- **Prasyarat:** data demo, `@today` = 2026-10-08.
- **Langkah:** jalankan ke-32 aturan Tahap 1 pada paket GENERAL.
- **Hasil yang diharapkan:** tidak ada hasil.
- **Test otomatis:** `backend/internal/rules/rules_test.go` › `TestBaseline/TC-RULE-BASE-2`
- **Status:** lulus

## TC-RULE-C01-1 — Sample size dan reaction plan kosong pada baris CP step 90.

- **Level:** integrasi (database) · **Kebutuhan:** aturan C01 (`docs/06-rules.md` §4)
- **Prasyarat:** data demo (`db/seed/demo.sql`), paket PS-07, `@today` = 2026-10-08.
- **Langkah:**
  1. Buka transaksi, jalankan `backend/internal/rules/testdata/C01.sql`.
  2. Jalankan `backend/internal/rules/sql/C01.sql` dengan `@package_id` = PS-07.
  3. Rollback transaksi.
- **Hasil yang diharapkan:** aturan C01 mengembalikan tepat 2 hasil:
  - `cp_lines` · `sampleSize` dengan charNo = "90-01"
  - `cp_lines` · `reactionPlan` dengan charNo = "90-01"
- **Test otomatis:** `backend/internal/rules/rules_test.go` › `TestRuleFixtures/TC-RULE-C01-1` (dibuat di M1)
- **Status:** lulus

## TC-RULE-F01-1 — Teks failure mode kosong dan D tidak diisi.

- **Level:** integrasi (database) · **Kebutuhan:** aturan F01 (`docs/06-rules.md` §4)
- **Prasyarat:** data demo (`db/seed/demo.sql`), paket PS-07, `@today` = 2026-10-08.
- **Langkah:**
  1. Buka transaksi, jalankan `backend/internal/rules/testdata/F01.sql`.
  2. Jalankan `backend/internal/rules/sql/F01.sql` dengan `@package_id` = PS-07.
  3. Rollback transaksi.
- **Hasil yang diharapkan:** aturan F01 mengembalikan tepat 2 hasil:
  - `failure_modes` · `text` dengan label = "Failure mode", opNo = "50"
  - `failure_chains` · `d` dengan label = "Detection (D)", opNo = "50"
- **Test otomatis:** `backend/internal/rules/rules_test.go` › `TestRuleFixtures/TC-RULE-F01-1` (dibuat di M1)
- **Status:** lulus

## TC-RULE-F02-1 — Dua failure mode diketik dalam satu sel.

- **Level:** integrasi (database) · **Kebutuhan:** aturan F02 (`docs/06-rules.md` §4)
- **Prasyarat:** data demo (`db/seed/demo.sql`), paket PS-07, `@today` = 2026-10-08.
- **Langkah:**
  1. Buka transaksi, jalankan `backend/internal/rules/testdata/F02.sql`.
  2. Jalankan `backend/internal/rules/sql/F02.sql` dengan `@package_id` = PS-07.
  3. Rollback transaksi.
- **Hasil yang diharapkan:** aturan F02 mengembalikan tepat 1 hasil:
  - `failure_modes` · `text` dengan text = "Burr / crack"
- **Test otomatis:** `backend/internal/rules/rules_test.go` › `TestRuleFixtures/TC-RULE-F02-1` (dibuat di M1)
- **Status:** lulus

## TC-RULE-F03-1 — Teks effect yang sama dinilai 6 sekali dan 7 dua kali.

- **Level:** integrasi (database) · **Kebutuhan:** aturan F03 (`docs/06-rules.md` §4)
- **Prasyarat:** data demo (`db/seed/demo.sql`), paket PS-07, `@today` = 2026-10-08.
- **Langkah:**
  1. Buka transaksi, jalankan `backend/internal/rules/testdata/F03.sql`.
  2. Jalankan `backend/internal/rules/sql/F03.sql` dengan `@package_id` = PS-07.
  3. Rollback transaksi.
- **Hasil yang diharapkan:** aturan F03 mengembalikan tepat 3 hasil:
  - `failure_effects` · `s` dengan s = 7, values = "6, 7"
  - `failure_effects` · `s` dengan s = 7, values = "6, 7"
  - `failure_effects` · `s` dengan s = 6, values = "6, 7"
- **Test otomatis:** `backend/internal/rules/rules_test.go` › `TestRuleFixtures/TC-RULE-F03-1` (dibuat di M1)
- **Status:** lulus

## TC-RULE-F07-1 — S dinaikkan menjadi 9 pada chain tanpa aksi atau justifikasi.

- **Level:** integrasi (database) · **Kebutuhan:** aturan F07 (`docs/06-rules.md` §4)
- **Prasyarat:** data demo (`db/seed/demo.sql`), paket PS-07, `@today` = 2026-10-08.
- **Langkah:**
  1. Buka transaksi, jalankan `backend/internal/rules/testdata/F07.sql`.
  2. Jalankan `backend/internal/rules/sql/F07.sql` dengan `@package_id` = PS-07.
  3. Rollback transaksi.
- **Hasil yang diharapkan:** aturan F07 mengembalikan tepat 1 hasil:
  - `failure_chains` · `s` dengan s = 9, opNo = "60"
- **Test otomatis:** `backend/internal/rules/rules_test.go` › `TestRuleFixtures/TC-RULE-F07-1` (dibuat di M1)
- **Status:** lulus

## TC-RULE-F08-1 — Customer B menetapkan ambang RPN 100; chain dengan RPN >= 100 tanpa aksi dilaporkan.

- **Level:** integrasi (database) · **Kebutuhan:** aturan F08 (`docs/06-rules.md` §4)
- **Prasyarat:** data demo (`db/seed/demo.sql`), paket PS-07, `@today` = 2026-10-08.
- **Langkah:**
  1. Buka transaksi, jalankan `backend/internal/rules/testdata/F08.sql`.
  2. Jalankan `backend/internal/rules/sql/F08.sql` dengan `@package_id` = PS-07.
  3. Rollback transaksi.
- **Hasil yang diharapkan:** aturan F08 mengembalikan tepat 4 hasil:
  - `failure_chains` · `rpn` dengan rpn = 126
  - `failure_chains` · `rpn` dengan rpn = 144
  - `failure_chains` · `rpn` dengan rpn = 105
  - `failure_chains` · `rpn` dengan rpn = 105
- **Test otomatis:** `backend/internal/rules/rules_test.go` › `TestRuleFixtures/TC-RULE-F08-1` (dibuat di M1)
- **Status:** lulus

## TC-RULE-F09-1 — Aksi selesai tanpa D baru.

- **Level:** integrasi (database) · **Kebutuhan:** aturan F09 (`docs/06-rules.md` §4)
- **Prasyarat:** data demo (`db/seed/demo.sql`), paket PS-07, `@today` = 2026-10-08.
- **Langkah:**
  1. Buka transaksi, jalankan `backend/internal/rules/testdata/F09.sql`.
  2. Jalankan `backend/internal/rules/sql/F09.sql` dengan `@package_id` = PS-07.
  3. Rollback transaksi.
- **Hasil yang diharapkan:** aturan F09 mengembalikan tepat 1 hasil:
  - `actions` · `newD` dengan label = "new D"
- **Test otomatis:** `backend/internal/rules/rules_test.go` › `TestRuleFixtures/TC-RULE-F09-1` (dibuat di M1)
- **Status:** lulus

## TC-RULE-F10-1 — S baru lebih rendah dari S saat ini tanpa catatan perubahan desain.

- **Level:** integrasi (database) · **Kebutuhan:** aturan F10 (`docs/06-rules.md` §4)
- **Prasyarat:** data demo (`db/seed/demo.sql`), paket PS-07, `@today` = 2026-10-08.
- **Langkah:**
  1. Buka transaksi, jalankan `backend/internal/rules/testdata/F10.sql`.
  2. Jalankan `backend/internal/rules/sql/F10.sql` dengan `@package_id` = PS-07.
  3. Rollback transaksi.
- **Hasil yang diharapkan:** aturan F10 mengembalikan tepat 1 hasil:
  - `actions` · `designChangeNote` dengan s = 5, newS = 4
- **Test otomatis:** `backend/internal/rules/rules_test.go` › `TestRuleFixtures/TC-RULE-F10-1` (dibuat di M1)
- **Status:** lulus

## TC-RULE-F11-1 — Aksi kedua yang lewat target.

- **Level:** integrasi (database) · **Kebutuhan:** aturan F11 (`docs/06-rules.md` §4)
- **Prasyarat:** data demo (`db/seed/demo.sql`), paket PS-07, `@today` = 2026-10-08.
- **Langkah:**
  1. Buka transaksi, jalankan `backend/internal/rules/testdata/F11.sql`.
  2. Jalankan `backend/internal/rules/sql/F11.sql` dengan `@package_id` = PS-07.
  3. Rollback transaksi.
- **Hasil yang diharapkan:** aturan F11 mengembalikan tepat 2 hasil:
  - `actions` · `targetDate` dengan daysLate = 13
  - `actions` · `targetDate` dengan daysLate = 7
- **Test otomatis:** `backend/internal/rules/rules_test.go` › `TestRuleFixtures/TC-RULE-F11-1` (dibuat di M1)
- **Status:** lulus

## TC-RULE-F12-1 — Nomor dokumen diketik di sel kontrol.

- **Level:** integrasi (database) · **Kebutuhan:** aturan F12 (`docs/06-rules.md` §4)
- **Prasyarat:** data demo (`db/seed/demo.sql`), paket PS-07, `@today` = 2026-10-08.
- **Langkah:**
  1. Buka transaksi, jalankan `backend/internal/rules/testdata/F12.sql`.
  2. Jalankan `backend/internal/rules/sql/F12.sql` dengan `@package_id` = PS-07.
  3. Rollback transaksi.
- **Hasil yang diharapkan:** aturan F12 mengembalikan tepat 2 hasil:
  - `failure_causes` · `text` dengan term = "operator error"
  - `controls` · `text` dengan term = "WI-"
- **Test otomatis:** `backend/internal/rules/rules_test.go` › `TestRuleFixtures/TC-RULE-F12-1` (dibuat di M1)
- **Status:** lulus

## TC-RULE-K01-1 — Step baru tanpa failure mode dilaporkan.

- **Level:** integrasi (database) · **Kebutuhan:** aturan K01 (`docs/06-rules.md` §4)
- **Prasyarat:** data demo (`db/seed/demo.sql`), paket PS-07, `@today` = 2026-10-08.
- **Langkah:**
  1. Buka transaksi, jalankan `backend/internal/rules/testdata/K01.sql`.
  2. Jalankan `backend/internal/rules/sql/K01.sql` dengan `@package_id` = PS-07.
  3. Rollback transaksi.
- **Hasil yang diharapkan:** aturan K01 mengembalikan tepat 1 hasil:
  - `process_steps` · `name` dengan opNo = "95"
- **Test otomatis:** `backend/internal/rules/rules_test.go` › `TestRuleFixtures/TC-RULE-K01-1` (dibuat di M1)
- **Status:** lulus

## TC-RULE-K02-1 — Menghapus baris CP 50-03 menambah K02 kedua di samping baseline demo 60-03.

- **Level:** integrasi (database) · **Kebutuhan:** aturan K02 (`docs/06-rules.md` §4)
- **Prasyarat:** data demo (`db/seed/demo.sql`), paket PS-07, `@today` = 2026-10-08.
- **Langkah:**
  1. Buka transaksi, jalankan `backend/internal/rules/testdata/K02.sql`.
  2. Jalankan `backend/internal/rules/sql/K02.sql` dengan `@package_id` = PS-07.
  3. Rollback transaksi.
- **Hasil yang diharapkan:** aturan K02 mengembalikan tepat 2 hasil:
  - `characteristics` · `name` dengan charNo = "60-03"
  - `characteristics` · `name` dengan charNo = "50-03"
- **Test otomatis:** `backend/internal/rules/rules_test.go` › `TestRuleFixtures/TC-RULE-K02-1` (dibuat di M1)
- **Status:** lulus

## TC-RULE-K03-1 — Failure mode step 50 yang menunjuk karakteristik milik step 60.

- **Level:** integrasi (database) · **Kebutuhan:** aturan K03 (`docs/06-rules.md` §4)
- **Prasyarat:** data demo (`db/seed/demo.sql`), paket PS-07, `@today` = 2026-10-08.
- **Langkah:**
  1. Buka transaksi, jalankan `backend/internal/rules/testdata/K03.sql`.
  2. Jalankan `backend/internal/rules/sql/K03.sql` dengan `@package_id` = PS-07.
  3. Rollback transaksi.
- **Hasil yang diharapkan:** aturan K03 mengembalikan tepat 1 hasil:
  - `failure_modes` · `characteristicId` dengan opNo = "50", charNo = "60-01", charOpNo = "60"
- **Test otomatis:** `backend/internal/rules/rules_test.go` › `TestRuleFixtures/TC-RULE-K03-1` (dibuat di M1)
- **Status:** lulus

## TC-RULE-K05-1 — Mesin tak terdaftar kedua pada baris CP.

- **Level:** integrasi (database) · **Kebutuhan:** aturan K05 (`docs/06-rules.md` §4)
- **Prasyarat:** data demo (`db/seed/demo.sql`), paket PS-07, `@today` = 2026-10-08.
- **Langkah:**
  1. Buka transaksi, jalankan `backend/internal/rules/testdata/K05.sql`.
  2. Jalankan `backend/internal/rules/sql/K05.sql` dengan `@package_id` = PS-07.
  3. Rollback transaksi.
- **Hasil yang diharapkan:** aturan K05 mengembalikan tepat 2 hasil:
  - `cp_lines` · `machines` dengan opNo = "90", unknown = "Lux meter"
  - `cp_lines` · `machines` dengan opNo = "50", unknown = "Glue dispenser"
- **Test otomatis:** `backend/internal/rules/rules_test.go` › `TestRuleFixtures/TC-RULE-K05-1` (dibuat di M1)
- **Status:** lulus

## TC-RULE-K06-1 — Header PFMEA kini juga berbeda dengan data part (part number).

- **Level:** integrasi (database) · **Kebutuhan:** aturan K06 (`docs/06-rules.md` §4)
- **Prasyarat:** data demo (`db/seed/demo.sql`), paket PS-07, `@today` = 2026-10-08.
- **Langkah:**
  1. Buka transaksi, jalankan `backend/internal/rules/testdata/K06.sql`.
  2. Jalankan `backend/internal/rules/sql/K06.sql` dengan `@package_id` = PS-07.
  3. Rollback transaksi.
- **Hasil yang diharapkan:** aturan K06 mengembalikan tepat 2 hasil:
  - `documents` · `header` dengan docType = "CP", docChangeLevel = "A"
  - `documents` · `header` dengan docType = "PFMEA", docPartNo = "PS-08"
- **Test otomatis:** `backend/internal/rules/rules_test.go` › `TestRuleFixtures/TC-RULE-K06-1` (dibuat di M1)
- **Status:** lulus

## TC-RULE-R01-1 — Menghapus baris CP 60-01 membuat detection control-nya tidak tercakup.

- **Level:** integrasi (database) · **Kebutuhan:** aturan R01 (`docs/06-rules.md` §4)
- **Prasyarat:** data demo (`db/seed/demo.sql`), paket PS-07, `@today` = 2026-10-08.
- **Langkah:**
  1. Buka transaksi, jalankan `backend/internal/rules/testdata/R01.sql`.
  2. Jalankan `backend/internal/rules/sql/R01.sql` dengan `@package_id` = PS-07.
  3. Rollback transaksi.
- **Hasil yang diharapkan:** aturan R01 mengembalikan tepat 2 hasil:
  - `controls` · `text` dengan charNo = "60-03", control = "AOI 100%"
  - `controls` · `text` dengan charNo = "60-01", control = "AOI 100% + ICT"
- **Test otomatis:** `backend/internal/rules/rules_test.go` › `TestRuleFixtures/TC-RULE-R01-1` (dibuat di M1)
- **Status:** lulus

## TC-RULE-R02-1 — Menandai zone temperature alarm sebagai kontrol sistem menghilangkan temuan demo.

- **Level:** integrasi (database) · **Kebutuhan:** aturan R02 (`docs/06-rules.md` §4)
- **Prasyarat:** data demo (`db/seed/demo.sql`), paket PS-07, `@today` = 2026-10-08.
- **Langkah:**
  1. Buka transaksi, jalankan `backend/internal/rules/testdata/R02.sql`.
  2. Jalankan `backend/internal/rules/sql/R02.sql` dengan `@package_id` = PS-07.
  3. Rollback transaksi.
- **Hasil yang diharapkan:** aturan R02 tidak mengembalikan hasil apa pun (temuan demo hilang).
- **Test otomatis:** `backend/internal/rules/rules_test.go` › `TestRuleFixtures/TC-RULE-R02-1` (dibuat di M1)
- **Status:** lulus

## TC-RULE-R03-1 — Memutus tautan baris CP AOI dari PFMEA menambah R03 kedua.

- **Level:** integrasi (database) · **Kebutuhan:** aturan R03 (`docs/06-rules.md` §4)
- **Prasyarat:** data demo (`db/seed/demo.sql`), paket PS-07, `@today` = 2026-10-08.
- **Langkah:**
  1. Buka transaksi, jalankan `backend/internal/rules/testdata/R03.sql`.
  2. Jalankan `backend/internal/rules/sql/R03.sql` dengan `@package_id` = PS-07.
  3. Rollback transaksi.
- **Hasil yang diharapkan:** aturan R03 mengembalikan tepat 2 hasil:
  - `cp_lines` · `controlId` dengan charNo = "75-01"
  - `cp_lines` · `controlId` dengan charNo = "70-01"
- **Test otomatis:** `backend/internal/rules/rules_test.go` › `TestRuleFixtures/TC-RULE-R03-1` (dibuat di M1)
- **Status:** lulus

## TC-RULE-R04-1 — D di step 90 diperbaiki; D terlalu optimis di step 50 (AOI hanya membenarkan D >= 3).

- **Level:** integrasi (database) · **Kebutuhan:** aturan R04 (`docs/06-rules.md` §4)
- **Prasyarat:** data demo (`db/seed/demo.sql`), paket PS-07, `@today` = 2026-10-08.
- **Langkah:**
  1. Buka transaksi, jalankan `backend/internal/rules/testdata/R04.sql`.
  2. Jalankan `backend/internal/rules/sql/R04.sql` dengan `@package_id` = PS-07.
  3. Rollback transaksi.
- **Hasil yang diharapkan:** aturan R04 mengembalikan tepat 1 hasil:
  - `failure_chains` · `d` dengan opNo = "50", d = 2, dMin = 3
- **Test otomatis:** `backend/internal/rules/rules_test.go` › `TestRuleFixtures/TC-RULE-R04-1` (dibuat di M1)
- **Status:** lulus

## TC-RULE-R05-1 — O = 3 dengan satu-satunya prevention control dihapus.

- **Level:** integrasi (database) · **Kebutuhan:** aturan R05 (`docs/06-rules.md` §4)
- **Prasyarat:** data demo (`db/seed/demo.sql`), paket PS-07, `@today` = 2026-10-08.
- **Langkah:**
  1. Buka transaksi, jalankan `backend/internal/rules/testdata/R05.sql`.
  2. Jalankan `backend/internal/rules/sql/R05.sql` dengan `@package_id` = PS-07.
  3. Rollback transaksi.
- **Hasil yang diharapkan:** aturan R05 mengembalikan tepat 1 hasil:
  - `failure_chains` · `o` dengan o = 3, opNo = "60"
- **Test otomatis:** `backend/internal/rules/rules_test.go` › `TestRuleFixtures/TC-RULE-R05-1` (dibuat di M1)
- **Status:** lulus

## TC-RULE-R06-1 — Menambah frekuensi verifikasi menghilangkan temuan demo.

- **Level:** integrasi (database) · **Kebutuhan:** aturan R06 (`docs/06-rules.md` §4)
- **Prasyarat:** data demo (`db/seed/demo.sql`), paket PS-07, `@today` = 2026-10-08.
- **Langkah:**
  1. Buka transaksi, jalankan `backend/internal/rules/testdata/R06.sql`.
  2. Jalankan `backend/internal/rules/sql/R06.sql` dengan `@package_id` = PS-07.
  3. Rollback transaksi.
- **Hasil yang diharapkan:** aturan R06 tidak mengembalikan hasil apa pun (temuan demo hilang).
- **Test otomatis:** `backend/internal/rules/rules_test.go` › `TestRuleFixtures/TC-RULE-R06-1` (dibuat di M1)
- **Status:** lulus

## TC-RULE-S01-1 — Karakteristik SC baru yang belum ada di PFMEA maupun CP.

- **Level:** integrasi (database) · **Kebutuhan:** aturan S01 (`docs/06-rules.md` §4)
- **Prasyarat:** data demo (`db/seed/demo.sql`), paket PS-07, `@today` = 2026-10-08.
- **Langkah:**
  1. Buka transaksi, jalankan `backend/internal/rules/testdata/S01.sql`.
  2. Jalankan `backend/internal/rules/sql/S01.sql` dengan `@package_id` = PS-07.
  3. Rollback transaksi.
- **Hasil yang diharapkan:** aturan S01 mengembalikan tepat 2 hasil:
  - `characteristics` · `scSymbolId` dengan charNo = "60-05", missingIn = "PFMEA"
  - `characteristics` · `scSymbolId` dengan charNo = "60-05", missingIn = "Control Plan"
- **Test otomatis:** `backend/internal/rules/rules_test.go` › `TestRuleFixtures/TC-RULE-S01-1` (dibuat di M1)
- **Status:** lulus

## TC-RULE-S02-1 — Karakteristik CC kedua di paket Customer B (Customer B tidak punya pemetaan CC).

- **Level:** integrasi (database) · **Kebutuhan:** aturan S02 (`docs/06-rules.md` §4)
- **Prasyarat:** data demo (`db/seed/demo.sql`), paket PS-07, `@today` = 2026-10-08.
- **Langkah:**
  1. Buka transaksi, jalankan `backend/internal/rules/testdata/S02.sql`.
  2. Jalankan `backend/internal/rules/sql/S02.sql` dengan `@package_id` = PS-07.
  3. Rollback transaksi.
- **Hasil yang diharapkan:** aturan S02 mengembalikan tepat 2 hasil:
  - `characteristics` · `scSymbolId` dengan charNo = "60-02", symbol = "CC"
  - `characteristics` · `scSymbolId` dengan charNo = "50-01", symbol = "CC"
- **Test otomatis:** `backend/internal/rules/rules_test.go` › `TestRuleFixtures/TC-RULE-S02-1` (dibuat di M1)
- **Status:** lulus

## TC-RULE-S03-1 — Menandai 90-01 sebagai SC memperlihatkan kontrolnya yang lemah (visual 100%).

- **Level:** integrasi (database) · **Kebutuhan:** aturan S03 (`docs/06-rules.md` §4)
- **Prasyarat:** data demo (`db/seed/demo.sql`), paket PS-07, `@today` = 2026-10-08.
- **Langkah:**
  1. Buka transaksi, jalankan `backend/internal/rules/testdata/S03.sql`.
  2. Jalankan `backend/internal/rules/sql/S03.sql` dengan `@package_id` = PS-07.
  3. Rollback transaksi.
- **Hasil yang diharapkan:** aturan S03 mengembalikan tepat 1 hasil:
  - `cp_lines` · `controlMethod` dengan charNo = "90-01", symbol = "SC"
- **Test otomatis:** `backend/internal/rules/rules_test.go` › `TestRuleFixtures/TC-RULE-S03-1` (dibuat di M1)
- **Status:** lulus

## TC-RULE-S04-1 — Menurunkan syarat CC customer menjadi 8 menghilangkan temuan demo.

- **Level:** integrasi (database) · **Kebutuhan:** aturan S04 (`docs/06-rules.md` §4)
- **Prasyarat:** data demo (`db/seed/demo.sql`), paket PS-07, `@today` = 2026-10-08.
- **Langkah:**
  1. Buka transaksi, jalankan `backend/internal/rules/testdata/S04.sql`.
  2. Jalankan `backend/internal/rules/sql/S04.sql` dengan `@package_id` = PS-07.
  3. Rollback transaksi.
- **Hasil yang diharapkan:** aturan S04 tidak mengembalikan hasil apa pun (temuan demo hilang).
- **Test otomatis:** `backend/internal/rules/rules_test.go` › `TestRuleFixtures/TC-RULE-S04-1` (dibuat di M1)
- **Status:** lulus

## TC-RULE-T01-1 — Step Shipping tertaut (wajib di rev 1) dihapus dari paket model.

- **Level:** integrasi (database) · **Kebutuhan:** aturan T01 (`docs/06-rules.md` §4)
- **Prasyarat:** data demo (`db/seed/demo.sql`), paket PS-07, `@today` = 2026-10-08.
- **Langkah:**
  1. Buka transaksi, jalankan `backend/internal/rules/testdata/T01.sql`.
  2. Jalankan `backend/internal/rules/sql/T01.sql` dengan `@package_id` = PS-07.
  3. Rollback transaksi.
- **Hasil yang diharapkan:** aturan T01 mengembalikan tepat 1 hasil:
  - `packages` · `steps` dengan opNo = "110", rev = 1
- **Test otomatis:** `backend/internal/rules/rules_test.go` › `TestRuleFixtures/TC-RULE-T01-1` (dibuat di M1)
- **Status:** lulus

## TC-RULE-T02-1 — Template General merilis rev 2 tetapi paket model masih mengikuti rev 1.

- **Level:** integrasi (database) · **Kebutuhan:** aturan T02 (`docs/06-rules.md` §4)
- **Prasyarat:** data demo (`db/seed/demo.sql`), paket PS-07, `@today` = 2026-10-08.
- **Langkah:**
  1. Buka transaksi, jalankan `backend/internal/rules/testdata/T02.sql`.
  2. Jalankan `backend/internal/rules/sql/T02.sql` dengan `@package_id` = PS-07.
  3. Rollback transaksi.
- **Hasil yang diharapkan:** aturan T02 mengembalikan tepat 1 hasil:
  - `packages` · `templateRev` dengan syncedRev = 1, latestRev = 2
- **Test otomatis:** `backend/internal/rules/rules_test.go` › `TestRuleFixtures/TC-RULE-T02-1` (dibuat di M1)
- **Status:** lulus

## TC-RULE-T03-1 — Dua konflik sinkron yang belum diselesaikan: konflik override kolom pada baris CP 20-02 dan konflik baris pada step 20 (dihapus di template, masih dipakai baris lokal).

- **Level:** integrasi (database) · **Kebutuhan:** aturan T03 (`docs/06-rules.md` §4)
- **Prasyarat:** data demo (`db/seed/demo.sql`), paket PS-07, `@today` = 2026-10-08.
- **Langkah:**
  1. Buka transaksi, jalankan `backend/internal/rules/testdata/T03.sql`.
  2. Jalankan `backend/internal/rules/sql/T03.sql` dengan `@package_id` = PS-07.
  3. Rollback transaksi.
- **Hasil yang diharapkan:** aturan T03 mengembalikan tepat 2 hasil:
  - `cp_lines` · `sampleFreq` dengan column = "sampleFreq", rev = 2, localValue = "Every 3 reels", templateValue = "Every reel"
  - `process_steps` · `(seluruh baris)` dengan column = "", rev = 2, label = "20 IQC"
- **Test otomatis:** `backend/internal/rules/rules_test.go` › `TestRuleFixtures/TC-RULE-T03-1` (dibuat di M1)
- **Status:** lulus

## TC-RULE-T04-1 — Baris tertaut dilepas dari template tanpa alasan.

- **Level:** integrasi (database) · **Kebutuhan:** aturan T04 (`docs/06-rules.md` §4)
- **Prasyarat:** data demo (`db/seed/demo.sql`), paket PS-07, `@today` = 2026-10-08.
- **Langkah:**
  1. Buka transaksi, jalankan `backend/internal/rules/testdata/T04.sql`.
  2. Jalankan `backend/internal/rules/sql/T04.sql` dengan `@package_id` = PS-07.
  3. Rollback transaksi.
- **Hasil yang diharapkan:** aturan T04 mengembalikan tepat 1 hasil:
  - `cp_lines` · `detachReason`
- **Test otomatis:** `backend/internal/rules/rules_test.go` › `TestRuleFixtures/TC-RULE-T04-1` (dibuat di M1)
- **Status:** lulus

## TC-RULE-W01-1 — Menganalisis step rework menghilangkan temuan demo.

- **Level:** integrasi (database) · **Kebutuhan:** aturan W01 (`docs/06-rules.md` §4)
- **Prasyarat:** data demo (`db/seed/demo.sql`), paket PS-07, `@today` = 2026-10-08.
- **Langkah:**
  1. Buka transaksi, jalankan `backend/internal/rules/testdata/W01.sql`.
  2. Jalankan `backend/internal/rules/sql/W01.sql` dengan `@package_id` = PS-07.
  3. Rollback transaksi.
- **Hasil yang diharapkan:** aturan W01 tidak mengembalikan hasil apa pun (temuan demo hilang).
- **Test otomatis:** `backend/internal/rules/rules_test.go` › `TestRuleFixtures/TC-RULE-W01-1` (dibuat di M1)
- **Status:** lulus

## TC-RULE-W02-1 — Final inspection tanpa cabang NG.

- **Level:** integrasi (database) · **Kebutuhan:** aturan W02 (`docs/06-rules.md` §4)
- **Prasyarat:** data demo (`db/seed/demo.sql`), paket PS-07, `@today` = 2026-10-08.
- **Langkah:**
  1. Buka transaksi, jalankan `backend/internal/rules/testdata/W02.sql`.
  2. Jalankan `backend/internal/rules/sql/W02.sql` dengan `@package_id` = PS-07.
  3. Rollback transaksi.
- **Hasil yang diharapkan:** aturan W02 mengembalikan tepat 1 hasil:
  - `process_steps` · `ngFlow` dengan opNo = "90"
- **Test otomatis:** `backend/internal/rules/rules_test.go` › `TestRuleFixtures/TC-RULE-W02-1` (dibuat di M1)
- **Status:** lulus

## TC-RULE-W04-1 — Review terbaru menghilangkan temuan demo.

- **Level:** integrasi (database) · **Kebutuhan:** aturan W04 (`docs/06-rules.md` §4)
- **Prasyarat:** data demo (`db/seed/demo.sql`), paket PS-07, `@today` = 2026-10-08.
- **Langkah:**
  1. Buka transaksi, jalankan `backend/internal/rules/testdata/W04.sql`.
  2. Jalankan `backend/internal/rules/sql/W04.sql` dengan `@package_id` = PS-07.
  3. Rollback transaksi.
- **Hasil yang diharapkan:** aturan W04 tidak mengembalikan hasil apa pun (temuan demo hilang).
- **Test otomatis:** `backend/internal/rules/rules_test.go` › `TestRuleFixtures/TC-RULE-W04-1` (dibuat di M1)
- **Status:** lulus
