# M00 · Scaffold — Test case

Milestone: M0 · Kebutuhan: P1-13 (konfigurasi lewat environment variable, endpoint health),
`docs/10-milestones.md` M0, `docs/03-architecture.md` §2–3, §8 · User story: — (E2E Playwright
baru dipakai mulai M2)
Ditulis sebelum kode pada: 9 Oktober 2026 · Status terakhir: 0 dari 26 lulus

## Ringkasan

| ID | Judul | Level | Status |
| --- | --- | --- | --- |
| TC-M00-001 | `/healthz` menjawab 200 selama proses berjalan | unit | belum dibuat |
| TC-M00-002 | `/readyz` menjawab 200 bila tidak ada pemeriksaan terdaftar | unit | belum dibuat |
| TC-M00-003 | `/readyz` menjawab 200 bila semua pemeriksaan lulus | unit | belum dibuat |
| TC-M00-004 | `/readyz` menjawab 503 bila ada pemeriksaan gagal atau melewati batas waktu | unit | belum dibuat |
| TC-M00-005 | Konfigurasi memasang semua nilai bawaan | unit | belum dibuat |
| TC-M00-006 | `DATABASE_URL` dan `APP_BASE_URL` kosong dilaporkan sekaligus | unit | belum dibuat |
| TC-M00-007 | Semua nilai konfigurasi yang salah dilaporkan sekaligus | unit | belum dibuat |
| TC-M00-008 | `DEV_FAKE_TODAY` hanya dipakai bila `DEV_MODE=true` | unit | belum dibuat |
| TC-M00-009 | SPA belum di-build → "UI not built yet" | unit | belum dibuat |
| TC-M00-010 | Path non-API yang tidak dikenal mendapat `index.html` | unit | belum dibuat |
| TC-M00-011 | Path `/api/...` yang tidak dikenal mendapat 404 Problem, bukan `index.html` | unit | belum dibuat |
| TC-M00-012 | Aset ber-hash mendapat header cache permanen; aset yang tidak ada 404 | unit | belum dibuat |
| TC-M00-013 | Binary `pfmea`: usage, subcommand milestone berikutnya, konfigurasi salah | unit | belum dibuat |
| TC-M00-014 | `pfmea serve` berjalan, menjawab `/healthz`, lalu berhenti dengan rapi | integrasi | belum dibuat |
| TC-M00-015 | Target Makefile yang berjalan di M0 tersedia | integrasi | belum dibuat |
| TC-M00-016 | Target Makefile milestone berikutnya mencetak "available from M<n>" dengan kode 0 | integrasi | belum dibuat |
| TC-M00-017 | `.gitignore` mengabaikan rahasia dan hasil build | integrasi | belum dibuat |
| TC-M00-018 | App shell menampilkan teks bahasa Inggris dari `en.ts` | unit | belum dibuat |
| TC-M00-019 | Menu aktif dan breadcrumb sesuai halaman | unit | belum dibuat |
| TC-M00-020 | Proxy dev Vite meneruskan `/api` (termasuk WebSocket) ke server Go | unit | belum dibuat |
| TC-M00-021 | `make build` lalu `./bin/pfmea serve` menyajikan SPA di :8080 | integrasi | belum dibuat |
| TC-M00-022 | `make db` menjalankan PostgreSQL 18 yang sehat dengan JIT mati | integrasi | belum dibuat |
| TC-M00-023 | `make tools` + `make check` hijau di CI dan tidak mengubah file | integrasi | belum dibuat |
| TC-M00-024 | `make dev` menampilkan app shell di :5173 dan `/healthz` di :8080 | manual | belum dibuat |
| TC-M00-025 | Clone baru mengikuti panduan instalasi sampai `make check` hijau | manual | belum dibuat |
| TC-M00-026 | Komentar bahasa Indonesia dan teks tampilan bahasa Inggris | manual | belum dibuat |

## TC-M00-001 — `/healthz` menjawab 200 selama proses berjalan

