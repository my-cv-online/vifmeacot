// Pemeriksaan versi skema tanpa goose: dipakai /readyz dan `pfmea seed-demo`. goose sendiri
// membuat tabel goose_db_version saat ditanya versinya, padahal pemeriksaan tidak boleh
// mengubah database.

package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// undefinedTable adalah SQLSTATE PostgreSQL untuk tabel yang tidak ada (42P01).
const undefinedTable = "42P01"

// RowQuerier adalah bagian pgx yang dibutuhkan pemeriksaan versi; dipenuhi *pgx.Conn,
// *pgxpool.Pool dan pgx.Tx.
type RowQuerier interface {
	// QueryRow menjalankan query yang menghasilkan satu baris.
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// SchemaError berarti versi skema database tidak sama dengan migrasi terbaru di binary.
type SchemaError struct {
	// Current adalah versi di database (0 bila belum pernah dimigrasi).
	Current int64
	// Latest adalah versi migrasi terbaru yang tertanam di binary.
	Latest int64
}

// Error menjelaskan selisih versi (bahasa Inggris, untuk log dan output CLI).
func (e *SchemaError) Error() string {
	return fmt.Sprintf("database schema is at version %d, this binary expects version %d", e.Current, e.Latest)
}

// SchemaVersion membaca versi migrasi tertinggi di goose_db_version dengan query yang sama
// seperti goose. Database yang belum pernah dimigrasi (tabel tidak ada) berversi 0.
func SchemaVersion(ctx context.Context, q RowQuerier) (int64, error) {
	var v *int64
	err := q.QueryRow(ctx, "SELECT max(version_id) FROM goose_db_version").Scan(&v)
	var pgErr *pgconn.PgError
	switch {
	case errors.As(err, &pgErr) && pgErr.Code == undefinedTable:
		return 0, nil
	case err != nil:
		return 0, fmt.Errorf("read schema version: %w", err)
	case v == nil:
		return 0, nil
	default:
		return *v, nil
	}
}

// CheckSchema memastikan database sudah dimigrasi tepat sampai migrasi terbaru di binary.
// Selisih versi dilaporkan sebagai *SchemaError; error lain berarti database tidak bisa dibaca.
func CheckSchema(ctx context.Context, q RowQuerier) error {
	latest, err := LatestVersion()
	if err != nil {
		return err
	}
	current, err := SchemaVersion(ctx, q)
	if err != nil {
		return err
	}
	if current != latest {
		return &SchemaError{Current: current, Latest: latest}
	}
	return nil
}
