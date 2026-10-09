# M<nn> · <Nama milestone> — Test case

Milestone: M<nn> · Kebutuhan: P1-xx · User story: US-xx
Ditulis sebelum kode pada: <tanggal> · Status terakhir: 0 dari N lulus

## Ringkasan

| ID | Judul | Level | Status |
| --- | --- | --- | --- |
| TC-M<nn>-001 | <apa yang dibuktikan> | E2E | belum dibuat |

## TC-M<nn>-001 — <judul singkat>

- **Level:** E2E (Playwright) · **Kebutuhan:** US-xx, P1-xx
- **Prasyarat:** data demo (`db/seed/demo.sql`); login sebagai `apratama` (author); paket
  PS-07 terbuka di halaman PFD.
- **Langkah:**
  1. Klik "+ Add step", isi op no `30`, nama `Solder paste printing`, simbol Operation.
  2. Simpan.
- **Hasil yang diharapkan:**
  - Step 30 muncul di tabel di antara step 25 dan 50; diagram tergambar ulang tanpa reload.
  - Pengguna kedua yang membuka paket yang sama melihat step baru dalam waktu < 1 detik.
- **Test otomatis:** `web/tests/e2e/us02-edit-pfd.spec.ts` › `TC-M<nn>-001 add step 30 between 25 and 50`
- **Status:** belum dibuat