- **Level:** unit (Go, `httptest`) · **Kebutuhan:** M0 Deliver (`/healthz`), M0 Done when
  (`curl -i localhost:8080/healthz` → 200), P1-13
- **Prasyarat:** handler server dibuat tanpa pemeriksaan kesiapan dan tanpa build UI.
- **Langkah:**
  1. Kirim `GET /healthz`.
  2. Kirim `HEAD /healthz`.
- **Hasil yang diharapkan:**
  - Status 200 untuk keduanya; body GET `{"status":"ok"}` dengan `Content-Type:
    application/json`.
  - Header `Cache-Control: no-store` (status tidak boleh di-cache).
- **Test otomatis:** `backend/internal/httpapi/health_test.go` › `TestHealthz_TC_M00_001`
- **Status:** belum dibuat

## TC-M00-002 — `/readyz` menjawab 200 bila tidak ada pemeriksaan terdaftar

- **Level:** unit (Go, `httptest`) · **Kebutuhan:** M0 Deliver (`/readyz`; M0 tidak
  mendaftarkan pemeriksaan, M1 menambahkan "database reachable and migrations current")
- **Prasyarat:** handler server dibuat dengan daftar pemeriksaan kosong (kondisi M0).
- **Langkah:**
  1. Kirim `GET /readyz`.
- **Hasil yang diharapkan:**
  - Status 200, body `{"status":"ready","checks":[]}`, `Cache-Control: no-store`.
- **Test otomatis:** `backend/internal/httpapi/health_test.go` › `TestReadyz_TC_M00_002`
- **Status:** belum dibuat

## TC-M00-003 — `/readyz` menjawab 200 bila semua pemeriksaan lulus

- **Level:** unit (Go, `httptest`) · **Kebutuhan:** M0 Deliver (`/readyz` 200 bila semua lulus)
- **Prasyarat:** dua pemeriksaan tiruan terdaftar, `alpha` dan `beta`, keduanya mengembalikan
  `nil`.
- **Langkah:**
  1. Kirim `GET /readyz`.
- **Hasil yang diharapkan:**
  - Status 200, `status` = `ready`.
  - `checks` berisi `alpha` dan `beta` (urutan pendaftaran) dengan status `ok`.
- **Test otomatis:** `backend/internal/httpapi/health_test.go` › `TestReadyz_TC_M00_003`
- **Status:** belum dibuat

## TC-M00-004 — `/readyz` menjawab 503 bila ada pemeriksaan gagal atau melewati batas waktu

- **Level:** unit (Go, `httptest`) · **Kebutuhan:** M0 Deliver (`/readyz` 503 bila ada yang
  gagal)
- **Prasyarat:** tiga pemeriksaan tiruan: `alpha` lulus, `beta` mengembalikan error
  "secret connection detail", `gamma` menunggu sampai konteksnya habis; batas waktu per
  pemeriksaan dibuat pendek (50 ms) untuk test.
- **Langkah:**
  1. Kirim `GET /readyz`.
- **Hasil yang diharapkan:**
  - Status 503, `status` = `not_ready`.
  - `alpha` = `ok`, `beta` = `failed`, `gamma` = `failed`.
  - Body tidak memuat teks error internal ("secret connection detail"); error hanya dicatat di
    log.
  - Respons datang tidak lama setelah batas waktu (pemeriksaan dijalankan paralel, tidak
    menggantung).
- **Test otomatis:** `backend/internal/httpapi/health_test.go` › `TestReadyz_TC_M00_004`
- **Status:** belum dibuat

## TC-M00-005 — Konfigurasi memasang semua nilai bawaan

- **Level:** unit (Go) · **Kebutuhan:** M0 Deliver (`backend/internal/config`: env parsing,
  defaults), `docs/03-architecture.md` §8
- **Prasyarat:** hanya `DATABASE_URL=postgres://pfmea:secret@localhost:5432/pfmea` dan
  `APP_BASE_URL=http://localhost:5173` yang di-set.
- **Langkah:**
  1. Panggil `config.Load` dengan environment tersebut.
