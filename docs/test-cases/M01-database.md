# M01 · Database, seed, sqlc, test harness — Test case

Milestone: M1 · Kebutuhan: P1-13 (operasional), `docs/10-milestones.md` M1,
`docs/04-data-model.md`, `docs/06-rules.md` §1, §5, §7, `docs/11-testing.md` §2–3 · User story: —
Ditulis sebelum kode pada: 9 Oktober 2026 · Status terakhir: 18 dari 30 lulus

Test aturan (fixture 32 aturan dan baseline) ada di `rules.md` (`TC-RULE-*`) dan diotomatisasi di
milestone ini. Test karakterisasi trigger (TC-M01-021 sampai 028) menguji skema yang sudah ada
di paket spesifikasi; agar terbukti bisa gagal, setiap test dijalankan sekali dengan mutasi
sementara (misalnya trigger dimatikan) sebelum di-commit.

## Ringkasan

| ID | Judul | Level | Status |
| --- | --- | --- | --- |
| TC-M01-001 | File migrasi dan seed tertanam di binary | unit | lulus |
| TC-M01-002 | `pfmea migrate up` dan `status` di database kosong | integrasi | lulus |
| TC-M01-003 | `pfmea migrate down` hanya menyisakan tabel goose dan pg_trgm | integrasi | lulus |
| TC-M01-004 | Argumen dan konfigurasi `pfmea migrate` | unit | lulus |
| TC-M01-005 | `LoadDatabase` hanya mewajibkan `DATABASE_URL` | unit | lulus |
| TC-M01-006 | `seed-demo` menolak tanpa `DEV_MODE=true` | integrasi | lulus |
| TC-M01-007 | `seed-demo` memuat data demo sekali | integrasi | lulus |
| TC-M01-008 | `seed-demo --reset` memulai dari nol | integrasi | lulus |
| TC-M01-009 | `/readyz` memeriksa database dan migrasi | integrasi | lulus |
| TC-M01-010 | Target Makefile M1 berjalan | integrasi | lulus |
| TC-M01-011 | `make db migrate seed` dan `make migrate-down` di CI | integrasi | belum dibuat |
| TC-M01-012 | `testdb.New` memberi database demo yang terisolasi | integrasi | lulus |
| TC-M01-013 | `testdb.NewEmpty` memberi database kosong | integrasi | lulus |
| TC-M01-014 | `WithTx` mengisi `app.*` tanpa bocor | integrasi | lulus |
| TC-M01-015 | `WithTx` commit, rollback dan panic | integrasi | lulus |
| TC-M01-016 | `WithTx` mengulang 40001/40P01 | integrasi | lulus |
| TC-M01-017 | Pemetaan error PostgreSQL, termasuk saat COMMIT | integrasi | lulus |
| TC-M01-018 | `LockPackage` mengunci baris paket | integrasi | lulus |
| TC-M01-019 | `store.Open` membuat pool dengan pengaturan sesi | integrasi | lulus |
| TC-M01-020 | Query sqlc pengguna dan paket | integrasi | lulus |
| TC-M01-021 | `failure_chains.s` mengikuti severity effect | integrasi | lulus |
| TC-M01-022 | `rpn` dan `new_rpn` dihitung database | integrasi | lulus |
| TC-M01-023 | `version` hanya naik untuk perubahan bermakna | integrasi | lulus |
| TC-M01-024 | `content_version` naik satu per statement per paket | integrasi | lulus |
| TC-M01-025 | `audit_log` hanya menyimpan kolom yang berubah | integrasi | lulus |
| TC-M01-026 | Hapus baris yang masih dipakai ditolak (RESTRICT, SQLSTATE 23001) | integrasi | lulus |
| TC-M01-027 | Referensi lintas paket ditolak | integrasi | lulus |
| TC-M01-028 | `snake_to_camel` | integrasi | lulus |
| TC-M01-029 | Parser fixture dan pencocokan hasil aturan | unit | lulus |
| TC-M01-030 | Satu file SQL dan satu fixture untuk setiap aturan Tahap 1 | unit | lulus |

