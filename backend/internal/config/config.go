// Package config membaca konfigurasi aplikasi dari variabel lingkungan
// (docs/03-architecture.md §8), mengisi nilai bawaan, dan memvalidasi semuanya sekaligus supaya
// operator melihat seluruh kesalahan dalam satu kali start.
package config

import (
	"fmt"
	"log/slog"
	"net"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"time"

	"pfmea/backend/internal/i18n"
)

// Nilai bawaan sesuai tabel konfigurasi di docs/03-architecture.md §8.
const (
	// defaultHTTPAddr adalah alamat listen server HTTP.
	defaultHTTPAddr = ":8080"
	// defaultSessionTTL adalah umur sesi yang diperpanjang setiap kali dipakai.
	defaultSessionTTL = 12 * time.Hour
	// defaultExportDir adalah folder file ekspor yang dihasilkan.
	defaultExportDir = "/data/exports"
	// defaultExportTemplateDir adalah folder layout Excel milik perusahaan.
	defaultExportTemplateDir = "/data/templates"
	// defaultDBMaxConns adalah ukuran pool koneksi utama (API dan River).
	defaultDBMaxConns = 20
	// defaultRiverWorkers adalah jumlah worker antrean bawaan River.
	defaultRiverWorkers = 10
	// defaultRulePoolSize adalah ukuran pool koneksi terpisah untuk mesin aturan.
	defaultRulePoolSize = 8
	// defaultRuleParallelism adalah jumlah koneksi yang dipakai satu kali cek.
	defaultRuleParallelism = 4
	// defaultMetricsAllow adalah CIDR yang boleh membaca /metrics.
	defaultMetricsAllow = "127.0.0.1/32"
	// dateLayout adalah format tanggal DEV_FAKE_TODAY (YYYY-MM-DD).
	dateLayout = "2006-01-02"
)

// Config adalah konfigurasi aplikasi yang sudah divalidasi. Nilainya dibaca sekali saat start
// dan tidak diubah lagi.
type Config struct {
	// DatabaseURL adalah URL koneksi PostgreSQL (wajib).
	DatabaseURL string
	// HTTPAddr adalah alamat listen server HTTP, misalnya ":8080".
	HTTPAddr string
	// AppBaseURL adalah URL publik aplikasi (wajib); dipakai untuk cek Origin dan flag Secure
	// cookie mulai M2.
	AppBaseURL string
	// SessionTTL adalah umur sesi yang diperpanjang setiap kali dipakai.
	SessionTTL time.Duration
	// ExportDir adalah folder file ekspor.
	// TODO(M10): periksa bahwa folder ada dan bisa ditulis saat start.
	ExportDir string
	// ExportTemplateDir adalah folder layout Excel perusahaan.
	ExportTemplateDir string
	// DBMaxConns adalah ukuran pool koneksi utama.
	DBMaxConns int
	// RiverWorkers adalah jumlah worker antrean bawaan River.
	RiverWorkers int
	// RulePoolSize adalah ukuran pool koneksi mesin aturan.
	RulePoolSize int
	// RuleParallelism adalah jumlah koneksi dari pool aturan yang dipakai satu kali cek.
	RuleParallelism int
	// LogLevel adalah level log minimum.
	LogLevel slog.Level
	// MetricsAllow adalah daftar CIDR yang boleh membaca /metrics.
	MetricsAllow []netip.Prefix
	// DevMode mengaktifkan fitur pengembangan (seed-demo, log lebih rinci, proxy Vite).
	DevMode bool
	// DevFakeToday adalah tanggal "hari ini" palsu (YYYY-MM-DD) untuk E2E; kosong bila tidak
	// dipakai. Hanya terisi bila DevMode = true.
	DevFakeToday string
	// Warnings berisi peringatan yang tidak menghentikan start (misalnya variabel yang
	// diabaikan); dicatat ke log oleh pemanggil.
	Warnings []string

	// TODO(M13): LDAP_URL, LDAP_BIND_DN, LDAP_BIND_PASSWORD, LDAP_BASE_DN, LDAP_USER_FILTER,
	// LDAP_START_TLS dibaca saat login LDAP dibuat.
}

