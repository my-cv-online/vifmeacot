// Fungsi bantu dan konstanta bersama untuk test paket store (id deterministik dari
// db/seed/demo.sql).

package store_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"pfmea/backend/internal/store"
	"pfmea/backend/internal/testdb"
)

// Id baris data demo yang dipakai test.
var (
	// ps07 adalah paket model PS-07.
	ps07 = uuid.MustParse("3c91e2ed-8941-5f73-a84f-8be64945878f")
	// apratama adalah pengguna demo berperan author.
	apratama = uuid.MustParse("6b134e0d-7045-544f-8bbc-67886fbad833")
)

// TestMain menjalankan test paket lalu mematikan container database test.
func TestMain(m *testing.M) { testdb.Main(m) }

// newStore membuat database demo baru dan Store di atasnya.
func newStore(t *testing.T) (*testdb.DB, *store.Store) {
	t.Helper()
	db := testdb.New(t)
	return db, store.New(db.Pool)
}

// stepID mencari id step berdasarkan paket dan op no.
func stepID(t *testing.T, pool *pgxpool.Pool, pkg uuid.UUID, opNo string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	if err := pool.QueryRow(context.Background(),
		"SELECT id FROM process_steps WHERE package_id = $1 AND op_no = $2", pkg, opNo).Scan(&id); err != nil {
		t.Fatalf("step %s: %v", opNo, err)
	}
	return id
}

// scalar menjalankan query satu nilai dan menggagalkan test bila error.
func scalar[T any](t *testing.T, pool *pgxpool.Pool, sql string, args ...any) T {
	t.Helper()
	var v T
	if err := pool.QueryRow(context.Background(), sql, args...).Scan(&v); err != nil {
		t.Fatalf("query %q: %v", sql, err)
	}
	return v
}
