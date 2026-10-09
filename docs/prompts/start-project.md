# Prompt sesi pertama Claude Code

Cara pakai: buka folder paket ini dengan Claude Code, ganti `OWNER/pfmea-system` pada baris
pertama prompt (di bawah garis) dengan akun atau organisasi GitHub dan nama repository Anda,
lalu salin semua teks di bawah garis sebagai pesan pertama. Tetap di depan layar sampai Claude
menunjukkan rencana test case M0, karena rencana itu perlu Anda setujui. Sesi berikutnya cukup
`/milestone M1`, `/milestone M2`, dan seterusnya (satu milestone per sesi).

\---

Repository GitHub: https://github.com/my-cv-online/vifmeacot.git (public)

Ini sesi pertama project "PFMEA · Control Plan · PFD system". Folder ini berisi paket
spesifikasi Tahap 1 (`README.md`, `CLAUDE.md`, `docs/`). Pakai nama repository di baris
pertama untuk setiap perintah di bawah yang memuat `OWNER/pfmea-system`. Ada lima langkah:

* Langkah 1–2: cek prasyarat, hubungkan folder ini ke repository GitHub di atas, lalu push
semua isinya ke branch `main`;
* Langkah 3: tulis `docs/build-guide.md`, rangkuman tahapan pembuatan web mulai dari instalasi
awal project sampai tahap-tahap berikutnya;
* Langkah 4: kerjakan milestone M0 (Scaffold) dengan cara test-first, lalu push hasilnya sampai
CI hijau;
* Langkah 5: periksa semuanya dan laporkan.

Jangan mengerjakan M1 atau milestone lain di sesi ini.

## Aturan untuk seluruh project

1. **Test case dulu.** Setiap fitur baru, pembaruan atau perbaikan bug dimulai dengan
menentukan test case-nya di `docs/test-cases/` (format di `docs/test-cases/README.md`, isi
berbahasa Indonesia), lalu di-commit dan di-push sebelum kode apa pun. Setelah itu tulis test
otomatisnya (nama test memuat ID test case, misalnya `TestHealthz\_TC\_M00\_001` atau
`test('TC-M00-006 …')`), jalankan dan pastikan gagal karena alasan yang benar, baru tulis
kode sampai lulus. Test tidak boleh dilemahkan atau dihapus supaya lulus.
2. **Semua masuk ke `main` di GitHub.** Hanya satu branch, `main`, dengan remote `origin`.
Commit kecil memakai Conventional Commits berbahasa Inggris (`feat:`, `fix:`, `test:`,
`docs:`, `chore:`, `refactor:`, `ci:`). Setiap commit harus lolos `make check`, lalu
langsung di-push. Test yang masih gagal tidak di-commit sendirian; test itu di-commit bersama
kode yang membuatnya lulus, jadi `main` di GitHub tidak pernah merah. Sebelum Makefile ada
(awal M0), commit yang hanya berisi dokumen boleh di-push tanpa `make check`. Dilarang
force-push, mengubah commit yang sudah di-push, membuat branch lain, atau meng-commit rahasia
(`.env`, token, kata sandi asli) dan hasil build.
3. **Komentar kode berbahasa Indonesia.** Semua kode yang ditulis tangan (Go, TypeScript,
JavaScript, Svelte, HTML, CSS, SQL, Makefile, YAML, Dockerfile, shell) diberi komentar bahasa
Indonesia supaya mudah dipahami tim: komentar singkat di awal setiap file tentang gunanya,
komentar untuk setiap fungsi, tipe, konstanta dan test, serta komentar pada langkah yang
tidak jelas di dalam fungsi. Jelaskan alasannya, jangan mengulang isi kode. Doc comment Go
diawali nama identifier, misalnya `// LoadConfig membaca variabel lingkungan lalu memvalidasinya.` Kode hasil generator dan file JSON (tidak mendukung komentar) dikecualikan.
4. **Bahasa Inggris untuk tampilan dan penamaan.** Semua teks yang dilihat pengguna (label,
pesan, error, pesan aturan cek, judul halaman, teks buatan aplikasi di file Excel dan nama
file ekspor) berbahasa Inggris. Teks UI disimpan di `web/src/lib/i18n/en.ts`, teks server di
`backend/internal/i18n/en.go`, dan pesan aturan cek di baris `-- message:` file SQL aturan.
Nama file, folder, variabel, fungsi, tabel, endpoint, log dan commit message juga berbahasa
Inggris.
5. Dokumen untuk tim (`README.md`, `docs/build-guide.md`, `docs/test-cases/`) berbahasa
Indonesia; dokumen spesifikasi untuk AI (`CLAUDE.md`, `docs/00`–`docs/11`) tetap berbahasa
Inggris.
6. Jika spesifikasi tidak menjawab atau saling bertentangan, tanya saya dulu, lalu catat
keputusannya di dokumen yang relevan.