// ValidationError berisi semua masalah konfigurasi yang ditemukan dalam satu kali Load.
type ValidationError struct {
	// Problems adalah daftar masalah, satu kalimat bahasa Inggris per masalah, sesuai urutan
	// variabel di docs/03-architecture.md §8.
	Problems []string
}

// Error menggabungkan semua masalah menjadi teks multi-baris yang siap dicetak saat start gagal.
func (e *ValidationError) Error() string {
	var b strings.Builder
	b.WriteString(i18n.ConfigInvalid + ":")
	for _, p := range e.Problems {
		b.WriteString("\n  - ")
		b.WriteString(p)
	}
	return b.String()
}

// loader mengumpulkan nilai dan masalah selama Load supaya setiap variabel divalidasi tanpa
// menghentikan pemeriksaan variabel berikutnya.
type loader struct {
	// getenv adalah sumber nilai variabel (os.Getenv di produksi, map di test).
	getenv func(string) string
	// problems menampung masalah validasi.
	problems []string
}

// get membaca variabel dan membuang spasi di awal/akhir; nilai yang hanya berisi spasi
// dianggap kosong.
func (l *loader) get(key string) string {
	return strings.TrimSpace(l.getenv(key))
}

// fail mencatat satu masalah validasi.
func (l *loader) fail(format string, args ...any) {
	l.problems = append(l.problems, fmt.Sprintf(format, args...))
}

// required membaca variabel wajib dan mencatat masalah bila kosong.
func (l *loader) required(key string) (string, bool) {
	v := l.get(key)
	if v == "" {
		l.fail(i18n.ConfigRequired, key)
		return "", false
	}
	return v, true
}

// stringOr membaca variabel teks dengan nilai bawaan.
func (l *loader) stringOr(key, def string) string {
	if v := l.get(key); v != "" {
		return v
	}
	return def
}

