# Sistem PFD · PFMEA · Control Plan — Paket spesifikasi Tahap 1

> Dokumen ini dulunya `README.md` di root repository (isi paket spesifikasi yang menjadi titik
> awal project). Sejak 9 Oktober 2026 `README.md` menjelaskan aplikasi web yang dikembangkan,
> dan penjelasan paket dipindahkan ke sini tanpa perubahan isi.

Paket ini adalah titik awal repository untuk membangun aplikasi web pembuatan, pengecekan dan
monitoring **PFD, PFMEA dan Control Plan** dengan bantuan Claude Code. Isinya ditulis supaya AI
bisa langsung bekerja: hasil akhir yang diinginkan (4 tahap), lingkup Tahap 1 yang rinci,
kontrak API, skema database yang sudah diuji, 32 aturan cek konsistensi beserta data uji, dan
urutan pekerjaan (milestone) dengan kriteria selesai yang bisa diperiksa.

## Isi paket

| File / folder | Fungsi |
| --- | --- |
| `CLAUDE.md` | Instruksi utama untuk Claude Code (dibaca otomatis setiap sesi) |
| `docs/00-vision.md` | Hasil akhir setelah Tahap 4, roadmap, gate, kebutuhan non-fungsional |
| `docs/01-domain-glossary.md` | Penjelasan domain PFD/PFMEA/CP dan pemetaan kolom form AIAG ke data |
| `docs/02-phase1-scope.md` | Lingkup Tahap 1, matriks hak akses, user story, kriteria Gate 1 |
| `docs/03-architecture.md` | Stack (Go + PostgreSQL 18 + SvelteKit), struktur repo, target performa, alasan keputusan |
| `docs/04-data-model.md` | Penjelasan skema database |
| `docs/05-api.md` + `api/openapi.yaml` | Aturan dan kontrak API |
| `docs/06-rules.md` | Mesin aturan dan 32 aturan Tahap 1 |
| `docs/07-template-general.md` | Template General: buat paket, rilis, sinkron, override, konflik |
| `docs/08-screens.md` + `docs/mockups/` | Perilaku layar dan mockup |
| `docs/09-excel-export.md` | Ekspor Excel A3 |
| `docs/10-milestones.md` | Urutan kerja M0–M13 dengan checklist |
| `docs/11-testing.md` | Aturan test-first, strategi uji, uji performa, checklist Gate 1 |
| `docs/test-cases/` | Test case yang ditulis **sebelum** kode (format, template, dan test case 32 aturan) |
| `docs/prompts/start-project.md` | Prompt untuk sesi pertama Claude Code (GitHub, panduan tahapan, M0) |
| `db/migrations/00001_init.sql`, `db/seed/demo.sql` | Skema database dan data demo |
| `backend/internal/rules/sql`, `.../testdata` | 32 query aturan dan data ujinya |
| `.claude/` | Aturan per folder, skill (`/milestone`, `/add-rule`, `fmea-domain`) dan agen `spec-reviewer` |

## Hasil akhir yang dituju

| Tahap | Isi | Gate |
| --- | --- | --- |
| **1** (paket ini) | Master data, paket, Template General (rilis + sinkron otomatis), PFD tabel + diagram, PFMEA AIAG 4th dengan RPN, Control Plan Template A, 32 aturan, ekspor Excel A3, dashboard, tempel dari Excel | Satu paket nyata dibuat ulang di sistem dan ekspor Excel-nya sama dengan Excel lama |
| 2 | AIAG-VDA (struktur 4M, AP), CP-1 (Template B, Safe Launch), 42 aturan, review & approval, revisi, PDF, impor Excel lama | Satu paket AIAG-VDA lolos review QA tanpa Error |
| 3 | Pelacakan aksi + bukti, notifikasi e-mail, KPI dan tren, matriks keterlacakan | Semua paket aktif termonitor di dashboard |
| 4 | Template per varian/family, konversi 4th → AIAG-VDA, ekspor untuk AI, data NG dari lini | Diputuskan setelah Tahap 3 |

## Cara memakai dengan Claude Code

1. Siapkan komputer pengembang: Git, GitHub CLI (`gh`) yang sudah login (`gh auth login`),
   Docker, Go 1.27, Node.js 24 LTS (minimal 22.17), dan `make` (di Windows pakai WSL2).
2. Ekstrak paket ini ke folder kosong dan buka folder itu dengan Claude Code. `CLAUDE.md`
   dibaca otomatis.
3. Sesi pertama: salin teks prompt dari `docs/prompts/start-project.md` (semua teks di bawah
   garis; ganti dulu `OWNER/pfmea-system` di baris pertamanya dengan repository GitHub Anda).
   Claude akan menghubungkan folder ke GitHub, mem-push semua isi ke branch `main`, menulis
   `docs/build-guide.md` (rangkuman tahapan, mulai dari instalasi awal), lalu mengerjakan M0
   dengan cara test-first dan mem-push hasilnya sampai CI hijau. Daftar test case M0 akan
   ditunjukkan dulu untuk Anda setujui.
