// Migrator goose untuk file migrasi yang tertanam.

package db

import (
	"database/sql"
	"fmt"
	"io/fs"

	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

// NewMigrator membuat migrator goose untuk migrasi tertanam di atas koneksi sqlDB. Session lock
// PostgreSQL mencegah dua proses menjalankan migrasi bersamaan; registry global goose dimatikan
// supaya hanya file SQL tertanam yang dipakai. Menutup migrator juga menutup sqlDB.
// TODO(M8): jalankan migrasi River sesudah (up) dan sebelum (down) migrasi goose.
func NewMigrator(sqlDB *sql.DB) (*goose.Provider, error) {
	fsys, err := fs.Sub(Files, migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("open embedded migrations: %w", err)
	}
	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return nil, fmt.Errorf("create migration lock: %w", err)
	}
	return goose.NewProvider(goose.DialectPostgres, sqlDB, fsys,
		goose.WithSessionLocker(locker),
		goose.WithDisableGlobalRegistry(true),
	)
}