## Langkah 1 — Baca dan cek prasyarat

* Baca `CLAUDE.md`, `README.md`, `docs/00-vision.md` §4, `docs/02-phase1-scope.md` §4–5,
`docs/03-architecture.md` §2, §3 dan §8, `docs/10-milestones.md`, `docs/11-testing.md`,
`docs/test-cases/README.md`, `docs/test-cases/\_template.md` dan
`.claude/skills/milestone/SKILL.md`.
* Cek alat di komputer ini dan tampilkan versinya dalam satu tabel: `git`, `gh`
(`gh auth status`), Docker (`docker info`), Go 1.27 atau lebih baru, Node.js 24 LTS (minimal
22.17), `make`. Cek juga `git config user.name` dan `git config user.email`.
* Token `gh` harus punya scope `repo` dan `workflow` (tanpa `workflow`, push
`.github/workflows/ci.yml` lewat HTTPS ditolak). Jika belum ada, minta saya menjalankan
`gh auth refresh -h github.com -s workflow`. Jalankan `gh auth setup-git` supaya `git push`
memakai login `gh`.
* Jika Git, `gh`, Go, Node.js, `make` atau identitas Git belum siap, berhenti dan beri tahu saya
apa yang perlu dipasang atau diatur. Jangan memasang software sistem sendiri. Jika hanya
Docker yang belum berjalan, lanjutkan Langkah 2–3 dan minta saya menyalakannya sebelum
Langkah 4.

## Langkah 2 — Hubungkan ke GitHub dan push awal

* Jika baris "Repository GitHub" di atas masih berisi `OWNER`, tanyakan nama repository-nya
dulu.
* Jika folder ini belum repository git, jalankan `git init -b main`. Jika sudah, pastikan
branch aktifnya `main` dan working tree bersih.
* Buat `.gitignore` awal (M0 akan melengkapinya): `.env` dan `.env.\*` kecuali `.env.example`,
`/bin/`, `node\_modules/`, `.svelte-kit/`, `web/build/`, `backend/internal/webui/dist/\*`
kecuali `backend/internal/webui/dist/.keep`, `test-results/`, `playwright-report/`,
`coverage/`, `perf-report.json`, `\*.log`, `.DS\_Store`, `.claude/settings.local.json`.
* Sebelum commit pertama, periksa `git status`. Jika ada file yang tampak rahasia atau file di
atas 5 MB, berhenti dan tanya saya. Commit pertama: `chore: import phase 1 specification`.
* Repository di GitHub:

  * belum ada: `gh repo create OWNER/pfmea-system --private --source=. --remote=origin`;
  * sudah ada: cek `gh repo view OWNER/pfmea-system --json visibility,isEmpty,defaultBranchRef`.
Jika public, berhenti dan tanya saya (data PFMEA bersifat rahasia, repository harus
private). Jika remote `origin` belum ada, tambahkan dengan
`git remote add origin https://github.com/OWNER/pfmea-system.git`. Jika repository sudah
berisi commit (misalnya README atau LICENSE buatan GitHub), jangan force-push: tunjukkan
isinya dan minta izin saya sebelum menggabungkannya dengan
`git pull --no-rebase --no-commit --allow-unrelated-histories origin main` (jika bentrok,
isi paket ini yang dipakai), lalu commit `chore: merge initial GitHub files`.
* `git push -u origin main`. Pastikan default branch di GitHub adalah `main`
(`gh repo edit OWNER/pfmea-system --default-branch main` bila perlu).
* Lindungi `main` dari force-push dan penghapusan dengan ruleset (push biasa tetap boleh):