- **Hasil yang diharapkan:**
  - Tidak ada error.
  - `HTTP_ADDR` = `:8080`, `SESSION_TTL` = 12 jam, `EXPORT_DIR` = `/data/exports`,
    `EXPORT_TEMPLATE_DIR` = `/data/templates`, `DB_MAX_CONNS` = 20, `RIVER_WORKERS` = 10,
    `RULE_POOL_SIZE` = 8, `RULE_PARALLELISM` = 4, `LOG_LEVEL` = info, `METRICS_ALLOW` =
    `127.0.0.1/32`, `DEV_MODE` = false, `DEV_FAKE_TODAY` kosong.
  - Representasi log konfigurasi tidak memuat kata sandi dari `DATABASE_URL`.
- **Test otomatis:** `backend/internal/config/config_test.go` › `TestLoad_TC_M00_005`
- **Status:** belum dibuat

## TC-M00-006 — `DATABASE_URL` dan `APP_BASE_URL` kosong dilaporkan sekaligus

- **Level:** unit (Go) · **Kebutuhan:** M0 Deliver ("validation errors listed together")
- **Prasyarat:** environment kosong (kedua variabel wajib tidak di-set; variabel berisi spasi
  saja juga dianggap kosong).
- **Langkah:**
  1. Panggil `config.Load`.
- **Hasil yang diharapkan:**
  - Satu error bertipe `*config.ValidationError` berisi tepat dua masalah:
    "DATABASE_URL is required" dan "APP_BASE_URL is required".
  - Teks error memuat keduanya, satu baris per masalah.
- **Test otomatis:** `backend/internal/config/config_test.go` › `TestLoad_TC_M00_006`
- **Status:** belum dibuat

## TC-M00-007 — Semua nilai konfigurasi yang salah dilaporkan sekaligus

- **Level:** unit (Go) · **Kebutuhan:** M0 Deliver ("validation errors listed together"),
  `docs/03-architecture.md` §8
- **Prasyarat:** environment berisi: `DATABASE_URL=mysql://x`, `APP_BASE_URL=/relative`,
  `HTTP_ADDR=8080`, `SESSION_TTL=abc`, `DB_MAX_CONNS=0`, `RIVER_WORKERS=-1`,
  `RULE_POOL_SIZE=2`, `RULE_PARALLELISM=3`, `LOG_LEVEL=verbose`, `METRICS_ALLOW=10.0.0.0/33`,
  `DEV_MODE=maybe`.
- **Langkah:**
  1. Panggil `config.Load`.
- **Hasil yang diharapkan:**
  - Satu `*config.ValidationError` yang memuat satu masalah untuk setiap variabel di atas
    (10 masalah: `DATABASE_URL`, `APP_BASE_URL`, `HTTP_ADDR`, `SESSION_TTL`, `DB_MAX_CONNS`,
    `RIVER_WORKERS`, `RULE_PARALLELISM` > `RULE_POOL_SIZE`, `LOG_LEVEL`, `METRICS_ALLOW`,
    `DEV_MODE`), setiap pesan menyebut nama variabelnya, dalam bahasa Inggris.
- **Test otomatis:** `backend/internal/config/config_test.go` › `TestLoad_TC_M00_007`
- **Status:** belum dibuat

## TC-M00-008 — `DEV_FAKE_TODAY` hanya dipakai bila `DEV_MODE=true`

- **Level:** unit (Go) · **Kebutuhan:** `docs/03-architecture.md` §8 (`DEV_FAKE_TODAY`
  "honoured only with `DEV_MODE=true`"), `docs/11-testing.md` §2
- **Prasyarat:** variabel wajib di-set dengan nilai benar.
- **Langkah:**
  1. `DEV_MODE=true`, `DEV_FAKE_TODAY=2026-10-08`.
  2. `DEV_MODE=false`, `DEV_FAKE_TODAY=2026-10-08`.
  3. `DEV_MODE=true`, `DEV_FAKE_TODAY=08-10-2026`.
