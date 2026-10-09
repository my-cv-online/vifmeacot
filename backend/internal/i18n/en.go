// Package i18n menyimpan semua teks bahasa Inggris yang dibuat server: pesan validasi
// konfigurasi, judul Problem Details, pesan CLI dan halaman fallback. Teks yang dilihat
// pengguna tidak boleh ditulis langsung di paket lain (CLAUDE.md, Language policy).
package i18n

// Pesan validasi konfigurasi. Placeholder pertama selalu nama variabel lingkungan supaya
// operator langsung tahu variabel mana yang harus diperbaiki.
const (
	// ConfigInvalid adalah judul daftar masalah konfigurasi saat server gagal start.
	ConfigInvalid = "invalid configuration"
	// ConfigRequired dipakai untuk variabel wajib yang kosong.
	ConfigRequired = "%s is required"
	// ConfigPostgresURL dipakai bila DATABASE_URL bukan URL PostgreSQL.
	ConfigPostgresURL = "%s must be a postgres:// or postgresql:// URL"
	// ConfigAbsoluteURL dipakai bila APP_BASE_URL bukan origin http(s) (skema dan host saja).
	ConfigAbsoluteURL = "%s must be an origin such as https://pfmea.example.com or http://localhost:5173 (no path or query)"
	// ConfigListenAddr dipakai bila alamat listen tidak berbentuk host:port.
	ConfigListenAddr = "%s must be host:port, for example :8080"
	// ConfigDuration dipakai bila durasi tidak bisa dibaca atau tidak positif.
	ConfigDuration = "%s must be a positive duration such as 12h or 30m"
	// ConfigMinInt dipakai bila bilangan bulat tidak valid atau di bawah batas minimum.
	ConfigMinInt = "%s must be a whole number of at least %d"
	// ConfigNotGreater dipakai bila satu nilai melebihi nilai lain yang menjadi batasnya.
	ConfigNotGreater = "%s must not be greater than %s"
	// ConfigLogLevel dipakai bila LOG_LEVEL tidak dikenal.
	ConfigLogLevel = "%s must be one of debug, info, warn, error"
	// ConfigCIDRList dipakai bila daftar CIDR tidak valid.
	ConfigCIDRList = "%s must be a comma-separated list of CIDR ranges such as 127.0.0.1/32"
	// ConfigBool dipakai bila nilai boolean tidak dikenal.
	ConfigBool = "%s must be true or false"
	// ConfigDate dipakai bila tanggal tidak berformat YYYY-MM-DD.
	ConfigDate = "%s must be a date in the form YYYY-MM-DD"
	// ConfigFakeTodayIgnored adalah peringatan bila DEV_FAKE_TODAY di-set tanpa DEV_MODE=true.
	ConfigFakeTodayIgnored = "%s is ignored because DEV_MODE is not true"
)

// Teks HTTP yang dikirim server ke browser atau klien API.
const (
	// ProblemNotFoundTitle adalah judul Problem Details untuk kode not_found (docs/05-api.md §2).
	ProblemNotFoundTitle = "Not found"
	// NotFound adalah body teks biasa untuk file atau path yang tidak ada (misalnya aset SPA).
	NotFound = "Not found"
	// UINotBuilt ditampilkan bila binary dibangun tanpa hasil build web (dist/ kosong).
	UINotBuilt = "UI not built yet. Run make build to embed the web application."
	// MethodNotAllowed ditampilkan bila halaman web diminta dengan method selain GET/HEAD.
	MethodNotAllowed = "Method not allowed"
)

// Teks antarmuka baris perintah binary pfmea.
const (
	// CLIUsage adalah bantuan singkat binary; subcommand milestone berikutnya ditambahkan saat
	// milestone itu mengimplementasikannya.
	CLIUsage = `Usage: pfmea <command>

Commands:
  serve                    Start the HTTP server
  migrate up|down|status   Apply, roll back (DEV_MODE=true only) or list database migrations
  help                     Show this help
`
	// CLIUnknownCommand dipakai untuk subcommand yang tidak dikenal.
	CLIUnknownCommand = "pfmea: unknown command %q"
	// CLINotAvailable dipakai untuk subcommand yang baru dibuat di milestone berikutnya.
	CLINotAvailable = "pfmea %s: available from %s"
	// CLICommandFailed dipakai bila subcommand gagal sebelum log JSON aktif (misalnya
	// konfigurasi salah); argumen kedua adalah pesan error.
	CLICommandFailed = "pfmea %s: %v"
)

// Teks perintah `pfmea migrate` (docs/03-architecture.md §3.2).
const (
	// MigrateUsage ditampilkan bila sub-perintah migrate tidak ada atau tidak dikenal.
	MigrateUsage = "Usage: pfmea migrate up|down|status"
	// MigrateDownNeedsDevMode menolak migrate down di luar mode pengembangan karena semua data
	// ikut terhapus.
	MigrateDownNeedsDevMode = "pfmea migrate down deletes all data and requires DEV_MODE=true"
	// MigrateDatabaseURLInvalid dipakai bila URL database tidak bisa dibaca driver; URL tidak
	// ikut dicetak supaya kata sandi tidak bocor.
	MigrateDatabaseURLInvalid = "DATABASE_URL cannot be parsed"
	// MigrateApplied melaporkan satu file migrasi yang baru diterapkan.
	MigrateApplied = "applied %s"
	// MigrateNoPending dipakai bila semua migrasi sudah diterapkan.
	MigrateNoPending = "no pending migrations"
	// MigrateRolledBack melaporkan satu file migrasi yang dibatalkan.
	MigrateRolledBack = "rolled back %s"
	// MigrateNothingToRollBack dipakai bila belum ada migrasi yang diterapkan.
	MigrateNothingToRollBack = "nothing to roll back"
	// MigrateStatusLine adalah satu baris `migrate status`: nama file dan statusnya
	// (applied atau pending).
	MigrateStatusLine = "%-28s %s"
	// MigrateVersion adalah baris terakhir `migrate status`: versi skema di database.
	MigrateVersion = "version %d"
)

// Status migrasi yang dicetak `pfmea migrate status`.
const (
	// MigrateStateApplied berarti file migrasi sudah diterapkan di database.
	MigrateStateApplied = "applied"
	// MigrateStatePending berarti file migrasi belum diterapkan.
	MigrateStatePending = "pending"
)
