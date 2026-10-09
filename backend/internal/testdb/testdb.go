// Package testdb menyediakan database PostgreSQL 18 untuk test (docs/11-testing.md §2): satu
// server per proses test (container testcontainers, atau TEST_DATABASE_URL), satu template yang
// sudah dimigrasi dan berisi data demo, lalu setiap test mendapat salinan sendiri lewat
// CREATE DATABASE … TEMPLATE (puluhan milidetik) yang dihapus setelah test selesai.
package testdb

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"pfmea/db"
)

// Pengaturan harness.
const (
	// harnessVersion ikut di-hash ke nama template; naikkan bila cara membangun template berubah
	// supaya server bersama (TEST_DATABASE_URL) tidak memakai template lama.
	harnessVersion = "1"
	// pgImage adalah image PostgreSQL untuk test, sama dengan produksi.
	pgImage = "postgres:18"
	// plantTimeZone adalah zona waktu sesi test: data demo dan fixture mengasumsikan tanggal di
	// zona waktu pabrik (docs/11-testing.md §2).
	plantTimeZone = "Asia/Jakarta"
	// poolMaxConns membatasi koneksi per database test supaya banyak test paralel tetap muat.
	poolMaxConns = 4
	// connectTimeout membatasi waktu menunggu server saat start.
	connectTimeout = 30 * time.Second
)

// DB adalah satu database test beserta pool koneksinya.
type DB struct {
	// Name adalah nama database (pfmea_t_<acak>).
	Name string
	// URL adalah connection string ke database ini, untuk perintah CLI yang diuji.
	URL string
	// Pool adalah pool pgx ke database ini (jit off, zona waktu pabrik).
	Pool *pgxpool.Pool
}

// server adalah server PostgreSQL yang dipakai satu proses test.
type server struct {
	// adminURL menunjuk database administratif server (untuk CREATE/DROP DATABASE).
	adminURL string
	// admin adalah pool ke database administratif.
	admin *pgxpool.Pool
	// container terisi bila server dijalankan oleh harness (bukan TEST_DATABASE_URL).
	container *tcpostgres.PostgresContainer
}

// State per proses test. Ini satu-satunya state global harness: server dan template sengaja
// dibuat sekali per proses (sync.Once) karena mahal, dan dibersihkan oleh Main.
var (
	srvOnce sync.Once
	srv     *server
	srvErr  error
	tplOnce sync.Once
	tplErr  error
)

// Main menjalankan test paket lalu mematikan container yang dibuat harness. Dipanggil dari
// TestMain setiap paket yang memakai testdb: func TestMain(m *testing.M) { testdb.Main(m) }.
func Main(m *testing.M) {
	code := m.Run()
	if srv != nil {
		srv.admin.Close()
		if srv.container != nil {
			_ = testcontainers.TerminateContainer(srv.container)
		}
	}
	os.Exit(code)
}

// New membuat database baru berisi skema dan data demo untuk satu test. Database dan pool-nya
// ditutup serta dihapus otomatis saat test selesai. Tanpa Docker dan tanpa TEST_DATABASE_URL
// test gagal (bukan dilewati) supaya `make check` tidak hijau tanpa test database.
func New(t testing.TB) *DB {
	t.Helper()
	s := getServer(t)
	tplOnce.Do(func() { tplErr = ensureTemplate(context.Background(), s) })
	if tplErr != nil {
		t.Fatalf("testdb: build template database: %v", tplErr)
	}
	return create(t, s, "TEMPLATE "+pgx.Identifier{templateName()}.Sanitize())
}

// NewEmpty membuat database baru yang benar-benar kosong (tanpa migrasi), untuk menguji
// `pfmea migrate` dan perintah lain dari nol.
func NewEmpty(t testing.TB) *DB {
	t.Helper()
	return create(t, getServer(t), "TEMPLATE template0")
}

// getServer mengembalikan server proses ini dan menggagalkan test bila server tidak tersedia.
func getServer(t testing.TB) *server {
	t.Helper()
	srvOnce.Do(func() { srv, srvErr = startServer(context.Background(), os.Getenv) })
	if srvErr != nil {
		t.Fatalf("testdb: %v", srvErr)
	}
	return srv
}

