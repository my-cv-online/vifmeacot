# Panduan pembangunan — Sistem PFD · PFMEA · Control Plan

Panduan ini merangkum cara membangun aplikasi dari instalasi awal sampai tahap-tahap berikutnya.
Sumbernya adalah spesifikasi di `docs/00`–`docs/11` dan `CLAUDE.md`; bila ada perbedaan,
spesifikasi itulah yang berlaku. Tabel milestone (§4) diperbarui di setiap milestone.

## 1. Gambaran singkat

Kita membangun aplikasi web untuk pabrik elektronik otomotif (PCBA, IATF 16949) yang dipakai
untuk membuat, mengecek dan memonitor tiga dokumen mutu per produk: **PFD**, **PFMEA** dan
**Control Plan**. Ketiganya adalah tiga tampilan dari satu data proses: langkah, karakteristik,
kelas karakteristik khusus dan kontrol dimasukkan sekali lalu tampil *read-only* di dokumen
lain. Hal yang masih tidak konsisten ditangkap otomatis oleh mesin aturan, dan dashboard
menunjukkan kondisi setiap paket. Aplikasi berjalan on-premise: satu binary Go, satu database
PostgreSQL 18, dan SPA SvelteKit yang tertanam di binary. **Tahap 1** berisi master data, paket,
Template General (rilis + sinkron otomatis), editor PFD (tabel + diagram), worksheet PFMEA
AIAG 4th dengan RPN, Control Plan Template A, 32 aturan cek, ekspor Excel A3, dashboard dan
tempel dari Excel. Tahap 2–4 dikerjakan berurutan; setiap tahap baru dimulai setelah gate tahap
sebelumnya lulus.

| Tahap | Isi | Gate |
| --- | --- | --- |
| **1** (sekarang) | Master data, paket, Template General (rilis + sinkron), PFD tabel + diagram, PFMEA AIAG 4th dengan RPN, CP Template A, 32 aturan, ekspor Excel A3, dashboard, tempel dari Excel | Satu paket nyata dibuat ulang di sistem dan ekspor Excel-nya sama dengan Excel lama |
| 2 | AIAG-VDA (struktur 4M, Action Priority), CP Template B (CP-1, Safe Launch), reaction plan terstruktur dengan pemilik, 42 aturan, review & approval, revisi dan propagasi ke paket yang sudah dirilis, PDF, impor Excel lama | Satu paket AIAG-VDA lolos review QA tanpa Error |
| 3 | Pelacakan aksi + bukti, notifikasi e-mail, KPI dan tren bulanan, matriks keterlacakan, diff revisi, pemicu out-of-sync dari perubahan PFD/WI | Semua paket aktif termonitor di dashboard |
| 4 | Template per varian/family, konversi 4th → AIAG-VDA, ekspor Markdown/JSON untuk AI, data NG dari lini, editor PFD drag-and-drop | Diputuskan setelah Tahap 3 |

## 2. Instalasi awal project

### 2.1 Prasyarat

| Alat | Versi | Catatan |
| --- | --- | --- |
| Git | — | Atur identitas: `git config --global user.name "Nama Anda"` dan `git config --global user.email "email@perusahaan"` |
| GitHub CLI (`gh`) | — | `gh auth login`; token wajib punya scope `repo` dan `workflow` (`gh auth refresh -h github.com -s workflow`), lalu `gh auth setup-git` supaya `git push` memakai login `gh` |
| Docker | Docker Engine atau Docker Desktop dengan Docker Compose v2 (perintah `docker compose`) | Untuk PostgreSQL 18 (`make db`) dan untuk test database (testcontainers, mulai M1) |
| Go | 1.27 atau lebih baru | Versi minimum ditulis di `go.mod` |
| Node.js | 24 LTS (minimal 22.17) | Untuk `web/` (SvelteKit) |
| `make` | GNU Make | Semua perintah project dijalankan lewat `Makefile` |
| Compiler C (`gcc`) | — | Dibutuhkan `go test -race` (cgo) yang dijalankan `make test`; di Ubuntu/WSL2: `sudo apt install build-essential` (sekaligus memasang `make`) |
| Windows | WSL2 (misalnya Ubuntu) | Jalankan semua perintah di dalam WSL2 |

Periksa dengan:

```bash
git --version && gh auth status && docker info && docker compose version
go version && node --version && make --version && gcc --version
git config user.name && git config user.email
```

### 2.2 Langkah instalasi

| # | Perintah | Hasil | Tersedia mulai |
| --- | --- | --- | --- |
| 1 | `gh repo clone my-cv-online/vifmeacot && cd vifmeacot` (atau `git clone https://github.com/my-cv-online/vifmeacot.git && cd vifmeacot`) | Salinan repository di komputer | — |
| 2 | `make tools` | Alat Go yang versinya dipatok (sqlc, goose, oapi-codegen, golangci-lint, dan air untuk live reload `make dev`) terpasang di `bin/` project, dependensi web terpasang dengan `npm ci` di `web/`; mulai M2 juga Playwright Chromium. Pertama kali ±3–5 menit; berikutnya hanya alat yang versinya berubah yang dipasang ulang | M0 |
| 3 | `cp .env.example .env` | File konfigurasi lokal (tidak pernah di-commit); wajib sebelum `make dev` | M0 |
| 4 | `make db` | PostgreSQL 18 berjalan di Docker (dengan healthcheck dan volume data), hanya terbuka di `127.0.0.1:5432`; perintah selesai setelah database sehat | M0 |
| 5 | `make migrate` | Skema database terpasang (`db/migrations`); mencetak `applied 00001_init.sql`, atau `no pending migrations` bila sudah terpasang. Hanya butuh `DATABASE_URL` dari `.env` | M1 |
| 6 | `make seed` | Data demo dimuat (`db/seed/demo.sql`): `loaded demo data: 6 users, 2 packages`. Butuh `DEV_MODE=true` (sudah di `.env.example`); ditolak bila database sudah berisi data. Mulai ulang dari nol: `make seed RESET=1` (skema dihapus, migrasi dijalankan ulang, data demo dimuat; restart `make dev` sesudahnya) | M1 |
| 7 | `make dev` | Server Go di http://localhost:8080 (dibangun ulang otomatis bila file Go berubah) dan Vite di http://localhost:5173 (proxy `/api` ke :8080); hentikan dengan Ctrl+C | M0 |
| 8 | Buka http://localhost:5173 | M0: kerangka aplikasi (app shell) dengan halaman placeholder; mulai M2: halaman login | M0 |
| 9 | `make check` | Semua pemeriksaan hijau (wajib sebelum setiap commit) | M0 |

Cek server berjalan: `curl -i localhost:8080/healthz` → `200`. Cek server siap:
`curl -i localhost:8080/readyz` → `200` bila database terjangkau dan sudah dimigrasi sampai versi
terbaru (mulai M1), `503` bila belum (detailnya di log server).

Perintah database lain (mulai M1; `.env` dimuat otomatis oleh `make`):

```bash
make migrate-down                                    # hapus semua tabel dan data (butuh DEV_MODE=true)
set -a && . ./.env && set +a                         # muat .env ke shell untuk perintah di bawah
go run ./backend/cmd/pfmea migrate status            # daftar migrasi: applied/pending dan versi skema
```

Versi produksi lokal (binary dengan SPA tertanam, tersedia mulai M0). `./bin/pfmea serve`
membaca konfigurasi dari environment variable; `DATABASE_URL` dan `APP_BASE_URL` wajib di-set
(`docs/03-architecture.md` §8). Contoh dengan memuat `.env`:

```bash
make build                                             # build web → backend/internal/webui/dist → go build
set -a && . ./.env && set +a                           # muat variabel dari .env ke shell
APP_BASE_URL=http://localhost:8080 ./bin/pfmea serve   # buka http://localhost:8080
```