- **Hasil yang diharapkan:**
  1. Tanggal palsu = `2026-10-08`, tanpa peringatan.
  2. Tanggal palsu kosong (diabaikan) dan ada satu peringatan yang menyebut `DEV_FAKE_TODAY`.
  3. Error validasi yang menyebut `DEV_FAKE_TODAY` dan format `YYYY-MM-DD`.
- **Test otomatis:** `backend/internal/config/config_test.go` › `TestLoad_TC_M00_008`
- **Status:** belum dibuat

## TC-M00-009 — SPA belum di-build → "UI not built yet"

- **Level:** unit (Go, `httptest`) · **Kebutuhan:** M0 Deliver (`backend/internal/webui`: bila
  `dist/` kosong server menjawab "UI not built yet")
- **Prasyarat:** file system UI hanya berisi `.keep` (kondisi repository sebelum `make build`).
- **Langkah:**
  1. Kirim `GET /` dan `GET /packages`.
- **Hasil yang diharapkan:**
  - Status 503, `Content-Type: text/plain; charset=utf-8`, body memuat "UI not built yet".
- **Test otomatis:** `backend/internal/webui/webui_test.go` › `TestHandler_TC_M00_009`
- **Status:** belum dibuat

## TC-M00-010 — Path non-API yang tidak dikenal mendapat `index.html`

- **Level:** unit (Go, `httptest`) · **Kebutuhan:** M0 Deliver (webui menyajikan `index.html`
  untuk path non-API yang tidak dikenal)
- **Prasyarat:** file system UI tiruan berisi `index.html`, `robots.txt` dan
  `_app/immutable/entry/start.abc123.js`.
- **Langkah:**
  1. `GET /`, `GET /packages/PS-07/pfmea`, `GET /findings?rule=R04`.
  2. `GET /index.html`.
  3. `GET /robots.txt`.
  4. `POST /packages`.
- **Hasil yang diharapkan:**
  1. Status 200, isi `index.html`, `Content-Type: text/html; charset=utf-8`,
     `Cache-Control: no-cache`.
  2. Status 200 dengan isi `index.html` (tidak di-redirect).
  3. Status 200 dengan isi `robots.txt` (file yang ada disajikan apa adanya).
  4. Status 405 dengan header `Allow: GET, HEAD`.
- **Test otomatis:** `backend/internal/webui/webui_test.go` › `TestHandler_TC_M00_010`
- **Status:** belum dibuat

## TC-M00-011 — Path `/api/...` yang tidak dikenal mendapat 404 Problem, bukan `index.html`

- **Level:** unit (Go, `httptest`) · **Kebutuhan:** M0 Deliver (`index.html` hanya untuk path
  non-API), `docs/05-api.md` §2 (Problem Details, kode `not_found`)
- **Prasyarat:** handler server dengan UI tiruan yang sudah di-build.
- **Langkah:**
  1. `GET /api`, `GET /api/`, `GET /api/v1/nothing`, `POST /api/v1/packages`.
- **Hasil yang diharapkan:**
  - Status 404, `Content-Type: application/problem+json`, body berisi
    `"type":"urn:pfmea:problem:not_found"`, `"title":"Not found"`, `"status":404`,
    `"code":"not_found"`.
  - Body tidak memuat isi `index.html`; tidak ada redirect 301 untuk `/api`.
- **Test otomatis:** `backend/internal/httpapi/server_test.go` › `TestRouter_TC_M00_011`
- **Status:** belum dibuat

## TC-M00-012 — Aset ber-hash mendapat header cache permanen; aset yang tidak ada 404

- **Level:** unit (Go, `httptest`) · **Kebutuhan:** M0 Deliver (aset ber-hash mendapat
  `Cache-Control: public, max-age=31536000, immutable`)
- **Prasyarat:** file system UI tiruan seperti TC-M00-010, ditambah `.keep`.
- **Langkah:**
  1. `GET /_app/immutable/entry/start.abc123.js`.
  2. `GET /_app/immutable/entry/missing.js` dan `GET /_app/immutable/`.
  3. `GET /.keep`.