// adminPool mengembalikan pool administratif; dipakai test harness sendiri.
func adminPool(t testing.TB) *pgxpool.Pool {
	t.Helper()
	return getServer(t).admin
}

// startServer menyiapkan server: TEST_DATABASE_URL bila di-set, selain itu container postgres:18
// dengan pengaturan cepat untuk test (fsync dan synchronous_commit mati; data dibuang).
func startServer(ctx context.Context, getenv func(string) string) (*server, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	if raw := getenv("TEST_DATABASE_URL"); raw != "" {
		admin, err := connect(ctx, raw)
		if err != nil {
			return nil, fmt.Errorf("TEST_DATABASE_URL is set but the server cannot be reached (%s): %w", redact(raw), err)
		}
		return &server{adminURL: raw, admin: admin}, nil
	}

	// Image bawaan modul postgres sudah menjalankan "postgres -c fsync=off"; argumen berikut
	// ditambahkan di belakangnya.
	ctr, err := tcpostgres.Run(ctx, pgImage,
		tcpostgres.WithDatabase("postgres"),
		tcpostgres.WithUsername("pfmea"),
		tcpostgres.WithPassword("pfmea-test"),
		tcpostgres.BasicWaitStrategies(),
		testcontainers.WithCmdArgs(
			"-c", "jit=off",
			"-c", "synchronous_commit=off",
			"-c", "full_page_writes=off",
			"-c", "max_connections=300",
		),
	)
	if err != nil {
		return nil, fmt.Errorf("start %s test container (is Docker running? otherwise set TEST_DATABASE_URL): %w", pgImage, err)
	}
	raw, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = testcontainers.TerminateContainer(ctr)
		return nil, fmt.Errorf("test container connection string: %w", err)
	}
	admin, err := connect(ctx, raw)
	if err != nil {
		_ = testcontainers.TerminateContainer(ctr)
		return nil, fmt.Errorf("connect to the test container: %w", err)
	}
	return &server{adminURL: raw, admin: admin, container: ctr}, nil
}

// connect membuka pool administratif kecil dan memastikan server menjawab.
func connect(ctx context.Context, raw string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(raw)
	if err != nil {
		return nil, errors.New("invalid connection URL")
	}
	cfg.MaxConns = 4
	cfg.ConnConfig.ConnectTimeout = 10 * time.Second
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	pingCtx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

// create membuat database baru dari klausa template yang diberikan lalu membuka pool ke sana.
// Cleanup penghapusan didaftarkan sebelum cleanup pool, karena cleanup berjalan terbalik: pool
// ditutup dulu, baru database dihapus.
func create(t testing.TB, s *server, templateClause string) *DB {
	t.Helper()
	ctx := context.Background()
	name := "pfmea_t_" + randomHex(6)
	ident := pgx.Identifier{name}.Sanitize()
	if _, err := s.admin.Exec(ctx, "CREATE DATABASE "+ident+" "+templateClause); err != nil {
		t.Fatalf("testdb: create database %s: %v", name, err)
	}
	t.Cleanup(func() {
		// t.Context() sudah dibatalkan saat cleanup, jadi pakai konteks baru.
		if _, err := s.admin.Exec(context.Background(), "DROP DATABASE IF EXISTS "+ident+" WITH (FORCE)"); err != nil {
			t.Errorf("testdb: drop database %s: %v", name, err)
		}
	})

	dbURL, err := withDatabase(s.adminURL, name)
	if err != nil {
		t.Fatalf("testdb: %v", err)
	}
	cfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		t.Fatalf("testdb: parse database URL: %v", err)
	}
	cfg.MaxConns = poolMaxConns
	cfg.ConnConfig.RuntimeParams["jit"] = "off"
	cfg.ConnConfig.RuntimeParams["timezone"] = plantTimeZone
	cfg.ConnConfig.RuntimeParams["application_name"] = "pfmea-test"
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("testdb: open pool for %s: %v", name, err)
	}
	t.Cleanup(pool.Close)
	return &DB{Name: name, URL: dbURL, Pool: pool}
}