## TC-M01-001 — File migrasi dan seed tertanam di binary

- **Level:** unit · **Kebutuhan:** M1 Deliver (`db/embed.go`)
- **Prasyarat:** paket `db`.
- **Langkah:** baca file tertanam dan versi migrasi terakhir.
- **Hasil yang diharapkan:** `migrations/00001_init.sql` dan `seed/demo.sql` ada; versi migrasi
  terakhir = 1.
- **Test otomatis:** `db/embed_test.go` › `TestFiles_TC_M01_001`
- **Status:** lulus

## TC-M01-002 — `pfmea migrate up` dan `status` di database kosong

- **Level:** integrasi · **Kebutuhan:** M1 Deliver (`pfmea migrate up|down|status`, goose, file
  tertanam), Done when (`make migrate`)
- **Prasyarat:** database kosong (`testdb.NewEmpty`); hanya `DATABASE_URL` di-set.
- **Langkah:**
  1. `pfmea migrate up`.
  2. `pfmea migrate up` lagi.
  3. `pfmea migrate status`.
- **Hasil yang diharapkan:**
  1. Kode 0, output memuat "applied 00001_init.sql"; tabel `packages` ada.
  2. Kode 0, output "no pending migrations".
  3. Kode 0, `00001_init.sql` berstatus applied, versi 1.
- **Test otomatis:** `backend/cmd/pfmea/migrate_test.go` › `TestMigrateUp_TC_M01_002`
- **Status:** lulus

## TC-M01-003 — `pfmea migrate down` hanya menyisakan tabel goose dan pg_trgm