- **Hasil yang diharapkan:**
  1. Status 200, `Cache-Control: public, max-age=31536000, immutable`, `Content-Type` memuat
     "javascript".
  2. Status 404 (bukan `index.html`, bukan daftar isi folder).
  3. Isi `index.html` (dotfile tidak pernah disajikan sebagai file).
- **Test otomatis:** `backend/internal/webui/webui_test.go` › `TestHandler_TC_M00_012`
- **Status:** belum dibuat

## TC-M00-013 — Binary `pfmea`: usage, subcommand milestone berikutnya, konfigurasi salah

- **Level:** unit (Go) · **Kebutuhan:** M0 Deliver (`backend/cmd/pfmea/main.go` dengan `serve`),
  `docs/03-architecture.md` §3.2
- **Prasyarat:** fungsi `run` dipanggil langsung dengan argumen, environment dan output tiruan.
- **Langkah:**
  1. Tanpa argumen.
  2. `pfmea help`.
  3. `pfmea migrate up`.
  4. `pfmea unknown`.
  5. `pfmea serve` dengan environment kosong.
- **Hasil yang diharapkan:**
  1. Usage (memuat `serve`) di stderr, kode keluar 2.
  2. Usage di stdout, kode keluar 0.
  3. "pfmea migrate: available from M1" di stderr, kode keluar 2.
  4. Pesan "unknown command" + usage, kode keluar 2.
  5. Kode keluar 1; stderr memuat "DATABASE_URL is required" dan "APP_BASE_URL is required"
     sekaligus.
- **Test otomatis:** `backend/cmd/pfmea/main_test.go` › `TestRun_TC_M00_013`
- **Status:** belum dibuat

## TC-M00-014 — `pfmea serve` berjalan, menjawab `/healthz`, lalu berhenti dengan rapi

- **Level:** integrasi (Go, proses `serve` sungguhan di port acak) · **Kebutuhan:** M0 Deliver
  (`serve`), M0 Done when (`/healthz` → 200)
- **Prasyarat:** `DATABASE_URL` dan `APP_BASE_URL` di-set, `HTTP_ADDR=127.0.0.1:0`.
- **Langkah:**
  1. Jalankan `run(ctx, ["serve"], …)` di goroutine.
  2. Baca alamat dari log JSON "listening".
  3. `GET http://<alamat>/healthz`.
  4. Batalkan konteks (setara SIGINT/SIGTERM).
- **Hasil yang diharapkan:**
  - Log "listening" berisi alamat; `/healthz` → 200.
  - Setelah konteks dibatalkan `run` selesai dengan kode 0 dalam waktu batas shutdown.
  - Log tidak memuat kata sandi dari `DATABASE_URL`.
- **Test otomatis:** `backend/cmd/pfmea/main_test.go` › `TestServe_TC_M00_014`
- **Status:** belum dibuat

## TC-M00-015 — Target Makefile yang berjalan di M0 tersedia

- **Level:** integrasi (Go menjalankan `make -n`) · **Kebutuhan:** M0 Deliver (Makefile dengan
  semua target `docs/03-architecture.md` §3.3; yang berjalan di M0: `tools`, `db`, `dev`,
  `test`, `lint`, `build`, `check`)
- **Prasyarat:** root repository; `make` terpasang.
- **Langkah:**
  1. Jalankan `make -n <target>` untuk `tools`, `db`, `dev`, `test`, `lint`, `build`, `check`.
- **Hasil yang diharapkan:**
  - Semua selesai dengan kode 0 dan tidak satu pun mencetak "available from".
  - `tools` memasang sqlc, goose, oapi-codegen, golangci-lint dan menjalankan `npm ci` di
    `web/`; `test` menjalankan `go test` dengan `-race` dan Vitest; `lint` menjalankan
    golangci-lint, `svelte-check`, eslint dan `redocly lint`; `build` mem-build web, menyalin ke
    `backend/internal/webui/dist` dan menjalankan `go build`.
  - `check` menjalankan `gen`, `lint`, `test`, `test-rules` dalam urutan itu.
- **Test otomatis:** `backend/internal/repotest/makefile_test.go` › `TestTargets_TC_M00_015`
- **Status:** belum dibuat

