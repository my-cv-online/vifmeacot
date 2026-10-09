// Pemuat data demo (db/seed/demo.sql).

package db

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/jackc/pgx/v5"
)

// demoSeed adalah path file data demo di dalam Files.
const demoSeed = "seed/demo.sql"

// LoadDemoSeed menjalankan data demo pada conn. File seed punya BEGIN…COMMIT sendiri, jadi conn
// harus koneksi biasa (autocommit), bukan transaksi yang sedang terbuka; pgx menjalankan teks
// banyak-statement lewat simple protocol karena tidak ada argumen.
func LoadDemoSeed(ctx context.Context, conn *pgx.Conn) error {
	b, err := fs.ReadFile(Files, demoSeed)
	if err != nil {
		return fmt.Errorf("read embedded seed: %w", err)
	}
	if _, err := conn.Exec(ctx, string(b)); err != nil {
		return fmt.Errorf("load demo seed: %w", err)
	}
	return nil
}