4. Sesi berikutnya: ketik `/milestone M1`, lalu `/milestone M2`, dan seterusnya. **Satu
   milestone per sesi.** Setelah selesai, coba hasilnya (`make dev`, buka
   http://localhost:5173), baru lanjut ke milestone berikutnya di sesi baru.
5. Jika Claude bertanya karena spesifikasi kurang jelas, jawab, lalu minta ia mencatat
   keputusan itu di dokumen yang relevan supaya sesi berikutnya ikut tahu.
6. Jangan izinkan pengujian dilemahkan atau dihapus hanya agar lulus; kriteria selesai di
   `docs/10-milestones.md` adalah ukuran kemajuan.

## Aturan kerja yang selalu berlaku

- **Test case dulu.** Setiap fitur baru, pembaruan atau perbaikan bug diawali dengan menulis
  test case di `docs/test-cases/`, di-commit dan di-push, baru kemudian test otomatis dan kode.
- **Semua masuk ke `main` di GitHub.** Satu branch (`main`), commit kecil, push setiap langkah
  yang `make check`-nya hijau, tidak ada force-push. Test yang masih gagal tidak di-commit
  sendirian, tetapi bersama kode yang membuatnya lulus, jadi `main` selalu hijau. GitHub Actions
  menjalankan pengujian di setiap push; sesi dianggap selesai bila CI hijau.
- **Bahasa:** tampilan aplikasi, nama file dan penamaan lain berbahasa Inggris; semua kode
  diberi komentar bahasa Indonesia (di awal file, di setiap fungsi, tipe dan test, dan di
  langkah yang tidak jelas).

## Yang perlu disiapkan perusahaan

| Input | Dibutuhkan saat |
| --- | --- |
| Daftar customer, tabel konversi simbol, ambang CSR (RPN, S minimum untuk CC) | M3 |
| Teks kriteria S/O/D AIAG 4th dari manual berlisensi | M3 |
| Library kontrol dengan rentang D per metode deteksi | M8 |
| Daftar proses general (wajib/opsional) beserta isi PFD/PFMEA/CP-nya | M9 |
| Form Excel PFD (SH), PFMEA (SF) dan CP (SC) yang dipakai sekarang | M10 |
| Satu paket nyata (Excel) untuk Gate 1 | M12 |
| Server, jaringan, sertifikat TLS, detail Active Directory (jika dipakai) | M13 |

Sebelum input asli tersedia, pengembangan dan pengujian memakai data demo.

## Data demo (hanya untuk pengembangan)

Pengguna: `admin`, `rsaputri` (approver), `apratama` dan `dhidayat` (author), `swulandari`
(reviewer), `operator1` (viewer). Kata sandi semua: `pfmea-dev-2026` — **jangan pernah dipakai di
server produksi.** Data demo berisi Template General rev 1 dan paket model PS-07 (Customer B)
yang sengaja mengandung 14 temuan agar mesin aturan bisa diuji. "Hari ini" pada data demo =
8 Oktober 2026.

## Keputusan yang masih terbuka

- [ ] Login Active Directory di Tahap 1 atau akun lokal saja.
- [ ] Pola nomor dokumen (bawaan `SF-{customer}-{seq:03}`).
- [ ] Pemilik Template General dan siapa yang menyetujui rilis template.
- [ ] Paket nyata mana yang dipakai untuk Gate 1.
- [ ] Spesifikasi server dan tujuan backup.

## Catatan bahasa

| Bagian | Bahasa |
| --- | --- |
| Seluruh tampilan aplikasi (label, pesan, error, pesan aturan, file Excel, nama file) | Inggris (teks UI di `web/src/lib/i18n/en.ts`) |
| Komentar di semua kode yang ditulis tangan | Indonesia |
| Nama file, folder, variabel, fungsi, tabel, commit message | Inggris |
| Dokumen spesifikasi untuk AI (`CLAUDE.md`, `docs/00`–`11`) | Inggris, karena lebih presisi untuk kode dan istilah AIAG |
| Dokumen untuk tim (`README.md`, `docs/build-guide.md`, `docs/test-cases/`) | Indonesia |

Mockup di `docs/mockups/` masih memakai label bahasa Indonesia dan hanya menjadi acuan tata
letak; label bahasa Inggris yang dipakai aplikasi ada di `docs/08-screens.md`.

## Status verifikasi paket

- Migrasi naik/turun, data demo, perilaku trigger, 32 aturan dan 32 data uji sudah dijalankan
  di PostgreSQL 16 dengan fungsi pengganti `uuidv7()`; M1 mengulanginya di PostgreSQL 18.
  Hasil: paket GENERAL 0 temuan, PS-07 tepat 14 temuan (pesan temuan sudah bahasa Inggris).
- `api/openapi.yaml` (90 operasi) lolos `redocly lint`, dan tipe TypeScript-nya berhasil dibuat
  serta dicek dengan TypeScript 6. Generator Go (oapi-codegen) dijalankan pertama kali di M2.