## TC-M00-016 — Target Makefile milestone berikutnya mencetak "available from M<n>" dengan kode 0

- **Level:** integrasi (Go menjalankan `make`) · **Kebutuhan:** M0 Deliver (target lain mencetak
  "available from M<n>" dan keluar 0 sampai milestone-nya)
- **Prasyarat:** root repository; `make` terpasang.
- **Langkah:**
  1. Jalankan `make migrate`, `make migrate-down`, `make seed`, `make test-rules`, `make gen`.
  2. Jalankan `make e2e`.
  3. Jalankan `make perf`.
- **Hasil yang diharapkan:**
  1. Masing-masing mencetak "available from M1", kode keluar 0.
  2. Mencetak "available from M2", kode keluar 0.
  3. Mencetak "available from M13", kode keluar 0.
- **Test otomatis:** `backend/internal/repotest/makefile_test.go` › `TestPlaceholders_TC_M00_016`
- **Status:** belum dibuat

## TC-M00-017 — `.gitignore` mengabaikan rahasia dan hasil build

- **Level:** integrasi (Go menjalankan `git check-ignore`) · **Kebutuhan:** M0 Deliver
  (`.gitignore` mengabaikan salinan build web dengan `backend/internal/webui/dist/*` dan
  menyimpan `!backend/internal/webui/dist/.keep`), aturan "never commit secrets or build
  output" di `CLAUDE.md`
- **Prasyarat:** root repository (repository git).
- **Langkah:**
  1. `git check-ignore` untuk `.env`, `.env.local`, `bin/pfmea`, `web/build/index.html`,
     `web/node_modules/x`, `web/.svelte-kit/x`, `backend/internal/webui/dist/index.html`,
     `tmp/pfmea`, `perf-report.json`.
  2. `git check-ignore` untuk `backend/internal/webui/dist/.keep` dan `.env.example`.
- **Hasil yang diharapkan:**
  1. Semua diabaikan.
  2. Keduanya tidak diabaikan.
- **Test otomatis:** `backend/internal/repotest/gitignore_test.go` › `TestGitignore_TC_M00_017`
- **Status:** belum dibuat

## TC-M00-018 — App shell menampilkan teks bahasa Inggris dari `en.ts`

- **Level:** unit (Vitest + Testing Library, jsdom) · **Kebutuhan:** M0 Deliver (app shell
  `docs/08-screens.md` §1, semua teks UI di `src/lib/i18n/en.ts`), `docs/08-screens.md` §14
- **Prasyarat:** komponen `AppShell` dirender untuk path `/`.
- **Langkah:**
  1. Render `AppShell` dengan `pathname = '/'`.
- **Hasil yang diharapkan:**
  - Top bar: tautan "PFD · PFMEA · CP System" ke `/`; kotak pencarian dengan placeholder
    "Search packages, steps, characteristics" (nonaktif sampai M11).
  - Sidebar: grup "MONITORING" (Dashboard → `/`, Consistency check → `/findings`), "PACKAGES"
    (All packages → `/packages`), "TEMPLATE" (Template General → `/template-general`).
  - Semua teks itu berasal dari `en.ts` (nilai dibandingkan dengan objek `en`).
  - Tidak ada tombol notifikasi (lonceng, Tahap 3) dan tidak ada label bahasa Indonesia dari
    mockup (misalnya "Cek konsistensi", "Paket").
- **Test otomatis:** `web/src/lib/components/AppShell.test.ts` ›
  `TC-M00-018 app shell shows the English texts from en.ts`
- **Status:** belum dibuat

## TC-M00-019 — Menu aktif dan breadcrumb sesuai halaman

- **Level:** unit (Vitest) · **Kebutuhan:** M0 Deliver (app shell + halaman placeholder),
  `docs/08-screens.md` §1 (breadcrumb di setiap halaman)
- **Prasyarat:** fungsi navigasi di `src/lib/navigation.ts` dan komponen `AppShell`.
- **Langkah:**
  1. Hitung breadcrumb untuk `/`, `/findings`, `/packages`, `/template-general` dan
     `/packages/PS-07`.
  2. Render `AppShell` dengan `pathname = '/findings'`.