```bash
  gh api repos/OWNER/pfmea-system/rulesets -X POST --input - <<'JSON'
  {"name": "protect-main", "target": "branch", "enforcement": "active",
   "conditions": {"ref\_name": {"include": \["\~DEFAULT\_BRANCH"], "exclude": \[]}},
   "rules": \[{"type": "deletion"}, {"type": "non\_fast\_forward"}]}
  JSON
  ```

  Periksa hasilnya dengan `gh api repos/OWNER/pfmea-system/rules/branches/main`: daftar itu
harus memuat `deletion` dan `non\_fast\_forward`. Jika ruleset gagal dibuat atau tidak aktif
(paket langganan GitHub, misalnya Free, belum tentu mendukung ruleset di repository
private), lanjutkan tanpa ruleset dan sebutkan di laporan akhir.

## Langkah 3 — Tulis `docs/build-guide.md` (bahasa Indonesia)

Rangkum dari dokumen yang ada (`README.md`, `docs/00`, `docs/02`, `docs/03`, `docs/10`,
`docs/11`); jangan menambahkan hal yang tidak ada di spesifikasi, kecuali di bagian pemecahan
masalah. Isinya:

1. **Gambaran singkat:** apa yang dibangun, isi Tahap 1, dan Tahap 2–4 sesudahnya (satu
paragraf dan satu tabel).
2. **Instalasi awal project:** prasyarat beserta versinya (termasuk WSL2 untuk Windows dan
login `gh`), clone repository, `make tools` (alat Go dan `npm ci` untuk `web/`),
`cp .env.example .env`, `make db`, `make migrate`, `make seed`, `make dev`, alamat yang
dibuka di browser, dan pengguna demo (kata sandi `pfmea-dev-2026` hanya untuk
pengembangan). Tandai perintah yang baru tersedia di milestone tertentu, misalnya
"tersedia mulai M1".
3. **Alur kerja test-first** untuk setiap perubahan, langkah demi langkah: test case → commit
dan push → test otomatis (dijalankan, harus gagal) → kode → `make check` → commit test dan
kode bersama-sama → push → CI hijau. Beri contoh untuk empat jenis perubahan: fitur milestone
(`docs/test-cases/M05-pfd-editor.md`, `TC-M05-001`), pembaruan fitur yang sudah ada
(lanjutkan nomor di file milestone pemiliknya), perbaikan bug (`docs/test-cases/bugs.md`,
`TC-BUG-001`) dan aturan cek baru (`/add-rule`, `TC-RULE-<KODE>-1`).
4. **Tabel tahapan M0–M13** dengan kolom: Tahap, Isi singkat, File test case, User story E2E
(`docs/11-testing.md` §4), Input dari perusahaan (`docs/02-phase1-scope.md` §5), Status.
Status awal "belum mulai"; milestone yang selesai diisi "selesai (<tanggal>)".
5. **Gate 1:** kriteria penerimaan Tahap 1 (`docs/11-testing.md` §6).
6. **Menjalankan test:** `make check`, `make test`, `make test-rules`, `make e2e`, `make perf`,
dan cara menjalankan satu test berdasarkan ID (misalnya
`go test ./backend/... -run TC\_M00\_001`, atau dari folder `web/`
`npx playwright test -g "TC-M04-001"`).
7. **Git dan GitHub:** aturan branch `main`, format commit, kapan push, cara melihat CI
(`gh run list`, `gh run view <id> --log-failed`), dan yang dilakukan bila push ditolak
(`git pull --rebase` lalu push lagi, tidak pernah force-push).
8. **Aturan bahasa:** tabel seperti di `CLAUDE.md`.
9. **Pemecahan masalah:** Docker belum berjalan, port 5432/8080/5173 sudah terpakai,
`make gen` mengubah file, CI merah, token `gh` tanpa scope `workflow`, folder project di WSL2
yang lambat.

Commit `docs: add build guide`, lalu push.

## Langkah 4 — Kerjakan M0 dengan test-first

Ikuti `.claude/skills/milestone/SKILL.md` untuk M0 (sama seperti `/milestone M0`). M0 tidak
mengubah spesifikasi yang sudah ada (skema, aturan cek, kontrak API).

1. Susun rencana di plan mode dan tunjukkan ke saya: daftar test case M0 (ID `TC-M00-001` dan
seterusnya, level, skenario, hasil yang diharapkan, butir "Deliver" atau "Done when" yang
dibuktikan) serta file yang akan dibuat. Minimal mencakup:

   * `GET /healthz` menjawab 200; `GET /readyz` menjawab 200 bila semua pemeriksaan kesiapan
