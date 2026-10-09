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
  serve    Start the HTTP server
  help     Show this help
`
	// CLIUnknownCommand dipakai untuk subcommand yang tidak dikenal.
	CLIUnknownCommand = "pfmea: unknown command %q"
	// CLINotAvailable dipakai untuk subcommand yang baru dibuat di milestone berikutnya.
	CLINotAvailable = "pfmea %s: available from %s"
	// CLICommandFailed dipakai bila subcommand gagal sebelum log JSON aktif (misalnya
	// konfigurasi salah); argumen kedua adalah pesan error.
	CLICommandFailed = "pfmea %s: %v"
)
