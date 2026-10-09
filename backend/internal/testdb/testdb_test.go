// Test harness database (test case TC-M01-012 dan TC-M01-013, docs/test-cases/M01-database.md).

package testdb

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

// TestMain menjalankan test paket ini lalu mematikan container PostgreSQL yang dibuat harness.
func TestMain(m *testing.M) { Main(m) }

// TestNew_TC_M01_012 memastikan New memberi database PostgreSQL 18 berisi data demo dengan JIT
// mati, dua test paralel mendapat salinan sendiri, database dihapus setelah test, dan template
// tidak dibangun ulang oleh New berikutnya.
func TestNew_TC_M01_012(t *testing.T) {
	ctx := context.Background()
	admin := adminPool(t)
	name := templateName()
	if !strings.HasPrefix(name, "pfmea_tpl_") {
		t.Errorf("template name = %s, want pfmea_tpl_<hash>", name)
	}
	// templateOID membaca oid database template; berubah bila template dibangun ulang.
	templateOID := func() uint32 {
		t.Helper()
		var oid uint32
		if err := admin.QueryRow(ctx, "SELECT oid FROM pg_database WHERE datname = $1", name).Scan(&oid); err != nil {
			t.Fatalf("template %s not found: %v", name, err)
		}
		return oid
	}

	var firstName string
	var firstOID uint32
	// "isolated" menghapus PS-07 di salinannya sendiri; "untouched" berjalan paralel dan setelah
	// penghapusan di-commit memastikan salinannya masih utuh. Kanal ditutup lewat defer supaya
	// "untouched" tidak menggantung bila "isolated" gagal di tengah jalan.
	deleted := make(chan struct{})
	t.Run("parallel", func(t *testing.T) {
		t.Run("isolated", func(t *testing.T) {
			t.Parallel()
			defer close(deleted)
			db := New(t)
			firstName = db.Name
			firstOID = templateOID()
			if _, err := db.Pool.Exec(ctx, "DELETE FROM packages WHERE code = 'PS-07'"); err != nil {
				t.Fatalf("delete PS-07: %v", err)
			}
		})
		t.Run("untouched", func(t *testing.T) {
			t.Parallel()
			db := New(t)
			<-deleted
			if db.Name == firstName {
				t.Fatalf("both tests got database %s", db.Name)
			}
			var version, jit string
			var users, ps07 int
			err := db.Pool.QueryRow(ctx, `SELECT current_setting('server_version_num'), current_setting('jit'),
				(SELECT count(*) FROM users), (SELECT count(*) FROM packages WHERE code = 'PS-07')`).Scan(&version, &jit, &users, &ps07)
			if err != nil {
				t.Fatalf("query: %v", err)
			}
			if !strings.HasPrefix(version, "18") {
				t.Errorf("server_version_num = %s, want PostgreSQL 18", version)
			}
			if jit != "off" {
				t.Errorf("jit = %s, want off", jit)
			}
			if users != 6 || ps07 != 1 {
				t.Errorf("users = %d, PS-07 rows = %d; want 6 and 1 (copy must be isolated)", users, ps07)
			}
		})
	})

	// Database sub-test pertama sudah dihapus oleh cleanup-nya; template tetap ada, ditandai
	// sebagai template, dan masih database yang sama (tidak dibangun ulang).
	var exists bool
	if err := admin.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)", firstName).Scan(&exists); err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Errorf("database %s still exists after its test ended", firstName)
	}
	New(t)
	if oid := templateOID(); oid != firstOID {
		t.Errorf("template %s was rebuilt (oid %d, then %d)", name, firstOID, oid)
	}
	var isTemplate bool
	if err := admin.QueryRow(ctx, "SELECT datistemplate FROM pg_database WHERE datname = $1", name).Scan(&isTemplate); err != nil {
		t.Fatal(err)
	}
	if !isTemplate {
		t.Errorf("database %s is not marked as a template", name)
	}
}

