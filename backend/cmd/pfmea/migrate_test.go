// Test perintah `pfmea migrate` (test case TC-M01-002 sampai TC-M01-004,
// docs/test-cases/M01-database.md).

package main

import (
	"context"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"pfmea/backend/internal/testdb"
	"pfmea/db"
)

// TestMain menjalankan test paket lalu mematikan container database test.
func TestMain(m *testing.M) { testdb.Main(m) }

// runCmd menjalankan binary dengan environment tiruan dan mengembalikan kode keluar serta output.
func runCmd(t *testing.T, env map[string]string, args ...string) (int, string, string) {
	t.Helper()
	var stdout, stderr syncBuffer
	code := run(context.Background(), args, envOf(env), &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

// count menjalankan query hitung di pool test.
func count(t *testing.T, pool *pgxpool.Pool, sql string) int {
	t.Helper()
	var n int
	if err := pool.QueryRow(context.Background(), sql).Scan(&n); err != nil {
		t.Fatalf("query %q: %v", sql, err)
	}
	return n
}

// TestMigrateUp_TC_M01_002 memastikan `migrate up` membuat skema di database kosong hanya dengan
// DATABASE_URL, menjalankan ulang tanpa efek, dan `migrate status` melaporkan versi 1.
func TestMigrateUp_TC_M01_002(t *testing.T) {
	empty := testdb.NewEmpty(t)
	env := map[string]string{"DATABASE_URL": empty.URL}

	code, out, errOut := runCmd(t, env, "migrate", "up")
	if code != 0 || !strings.Contains(out, "applied 00001_init.sql") {
		t.Fatalf("migrate up: code %d\nstdout: %s\nstderr: %s", code, out, errOut)
	}
	if n := count(t, empty.Pool, "SELECT count(*) FROM pg_tables WHERE schemaname = 'public' AND tablename = 'packages'"); n != 1 {
		t.Errorf("table packages missing after migrate up")
	}

	code, out, _ = runCmd(t, env, "migrate", "up")
	if code != 0 || !strings.Contains(out, "no pending migrations") {
		t.Errorf("second migrate up: code %d, stdout %q", code, out)
	}

	code, out, _ = runCmd(t, env, "migrate", "status")
	if code != 0 || !strings.Contains(out, "00001_init.sql") || !strings.Contains(out, "applied") || !strings.Contains(out, "version 1") {
		t.Errorf("migrate status: code %d, stdout %q", code, out)
	}
}

// TestMigrateDown_TC_M01_003 memastikan `migrate down` ditolak tanpa DEV_MODE=true, dan dengan
// DEV_MODE menghapus semua objek aplikasi sehingga hanya tabel goose dan extension pg_trgm yang
// tersisa; setelah itu migrate up dan seed bisa dijalankan lagi.
func TestMigrateDown_TC_M01_003(t *testing.T) {
	ctx := context.Background()
	d := testdb.New(t) // sudah dimigrasi dan berisi seed

	code, _, errOut := runCmd(t, map[string]string{"DATABASE_URL": d.URL}, "migrate", "down")
	if code != 1 || !strings.Contains(errOut, "requires DEV_MODE=true") {
		t.Errorf("migrate down without DEV_MODE: code %d, stderr %q", code, errOut)
	}
	if n := count(t, d.Pool, "SELECT count(*) FROM packages"); n != 2 {
		t.Fatalf("schema changed although migrate down was refused")
	}

	env := map[string]string{"DATABASE_URL": d.URL, "DEV_MODE": "true"}
	code, out, errOut := runCmd(t, env, "migrate", "down")
	if code != 0 || !strings.Contains(out, "rolled back 00001_init.sql") {
		t.Fatalf("migrate down: code %d\nstdout: %s\nstderr: %s", code, out, errOut)
	}

	// Objek di skema public yang bukan milik extension.
	const notExtension = `NOT EXISTS (SELECT 1 FROM pg_depend d WHERE d.objid = %s AND d.deptype = 'e')`
	tables := count(t, d.Pool, `SELECT count(*) FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = 'public' AND c.relkind IN ('r', 'p', 'v', 'm', 'f') AND c.relname <> 'goose_db_version'
		AND `+strings.ReplaceAll(notExtension, "%s", "c.oid"))
	funcs := count(t, d.Pool, `SELECT count(*) FROM pg_proc p JOIN pg_namespace n ON n.oid = p.pronamespace
		WHERE n.nspname = 'public' AND `+strings.ReplaceAll(notExtension, "%s", "p.oid"))
	types := count(t, d.Pool, `SELECT count(*) FROM pg_type t JOIN pg_namespace n ON n.oid = t.typnamespace
		WHERE n.nspname = 'public' AND t.typtype IN ('e', 'd')
		AND `+strings.ReplaceAll(notExtension, "%s", "t.oid"))
	if tables != 0 || funcs != 0 || types != 0 {
		t.Errorf("after migrate down: %d tables, %d functions, %d enum/domain types left; want none", tables, funcs, types)
	}
	if n := count(t, d.Pool, "SELECT count(*) FROM pg_tables WHERE schemaname = 'public' AND tablename = 'goose_db_version'"); n != 1 {
		t.Errorf("goose_db_version should remain")
	}
	if n := count(t, d.Pool, "SELECT count(*) FROM pg_extension WHERE extname = 'pg_trgm'"); n != 1 {
		t.Errorf("extension pg_trgm should remain")
	}

	// Siklus naik lagi: migrate up dan seed demo berjalan di atas database yang sudah turun.
	if code, out, errOut := runCmd(t, env, "migrate", "up"); code != 0 {
		t.Fatalf("migrate up after down: code %d\nstdout: %s\nstderr: %s", code, out, errOut)
	}
	conn, err := pgx.Connect(ctx, d.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close(ctx) }()
	if err := db.LoadDemoSeed(ctx, conn); err != nil {
		t.Errorf("seed after down/up: %v", err)
	}
}

// TestMigrateArgs_TC_M01_004 memastikan argumen yang salah menampilkan usage, `migrate` hanya
// butuh DATABASE_URL (tanpa APP_BASE_URL), dan error koneksi tidak membocorkan kata sandi.
func TestMigrateArgs_TC_M01_004(t *testing.T) {
	for _, args := range [][]string{{"migrate"}, {"migrate", "sideways"}} {
		code, _, errOut := runCmd(t, nil, args...)
		if code != 2 || !strings.Contains(errOut, "migrate up|down|status") {
			t.Errorf("%v: code %d, stderr %q; want usage and code 2", args, code, errOut)
		}
	}

	code, _, errOut := runCmd(t, nil, "migrate", "up")
	if code != 1 || !strings.Contains(errOut, "DATABASE_URL is required") || strings.Contains(errOut, "APP_BASE_URL") {
		t.Errorf("migrate up without env: code %d, stderr %q", code, errOut)
	}

	code, _, errOut = runCmd(t, map[string]string{"DATABASE_URL": "postgres://pfmea:secret@127.0.0.1:1/pfmea?connect_timeout=2"}, "migrate", "up")
	if code != 1 || errOut == "" || strings.Contains(errOut, "secret") {
		t.Errorf("unreachable database: code %d, stderr %q; want code 1 without the password", code, errOut)
	}
}