- **Hasil yang diharapkan:**
  1. `/` → "Dashboard"; `/findings` → "Dashboard / Consistency check"; `/packages` →
     "Dashboard / Packages"; `/template-general` → "Dashboard / Template General";
     `/packages/PS-07` → "Dashboard / Packages" (detail paket baru ada di M4).
  2. Hanya menu "Consistency check" yang punya `aria-current="page"`; breadcrumb tampil dengan
     item terakhir sebagai halaman saat ini.
- **Test otomatis:** `web/src/lib/navigation.test.ts` ›
  `TC-M00-019 active navigation item and breadcrumb follow the path`
- **Status:** belum dibuat

## TC-M00-020 — Proxy dev Vite meneruskan `/api` (termasuk WebSocket) ke server Go

- **Level:** unit (Vitest) · **Kebutuhan:** M0 Deliver (Vite dev proxy `/api` → `:8080`,
  termasuk WebSocket)
- **Prasyarat:** fungsi `apiProxy` di `web/vite.proxy.ts`.
- **Langkah:**
  1. `apiProxy(undefined)` dan `apiProxy(':8080')`.
  2. `apiProxy(':8081')` dan `apiProxy('0.0.0.0:9090')`.
- **Hasil yang diharapkan:**
  1. `/api` → target `http://127.0.0.1:8080`, `ws: true`, `changeOrigin: false` (header
     `Origin` asli tetap dikirim untuk cek Origin mulai M2).
  2. Target mengikuti port `HTTP_ADDR`: `http://127.0.0.1:8081` dan `http://127.0.0.1:9090`.
- **Test otomatis:** `web/vite.proxy.test.ts` ›
  `TC-M00-020 dev proxy forwards /api and WebSocket to the Go server`
- **Status:** belum dibuat

## TC-M00-021 — `make build` lalu `./bin/pfmea serve` menyajikan SPA di :8080

- **Level:** integrasi (CI dan lokal) · **Kebutuhan:** M0 Done when (`make build &&
  ./bin/pfmea serve` serves the SPA on :8080)
- **Prasyarat:** `make tools` sudah dijalankan.
- **Langkah:**
  1. `make build`.
  2. Jalankan `./bin/pfmea serve` dengan `DATABASE_URL` dan
     `APP_BASE_URL=http://localhost:8080`.
  3. `curl` ke `/`, `/packages/PS-07/pfmea`, `/healthz`, dan satu aset `/_app/immutable/...`
     yang dirujuk `index.html`.
- **Hasil yang diharapkan:**
  - `/` dan deep link → 200 HTML SvelteKit (`index.html` hasil build) yang merujuk aset
    dengan path absolut `/_app/`.
  - Aset → 200 dengan `Cache-Control: public, max-age=31536000, immutable`.
  - `/healthz` → 200.
- **Test otomatis:** `.github/workflows/ci.yml` › langkah
  `TC-M00-021 build binary and serve SPA`
- **Status:** belum dibuat

## TC-M00-022 — `make db` menjalankan PostgreSQL 18 yang sehat dengan JIT mati

- **Level:** integrasi (CI dan lokal, Docker) · **Kebutuhan:** M0 Deliver (`docker-compose.yml`
  service `db` `postgres:18` dengan healthcheck dan volume), `docs/03-architecture.md` §4
  (`jit = off`)
- **Prasyarat:** Docker berjalan; `cp .env.example .env`.
- **Langkah:**
  1. `make db`.
  2. `docker compose exec db psql -U pfmea -d pfmea -Atc 'show server_version' -c 'show jit'`.
  3. `docker compose down -v`.
- **Hasil yang diharapkan:**
  - `make db` selesai setelah container `healthy`.
  - Versi server diawali `18`, `jit` = `off`.
- **Test otomatis:** `.github/workflows/ci.yml` › langkah
  `TC-M00-022 make db starts PostgreSQL 18`
- **Status:** belum dibuat