// TestNoDocker_TC_M01_012 menjalankan test ini lagi di proses anak tanpa TEST_DATABASE_URL dan
// dengan start container yang gagal seperti saat Docker tidak berjalan: New harus menggagalkan
// test dengan pesan yang menyebut Docker dan TEST_DATABASE_URL, bukan melewatinya atau
// menggantung. Proses anak dipakai karena server test dibuat sekali per proses.
func TestNoDocker_TC_M01_012(t *testing.T) {
	if os.Getenv(childEnv) == "1" {
		runContainer = func(context.Context) (*tcpostgres.PostgresContainer, error) {
			return nil, errors.New("Cannot connect to the Docker daemon at unix:///var/run/docker.sock")
		}
		New(t)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestNoDocker_TC_M01_012$", "-test.count=1", "-test.v")
	cmd.Env = []string{childEnv + "=1", "HOME=" + t.TempDir(), "PATH=" + os.Getenv("PATH")}
	out, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("child test without Docker did not finish in time:\n%s", out)
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
		t.Fatalf("child test without Docker: err = %v, want exit code 1\n%s", err, out)
	}
	text := string(out)
	for _, want := range []string{"--- FAIL: TestNoDocker_TC_M01_012", "is Docker running?", "TEST_DATABASE_URL"} {
		if !strings.Contains(text, want) {
			t.Errorf("child output lacks %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "--- SKIP") {
		t.Errorf("child test was skipped instead of failing:\n%s", text)
	}
}

// childEnv menandai proses anak TestNoDocker_TC_M01_012.
const childEnv = "PFMEA_TESTDB_NO_DOCKER_CHILD"

// TestNewEmpty_TC_M01_013 memastikan NewEmpty memberi database baru tanpa relasi apa pun, untuk
// test migrasi dan perintah CLI.
func TestNewEmpty_TC_M01_013(t *testing.T) {
	db := NewEmpty(t)
	var relations int
	err := db.Pool.QueryRow(context.Background(),
		"SELECT count(*) FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace WHERE n.nspname = 'public'").Scan(&relations)
	if err != nil {
		t.Fatal(err)
	}
	if relations != 0 {
		t.Errorf("empty database has %d relations, want 0", relations)
	}
	if !strings.Contains(db.URL, db.Name) {
		t.Errorf("URL %q does not point to database %s", db.URL, db.Name)
	}
}

// TestUnreachableServer_TC_M01_012 memastikan TEST_DATABASE_URL yang tidak terjangkau menjadi
// error yang jelas (test gagal), bukan test yang dilewati, dan kata sandinya tidak pernah ikut
// tercetak, di bagian user:sandi@, di parameter query, maupun dalam bentuk key=value.
func TestUnreachableServer_TC_M01_012(t *testing.T) {
	urls := map[string]string{
		"userinfo":      "postgres://pfmea:secret@127.0.0.1:1/pfmea?connect_timeout=2",
		"query":         "postgres://pfmea@127.0.0.1:1/pfmea?connect_timeout=2&password=secret",
		"sslpassword":   "postgres://pfmea@127.0.0.1:1/pfmea?connect_timeout=2&sslpassword=secret",
		"keyword/value": "host=127.0.0.1 port=1 user=pfmea password=secret dbname=pfmea connect_timeout=2",
	}
	for name, raw := range urls {
		t.Run(name, func(t *testing.T) {
			getenv := func(k string) string {
				if k == "TEST_DATABASE_URL" {
					return raw
				}
				return ""
			}
			_, err := startServer(context.Background(), getenv)
			if err == nil {
				t.Fatal("startServer with an unreachable TEST_DATABASE_URL should fail")
			}
			if !strings.Contains(err.Error(), "TEST_DATABASE_URL") {
				t.Errorf("error should name TEST_DATABASE_URL: %v", err)
			}
			if strings.Contains(err.Error(), "secret") {
				t.Errorf("error leaks the password: %v", err)
			}
		})
	}
}
