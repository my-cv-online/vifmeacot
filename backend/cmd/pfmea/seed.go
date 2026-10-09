// Perintah `pfmea seed-demo [--reset]`: memuat data demo (db/seed/demo.sql) untuk pengembangan,
// test E2E dan latihan Gate 1. Hanya berjalan dengan DEV_MODE=true karena data demo tidak boleh
// masuk ke database produksi (docs/03-architecture.md §3.2).

package main

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/jackc/pgx/v5"

	"pfmea/backend/internal/config"
	"pfmea/backend/internal/i18n"
	"pfmea/db"
)

// resetLockTimeout membatasi tunggu lock saat --reset menghapus skema. Server yang masih
// memegang transaksi terbuka membuat DROP SCHEMA menunggu; lebih baik gagal dengan jelas
// daripada menggantung.
const resetLockTimeout = "10s"

// seedDemoCmd menjalankan `pfmea seed-demo [--reset]`. Tanpa --reset, data demo hanya dimuat
// ke database yang sudah dimigrasi ke versi terbaru dan belum berisi data. Dengan --reset,
// skema public dihapus, semua migrasi dijalankan ulang, lalu data demo dimuat.
func seedDemoCmd(ctx context.Context, args []string, getenv func(string) string, stdout, stderr io.Writer) int {
	reset, ok := parseSeedArgs(args)
	if !ok {
		_, _ = fmt.Fprintln(stderr, i18n.SeedUsage)
		return exitUsage
	}
	cfg, err := config.LoadDatabase(getenv)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, i18n.CLICommandFailed+"\n", "seed-demo", err)
		return exitError
	}
	if !cfg.DevMode {
		_, _ = fmt.Fprintln(stderr, i18n.SeedNeedsDevMode)
		return exitError
	}
	if err := seedDemo(ctx, cfg.URL, reset, stdout, stderr); err != nil {
		if !errors.Is(err, errRefused) {
			_, _ = fmt.Fprintf(stderr, i18n.CLICommandFailed+"\n", "seed-demo", err)
		}
		return exitError
	}
	return exitOK
}

// errRefused menandai penolakan yang pesannya sudah dicetak ke stderr.
var errRefused = errors.New("refused")

// seedDemo menjalankan langkah seed-demo setelah konfigurasi diperiksa. Penolakan yang sudah
// dicetak dikembalikan sebagai errRefused dan tidak dicetak ulang oleh pemanggil.
func seedDemo(ctx context.Context, databaseURL string, reset bool, stdout, stderr io.Writer) error {
	connCfg, err := connConfig(databaseURL, "pfmea-seed")
	if err != nil {
		return err
	}
	if reset {
		if err := resetSchema(ctx, connCfg); err != nil {
			return err
		}
		_, _ = fmt.Fprintln(stdout, i18n.SeedSchemaReset)
		if err := migrateAll(ctx, databaseURL, stdout); err != nil {
			return err
		}
	}

	conn, err := pgx.ConnectConfig(ctx, connCfg)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close(context.WithoutCancel(ctx)) }()

	// Versi diperiksa tanpa goose supaya database yang belum dimigrasi tidak diubah.
	if err := db.CheckSchema(ctx, conn); err != nil {
		var schemaErr *db.SchemaError
		if !errors.As(err, &schemaErr) {
			return err
		}
		_, _ = fmt.Fprintf(stderr, i18n.SeedNotMigrated+"\n", schemaErr.Current, schemaErr.Latest)
		return errRefused
	}
	var hasData bool
	if err := conn.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM users) OR EXISTS (SELECT 1 FROM customers)
		OR EXISTS (SELECT 1 FROM packages)`).Scan(&hasData); err != nil {
		return fmt.Errorf("check existing data: %w", err)
	}
	if hasData {
		_, _ = fmt.Fprintln(stderr, i18n.SeedHasData)
		return errRefused
	}

	if err := db.LoadDemoSeed(ctx, conn); err != nil {
		return err
	}
	// TODO(M8): jalankan cek penuh untuk setiap paket setelah data demo dimuat.
	var users, packages int
	if err := conn.QueryRow(ctx, "SELECT (SELECT count(*) FROM users), (SELECT count(*) FROM packages)").
		Scan(&users, &packages); err != nil {
		return fmt.Errorf("count demo data: %w", err)
	}
	_, _ = fmt.Fprintf(stdout, i18n.SeedLoaded+"\n", users, packages)
	if reset {
		_, _ = fmt.Fprintln(stdout, i18n.SeedRestartServer)
	}
	return nil
}

// parseSeedArgs membaca argumen seed-demo; ok = false bila ada argumen yang tidak dikenal.
func parseSeedArgs(args []string) (reset, ok bool) {
	switch {
	case len(args) == 0:
		return false, true
	case len(args) == 1 && args[0] == "--reset":
		return true, true
	default:
		return false, false
	}
}

// resetSchema menghapus skema public beserta semua isinya (tabel, fungsi, tipe, extension,
// tabel goose) lalu membuatnya lagi dengan hak USAGE untuk PUBLIC seperti skema bawaan
// PostgreSQL. Kedua statement dikirim sekaligus sehingga berjalan dalam satu transaksi.
func resetSchema(ctx context.Context, connCfg *pgx.ConnConfig) error {
	cfg := connCfg.Copy()
	cfg.RuntimeParams["lock_timeout"] = resetLockTimeout
	conn, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close(context.WithoutCancel(ctx)) }()
	if _, err := conn.Exec(ctx, `DROP SCHEMA public CASCADE;
		CREATE SCHEMA public;
		GRANT USAGE ON SCHEMA public TO PUBLIC`); err != nil {
		return fmt.Errorf("reset schema public: %w", err)
	}
	return nil
}

// migrateAll menjalankan semua migrasi (seperti `pfmea migrate up`) dan mencetak hasilnya.
func migrateAll(ctx context.Context, databaseURL string, stdout io.Writer) error {
	migrator, err := openMigrator(databaseURL)
	if err != nil {
		return err
	}
	defer func() { _ = migrator.Close() }()
	return migrateUp(ctx, migrator, stdout)
}