`APP_BASE_URL` di `.env` berisi alamat Vite (http://localhost:5173) untuk `make dev`; saat
membuka binary langsung di :8080 nilainya ditimpa seperti di atas, karena mulai M2 cek `Origin`
menolak request dari alamat lain (403).

### 2.3 Pengguna demo

Login tersedia mulai M2, data demo mulai M1. Semua pengguna demo memakai kata sandi
`pfmea-dev-2026` — **hanya untuk pengembangan, jangan pernah dipakai di server produksi.**

| Pengguna | Peran |
| --- | --- |
| `admin` | admin |
| `rsaputri` | approver |
| `apratama`, `dhidayat` | author |
| `swulandari` | reviewer |
| `operator1` | viewer |

Data demo berisi Template General rev 1 dan paket model PS-07 (Customer B) yang sengaja
mengandung 14 temuan agar mesin aturan bisa diuji. "Hari ini" pada data demo adalah
8 Oktober 2026.

## 3. Alur kerja test-first

Setiap perubahan (fitur baru, pembaruan, perbaikan bug, aturan baru) mengikuti urutan ini:

1. **Tulis test case** di `docs/test-cases/` (format di `docs/test-cases/README.md`, template
   `_template.md`, isi bahasa Indonesia): ID, level, kebutuhan, prasyarat, langkah, hasil yang
   diharapkan, nama test otomatis, status `belum dibuat` (alur status: `belum dibuat` →
   `gagal` → `lulus`).
2. **Commit dan push** test case itu lebih dulu, misalnya `test(M5): define test cases`.
3. **Tulis test otomatis.** Nama test memuat ID test case (`TestCreatePackage_TC_M04_001`,
   `test('TC-M04-001 …')`), dengan komentar bahasa Indonesia di atasnya.
4. **Jalankan dan pastikan gagal** karena alasan yang benar (fitur belum ada, bukan salah
   ketik); ubah status test case menjadi `gagal`. Test yang gagal **tidak** di-commit
   sendirian.
5. **Tulis kode** sampai test lulus. Test tidak boleh dilemahkan atau dihapus supaya lulus.
6. **`make check`** harus hijau.
7. **Perbarui status test case menjadi `lulus`, lalu commit test, kode dan file test case
   bersama-sama** (Conventional Commits bahasa Inggris).
8. **Push** ke `main`.
9. **Tunggu CI hijau** di GitHub Actions (`gh run list`, `gh run watch <id> --exit-status`).

Contoh untuk empat jenis perubahan:

| Jenis | File test case | ID | Catatan |
| --- | --- | --- | --- |
| Fitur milestone | `docs/test-cases/M05-pfd-editor.md` | `TC-M05-001`, `TC-M05-002`, … | Satu file per milestone; ditulis di awal sesi milestone (`/milestone M5`) |
| Pembaruan fitur yang sudah ada | File milestone pemilik fitur, misalnya `M05-pfd-editor.md` | Lanjutkan nomornya, misalnya `TC-M05-013` | ID tidak pernah dipakai ulang atau dinomori ulang; case yang tidak berlaku ditandai "dihapus" beserta alasannya |
| Perbaikan bug | `docs/test-cases/bugs.md` | `TC-BUG-001`, … | Bug direproduksi dulu dengan test yang gagal |
| Aturan cek baru atau berubah | `docs/test-cases/rules.md` (pakai skill `/add-rule`) | `TC-RULE-<KODE>-1`, misalnya `TC-RULE-K01-1` | Test otomatisnya adalah fixture `backend/internal/rules/testdata/<KODE>.sql`; baseline demo (PS-07 = 14 temuan, GENERAL = 0) harus tetap hijau |

## 4. Milestone M0–M13

Satu milestone per sesi Claude Code (`/milestone M<n>`), dikerjakan berurutan. Rincian
"Deliver" dan "Done when" setiap milestone ada di `docs/10-milestones.md`. Nama file test case
mengikuti pola `M<nn>-<slug>.md`; nama di tabel untuk milestone yang belum dimulai adalah
rencana dan ditetapkan saat milestone itu dikerjakan.

| Milestone | Isi singkat | File test case | User story E2E | Input dari perusahaan | Status |
| --- | --- | --- | --- | --- | --- |
| M0 Scaffold | Repository kosong yang bisa dijalankan: `go.mod`, `pfmea serve` (`/healthz`, `/readyz`), konfigurasi, Makefile, Docker Compose, CI, kerangka SvelteKit, SPA tertanam | `M00-scaffold.md` | — (unit dengan Vitest) | — | selesai (9 Oktober 2026) |
| M1 Database | Migrasi (goose), seed demo, `store.WithTx`, sqlc, harness test database, test trigger, `make test-rules` | `M01-database.md`, `rules.md` | — | — | sedang dikerjakan (review) |
| M2 API dan auth | `make gen`, kerangka API hasil generator (operasi lain 501), login, sesi, pengguna, `pfmea init`, halaman login | `M02-api-auth.md` | Login/logout `apratama` (Playwright) | — | belum mulai |
| M3 Master data | Customer, kelas, part, library kontrol, kriteria, istilah terlarang, aturan, pengaturan; layar `/master/*` | `M03-master-data.md` | Admin mengedit tabel simbol Customer B (Playwright) | Daftar customer, tabel konversi simbol, ambang CSR (RPN, S minimum untuk CC); teks kriteria S/O/D AIAG 4th dari manual berlisensi | belum mulai |
| M4 Paket | Daftar, buat dari Template General, nomor dokumen, ringkasan paket, sidebar | `M04-packages.md` | US-01 (tanpa klausul cek) | — | belum mulai |
| M5 Editor PFD | `domain/entities.go`, mesin `/changes`, editor PFD + diagram, realtime | `M05-pfd-editor.md` | US-02, US-03, US-07 (PFD) | — | belum mulai |
| M6 PFMEA AIAG 4th | Worksheet PFMEA, effect/kontrol/aksi, `pfmea perf-gen` | `M06-pfmea-worksheet.md` | US-04, US-07 (chain dan field virtual) | — | belum mulai |
| M7 Control Plan | CP Template A, "+ Lines from PFMEA controls" | `M07-control-plan.md` | US-06 (tanpa R01) | — | belum mulai |
| M8 Mesin aturan | Loader, eksekutor, temuan, River, waive, quick fix, layar temuan | `M08-rule-engine.md`, `rules.md` | US-01 (cek), US-06 (R01), US-08, US-09 | Library kontrol dengan rentang D per metode deteksi | belum mulai |
| M9 Template General | `PlanSync`, rilis, sinkron, konflik, override, restore | `M09-template-general.md` | US-10, US-11 | Daftar proses general (wajib/opsional) beserta isi PFD/PFMEA/CP-nya | belum mulai |
| M10 Ekspor Excel | Layout, writer PFD/PFMEA/CP, cache, `pfmea xlsx-compare` | `M10-excel-export.md` | US-12 | Form Excel PFD (SH), PFMEA (SF) dan CP (SC) yang dipakai sekarang | belum mulai |
| M11 Dashboard | Dashboard, aktivitas, pencarian | `M11-dashboard.md` | US-13 | — | belum mulai |
| M12 Tempel dari Excel | `pasteRows`, dialog tempel di PFD, PFMEA, CP | `M12-bulk-paste.md` | US-05 | Satu paket nyata (Excel) untuk Gate 1 | belum mulai |
| M13 Hardening | `make perf`, `deploy/`, LDAP, `/metrics`, bagian operasional README, Gate 1 | `M13-hardening.md` | — (uji performa, checklist Gate 1) | Server, jaringan, sertifikat TLS, detail Active Directory (jika dipakai) | belum mulai |

Setelah milestone selesai, kolom Status diisi "selesai (<tanggal>)".

Keputusan 9 Oktober 2026: kolom "Input dari perusahaan" diisi dengan **data dummy** yang dirancang
di milestone tersebut (fiktif, tidak menyalin teks manual AIAG berlisensi). Repository tetap
public dan tidak pernah berisi data asli perusahaan.

## 5. Gate 1 (penerimaan Tahap 1)

Tahap 1 selesai bila semua kriteria ini terpenuhi (`docs/11-testing.md` §6):

| # | Kriteria | Cara verifikasi | Pemilik |
| --- | --- | --- | --- |
| 1 | Satu paket nyata dibuat ulang di sistem | QA memilih paketnya; engineer membangunnya ulang dengan tempel dari Excel dan editor | QA + process engineering |
| 2 | Ekspor Excel sama dengan file lama | Laporan `pfmea xlsx-compare` tanpa perbedaan yang tidak bisa dijelaskan; QA menandatangani layout A3 yang dicetak | QA |
| 3 | Ke-32 aturan benar pada paket itu | QA meninjau setiap temuan (benar/salah) dan masalah yang terlewat; perbaikannya menjadi fixture aturan | QA |
| 4 | Template General asli dirilis dan disinkronkan | Template berisi proses general perusahaan; rilis menunjukkan semua paket `done` | QA Manager (approver) |
| 5 | Target performa tercapai | `perf-report.json` dari server pabrik (profil perf, database `pfmea_perf`), semua target hijau | IT |
| 6 | Backup berhasil dipulihkan di mesin kedua | Log `deploy/restore.sh` dan login di salinan yang dipulihkan | IT |

## 6. Menjalankan test

| Perintah | Isi | Tersedia mulai |
| --- | --- | --- |
| `make check` | `make gen` (tidak boleh mengubah file yang di-track) + `make lint` + `make test` + `make test-rules`; wajib hijau sebelum setiap commit dan push | M0 |
| `make test` | Test Go (unit dan integrasi, dengan `-race`; tanpa fixture aturan) + Vitest. Test database butuh Docker (lihat di bawah) | M0 |
| `make lint` | golangci-lint, `svelte-check`, eslint, `redocly lint` (`docs/03-architecture.md` §3.3) | M0 |
| `make test-rules` | Fixture 32 aturan (hanya SQL aturan; pesan temuan diperiksa mulai M8) + baseline (PS-07 = 14 temuan, GENERAL = 0) | M1 |
| `make e2e` | Playwright (Chromium) untuk user story US-01 … US-13 | M2 |
| `make perf` | Target performa `docs/03-architecture.md` §9 | M13 |

Sebelum milestone-nya, target yang belum tersedia mencetak "available from M<n>" dan selesai
dengan kode 0.

**Database untuk test (mulai M1).** Test yang memakai database menjalankan PostgreSQL 18 sendiri
di Docker (testcontainers) sekali per paket Go yang diuji, membangun database template berisi
skema dan data demo sekali, lalu memberi setiap test salinannya sendiri yang dihapus setelah
test selesai. Database `make db` tidak dipakai dan tidak diubah. Docker harus berjalan; tanpa
Docker test database **gagal** (tidak dilewati). Bila Docker tidak bisa dipakai, arahkan test ke
server PostgreSQL 18 lain dengan user yang boleh membuat database:
`TEST_DATABASE_URL=postgres://user:sandi@host:5432/postgres make test`.

Menjalankan satu test berdasarkan ID (dari root repository; tanda kurung membuat `cd` tidak
mengubah folder terminal Anda):

```bash
go test ./backend/... -run TC_M00_001 -v                                    # test Go (nama fungsi memuat TC_M00_001)
go test ./backend/internal/rules/ -run 'TestRuleFixtures/TC-RULE-K01-1' -v   # subtest: sebutkan juga nama test induknya
(cd web && npx vitest run -t "TC-M00-018")                                   # test Vitest
(cd web && npx playwright test -g "TC-M04-001")                              # test Playwright (mulai M2)
```

Jika keluaran `go test` berisi `[no tests to run]`, pola tidak cocok dengan test mana pun — itu
bukan tanda lulus. Test Playwright memakai binary hasil `make build` dan database khusus
(`docs/11-testing.md` §4): jalankan `make db` dan `make build` dulu setiap kali kode Go atau web
berubah.

## 7. Git dan GitHub

- **Satu branch: `main`**, remote `origin` (https://github.com/my-cv-online/vifmeacot). Tidak ada
  branch lain dan tidak ada pull request.
- **Format commit:** Conventional Commits berbahasa Inggris, kecil dan fokus: `feat:`, `fix:`,
  `test:`, `docs:`, `chore:`, `refactor:`, `ci:` (misalnya `test(M5): define test cases`,
  `feat(M5): PFD editor with realtime updates`).
- **Kapan push:** setiap langkah yang `make check`-nya hijau langsung di-commit dan di-push.
  Test yang masih gagal tidak di-commit sendirian, jadi `main` di GitHub tidak pernah merah.
  Commit yang hanya berisi dokumen boleh di-push tanpa `make check` hanya sebelum Makefile ada
  (awal M0).
- **Dilarang:** force-push, mengubah commit yang sudah di-push, membuat branch lain,
  meng-commit rahasia (`.env`, token, kata sandi asli) atau hasil build, serta meng-commit data
  perusahaan asli (form Excel, paket nyata, teks kriteria S/O/D berlisensi): repository ini
  public dan hanya memakai data dummy (lihat `README.md`). Teks kriteria berlisensi tidak pernah masuk ke kode, seed
  atau test.
- **Melihat CI:** GitHub Actions menjalankan `make tools` dan `make check` di setiap push ke
  `main` (mulai M2 juga `make e2e`).

  ```bash
  gh run list --limit 5                     # run terakhir
  gh run watch <id> --exit-status           # tunggu sampai selesai
  gh run view <id> --log-failed             # baca log langkah yang gagal
  ```

- **Push ditolak** (ada commit baru di GitHub): jalankan `git pull --rebase --autostash`,
  selesaikan bentrok bila ada (`git add <file>` lalu `git rebase --continue`), jalankan
  `make check`, lalu `git push` lagi. Jangan pernah force-push.
- Sesi kerja selesai bila working tree bersih, `git log origin/main..main` kosong, dan run CI
  terakhir hijau. Pengecualian: bila sesi harus berhenti di tengah milestone, push yang sudah
  hijau, biarkan pekerjaan yang belum selesai tetap tidak di-commit (jangan dibuang) dan catat
  di laporan; sesi berikutnya melanjutkannya.

## 8. Aturan bahasa

| Bagian | Bahasa |
| --- | --- |
| Semua yang dilihat pengguna: label UI, pesan, error, pesan aturan cek, teks buatan aplikasi di file Excel, nama file ekspor | Inggris (`web/src/lib/i18n/en.ts`, `backend/internal/i18n/en.go`, header SQL aturan). Isi FMEA tetap dalam bahasa yang ditulis perusahaan (`docs/00-vision.md` §6) |
| Komentar di semua file yang ditulis tangan (Go, TypeScript, JavaScript, Svelte, HTML, CSS, SQL, Makefile, YAML, Dockerfile, shell) | Indonesia |
| Identifier, nama file dan folder, field API, nama database, pesan log, nama test dan pesan kegagalan test, output perkakas (`make help`, CLI), commit message | Inggris |
| Dokumen spesifikasi untuk AI (`CLAUDE.md`, `docs/00`–`docs/11`) | Inggris |
| Dokumen untuk tim (`README.md`, `docs/build-guide.md`, `docs/spec-package.md`, `docs/test-cases/`) | Indonesia (nama file bahasa Inggris) |

Komentar ditulis di mana-mana: komentar singkat di awal setiap file tentang gunanya, komentar
untuk setiap fungsi, tipe, konstanta dan test, serta komentar pada langkah yang tidak jelas.
Jelaskan alasannya, jangan mengulang isi kode. Doc comment Go diawali nama identifier, misalnya
`// LoadConfig membaca variabel lingkungan lalu memvalidasinya.` Kode hasil generator (`gen/`,
`sqlc/`, `schema.d.ts`) dan file JSON dikecualikan; teks `description` di `api/openapi.yaml`
tetap bahasa Inggris karena menjadi dokumentasi API.

## 9. Pemecahan masalah

| Masalah | Penyebab / solusi |
| --- | --- |
| `make db` gagal: *Cannot connect to the Docker daemon* | Docker belum berjalan. Nyalakan Docker Desktop (Windows/macOS; di Windows aktifkan integrasi WSL2 untuk distro Anda) atau `sudo systemctl start docker` (Linux; di WSL2 tanpa systemd: `sudo service docker start`), lalu ulangi. Cek dengan `docker info`. |
| Port 5432 sudah terpakai | Ada PostgreSQL lain di komputer. Hentikan (`sudo systemctl stop postgresql`, atau `sudo service postgresql stop` di WSL2 tanpa systemd), atau jalankan database project di port lain: ubah `PG_PORT` dan port di `DATABASE_URL` pada `.env`, lalu `make db` lagi. Cari pemakai port dengan `sudo lsof -i :5432` atau `sudo ss -ltnp 'sport = :5432'`. |
| Port 8080 sudah terpakai | Cari PID-nya dengan `sudo lsof -i :8080` (atau `sudo ss -ltnp 'sport = :8080'`), lalu hentikan dengan `kill <PID>`. Atau ubah `HTTP_ADDR` di `.env` (misalnya `:8081`); proxy `/api` Vite di `make dev` membaca `HTTP_ADDR` yang sama. |
| Port 5173 sudah terpakai | Biasanya `make dev` lama masih berjalan; Vite lalu pindah ke 5174 dan mencetak alamat barunya. Hentikan yang lama (Ctrl+C di terminalnya, atau `sudo lsof -i :5173` lalu `kill <PID>`) dan jalankan `make dev` lagi. |
| `make check` gagal karena `make gen` mengubah file | Hasil generator di repository tidak sama dengan kontrak/query. Jalankan `make gen`, periksa `git diff`, lalu commit file hasil generator bersama perubahan kontrak/query-nya. File hasil generator tidak boleh diedit manual. |
| Test Go gagal dengan *testdb: … Cannot connect to the Docker daemon* atau *failed to start PostgreSQL* | Test database butuh Docker. Nyalakan Docker (lihat baris `make db` di atas) lalu ulangi, atau pakai `TEST_DATABASE_URL` (bagian 6). Pertama kali image `postgres:18` dan `testcontainers/ryuk` diunduh, jadi butuh internet atau image yang sudah ada. |
| Test lama di mesin lambat, container test tertinggal | Container test dihapus otomatis setelah `go test` selesai (Ryuk menghapusnya walaupun proses test dihentikan paksa). Daftar yang tertinggal: `docker ps --filter label=org.testcontainers=true`. |
| `make migrate` / `make seed` gagal: *connection refused* | Database belum berjalan atau port di `DATABASE_URL` salah. Jalankan `make db`, cek `PG_PORT` dan `DATABASE_URL` di `.env`. |
| `make seed` ditolak: *requires DEV_MODE=true* | Set `DEV_MODE=true` di `.env` (hanya untuk pengembangan; jangan di server produksi). |
| `make seed` ditolak: *run pfmea migrate up first* | Skema belum dimigrasi atau versinya beda dengan kode. Jalankan `make migrate`, lalu `make seed` lagi. |
| `make seed` ditolak: *database already contains data* | Data demo (atau data lain) sudah ada. Pakai `make seed RESET=1` untuk menghapus semuanya dan memuat data demo dari nol, lalu restart `make dev`. |
| `/readyz` menjawab 503 | Database tidak terjangkau atau belum dimigrasi. Lihat log server (`readiness check failed`), lalu `make db` dan `make migrate`. |
| CI merah | `gh run list`, lalu `gh run view <id> --log-failed`. Reproduksi di lokal dengan `make check`, perbaiki, commit dan push lagi. Jangan menonaktifkan atau melemahkan test. |
| Push `.github/workflows/ci.yml` ditolak (*refusing to allow an OAuth App to create or update workflow … without `workflow` scope*) | Token `gh` belum punya scope `workflow`. Jalankan `gh auth refresh -h github.com -s workflow`, lalu `gh auth setup-git`, lalu push lagi. |
| Semua perintah lambat di WSL2 | Folder project berada di `/mnt/c/...` (file system Windows). Pindahkan/clone ulang ke file system Linux, misalnya `~/projects/vifmeacot`, dan buka dari sana (VS Code: "WSL: Open Folder"). |
| `make test` gagal: *-race requires cgo* | Compiler C belum terpasang. Ubuntu/WSL2: `sudo apt install build-essential`. |
| `go` mengunduh toolchain atau mengeluh versi | `go.mod` meminta Go 1.27. Go 1.21+ bisa mengunduh toolchain 1.27 otomatis bila `GOTOOLCHAIN=auto`. Go dari `apt` Ubuntu memakai `GOTOOLCHAIN=local` (pesan berisi *GOTOOLCHAIN=local*), jadi tidak mengunduh: jalankan `go env -w GOTOOLCHAIN=auto`, atau lebih baik pasang Go 1.27 dari https://go.dev/dl/ (juga bila jaringan memblokir unduhan). |
| `make dev` gagal: *missing .env* | Jalankan `cp .env.example .env` dulu (langkah 3). |
| Halaman di :8080 menampilkan "UI not built yet" | Binary dibangun tanpa hasil build web. Jalankan `make build` (bukan `go build` langsung), lalu jalankan `./bin/pfmea serve` lagi. Saat pengembangan, buka http://localhost:5173 (`make dev`). |
| `npm ci` gagal: versi Node terlalu lama | Pasang Node.js 24 LTS (minimal 22.17), misalnya dengan `nvm install 24`. |
