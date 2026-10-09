// Test harness database (test case TC-M01-012 dan TC-M01-013, docs/test-cases/M01-database.md).

package testdb

import (
	"context"
	"strings"
	"testing"
)

// TestMain menjalankan test paket ini lalu mematikan container PostgreSQL yang dibuat harness.
func TestMain(m *testing.M) { Main(m) }

// TestNew_TC_M01_012 memastikan New memberi database PostgreSQL 18 berisi data demo dengan JIT
// mati, setiap test mendapat salinan sendiri, database dihapus setelah test, dan template
// hanya dibuat sekali per isi migrasi+seed.
func TestNew_TC_M01_012(t *testing.T) {
	ctx := context.Background()
	var firstName string

	// "isolated" menghapus PS-07 di salinannya sendiri; "untouched" lalu memastikan salinannya
	// masih utuh. Sub-test kedua menunggu sinyal supaya penghapusan pasti sudah di-commit.
	deleted := make(chan struct{})
	t.Run("isolated", func(t *testing.T) {
		db := New(t)
		firstName = db.Name
		if _, err := db.Pool.Exec(ctx, "DELETE FROM packages WHERE code = 'PS-07'"); err != nil {
			t.Fatalf("delete PS-07: %v", err)
		}
		close(deleted)
	})
	t.Run("untouched", func(t *testing.T) {
		<-deleted
		db := New(t)
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

	// Database sub-test pertama sudah dihapus oleh cleanup-nya; template tetap ada dan ditandai
	// sebagai template.
	admin := adminPool(t)
	var exists bool
	if err := admin.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)", firstName).Scan(&exists); err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Errorf("database %s still exists after its test ended", firstName)
	}
	name := templateName()
	if !strings.HasPrefix(name, "pfmea_tpl_") {
		t.Errorf("template name = %s, want pfmea_tpl_<hash>", name)
	}
	var isTemplate bool
	if err := admin.QueryRow(ctx, "SELECT datistemplate FROM pg_database WHERE datname = $1", name).Scan(&isTemplate); err != nil {
		t.Fatalf("template %s not found: %v", name, err)
	}
	if !isTemplate {
		t.Errorf("database %s is not marked as a template", name)
	}
	var builds int
	if err := admin.QueryRow(ctx, "SELECT count(*) FROM pg_database WHERE datname LIKE 'pfmea_tpl_%'").Scan(&builds); err != nil {
		t.Fatal(err)
	}
	if builds < 1 {
		t.Errorf("no template database found")
	}
}

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
// error yang jelas (test gagal), bukan test yang dilewati.
func TestUnreachableServer_TC_M01_012(t *testing.T) {
	getenv := func(k string) string {
		if k == "TEST_DATABASE_URL" {
			return "postgres://pfmea:secret@127.0.0.1:1/pfmea?connect_timeout=2"
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
}