// intAtLeast membaca bilangan bulat dengan nilai bawaan dan batas minimum. Nilai ok = false
// berarti variabel salah sehingga pemeriksaan silang yang memakainya dilewati.
func (l *loader) intAtLeast(key string, def, minimum int) (int, bool) {
	v := l.get(key)
	if v == "" {
		return def, true
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < minimum {
		l.fail(i18n.ConfigMinInt, key, minimum)
		return def, false
	}
	return n, true
}

// duration membaca durasi positif dengan nilai bawaan.
func (l *loader) duration(key string, def time.Duration) time.Duration {
	v := l.get(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil || d <= 0 {
		l.fail(i18n.ConfigDuration, key)
		return def
	}
	return d
}

// boolean membaca nilai true/false (juga 1/0, t/f) dengan nilai bawaan.
func (l *loader) boolean(key string, def bool) bool {
	v := l.get(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		l.fail(i18n.ConfigBool, key)
		return def
	}
	return b
}

// postgresURL memvalidasi DATABASE_URL: wajib dan berskema postgres:// atau postgresql://.
// Host boleh kosong bila socket Unix diberikan lewat parameter host (bentuk yang sah di pgx,
// misalnya postgres:///pfmea?host=/var/run/postgresql).
func (l *loader) postgresURL(key string) string {
	v, ok := l.required(key)
	if !ok {
		return ""
	}
	u, err := url.Parse(v)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") ||
		(u.Host == "" && u.Query().Get("host") == "") {
		l.fail(i18n.ConfigPostgresURL, key)
		return ""
	}
	return v
}

// originURL memvalidasi APP_BASE_URL sebagai origin: wajib, http(s), dengan nama host, tanpa
// path, query atau fragment. Nilai disimpan sebagai skema://host[:port] karena mulai M2
// dibandingkan langsung dengan header Origin browser.
func (l *loader) originURL(key string) string {
	v, ok := l.required(key)
	if !ok {
		return ""
	}
	u, err := url.Parse(v)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" ||
		u.User != nil || (u.Path != "" && u.Path != "/") || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
		l.fail(i18n.ConfigAbsoluteURL, key)
		return ""
	}
	return u.Scheme + "://" + u.Host
}

// listenAddr memvalidasi alamat host:port; port 0 diizinkan (port acak, dipakai test).
func (l *loader) listenAddr(key, def string) string {
	v := l.get(key)
	if v == "" {
		return def
	}
	_, port, err := net.SplitHostPort(v)
	if err != nil {
		l.fail(i18n.ConfigListenAddr, key)
		return def
	}
	if p, perr := strconv.Atoi(port); perr != nil || p < 0 || p > 65535 {
		l.fail(i18n.ConfigListenAddr, key)
		return def
	}
	return v
}

// logLevel membaca LOG_LEVEL (debug, info, warn, error).
func (l *loader) logLevel(key string) slog.Level {
	switch strings.ToLower(l.get(key)) {
	case "", "info":
		return slog.LevelInfo
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		l.fail(i18n.ConfigLogLevel, key)
		return slog.LevelInfo
	}
}

// cidrList membaca daftar CIDR yang dipisah koma.
func (l *loader) cidrList(key, def string) []netip.Prefix {
	v := l.stringOr(key, def)
	var out []netip.Prefix
	for _, part := range strings.Split(v, ",") {
		p, err := netip.ParsePrefix(strings.TrimSpace(part))
		if err != nil {
			l.fail(i18n.ConfigCIDRList, key)
			return nil
		}
		out = append(out, p)
	}
	return out
}

// date membaca tanggal opsional berformat YYYY-MM-DD; kosong berarti tidak di-set.
func (l *loader) date(key string) string {
	v := l.get(key)
	if v == "" {
		return ""
	}
	if _, err := time.Parse(dateLayout, v); err != nil {
		l.fail(i18n.ConfigDate, key)
		return ""
	}
	return v
}

// Load membaca semua variabel lingkungan lewat getenv, mengisi nilai bawaan, dan memvalidasi
// semuanya. Bila ada masalah, Load mengembalikan *ValidationError yang memuat semua masalah
// sekaligus.
func Load(getenv func(string) string) (Config, error) {
	l := &loader{getenv: getenv}
	var c Config

	c.DatabaseURL = l.postgresURL("DATABASE_URL")
	c.HTTPAddr = l.listenAddr("HTTP_ADDR", defaultHTTPAddr)
	c.AppBaseURL = l.originURL("APP_BASE_URL")
	c.SessionTTL = l.duration("SESSION_TTL", defaultSessionTTL)
	c.ExportDir = l.stringOr("EXPORT_DIR", defaultExportDir)
	c.ExportTemplateDir = l.stringOr("EXPORT_TEMPLATE_DIR", defaultExportTemplateDir)
	c.DBMaxConns, _ = l.intAtLeast("DB_MAX_CONNS", defaultDBMaxConns, 1)
	c.RiverWorkers, _ = l.intAtLeast("RIVER_WORKERS", defaultRiverWorkers, 1)
	var poolOK, parOK bool
	c.RulePoolSize, poolOK = l.intAtLeast("RULE_POOL_SIZE", defaultRulePoolSize, 1)
	c.RuleParallelism, parOK = l.intAtLeast("RULE_PARALLELISM", defaultRuleParallelism, 1)
	// Satu kali cek tidak bisa memakai lebih banyak koneksi daripada isi pool aturan; diperiksa
	// hanya bila kedua nilai sendiri sudah benar supaya tidak ada masalah ganda.
	if poolOK && parOK && c.RuleParallelism > c.RulePoolSize {
		l.fail(i18n.ConfigNotGreater, "RULE_PARALLELISM", "RULE_POOL_SIZE")
	}
	c.LogLevel = l.logLevel("LOG_LEVEL")
	c.MetricsAllow = l.cidrList("METRICS_ALLOW", defaultMetricsAllow)
	c.DevMode = l.boolean("DEV_MODE", false)

	// Format DEV_FAKE_TODAY selalu diperiksa, tetapi nilainya hanya dipakai dalam mode
	// pengembangan supaya server produksi tidak pernah memakai tanggal palsu.
	if fake := l.date("DEV_FAKE_TODAY"); fake != "" {
		if c.DevMode {
			c.DevFakeToday = fake
		} else {
			c.Warnings = append(c.Warnings, fmt.Sprintf(i18n.ConfigFakeTodayIgnored, "DEV_FAKE_TODAY"))
		}
	}

	if len(l.problems) > 0 {
		return Config{}, &ValidationError{Problems: l.problems}
	}
	return c, nil
}

// Database adalah konfigurasi perintah yang hanya butuh database (migrate, seed-demo). Perintah
// ini tidak membaca APP_BASE_URL dan variabel server lain supaya bisa dijalankan sebelum
// konfigurasi server lengkap (docs/03-architecture.md §3.2).
type Database struct {
	// URL adalah URL koneksi PostgreSQL (DATABASE_URL, wajib).
	URL string
	// DevMode mengizinkan perintah yang menghapus data (migrate down, seed-demo).
	DevMode bool
	// LogLevel adalah level log minimum.
	LogLevel slog.Level
}

// LoadDatabase membaca dan memvalidasi DATABASE_URL, DEV_MODE dan LOG_LEVEL dengan aturan yang
// sama seperti Load. Semua masalah dilaporkan sekaligus dalam *ValidationError.
func LoadDatabase(getenv func(string) string) (Database, error) {
	l := &loader{getenv: getenv}
	var c Database
	c.URL = l.postgresURL("DATABASE_URL")
	c.LogLevel = l.logLevel("LOG_LEVEL")
	c.DevMode = l.boolean("DEV_MODE", false)
	if len(l.problems) > 0 {
		return Database{}, &ValidationError{Problems: l.problems}
	}
	return c, nil
}

// LogValue menyiapkan konfigurasi untuk dicatat ke log saat start; kata sandi di DATABASE_URL
// (bagian user:sandi@ maupun parameter query) disamarkan supaya tidak pernah tersimpan di log.
func (c Config) LogValue() slog.Value {
	dbURL := ""
	if u, err := url.Parse(c.DatabaseURL); err == nil {
		q := u.Query()
		changed := false
		// Parameter query berisi rahasia yang juga dibaca pgx.
		for _, k := range []string{"password", "sslpassword"} {
			if q.Has(k) {
				q.Set(k, "xxxxx")
				changed = true
			}
		}
		if changed {
			u.RawQuery = q.Encode()
		}
		dbURL = u.Redacted()
	}
	prefixes := make([]string, len(c.MetricsAllow))
	for i, p := range c.MetricsAllow {
		prefixes[i] = p.String()
	}
	return slog.GroupValue(
		slog.String("database_url", dbURL),
		slog.String("http_addr", c.HTTPAddr),
		slog.String("app_base_url", c.AppBaseURL),
		slog.Duration("session_ttl", c.SessionTTL),
		slog.String("export_dir", c.ExportDir),
		slog.String("export_template_dir", c.ExportTemplateDir),
		slog.Int("db_max_conns", c.DBMaxConns),
		slog.Int("river_workers", c.RiverWorkers),
		slog.Int("rule_pool_size", c.RulePoolSize),
		slog.Int("rule_parallelism", c.RuleParallelism),
		slog.String("log_level", c.LogLevel.String()),
		slog.String("metrics_allow", strings.Join(prefixes, ",")),
		slog.Bool("dev_mode", c.DevMode),
		slog.String("dev_fake_today", c.DevFakeToday),
	)
}
