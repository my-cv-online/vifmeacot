// Package store adalah satu-satunya jalur akses database: pool pgx, transaksi tulis dengan
// konteks audit (WithTx), retry konflik serialisasi, pemetaan error PostgreSQL, lock paket,
// dan query sqlc (subfolder sqlc). Lihat docs/03-architecture.md §3.1 dan §5.
package store

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Nilai app.source yang dicatat audit_log (docs/04-data-model.md §4).
const (
	// SourceAPI dipakai untuk perubahan dari request API.
	SourceAPI = "api"
	// SourceSync dipakai oleh job sinkron Template General (M9).
	SourceSync = "sync"
	// SourceSeed dipakai saat memuat data demo.
	SourceSeed = "seed"
)

// Pengaturan transaksi.
const (
	// maxRetries adalah jumlah ulangan transaksi setelah serialization failure atau deadlock
	// (docs/03-architecture.md §3.1); total percobaan = 1 + maxRetries.
	maxRetries = 3
	// retryBaseDelay adalah jeda dasar sebelum mengulang; diberi jitter supaya dua transaksi
	// yang bertabrakan tidak mengulang di saat yang sama.
	retryBaseDelay = 10 * time.Millisecond
	// applicationName terlihat di pg_stat_activity untuk membedakan koneksi aplikasi.
	applicationName = "pfmea"
)

// Open membuat pool pgx ke databaseURL dengan maksimal maxConns koneksi. Koneksi baru dibuat
// saat pertama dipakai, sehingga server bisa start walaupun database belum siap (status
// kesiapan dilaporkan /readyz). JIT dimatikan per sesi karena memperlambat query view dan
// aturan (docs/03-architecture.md §4).
func Open(ctx context.Context, databaseURL string, maxConns int32) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		// Error parse pgx bisa memuat potongan URL; jangan sampai kata sandi ikut ke log.
		return nil, errors.New("database URL cannot be parsed")
	}
	cfg.MaxConns = maxConns
	cfg.ConnConfig.RuntimeParams["jit"] = "off"
	cfg.ConnConfig.RuntimeParams["application_name"] = applicationName
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create database pool: %w", err)
	}
	return pool, nil
}

// Store membungkus pool database untuk service.
type Store struct {
	// pool adalah pool koneksi bersama.
	pool *pgxpool.Pool
}

// New membuat Store di atas pool yang sudah dibuka.
func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// Pool mengembalikan pool untuk query baca yang tidak butuh transaksi tulis.
func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}

// Actor adalah siapa dan apa yang melakukan perubahan; disimpan trigger audit_log.
type Actor struct {
	// UserID adalah pengguna yang login; uuid.Nil berarti tidak ada (proses sistem).
	UserID uuid.UUID
	// RequestID adalah id request HTTP (header X-Request-Id) atau id job.
	RequestID string
	// Source adalah asal perubahan: SourceAPI, SourceSync atau SourceSeed; kosong berarti
	// "system" di audit_log.
	Source string
}

// TxFunc adalah isi satu transaksi tulis.
type TxFunc func(ctx context.Context, tx pgx.Tx) error

// WithTx menjalankan fn dalam satu transaksi: mengisi app.user_id, app.request_id dan
// app.source (satu round trip, berlaku lokal di transaksi), menjalankan fn, lalu commit.
// Serialization failure (40001) dan deadlock (40P01) membuat seluruh transaksi diulang sampai
// tiga kali; setelah itu hasilnya ErrConflict. Semua error PostgreSQL, termasuk yang baru muncul
// saat COMMIT (unique deferrable), dipetakan ke jenis error store (lihat errors.go).
func (s *Store) WithTx(ctx context.Context, actor Actor, fn TxFunc) error {
	for attempt := 0; ; attempt++ {
		err := s.runTx(ctx, actor, fn)
		if err == nil {
			return nil
		}
		if !isRetryable(err) {
			return mapError(err)
		}
		if attempt == maxRetries {
			return conflictAfterRetries(err)
		}
		// Jeda bertambah per percobaan, ditambah jitter acak. Konteks yang sudah atau baru
		// dibatalkan menghentikan penantian dan tidak diulang.
		delay := retryBaseDelay*time.Duration(attempt+1) + time.Duration(rand.Int64N(int64(retryBaseDelay)))
		select {
		case <-ctx.Done():
			return mapError(err)
		case <-time.After(delay):
		}
	}
}

// runTx menjalankan satu percobaan transaksi. Transaksi yang tidak di-commit selalu di-rollback,
// juga bila fn panic (panic diteruskan ke pemanggil) atau mengakhiri goroutine dengan
// runtime.Goexit (misalnya t.Fatal di test), supaya koneksi dan lock-nya tidak tertahan.
func (s *Store) runTx(ctx context.Context, actor Actor, fn TxFunc) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	// Setelah Commit yang sukses, Rollback hanya mengembalikan pgx.ErrTxClosed. Konteks baru
	// dipakai karena ctx bisa sudah dibatalkan.
	defer func() { _ = tx.Rollback(context.Background()) }()

	userID := ""
	if actor.UserID != uuid.Nil {
		userID = actor.UserID.String()
	}
	// is_local = true sama dengan SET LOCAL: nilainya hilang setelah transaksi selesai, jadi
	// tidak bocor ke pemakai koneksi berikutnya di pool.
	if _, err = tx.Exec(ctx,
		"SELECT set_config('app.user_id', $1, true), set_config('app.request_id', $2, true), set_config('app.source', $3, true)",
		userID, actor.RequestID, actor.Source); err != nil {
		return fmt.Errorf("set audit context: %w", err)
	}
	if err = fn(ctx, tx); err != nil {
		return err
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}