## TC-M00-023 — `make tools` + `make check` hijau di CI dan tidak mengubah file

- **Level:** integrasi (GitHub Actions) · **Kebutuhan:** M0 Deliver (`.github/workflows/ci.yml`),
  M0 Done when (`make check` lokal dan run CI pertama hijau), `docs/11-testing.md` §1
  (`make gen` tidak boleh mengubah file yang di-track)
- **Prasyarat:** push ke `main` (atau pull request).
- **Langkah:**
  1. Workflow menjalankan `make tools` dan `make check` di `ubuntu-latest` dengan Go dari
     `go.mod` dan Node 24.
  2. Setelah semua langkah, `git status --porcelain` diperiksa.
- **Hasil yang diharapkan:**
  - Semua langkah hijau; working tree tetap bersih.
- **Test otomatis:** `.github/workflows/ci.yml` › job `check` (langkah `make check` dan
  `TC-M00-023 working tree stays clean`)
- **Status:** belum dibuat

## TC-M00-024 — `make dev` menampilkan app shell di :5173 dan `/healthz` di :8080

- **Level:** manual (E2E Playwright baru dipakai mulai M2) · **Kebutuhan:** M0 Done when
  (`make dev` → http://localhost:5173 shows the shell; `curl -i localhost:8080/healthz` → 200)
- **Prasyarat:** `make tools`, `cp .env.example .env`, `make db`.
- **Langkah:**
  1. `make dev`.
  2. Buka http://localhost:5173, klik setiap menu sidebar.
  3. `curl -i localhost:8080/healthz` dan `curl -i localhost:5173/api/v1/nothing`.
  4. Ubah satu file Go (misalnya komentar) dan simpan.
- **Hasil yang diharapkan:**
  - App shell tampil dengan teks bahasa Inggris; setiap menu membuka halaman placeholder-nya
    dan breadcrumb berubah.
  - `/healthz` → 200; `/api/v1/nothing` lewat proxy Vite → 404 Problem dari server Go.
  - Server Go di-build ulang dan dijalankan ulang otomatis (live reload).
- **Test otomatis:** — (manual)
- **Status:** belum dibuat

## TC-M00-025 — Clone baru mengikuti panduan instalasi sampai `make check` hijau

- **Level:** manual · **Kebutuhan:** M0 Done when (clone baru dari GitHub lolos langkah
  instalasi `docs/build-guide.md` yang sudah ada di M0 sampai `make check` hijau)
- **Prasyarat:** prasyarat `docs/build-guide.md` §2.1 terpenuhi; folder sementara kosong.
- **Langkah:**
  1. Clone repository dari GitHub ke folder sementara.
  2. Ikuti `docs/build-guide.md` §2.2 persis seperti tertulis untuk perintah yang tersedia di
     M0 (`make tools`, `cp .env.example .env`, `make db`, `make check`).
  3. Hapus folder sementara.
- **Hasil yang diharapkan:**
  - Setiap perintah berhasil tanpa langkah tambahan; `make check` hijau. Bila ada perbedaan,
    panduan diperbaiki.
- **Test otomatis:** — (manual)
- **Status:** belum dibuat

## TC-M00-026 — Komentar bahasa Indonesia dan teks tampilan bahasa Inggris

- **Level:** manual (review) · **Kebutuhan:** M0 Done when ("Every hand-written file has its
  comments in Indonesian; every visible text is English"), `CLAUDE.md` Language policy
- **Prasyarat:** semua file M0.
- **Langkah:**
  1. Jalankan agen `spec-reviewer` pada perubahan M0.
  2. Periksa setiap file buatan tangan: komentar awal file, komentar setiap fungsi, tipe,
     konstanta dan test; teks yang dilihat pengguna hanya di `en.ts` / `en.go`.
- **Hasil yang diharapkan:**
  - Tidak ada temuan blocker atau major tentang bahasa; semua teks UI dan server berbahasa
    Inggris, semua komentar berbahasa Indonesia.
- **Test otomatis:** — (manual; agen `spec-reviewer`)
- **Status:** belum dibuat
