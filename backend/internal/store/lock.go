// Lock baris paket sebelum konten paket ditulis.

package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// LockPackage mengunci baris paket sebelum konten ditulis (urutan lock: package_links → paket →
// baris konten, docs/03-architecture.md §3.1). Trigger statement menaikkan
// packages.content_version, jadi tanpa lock ini penulis paralel saling deadlock. FOR NO KEY
// UPDATE tetap mengizinkan sesi lain menyisipkan findings, check run dan ekspor yang merujuk
// paket. Paket yang tidak ada menghasilkan ErrNotFound.
func LockPackage(ctx context.Context, tx pgx.Tx, packageID uuid.UUID) error {
	var one int
	err := tx.QueryRow(ctx, "SELECT 1 FROM packages WHERE id = $1 FOR NO KEY UPDATE", packageID).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("package %s: %w", packageID, ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("lock package %s: %w", packageID, err)
	}
	return nil
}