// ensureTemplate membangun template berisi migrasi dan data demo bila belum ada. Advisory lock
// di database administratif mencegah beberapa proses test (go test menjalankan paket secara
// paralel) membangun template yang sama bersamaan di server bersama.
func ensureTemplate(ctx context.Context, s *server) error {
	name := templateName()
	conn, err := s.admin.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	key := lockKey(name)
	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock($1)", key); err != nil {
		return fmt.Errorf("lock template build: %w", err)
	}
	defer func() { _, _ = conn.Exec(context.Background(), "SELECT pg_advisory_unlock($1)", key) }()

	var exists bool
	if err := conn.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)", name).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return nil
	}

	// Template dibangun dengan nama sementara lalu diganti namanya, supaya proses lain tidak
	// pernah melihat template yang setengah jadi.
	build := name + "_build"
	buildIdent := pgx.Identifier{build}.Sanitize()
	if _, err := conn.Exec(ctx, "DROP DATABASE IF EXISTS "+buildIdent+" WITH (FORCE)"); err != nil {
		return err
	}
	if _, err := conn.Exec(ctx, "CREATE DATABASE "+buildIdent+" TEMPLATE template0"); err != nil {
		return err
	}
	buildURL, err := withDatabase(s.adminURL, build)
	if err != nil {
		return err
	}
	if err := migrateAndSeed(ctx, buildURL); err != nil {
		return err
	}
	ident := pgx.Identifier{name}.Sanitize()
	if _, err := conn.Exec(ctx, "ALTER DATABASE "+buildIdent+" RENAME TO "+ident); err != nil {
		return err
	}
	// Template tidak boleh menerima koneksi: CREATE DATABASE … TEMPLATE gagal bila ada sesi lain
	// yang terhubung ke template.
	_, err = conn.Exec(ctx, "ALTER DATABASE "+ident+" WITH IS_TEMPLATE true ALLOW_CONNECTIONS false")
	return err
}

// migrateAndSeed menjalankan semua migrasi, data demo dan ANALYZE (statistik ikut tersalin ke
// setiap database test) di database dbURL, lalu menutup semua koneksinya.
func migrateAndSeed(ctx context.Context, dbURL string) error {
	connCfg, err := pgx.ParseConfig(dbURL)
	if err != nil {
		return err
	}
	migrator, err := db.NewMigrator(stdlib.OpenDB(*connCfg))
	if err != nil {
		return err
	}
	if _, err := migrator.Up(ctx); err != nil {
		_ = migrator.Close()
		return fmt.Errorf("migrate template: %w", err)
	}
	if err := migrator.Close(); err != nil {
		return err
	}

	conn, err := pgx.ConnectConfig(ctx, connCfg)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close(context.Background()) }()
	if err := db.LoadDemoSeed(ctx, conn); err != nil {
		return err
	}
	_, err = conn.Exec(ctx, "ANALYZE")
	return err
}

// templateName adalah nama template untuk isi migrasi dan seed saat ini:
// pfmea_tpl_<12 karakter hash>. Isi yang berubah menghasilkan template baru.
func templateName() string {
	h := sha256.New()
	h.Write([]byte("harness:" + harnessVersion + "\n"))
	_ = fs.WalkDir(db.Files, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := fs.ReadFile(db.Files, p)
		if err != nil {
			return err
		}
		h.Write([]byte(p + "\n"))
		h.Write(b)
		return nil
	})
	return "pfmea_tpl_" + hex.EncodeToString(h.Sum(nil))[:12]
}

// lockKey menurunkan kunci advisory lock 64-bit dari nama template.
func lockKey(name string) int64 {
	sum := sha256.Sum256([]byte(name))
	return int64(binary.BigEndian.Uint64(sum[:8]))
}

// withDatabase mengganti nama database di connection string.
func withDatabase(raw, name string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", errors.New("invalid connection URL")
	}
	u.Path = "/" + name
	return u.String(), nil
}

// redact menyamarkan kata sandi di connection string untuk pesan error.
func redact(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return "invalid URL"
	}
	return u.Redacted()
}

// randomHex membuat akhiran acak untuk nama database test.
func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
