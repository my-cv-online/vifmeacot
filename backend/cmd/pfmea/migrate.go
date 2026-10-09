// Perintah `pfmea migrate up|down|status`: menerapkan, membatalkan atau menampilkan migrasi
// database yang tertanam di binary (db/migrations, goose). Hanya DATABASE_URL yang wajib
// (config.LoadDatabase), jadi perintah ini bisa dijalankan sebelum konfigurasi server lengkap.

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"pfmea/backend/internal/config"
	"pfmea/backend/internal/i18n"
	"pfmea/db"
)

// connectTimeout membatasi waktu membuka koneksi bila DATABASE_URL tidak menyebut
// connect_timeout, supaya perintah tidak menggantung ketika database tidak terjangkau.
const connectTimeout = 10 * time.Second

// migrateCmd menjalankan `pfmea migrate <sub>`. Sub-perintah down menghapus semua data, jadi
// hanya diizinkan dengan DEV_MODE=true.
func migrateCmd(ctx context.Context, args []string, getenv func(string) string, stdout, stderr io.Writer) int {
	if len(args) != 1 || !slices.Contains([]string{"up", "down", "status"}, args[0]) {
		_, _ = fmt.Fprintln(stderr, i18n.MigrateUsage)
		return exitUsage
	}
	sub := args[0]
	cfg, err := config.LoadDatabase(getenv)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, i18n.CLICommandFailed+"\n", "migrate", err)
		return exitError
	}
	name := "migrate " + sub
	if sub == "down" && !cfg.DevMode {
		_, _ = fmt.Fprintln(stderr, i18n.MigrateDownNeedsDevMode)
		return exitError
	}

	migrator, err := openMigrator(cfg.URL)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, i18n.CLICommandFailed+"\n", name, err)
		return exitError
	}
	defer func() { _ = migrator.Close() }()

	switch sub {
	case "up":
		err = migrateUp(ctx, migrator, stdout)
	case "down":
		err = migrateDown(ctx, migrator, stdout)
	default:
		err = migrateStatus(ctx, migrator, stdout)
	}
	if err != nil {
		_, _ = fmt.Fprintf(stderr, i18n.CLICommandFailed+"\n", name, err)
		return exitError
	}
	return exitOK
}

// connConfig membaca databaseURL menjadi konfigurasi koneksi pgx dengan batas waktu koneksi,
// JIT mati dan application_name yang terlihat di pg_stat_activity. Error parse pgx bisa memuat
// potongan URL, jadi diganti pesan tetap supaya kata sandi tidak tercetak.
func connConfig(databaseURL, applicationName string) (*pgx.ConnConfig, error) {
	cfg, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		return nil, errors.New(i18n.MigrateDatabaseURLInvalid)
	}
	if cfg.ConnectTimeout == 0 {
		cfg.ConnectTimeout = connectTimeout
	}
	cfg.RuntimeParams["jit"] = "off"
	cfg.RuntimeParams["application_name"] = applicationName
	return cfg, nil
}

// openMigrator membuka migrator goose di atas *sql.DB baru. Koneksi baru dibuat saat migrator
// pertama dipakai; menutup migrator juga menutup *sql.DB.
func openMigrator(databaseURL string) (*goose.Provider, error) {
	cfg, err := connConfig(databaseURL, "pfmea-migrate")
	if err != nil {
		return nil, err
	}
	sqlDB := stdlib.OpenDB(*cfg)
	migrator, err := db.NewMigrator(sqlDB)
	if err != nil {
		_ = sqlDB.Close()
		return nil, err
	}
	return migrator, nil
}

// migrateUp menerapkan semua migrasi yang pending dan mencetak satu baris per file. Bila satu
// migrasi gagal, file yang sudah diterapkan sebelumnya tetap dicetak.
func migrateUp(ctx context.Context, m *goose.Provider, stdout io.Writer) error {
	results, err := m.Up(ctx)
	var partial *goose.PartialError
	if errors.As(err, &partial) {
		results = partial.Applied
	}
	for _, r := range results {
		_, _ = fmt.Fprintf(stdout, i18n.MigrateApplied+"\n", path.Base(r.Source.Path))
	}
	if err != nil {
		return err
	}
	if len(results) == 0 {
		_, _ = fmt.Fprintln(stdout, i18n.MigrateNoPending)
	}
	return nil
}

// migrateDown membatalkan semua migrasi (versi 0) dan mencetak satu baris per file. Tabel
// goose_db_version tetap ada.
func migrateDown(ctx context.Context, m *goose.Provider, stdout io.Writer) error {
	results, err := m.DownTo(ctx, 0)
	var partial *goose.PartialError
	if errors.As(err, &partial) {
		results = partial.Applied
	}
	for _, r := range results {
		_, _ = fmt.Fprintf(stdout, i18n.MigrateRolledBack+"\n", path.Base(r.Source.Path))
	}
	if err != nil {
		return err
	}
	if len(results) == 0 {
		_, _ = fmt.Fprintln(stdout, i18n.MigrateNothingToRollBack)
	}
	return nil
}

// migrateStatus mencetak setiap file migrasi dengan statusnya, lalu versi skema di database.
func migrateStatus(ctx context.Context, m *goose.Provider, stdout io.Writer) error {
	statuses, err := m.Status(ctx)
	if err != nil {
		return err
	}
	for _, s := range statuses {
		state := string(s.State)
		switch s.State {
		case goose.StateApplied:
			state = i18n.MigrateStateApplied
		case goose.StatePending:
			state = i18n.MigrateStatePending
		}
		_, _ = fmt.Fprintf(stdout, i18n.MigrateStatusLine+"\n", path.Base(s.Source.Path), state)
	}
	version, err := m.GetDBVersion(ctx)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(stdout, i18n.MigrateVersion+"\n", version)
	return nil
}