- **Level:** integrasi · **Kebutuhan:** M1 Done when ("`pfmea migrate down` leaves no application
  tables, functions or types; only goose's version table and the `pg_trgm` extension remain")
- **Prasyarat:** database yang sudah di-`migrate up` dan diisi seed.
- **Langkah:**
  1. `pfmea migrate down` dengan `DEV_MODE=false`.
  2. `pfmea migrate down` dengan `DEV_MODE=true`.
  3. `pfmea migrate up` lalu muat seed lagi.
- **Hasil yang diharapkan:**
  1. Kode 1, pesan "migrate down deletes all data and requires DEV_MODE=true"; skema utuh.
  2. Kode 0, versi 0. Di skema `public` hanya ada tabel `goose_db_version` (dan objek miliknya);
     extension yang tersisa hanya `plpgsql` dan `pg_trgm`; tidak ada fungsi, tipe enum atau
     tabel aplikasi (objek milik extension tidak dihitung).
  3. Berhasil (siklus naik-turun-naik).
- **Test otomatis:** `backend/cmd/pfmea/migrate_test.go` › `TestMigrateDown_TC_M01_003`
- **Status:** lulus

## TC-M01-004 — Argumen dan konfigurasi `pfmea migrate`

- **Level:** unit · **Kebutuhan:** `docs/03-architecture.md` §3.2; keputusan 9 Okt 2026
  (`migrate` tidak mewajibkan `APP_BASE_URL`)
- **Prasyarat:** fungsi `run` dengan environment tiruan.
- **Langkah:**
  1. `pfmea migrate` tanpa sub-perintah dan `pfmea migrate sideways`.
  2. `pfmea migrate up` tanpa environment.
  3. `pfmea migrate up` dengan `DATABASE_URL` ke port yang tidak terbuka dan kata sandi `secret`.
- **Hasil yang diharapkan:**
  1. Usage `migrate up|down|status`, kode 2.
  2. Kode 1, stderr memuat "DATABASE_URL is required" dan tidak menyebut `APP_BASE_URL`.
  3. Kode 1, pesan error bahasa Inggris, tidak memuat `secret`.
- **Test otomatis:** `backend/cmd/pfmea/migrate_test.go` › `TestMigrateArgs_TC_M01_004`
- **Status:** lulus

## TC-M01-005 — `LoadDatabase` hanya mewajibkan `DATABASE_URL`

- **Level:** unit · **Kebutuhan:** keputusan 9 Okt 2026 (konfigurasi per perintah),
  `docs/03-architecture.md` §8
- **Prasyarat:** fungsi `config.LoadDatabase`.
- **Langkah:**
  1. Hanya `DATABASE_URL` di-set.
  2. `DATABASE_URL=mysql://x`, `DEV_MODE=maybe`, `LOG_LEVEL=verbose`.
- **Hasil yang diharapkan:**
  1. Sukses; `DevMode` false, `LogLevel` info; `APP_BASE_URL` tidak diminta.
  2. Satu `*ValidationError` dengan tiga masalah (satu per variabel).
  - `config.Load` (untuk `serve`) tetap mewajibkan `APP_BASE_URL` (TC-M00-006 tetap lulus).
- **Test otomatis:** `backend/internal/config/config_test.go` › `TestLoadDatabase_TC_M01_005`
- **Status:** lulus

## TC-M01-006 — `seed-demo` menolak tanpa `DEV_MODE=true`

- **Level:** integrasi · **Kebutuhan:** `docs/03-architecture.md` §3.2 ("Refuses unless
  `DEV_MODE=true`")
- **Prasyarat:** database yang sudah dimigrasi tanpa data.
- **Langkah:** `pfmea seed-demo` dengan `DEV_MODE` kosong dan dengan `DEV_MODE=false`.
- **Hasil yang diharapkan:** kode 1, pesan "seed-demo requires DEV_MODE=true"; tabel `users`
  tetap kosong.
- **Test otomatis:** `backend/cmd/pfmea/seed_test.go` › `TestSeedDemoRefuses_TC_M01_006`
- **Status:** lulus

## TC-M01-007 — `seed-demo` memuat data demo sekali

- **Level:** integrasi · **Kebutuhan:** M1 Deliver (`pfmea seed-demo`), Done when (`make seed`),
  `docs/04-data-model.md` §8
- **Prasyarat:** `DEV_MODE=true`.
- **Langkah:**
  1. `pfmea seed-demo` pada database yang belum dimigrasi.
  2. `pfmea migrate up`, lalu `pfmea seed-demo`.
  3. `pfmea seed-demo` sekali lagi.
- **Hasil yang diharapkan:**
  1. Kode 1, pesan "run pfmea migrate up first".
  2. Kode 0; 6 pengguna, paket `GENERAL` dan `PS-07`; baris `audit_log` dari seed bersumber
     `seed`.
  3. Kode 1, pesan "database already contains data; use --reset"; jumlah data tidak berubah.
- **Test otomatis:** `backend/cmd/pfmea/seed_test.go` › `TestSeedDemo_TC_M01_007`
- **Status:** lulus

## TC-M01-008 — `seed-demo --reset` memulai dari nol

- **Level:** integrasi · **Kebutuhan:** `docs/03-architecture.md` §3.2 (`--reset` menghapus skema
  dan menjalankan migrate up penuh), `docs/11-testing.md` §4
- **Prasyarat:** database berisi seed; nama paket PS-07 diubah dan satu step ditambahkan.
- **Langkah:** `pfmea seed-demo --reset` dengan `DEV_MODE=true`.
- **Hasil yang diharapkan:** kode 0, pesan mengingatkan untuk me-restart server yang sedang
  berjalan; data sama dengan seed baru (nama PS-07 asli, jumlah step seperti seed), versi
  migrasi 1.
- **Test otomatis:** `backend/cmd/pfmea/seed_test.go` › `TestSeedDemoReset_TC_M01_008`
- **Status:** lulus

## TC-M01-009 — `/readyz` memeriksa database dan migrasi

- **Level:** integrasi · **Kebutuhan:** M1 Deliver (`/readyz` mendaftarkan "database reachable
  and migrations current"), `docs/03-architecture.md` §8
- **Prasyarat:** `pfmea serve` di port acak.
- **Langkah:**
  1. `DATABASE_URL` ke database yang sudah dimigrasi → `GET /readyz`.
  2. `DATABASE_URL` ke database kosong → `GET /readyz`, lalu periksa tabel di database itu.
  3. `DATABASE_URL` ke port yang tidak terbuka → `GET /readyz`.
- **Hasil yang diharapkan:**
  1. 200, pemeriksaan `database` = ok.
  2. 503, `database` = failed; tabel `goose_db_version` tidak dibuat oleh pemeriksaan.
  3. 503 dalam waktu kurang dari 3 detik.
  - `/healthz` 200 di ketiga keadaan.
- **Test otomatis:** `backend/cmd/pfmea/readiness_test.go` › `TestReadyzDatabase_TC_M01_009`
- **Status:** lulus

## TC-M01-010 — Target Makefile M1 berjalan

- **Level:** integrasi (Go menjalankan `make -n` dan `make`) · **Kebutuhan:** M0 Deliver (target
  yang tersedia mulai M1), `docs/03-architecture.md` §3.3, `docs/11-testing.md` §1
- **Prasyarat:** root repository.
- **Langkah:** `make -n migrate migrate-down seed gen test test-rules`; `make e2e`, `make perf`.
- **Hasil yang diharapkan:** `migrate`/`migrate-down`/`seed` memuat `.env` lalu memanggil
  `pfmea migrate up|down` dan `pfmea seed-demo`; `gen` menjalankan sqlc; `test` melewati test
  fixture aturan dengan `-skip`; `test-rules` menjalankannya dengan `-run`. Hanya `e2e` (M2) dan
  `perf` (M13) yang masih mencetak "available from".
- **Test otomatis:** `backend/internal/repotest/makefile_test.go` › `TestTargets_TC_M01_010`
- **Status:** lulus

## TC-M01-011 — `make db migrate seed` dan `make migrate-down` di CI

- **Level:** integrasi (GitHub Actions) · **Kebutuhan:** M1 Done when (`make db migrate seed`
  works; `pfmea migrate down` leaves only goose's table and pg_trgm)
- **Prasyarat:** runner dengan Docker; `cp .env.example .env`.
- **Langkah:** `make db`, `make migrate`, `make seed`, hitung baris; `make migrate-down`, hitung
  tabel; `docker compose down -v`.
- **Hasil yang diharapkan:** setelah seed: 6 pengguna dan 2 paket; setelah migrate-down: tabel di
  `public` hanya `goose_db_version`.
- **Test otomatis:** `.github/workflows/ci.yml` › langkah `TC-M01-011 make db migrate seed`
- **Status:** belum dibuat

## TC-M01-012 — `testdb.New` memberi database demo yang terisolasi

- **Level:** integrasi · **Kebutuhan:** M1 Deliver (`backend/internal/testdb`), `docs/11-testing.md`
  §2
- **Prasyarat:** Docker atau `TEST_DATABASE_URL`.
- **Langkah:**
  1. Dua test paralel memanggil `testdb.New`; test pertama menghapus PS-07.
  2. Periksa versi server, `jit`, data demo, nama database.
- **Hasil yang diharapkan:**
  - Versi server 18, `jit` = off, 6 pengguna, PS-07 ada di test kedua (terisolasi).
  - Database test dihapus setelah test selesai; template dibuat sekali per hash isi migrasi+seed.
  - Tanpa Docker dan tanpa `TEST_DATABASE_URL` test gagal dengan pesan jelas (tidak di-skip).
- **Test otomatis:** `backend/internal/testdb/testdb_test.go` › `TestNew_TC_M01_012`
- **Status:** lulus

## TC-M01-013 — `testdb.NewEmpty` memberi database kosong

- **Level:** integrasi · **Kebutuhan:** M1 Deliver (test migrasi dan perintah CLI)
- **Prasyarat:** seperti TC-M01-012.
- **Langkah:** `testdb.NewEmpty`, hitung relasi di skema `public`.
- **Hasil yang diharapkan:** 0 relasi; URL bisa dipakai perintah CLI.
- **Test otomatis:** `backend/internal/testdb/testdb_test.go` › `TestNewEmpty_TC_M01_013`
- **Status:** lulus

## TC-M01-014 — `WithTx` mengisi `app.*` tanpa bocor

- **Level:** integrasi · **Kebutuhan:** `CLAUDE.md` ("Every write goes through `store.WithTx`
  (sets `app.user_id`, `app.request_id`, `app.source`)"), `docs/03-architecture.md` §3.1
- **Prasyarat:** database demo; pool dengan satu koneksi.
- **Langkah:**
  1. `WithTx` dengan actor `apratama`, request id `req-1`, source `api` mengubah nama step 50.
  2. `WithTx` dengan actor kosong mengubah step yang sama.
  3. Setelah transaksi, baca `current_setting('app.user_id', true)` di koneksi yang sama.
- **Hasil yang diharapkan:**
  1. Baris `audit_log` membawa user id `apratama`, `req-1`, `api`.
  2. user id NULL, request id NULL, source `system`.
  3. Kosong (tidak bocor).
- **Test otomatis:** `backend/internal/store/store_test.go` › `TestWithTxSettings_TC_M01_014`
- **Status:** lulus

## TC-M01-015 — `WithTx` commit, rollback dan panic

- **Level:** integrasi · **Kebutuhan:** `docs/03-architecture.md` §3.1, §5 (satu request = satu
  transaksi)
- **Prasyarat:** database demo.
- **Langkah:** fn sukses; fn mengembalikan error; fn panic.
- **Hasil yang diharapkan:** perubahan tersimpan; tidak tersimpan dan error diteruskan; tidak
  tersimpan dan panic diteruskan ke pemanggil.
- **Test otomatis:** `backend/internal/store/store_test.go` › `TestWithTxCommitRollback_TC_M01_015`
- **Status:** lulus

## TC-M01-016 — `WithTx` mengulang 40001/40P01

- **Level:** integrasi · **Kebutuhan:** M1 Deliver ("retries on 40001 and 40P01"),
  `docs/05-api.md` §2 (setelah 3 kali → `version_conflict`)
- **Prasyarat:** database demo.
- **Langkah:**
  1. fn mengembalikan error 40001 dua kali, lalu sukses.
  2. fn selalu mengembalikan error 40P01.
  3. fn mengembalikan error 23505; konteks dibatalkan saat menunggu retry.
  4. Deadlock nyata: dua transaksi mengubah dua baris dengan urutan terbalik.
- **Hasil yang diharapkan:**
  1. fn dipanggil 3 kali, hasil sukses.
  2. fn dipanggil 4 kali (1 + 3 retry), hasil `ErrConflict`.
  3. Tanpa retry.
  4. Kedua transaksi akhirnya sukses.
- **Test otomatis:** `backend/internal/store/store_test.go` › `TestWithTxRetry_TC_M01_016`
- **Status:** lulus

## TC-M01-017 — Pemetaan error PostgreSQL, termasuk saat COMMIT

- **Level:** integrasi · **Kebutuhan:** M1 Deliver ("Postgres error mapping that also covers
  errors raised at `COMMIT`"), `docs/05-api.md` §2, `docs/04-data-model.md` §1 (unique deferrable)
- **Prasyarat:** database demo, paket PS-07.
- **Langkah:**
  1. Ubah `op_no` step 60 menjadi `50` (duplikat, baru terdeteksi saat COMMIT).
  2. Tukar `op_no` step 50 dan 60 dalam satu transaksi.
  3. Set `o = 11` pada satu chain.
  4. Insert karakteristik dengan step id yang tidak ada.
  5. Insert step tanpa `name`.
  6. Hapus step 50 PS-07 (masih punya failure mode, `ON DELETE RESTRICT`).
- **Hasil yang diharapkan:**
  1. `ErrDuplicate` dengan constraint `process_steps_op_no_uq`.
  2. Sukses.
  3. `ErrCheckViolation`.
  4. `ErrForeignKey` dengan nama tabel.
  5. `ErrNotNull` dengan kolom `name`.
  6. `ErrForeignKey` (SQLSTATE 23001 `restrict_violation`, dipetakan sama dengan 23503).
  - Error asli `*pgconn.PgError` tetap bisa diambil dengan `errors.As`.
- **Test otomatis:** `backend/internal/store/store_test.go` › `TestErrorMapping_TC_M01_017`
- **Status:** lulus

## TC-M01-018 — `LockPackage` mengunci baris paket

- **Level:** integrasi · **Kebutuhan:** M1 Deliver (`LockPackage`), `docs/03-architecture.md` §3.1
  (`FOR NO KEY UPDATE` tetap mengizinkan insert findings, check runs, exports)
- **Prasyarat:** database demo, paket PS-07.
- **Langkah:**
  1. `LockPackage` dengan id yang tidak ada.
  2. Transaksi A mengunci PS-07; transaksi B (lock_timeout 200 ms) mencoba `LockPackage` dan
     mengubah `packages.name`.
  3. Transaksi C insert `check_runs` dan `findings` untuk PS-07 selama A masih memegang lock.
- **Hasil yang diharapkan:** 1. `ErrNotFound`. 2. Keduanya gagal dengan 55P03 (lock not
  available). 3. Sukses.
- **Test otomatis:** `backend/internal/store/store_test.go` › `TestLockPackage_TC_M01_018`
- **Status:** lulus

## TC-M01-019 — `store.Open` membuat pool dengan pengaturan sesi

- **Level:** integrasi · **Kebutuhan:** M1 Deliver (pgx pool), `docs/03-architecture.md` §4 (JIT
  off), §8 (`DB_MAX_CONNS`)
- **Prasyarat:** URL database demo dan URL port yang tidak terbuka.
- **Langkah:** `store.Open` ke port yang tidak terbuka; `store.Open` ke database demo lalu baca
  `jit`, `application_name` dan ukuran pool.
- **Hasil yang diharapkan:** Open ke port tertutup tidak gagal (koneksi dibuat saat dipakai);
  `jit` = off, `application_name` = `pfmea`, MaxConns sesuai argumen.
- **Test otomatis:** `backend/internal/store/store_test.go` › `TestOpen_TC_M01_019`
- **Status:** lulus

## TC-M01-020 — Query sqlc pengguna dan paket

- **Level:** integrasi · **Kebutuhan:** M1 Deliver (`sqlc.yaml` + first queries (users, packages)
  + generated code)
- **Prasyarat:** database demo.
- **Langkah:** `GetUserByUsername("apratama")`, `GetUserByID`, `ListUsers`, `GetPackageByID`
  (PS-07 dan id acak), `ListPackages`.
- **Hasil yang diharapkan:** `apratama` berperan author; id acak → `pgx.ErrNoRows`; `ListUsers`
  mengembalikan 6 pengguna urut username tanpa kolom hash kata sandi; `ListPackages` mengembalikan
  `GENERAL` (general) dan `PS-07` (model); id bertipe `uuid.UUID` dan waktu `time.Time`.
- **Test otomatis:** `backend/internal/store/queries_test.go` › `TestQueries_TC_M01_020`
- **Status:** lulus

## TC-M01-021 — `failure_chains.s` mengikuti severity effect

- **Level:** integrasi (karakterisasi skema) · **Kebutuhan:** M1 Deliver ("chain `s` follows effect
  S (insert, update, delete, move effect)"), `docs/04-data-model.md` §4
- **Prasyarat:** database demo; failure mode PS-07 dengan dua chain.
- **Langkah:** insert chain baru; insert effect S lebih tinggi; ubah S effect; hapus effect
  tertinggi; pindahkan effect ke failure mode lain; buat failure mode tanpa effect lalu chain-nya.
- **Hasil yang diharapkan:** setiap kali, `s` semua chain di failure mode yang terlibat = S
  effect tertinggi; failure mode tanpa effect → `s` NULL.
- **Test otomatis:** `backend/internal/store/schema_test.go` › `TestChainSeverity_TC_M01_021`
- **Status:** lulus

## TC-M01-022 — `rpn` dan `new_rpn` dihitung database

- **Level:** integrasi (karakterisasi skema) · **Kebutuhan:** M1 Deliver ("`rpn` and `new_rpn`
  generated")
- **Prasyarat:** database demo.
- **Langkah:** ubah `o`/`d` chain; kosongkan `d`; isi `new_s`/`new_o`/`new_d` pada action.
- **Hasil yang diharapkan:** `rpn` = s × o × d; NULL bila ada faktor NULL; `new_rpn` =
  new_s × new_o × new_d.
- **Test otomatis:** `backend/internal/store/schema_test.go` › `TestRPN_TC_M01_022`
- **Status:** lulus

## TC-M01-023 — `version` hanya naik untuk perubahan bermakna

- **Level:** integrasi (karakterisasi skema) · **Kebutuhan:** M1 Deliver ("version bumps only on
  user-meaningful change (not on `s`/`rpn`, not on no-op updates)")
- **Prasyarat:** database demo.
- **Langkah:** ubah `o` chain; `UPDATE … SET o = o`; ubah S effect (mengubah `s` dan `rpn` chain).
- **Hasil yang diharapkan:** version chain +1; tetap; tetap (version effect +1).
- **Test otomatis:** `backend/internal/store/schema_test.go` › `TestVersion_TC_M01_023`
- **Status:** lulus

## TC-M01-024 — `content_version` naik satu per statement per paket

- **Level:** integrasi (karakterisasi skema) · **Kebutuhan:** M1 Deliver ("`content_version` +1 per
  statement per touched package")
- **Prasyarat:** database demo.
- **Langkah:** update satu baris PS-07; update tiga baris PS-07 dalam satu statement; update
  baris PS-07 dan GENERAL dalam satu statement; update no-op; ubah S effect; insert
  `check_runs`/`findings`.
- **Hasil yang diharapkan:** +1; +1; +1 untuk masing-masing paket; +0; +1 total; +0.
- **Test otomatis:** `backend/internal/store/schema_test.go` › `TestContentVersion_TC_M01_024`
- **Status:** lulus

## TC-M01-025 — `audit_log` hanya menyimpan kolom yang berubah

- **Level:** integrasi (karakterisasi skema) · **Kebutuhan:** M1 Deliver ("audit rows store only
  changed columns and carry `app.*` settings")
- **Prasyarat:** database demo; `app.*` di-set lewat `WithTx`.
- **Langkah:** ubah `name` step; update no-op; ubah S effect (mengubah `s` chain); insert dan hapus
  step baru.
- **Hasil yang diharapkan:** satu baris audit update berisi hanya `name` di `old_row`/`new_row`;
  tidak ada baris untuk no-op, untuk perubahan `s` chain atau untuk kenaikan `content_version`;
  insert dan delete menyimpan baris lengkap.
- **Test otomatis:** `backend/internal/store/schema_test.go` › `TestAudit_TC_M01_025`
- **Status:** lulus

## TC-M01-026 — Hapus baris yang masih dipakai ditolak (RESTRICT)

- **Level:** integrasi (karakterisasi skema) · **Kebutuhan:** M1 Deliver ("delete of a step with
  failure modes fails (RESTRICT)"), US-03
- **Prasyarat:** database demo.
- **Langkah:** hapus step 50 PS-07; hapus karakteristik yang dipakai baris CP; hapus step baru
  tanpa failure mode yang punya karakteristik dan alur NG.
- **Hasil yang diharapkan:** dua yang pertama ditolak dengan SQLSTATE 23001
  (`restrict_violation`, kode PostgreSQL untuk `ON DELETE RESTRICT`; bukan 23503) dan tidak ada
  baris terhapus; yang ketiga terhapus beserta karakteristik dan alurnya. (Diperbarui saat M1:
  test menunjukkan RESTRICT memakai 23001.)
- **Test otomatis:** `backend/internal/store/schema_test.go` › `TestRestrict_TC_M01_026`
- **Status:** lulus

## TC-M01-027 — Referensi lintas paket ditolak

- **Level:** integrasi (karakterisasi skema) · **Kebutuhan:** M1 Deliver ("cross-package
  references fail (composite FKs)"), `docs/04-data-model.md` §1
- **Prasyarat:** database demo.
- **Langkah:** insert karakteristik PS-07 yang menunjuk step GENERAL; insert failure mode PS-07
  yang menunjuk karakteristik GENERAL.
- **Hasil yang diharapkan:** keduanya 23503.
- **Test otomatis:** `backend/internal/store/schema_test.go` › `TestCrossPackage_TC_M01_027`
- **Status:** lulus

## TC-M01-028 — `snake_to_camel`

- **Level:** integrasi (karakterisasi skema) · **Kebutuhan:** M1 Deliver (`snake_to_camel`),
  `docs/04-data-model.md` §7
- **Prasyarat:** database demo.
- **Langkah:** `snake_to_camel` untuk `sample_freq`, `ep_verify_freq`, `d`, `package_id`, NULL.
- **Hasil yang diharapkan:** `sampleFreq`, `epVerifyFreq`, `d`, `packageId`, NULL.
- **Test otomatis:** `backend/internal/store/schema_test.go` › `TestSnakeToCamel_TC_M01_028`
- **Status:** lulus

## TC-M01-029 — Parser fixture dan pencocokan hasil aturan

- **Level:** unit · **Kebutuhan:** `docs/06-rules.md` §7 (multiset match: jumlah sama, setiap item
  cocok dengan satu hasil berbeda, params mencakup nilai yang diharapkan)
- **Prasyarat:** fungsi parser dan matcher di test aturan.
- **Langkah:** parse fixture tanpa `-- expect:`, dengan JSON salah, dan yang benar; cocokkan
  daftar hasil: jumlah beda, params kurang, item duplikat, kasus yang gagal bila dicocokkan rakus
  (greedy) tetapi berhasil dengan backtracking.
- **Hasil yang diharapkan:** error yang jelas untuk fixture salah; matcher menolak/menerima sesuai
  aturan multiset, termasuk kasus backtracking.
- **Test otomatis:** `backend/internal/rules/fixture_test.go` › `TestFixtureMatcher_TC_M01_029`
- **Status:** lulus

## TC-M01-030 — Satu file SQL dan satu fixture untuk setiap aturan Tahap 1

- **Level:** unit · **Kebutuhan:** `CLAUDE.md` ("one SQL file plus one fixture per rule"),
  `docs/06-rules.md` §1
- **Prasyarat:** folder `backend/internal/rules/sql` dan `testdata`.
- **Langkah:** daftar file dan header.
- **Hasil yang diharapkan:** 32 file SQL, 32 fixture dengan nama sama; header `-- rule:` sama
  dengan nama file; kode aturan sesuai daftar Tahap 1 di `docs/06-rules.md` §4.
- **Test otomatis:** `backend/internal/rules/rules_test.go` › `TestRuleFiles_TC_M01_030`
- **Status:** lulus
