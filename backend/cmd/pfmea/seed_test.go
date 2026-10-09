// Test perintah `pfmea seed-demo` (test case TC-M01-006 sampai TC-M01-008,
// docs/test-cases/M01-database.md).

package main

import (
	"strings"
	"testing"

	"pfmea/backend/internal/testdb"
)

// migratedEmpty membuat database kosong lalu menjalankan `pfmea migrate up` di atasnya.
func migratedEmpty(t *testing.T) *testdb.DB {
	t.Helper()
	d := testdb.NewEmpty(t)
	if code, out, errOut := runCmd(t, map[string]string{"DATABASE_URL": d.URL}, "migrate", "up"); code != 0 {
		t.Fatalf("migrate up: code %d\nstdout: %s\nstderr: %s", code, out, errOut)
	}
	return d
}

// TestSeedDemoRefuses_TC_M01_006 memastikan seed-demo ditolak bila DEV_MODE kosong atau false,
// dan database tetap tanpa data.
func TestSeedDemoRefuses_TC_M01_006(t *testing.T) {
	d := migratedEmpty(t)
	for _, devMode := range []string{"", "false"} {
		t.Run("DEV_MODE="+devMode, func(t *testing.T) {
			code, _, errOut := runCmd(t, map[string]string{"DATABASE_URL": d.URL, "DEV_MODE": devMode}, "seed-demo")
			if code != 1 || !strings.Contains(errOut, "seed-demo requires DEV_MODE=true") {
				t.Errorf("code %d, stderr %q; want code 1 and the DEV_MODE message", code, errOut)
			}
		})
	}
	if n := count(t, d.Pool, "SELECT count(*) FROM users"); n != 0 {
		t.Errorf("users = %d after refused seed-demo, want 0", n)
	}
}

// TestSeedDemo_TC_M01_007 memastikan seed-demo menolak database yang belum dimigrasi, memuat
// data demo sekali (audit bersumber seed), dan menolak dijalankan lagi tanpa --reset.
func TestSeedDemo_TC_M01_007(t *testing.T) {
	empty := testdb.NewEmpty(t)
	code, _, errOut := runCmd(t, map[string]string{"DATABASE_URL": empty.URL, "DEV_MODE": "true"}, "seed-demo")
	if code != 1 || !strings.Contains(errOut, "run pfmea migrate up first") {
		t.Errorf("seed-demo before migrate: code %d, stderr %q", code, errOut)
	}
	// Pemeriksaan versi hanya membaca: tidak membuat tabel goose di database kosong.
	if n := count(t, empty.Pool, "SELECT count(*) FROM pg_tables WHERE schemaname = 'public'"); n != 0 {
		t.Errorf("seed-demo created %d tables in an unmigrated database", n)
	}

	d := migratedEmpty(t)
	env := map[string]string{"DATABASE_URL": d.URL, "DEV_MODE": "true"}
	code, out, errOut := runCmd(t, env, "seed-demo")
	if code != 0 {
		t.Fatalf("seed-demo: code %d\nstdout: %s\nstderr: %s", code, out, errOut)
	}
	if !strings.Contains(out, "6 users, 2 packages") {
		t.Errorf("stdout %q lacks the loaded counts", out)
	}
	if n := count(t, d.Pool, "SELECT count(*) FROM users"); n != 6 {
		t.Errorf("users = %d, want 6", n)
	}
	if n := count(t, d.Pool, "SELECT count(*) FROM packages WHERE code IN ('GENERAL', 'PS-07')"); n != 2 {
		t.Errorf("packages GENERAL and PS-07: %d, want 2", n)
	}
	if n := count(t, d.Pool, "SELECT count(*) FROM audit_log WHERE table_name = 'users' AND source = 'seed'"); n != 6 {
		t.Errorf("audit rows for seeded users with source seed = %d, want 6", n)
	}

	code, _, errOut = runCmd(t, env, "seed-demo")
	if code != 1 || !strings.Contains(errOut, "database already contains data; use --reset") {
		t.Errorf("second seed-demo: code %d, stderr %q", code, errOut)
	}
	if n := count(t, d.Pool, "SELECT count(*) FROM users"); n != 6 {
		t.Errorf("users = %d after refused second seed-demo, want 6", n)
	}
}

// TestSeedDemoReset_TC_M01_008 memastikan seed-demo --reset menghapus perubahan lokal dan
// membangun ulang skema serta data demo dari nol.
func TestSeedDemoReset_TC_M01_008(t *testing.T) {
	d := testdb.New(t)
	steps := count(t, d.Pool, "SELECT count(*) FROM process_steps")
	if _, err := d.Pool.Exec(t.Context(), "UPDATE packages SET name = 'Changed' WHERE code = 'PS-07'"); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Pool.Exec(t.Context(), `INSERT INTO process_steps (package_id, op_no, seq, name)
		SELECT id, '999', 999, 'Extra step' FROM packages WHERE code = 'PS-07'`); err != nil {
		t.Fatal(err)
	}

	code, out, errOut := runCmd(t, map[string]string{"DATABASE_URL": d.URL, "DEV_MODE": "true"}, "seed-demo", "--reset")
	if code != 0 {
		t.Fatalf("seed-demo --reset: code %d\nstdout: %s\nstderr: %s", code, out, errOut)
	}
	if !strings.Contains(out, "restart") {
		t.Errorf("stdout %q should remind to restart a running server", out)
	}

	// Koneksi lama di pool menyimpan statement untuk objek yang sudah dihapus; mulai baru.
	d.Pool.Reset()
	if n := count(t, d.Pool, "SELECT count(*) FROM packages WHERE code = 'PS-07' AND name <> 'Changed'"); n != 1 {
		t.Errorf("PS-07 name not restored")
	}
	if n := count(t, d.Pool, "SELECT count(*) FROM process_steps"); n != steps {
		t.Errorf("process_steps = %d after reset, want %d", n, steps)
	}
	if n := count(t, d.Pool, "SELECT max(version_id)::int FROM goose_db_version"); n != 1 {
		t.Errorf("migration version = %d after reset, want 1", n)
	}

	for _, args := range [][]string{{"seed-demo", "--force"}, {"seed-demo", "--reset", "extra"}} {
		code, _, errOut := runCmd(t, map[string]string{"DATABASE_URL": d.URL, "DEV_MODE": "true"}, args...)
		if code != 2 || !strings.Contains(errOut, "seed-demo [--reset]") {
			t.Errorf("%v: code %d, stderr %q; want usage and code 2", args, code, errOut)
		}
	}
}