lulus dan 503 bila ada yang gagal (di M0 diuji dengan pemeriksaan tiruan; pemeriksaan
database baru ditambahkan di M1);
   * konfigurasi: nilai bawaan terpasang, dan semua error validasi (misalnya `DATABASE\_URL` dan
`APP\_BASE\_URL` kosong) tampil sekaligus;
   * SPA yang tertanam: "UI not built yet" bila hasil build kosong, `index.html` untuk path
non-API yang tidak dikenal, path `/api/...` yang tidak dikenal tidak dijawab dengan
`index.html`, dan header `Cache-Control` untuk aset ber-hash;
   * target Makefile: yang berjalan di M0, dan yang mencetak "available from M<n>" lalu keluar
dengan kode 0 (pembagiannya ada di `docs/10-milestones.md`, bagian M0);
   * app shell tampil dengan teks bahasa Inggris dari `en.ts` (level unit dengan Vitest; E2E
Playwright baru dipakai mulai M2);
   * `make check` hijau di lokal dan di CI.
2. Setelah saya setujui, tulis `docs/test-cases/M00-scaffold.md` (bahasa Indonesia, memakai
`\_template.md`), commit `test(M0): define test cases`, lalu push.
3. Tulis test otomatisnya, jalankan, dan pastikan gagal karena alasan yang benar. Jangan
di-commit dulu; test di-commit bersama kode yang membuatnya lulus.
4. Implementasikan semua butir "Deliver" M0 di `docs/10-milestones.md`, termasuk
`.github/workflows/ci.yml`. Semua komentar berbahasa Indonesia, semua teks UI berbahasa
Inggris. Kerjakan dalam langkah kecil; setiap langkah yang `make check`-nya hijau langsung
di-commit (test beserta kodenya) dan di-push.
5. Jalankan semua perintah "Done when" M0 dan `make check`, lalu perbarui kolom status di file
test case.
6. Uji panduan instalasi: clone repository dari GitHub ke folder sementara, ikuti bagian
instalasi di `docs/build-guide.md` persis seperti tertulis untuk perintah yang sudah
tersedia di M0 sampai `make check` hijau, perbaiki panduan bila ada yang berbeda, lalu
hapus folder sementara itu.
7. Jalankan agen `spec-reviewer` untuk seluruh perubahan M0 (ia membandingkan mulai dari commit
`test(M0): define test cases`, jadi yang sudah di-push ikut diperiksa), lalu perbaiki semua
temuan blocker dan major.
8. Centang M0 di `docs/10-milestones.md`, isi status M0 di tabel tahapan
`docs/build-guide.md`, commit `feat(M0): scaffold server, web shell and CI`, lalu push.
9. Tunggu CI untuk commit terakhir: ambil id run dengan
`gh run list --commit "$(git rev-parse HEAD)" --limit 1 --json databaseId,status,conclusion`
(ulangi beberapa detik kemudian bila belum muncul), lalu jalankan
`gh run watch <id> --exit-status` dengan timeout 10 menit atau di background. Jika gagal,
baca `gh run view <id> --log-failed`, perbaiki, lalu commit dan push lagi sampai hijau.

Jika sesi harus berhenti sebelum M0 selesai: push semua yang sudah hijau, biarkan pekerjaan
yang belum selesai tetap tidak di-commit (jangan di-reset atau dibuang), dan tulis daftarnya di
laporan; saya akan melanjutkannya dengan `/milestone M0` di sesi baru.

## Langkah 5 — Penutup dan laporan

Sebelum selesai, pastikan `git status` bersih (kecuali sesi berhenti lebih awal),
`git log origin/main..main` kosong (tidak ada commit yang belum di-push) dan run CI terakhir di
`main` hijau. Lalu laporkan dalam bahasa Indonesia:

* URL repository, daftar commit yang sudah di-push, dan apakah ruleset `main` aktif;
* ringkasan tahapan dari `docs/build-guide.md` (tabel M0–M13 versi singkat);
* test case M0 beserta hasilnya, dan tautan run CI;
* cara mencobanya di komputer saya (`make dev`, alamat, pengguna demo);
* pertanyaan terbuka dan input perusahaan yang dibutuhkan untuk milestone berikutnya;
* langkah berikutnya: sesi baru, ketik `/milestone M1`.

