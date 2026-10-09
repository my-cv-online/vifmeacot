// Package db menanam file migrasi skema (goose) dan data demo ke binary, dan menyediakan
// migrator serta pemuat seed yang dipakai `pfmea migrate`, `pfmea seed-demo` dan test database.
package db

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strconv"
	"strings"
)

// Files berisi migrations/*.sql dan seed/*.sql. Isinya tidak pernah diubah saat program berjalan.
//
//go:embed migrations/*.sql seed/*.sql
var Files embed.FS

// migrationsDir adalah folder migrasi di dalam Files.
const migrationsDir = "migrations"

// LatestVersion mengembalikan versi migrasi tertinggi yang tertanam, dibaca dari awalan angka
// nama file (00001_init.sql → 1). Dipakai pemeriksaan kesiapan untuk tahu apakah database
// sudah dimigrasi sampai versi terbaru.
func LatestVersion() (int64, error) {
	entries, err := fs.ReadDir(Files, migrationsDir)
	if err != nil {
		return 0, fmt.Errorf("read embedded migrations: %w", err)
	}
	var latest int64
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || path.Ext(name) != ".sql" {
			continue
		}
		prefix, _, ok := strings.Cut(name, "_")
		if !ok {
			return 0, fmt.Errorf("migration %s has no version prefix", name)
		}
		v, err := strconv.ParseInt(prefix, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("migration %s has an invalid version prefix: %w", name, err)
		}
		latest = max(latest, v)
	}
	if latest == 0 {
		return 0, fmt.Errorf("no embedded migrations found")
	}
	return latest, nil
}
