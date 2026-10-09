// Pemetaan error PostgreSQL ke jenis error store. M2 memetakan jenis ini ke Problem Details
// (docs/05-api.md §2): ErrDuplicate dan ErrCheckViolation → validation_failed, ErrForeignKey →
// delete_blocked atau invalid_reference (tergantung operasinya), ErrConflict → version_conflict.

package store

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

// Jenis error database. Dicek dengan errors.Is(err, store.ErrDuplicate).
var (
	// ErrDuplicate: pelanggaran unique (23505), termasuk yang terdeteksi saat COMMIT.
	ErrDuplicate = errors.New("duplicate value")
	// ErrCheckViolation: pelanggaran check constraint (23514).
	ErrCheckViolation = errors.New("check constraint violated")
	// ErrForeignKey: pelanggaran foreign key: referensi tidak ada (23503), atau baris masih
	// dipakai baris lain saat dihapus (23001 restrict_violation untuk ON DELETE RESTRICT).
	ErrForeignKey = errors.New("foreign key violated")
	// ErrNotNull: kolom wajib bernilai NULL (23502).
	ErrNotNull = errors.New("required value missing")
	// ErrInvalidValue: nilai tidak valid untuk tipenya (kelas SQLSTATE 22, mis. enum salah).
	ErrInvalidValue = errors.New("invalid value")
	// ErrConflict: transaksi tetap gagal karena serialization failure (40001) atau deadlock
	// (40P01) setelah semua retry.
	ErrConflict = errors.New("concurrent update conflict")
	// ErrNotFound: baris yang dicari tidak ada.
	ErrNotFound = errors.New("not found")
)

// DBError adalah error database yang sudah dipetakan: jenisnya (Kind) beserta detail dari
// PostgreSQL untuk membangun pesan validasi. Error asli tetap bisa diambil dengan errors.As.
type DBError struct {
	// Kind adalah salah satu Err* di atas.
	Kind error
	// Code adalah SQLSTATE PostgreSQL.
	Code string
	// Table, Column dan Constraint menunjuk objek yang dilanggar (bisa kosong).
	Table, Column, Constraint string
	// Detail adalah keterangan tambahan dari PostgreSQL (tidak dikirim ke klien).
	Detail string
	// Err adalah error asli.
	Err error
}

// Error menyusun pesan untuk log (bahasa Inggris).
func (e *DBError) Error() string {
	var b strings.Builder
	b.WriteString("database: ")
	b.WriteString(e.Kind.Error())
	if e.Constraint != "" {
		fmt.Fprintf(&b, " (constraint %s)", e.Constraint)
	}
	if e.Table != "" {
		fmt.Fprintf(&b, " on table %s", e.Table)
	}
	if e.Column != "" {
		fmt.Fprintf(&b, " column %s", e.Column)
	}
	return b.String()
}

// Is membuat errors.Is(err, ErrDuplicate) dan seterusnya bekerja.
func (e *DBError) Is(target error) bool {
	return target == e.Kind
}

// Unwrap mengembalikan error asli (mis. *pgconn.PgError).
func (e *DBError) Unwrap() error {
	return e.Err
}

// kindOf memilih jenis error untuk SQLSTATE PostgreSQL; nil bila tidak dipetakan.
func kindOf(code string) error {
	switch {
	case code == "23505":
		return ErrDuplicate
	case code == "23514":
		return ErrCheckViolation
	case code == "23503", code == "23001":
		return ErrForeignKey
	case code == "23502":
		return ErrNotNull
	case code == "40001", code == "40P01":
		return ErrConflict
	case strings.HasPrefix(code, "22"):
		return ErrInvalidValue
	default:
		return nil
	}
}

// mapError membungkus error PostgreSQL yang dikenal menjadi *DBError; error lain dikembalikan
// apa adanya.
func mapError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}
	kind := kindOf(pgErr.Code)
	if kind == nil {
		return err
	}
	return &DBError{
		Kind:       kind,
		Code:       pgErr.Code,
		Table:      pgErr.TableName,
		Column:     pgErr.ColumnName,
		Constraint: pgErr.ConstraintName,
		Detail:     pgErr.Detail,
		Err:        err,
	}
}

// isRetryable melaporkan apakah transaksi boleh diulang dari awal.
func isRetryable(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && (pgErr.Code == "40001" || pgErr.Code == "40P01")
}

// conflictAfterRetries menandai bahwa retry sudah habis.
func conflictAfterRetries(err error) error {
	mapped := mapError(err)
	return fmt.Errorf("after %d retries: %w", maxRetries, mapped)
}
