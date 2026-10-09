// Test Store: transaksi, retry, pemetaan error, lock paket dan pool (test case TC-M01-014 sampai
// TC-M01-019, docs/test-cases/M01-database.md).

package store_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"pfmea/backend/internal/store"
)

// TestWithTxSettings_TC_M01_014 memastikan WithTx mengisi app.user_id, app.request_id dan
// app.source yang dicatat trigger audit, actor kosong menjadi NULL/NULL/system, dan nilainya
// tidak bocor ke transaksi berikutnya di koneksi yang sama.
func TestWithTxSettings_TC_M01_014(t *testing.T) {
	ctx := context.Background()
	db, _ := newStore(t)
	// Pool satu koneksi supaya transaksi berikutnya pasti memakai koneksi yang sama.
	pool, err := store.Open(ctx, db.URL, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	st := store.New(pool)
	step := stepID(t, pool, ps07, "50")

	// rename mengubah nama step 50 di dalam WithTx dengan actor yang diberikan.
	rename := func(actor store.Actor, name string) {
		t.Helper()
		err := st.WithTx(ctx, actor, func(ctx context.Context, tx pgx.Tx) error {
			_, err := tx.Exec(ctx, "UPDATE process_steps SET name = $1 WHERE id = $2", name, step)
			return err
		})
		if err != nil {
			t.Fatalf("WithTx: %v", err)
		}
	}
	// lastAudit membaca baris audit terakhir untuk step 50.
	lastAudit := func() (userID *uuid.UUID, requestID *string, source string) {
		t.Helper()
		err := pool.QueryRow(ctx, `SELECT user_id, request_id, source FROM audit_log
			WHERE table_name = 'process_steps' AND row_id = $1 ORDER BY id DESC LIMIT 1`, step).Scan(&userID, &requestID, &source)
		if err != nil {
			t.Fatalf("audit row: %v", err)
		}
		return userID, requestID, source
	}

	rename(store.Actor{UserID: apratama, RequestID: "req-1", Source: store.SourceAPI}, "Reflow soldering A")
	u, r, s := lastAudit()
	if u == nil || *u != apratama || r == nil || *r != "req-1" || s != "api" {
		t.Errorf("audit = (%v, %v, %q), want (apratama, req-1, api)", u, r, s)
	}

	rename(store.Actor{}, "Reflow soldering B")
	u, r, s = lastAudit()
	if u != nil || r != nil || s != "system" {
		t.Errorf("audit for empty actor = (%v, %v, %q), want (NULL, NULL, system)", u, r, s)
	}

	// set_config(..., true) berlaku lokal di transaksi; setelahnya nilainya kosong.
	rename(store.Actor{UserID: apratama, RequestID: "req-2", Source: store.SourceAPI}, "Reflow soldering C")
	leaked := scalar[*string](t, pool, "SELECT nullif(current_setting('app.user_id', true), '')")
	if leaked != nil {
		t.Errorf("app.user_id leaked into the next transaction: %q", *leaked)
	}
}

// TestWithTxCommitRollback_TC_M01_015 memastikan perubahan tersimpan bila fn sukses, dibatalkan
// bila fn mengembalikan error, dan dibatalkan lalu panic diteruskan bila fn panic.
func TestWithTxCommitRollback_TC_M01_015(t *testing.T) {
	ctx := context.Background()
	db, st := newStore(t)
	step := stepID(t, db.Pool, ps07, "50")
	// setName mengubah nama step 50 di dalam transaksi.
	setName := func(name string) func(context.Context, pgx.Tx) error {
		return func(ctx context.Context, tx pgx.Tx) error {
			_, err := tx.Exec(ctx, "UPDATE process_steps SET name = $1 WHERE id = $2", name, step)
			return err
		}
	}
	// name membaca nama step 50 saat ini.
	name := func() string {
		return scalar[string](t, db.Pool, "SELECT name FROM process_steps WHERE id = $1", step)
	}

	if err := st.WithTx(ctx, store.Actor{}, setName("Committed")); err != nil {
		t.Fatalf("WithTx: %v", err)
	}
	if got := name(); got != "Committed" {
		t.Errorf("name = %q after commit", got)
	}

	boom := errors.New("boom")
	err := st.WithTx(ctx, store.Actor{}, func(ctx context.Context, tx pgx.Tx) error {
		if err := setName("Rolled back")(ctx, tx); err != nil {
			return err
		}
		return boom
	})
	if !errors.Is(err, boom) {
		t.Errorf("err = %v, want boom", err)
	}
	if got := name(); got != "Committed" {
		t.Errorf("name = %q, the failed transaction was not rolled back", got)
	}

	func() {
		defer func() {
			if p := recover(); p != "panic in fn" {
				t.Errorf("recovered %v, want the original panic", p)
			}
		}()
		_ = st.WithTx(ctx, store.Actor{}, func(ctx context.Context, tx pgx.Tx) error {
			_ = setName("Panicked")(ctx, tx)
			panic("panic in fn")
		})
	}()
	if got := name(); got != "Committed" {
		t.Errorf("name = %q, the panicking transaction was not rolled back", got)
	}
}

// TestWithTxRetry_TC_M01_016 memastikan serialization failure dan deadlock diulang sampai tiga
// kali, lalu menjadi ErrConflict; error lain dan konteks yang batal tidak diulang; deadlock nyata
// berakhir dengan kedua transaksi sukses.
func TestWithTxRetry_TC_M01_016(t *testing.T) {
	ctx := context.Background()
	db, st := newStore(t)

	t.Run("serialization failure twice", func(t *testing.T) {
		var calls int
		err := st.WithTx(ctx, store.Actor{}, func(context.Context, pgx.Tx) error {
			calls++
			if calls <= 2 {
				return &pgconn.PgError{Code: "40001", Message: "could not serialize access"}
			}
			return nil
		})
		if err != nil || calls != 3 {
			t.Errorf("err = %v, calls = %d; want nil and 3", err, calls)
		}
	})

	t.Run("deadlock every time", func(t *testing.T) {
		var calls int
		err := st.WithTx(ctx, store.Actor{}, func(context.Context, pgx.Tx) error {
			calls++
			return &pgconn.PgError{Code: "40P01", Message: "deadlock detected"}
		})
		if !errors.Is(err, store.ErrConflict) || calls != 4 {
			t.Errorf("err = %v, calls = %d; want ErrConflict and 4", err, calls)
		}
	})

	t.Run("other errors are not retried", func(t *testing.T) {
		var calls int
		err := st.WithTx(ctx, store.Actor{}, func(context.Context, pgx.Tx) error {
			calls++
			return &pgconn.PgError{Code: "23505", ConstraintName: "x"}
		})
		if !errors.Is(err, store.ErrDuplicate) || calls != 1 {
			t.Errorf("err = %v, calls = %d; want ErrDuplicate and 1", err, calls)
		}
	})

	t.Run("cancelled context is not retried", func(t *testing.T) {
		cctx, cancel := context.WithCancel(ctx)
		var calls int
		err := st.WithTx(cctx, store.Actor{}, func(context.Context, pgx.Tx) error {
			calls++
			cancel()
			return &pgconn.PgError{Code: "40001"}
		})
		if calls != 1 || err == nil {
			t.Errorf("err = %v, calls = %d; want an error after 1 call", err, calls)
		}
	})

	t.Run("real deadlock", func(t *testing.T) {
		// Baris customers dipakai (bukan konten paket): update konten juga mengunci baris paket
		// lewat trigger content_version, sehingga transaksi kedua sudah tertahan sebelum
		// deadlock terjadi. Itulah alasan LockPackage wajib dipanggil lebih dulu.
		a := scalar[uuid.UUID](t, db.Pool, "SELECT id FROM customers WHERE code = 'CA'")
		b := scalar[uuid.UUID](t, db.Pool, "SELECT id FROM customers WHERE code = 'CB'")
		// Kedua transaksi mengunci baris dengan urutan terbalik. Pada percobaan pertama mereka
		// saling menunggu setelah update pertama sehingga PostgreSQL mendeteksi deadlock.
		var arrived sync.WaitGroup
		arrived.Add(2)
		bothArrived := make(chan struct{})
		go func() { arrived.Wait(); close(bothArrived) }()
		var retries atomic.Int32
		run := func(first, second uuid.UUID) error {
			attempt := 0
			return st.WithTx(ctx, store.Actor{}, func(ctx context.Context, tx pgx.Tx) error {
				attempt++
				if attempt > 1 {
					retries.Add(1)
				}
				if _, err := tx.Exec(ctx, "UPDATE customers SET notes = 'A' WHERE id = $1", first); err != nil {
					return err
				}
				if attempt == 1 {
					arrived.Done()
					// Batas waktu supaya test tidak pernah menggantung bila sinkronisasi gagal.
					select {
					case <-bothArrived:
					case <-time.After(10 * time.Second):
						return errors.New("the other transaction never reached the barrier")
					}
				}
				_, err := tx.Exec(ctx, "UPDATE customers SET notes = 'B' WHERE id = $1", second)
				return err
			})
		}
		errs := make(chan error, 2)
		go func() { errs <- run(a, b) }()
		go func() { errs <- run(b, a) }()
		for range 2 {
			if err := <-errs; err != nil {
				t.Errorf("transaction failed after retries: %v", err)
			}
		}
		if retries.Load() == 0 {
			t.Errorf("expected at least one retry after the deadlock")
		}
	})
}

// TestErrorMapping_TC_M01_017 memastikan error PostgreSQL dipetakan ke jenis store, termasuk
// pelanggaran unique yang baru terdeteksi saat COMMIT (constraint deferrable), dan error asli
// tetap bisa diambil.
func TestErrorMapping_TC_M01_017(t *testing.T) {
	ctx := context.Background()
	db, st := newStore(t)
	step50 := stepID(t, db.Pool, ps07, "50")
	step60 := stepID(t, db.Pool, ps07, "60")
	// exec menjalankan satu statement di dalam WithTx.
	exec := func(sql string, args ...any) error {
		return st.WithTx(ctx, store.Actor{}, func(ctx context.Context, tx pgx.Tx) error {
			_, err := tx.Exec(ctx, sql, args...)
			return err
		})
	}

	// 1. Duplikat op_no baru gagal saat COMMIT.
	err := exec("UPDATE process_steps SET op_no = '50' WHERE id = $1", step60)
	var dbErr *store.DBError
	if !errors.Is(err, store.ErrDuplicate) || !errors.As(err, &dbErr) || dbErr.Constraint != "process_steps_op_no_uq" {
		t.Errorf("duplicate op_no: err = %v, want ErrDuplicate on process_steps_op_no_uq", err)
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		t.Errorf("the original *pgconn.PgError is not reachable: %v", err)
	}

	// 2. Menukar op_no dalam satu transaksi diizinkan karena constraint dicek saat COMMIT.
	err = st.WithTx(ctx, store.Actor{}, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, "UPDATE process_steps SET op_no = '60' WHERE id = $1", step50); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, "UPDATE process_steps SET op_no = '50' WHERE id = $1", step60)
		return err
	})
	if err != nil {
		t.Errorf("swapping op_no in one transaction failed: %v", err)
	}

	// 3. Check constraint.
	err = exec("UPDATE failure_chains SET o = 11 WHERE package_id = $1", ps07)
	if !errors.Is(err, store.ErrCheckViolation) {
		t.Errorf("o = 11: err = %v, want ErrCheckViolation", err)
	}

	// 4. Foreign key.
	err = exec(`INSERT INTO characteristics (package_id, step_id, char_no, kind, name)
		VALUES ($1, $2, '99-01', 'product', 'Orphan')`, ps07, uuid.New())
	if !errors.Is(err, store.ErrForeignKey) || !errors.As(err, &dbErr) || dbErr.Table != "characteristics" {
		t.Errorf("foreign key: err = %v, want ErrForeignKey on characteristics", err)
	}

	// 5. NOT NULL.
	err = exec("INSERT INTO process_steps (package_id, op_no, seq) VALUES ($1, '97', 970)", ps07)
	if !errors.Is(err, store.ErrNotNull) || !errors.As(err, &dbErr) || dbErr.Column != "name" {
		t.Errorf("not null: err = %v, want ErrNotNull on column name", err)
	}
}

// TestLockPackage_TC_M01_018 memastikan LockPackage mengembalikan ErrNotFound untuk paket yang
// tidak ada, menahan lock paket dan update baris paket dari transaksi lain, tetapi tetap
// mengizinkan insert check_runs dan findings yang merujuk paket itu.
func TestLockPackage_TC_M01_018(t *testing.T) {
	ctx := context.Background()
	db, st := newStore(t)

	err := st.WithTx(ctx, store.Actor{}, func(ctx context.Context, tx pgx.Tx) error {
		return store.LockPackage(ctx, tx, uuid.New())
	})
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("unknown package: err = %v, want ErrNotFound", err)
	}

	holder, err := db.Pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = holder.Rollback(ctx) }()
	if err := store.LockPackage(ctx, holder, ps07); err != nil {
		t.Fatalf("LockPackage: %v", err)
	}

	// blocked menjalankan fn di transaksi lain dengan lock_timeout pendek dan mengembalikan
	// kode SQLSTATE error-nya.
	blocked := func(fn func(pgx.Tx) error) string {
		t.Helper()
		tx, err := db.Pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		if _, err := tx.Exec(ctx, "SET LOCAL lock_timeout = '200ms'"); err != nil {
			t.Fatal(err)
		}
		var pgErr *pgconn.PgError
		if err := fn(tx); errors.As(err, &pgErr) {
			return pgErr.Code
		} else if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return ""
	}
	if code := blocked(func(tx pgx.Tx) error { return store.LockPackage(ctx, tx, ps07) }); code != "55P03" {
		t.Errorf("second LockPackage: code %q, want 55P03 (lock not available)", code)
	}
	if code := blocked(func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, "UPDATE packages SET name = name || ' x' WHERE id = $1", ps07)
		return err
	}); code != "55P03" {
		t.Errorf("package update: code %q, want 55P03", code)
	}
	if code := blocked(func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, "INSERT INTO check_runs (package_id, trigger, rule_codes) VALUES ($1, 'manual', '{K01}')", ps07); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `INSERT INTO findings (package_id, rule_code, level, object_type, message, fingerprint)
			VALUES ($1, 'K01', 'error', 'process_steps', 'Test finding', 'fp-lock-test')`, ps07)
		return err
	}); code != "" {
		t.Errorf("check_runs/findings insert while the package is locked: code %q, want success", code)
	}
}

// TestOpen_TC_M01_019 memastikan Open tidak langsung terhubung, koneksi memakai jit off dan
// application_name pfmea, ukuran pool mengikuti argumen, dan URL yang salah tidak membocorkan
// kata sandi di pesan error.
func TestOpen_TC_M01_019(t *testing.T) {
	ctx := context.Background()

	lazy, err := store.Open(ctx, "postgres://pfmea:secret@127.0.0.1:1/pfmea", 2)
	if err != nil {
		t.Errorf("Open should not connect eagerly: %v", err)
	} else {
		lazy.Close()
	}

	_, err = store.Open(ctx, "postgres://pfmea:secret@[::1/pfmea", 2)
	if err == nil || strings.Contains(err.Error(), "secret") {
		t.Errorf("invalid URL: err = %v, want an error without the password", err)
	}

	db, _ := newStore(t)
	pool, err := store.Open(ctx, db.URL, 3)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if pool.Config().MaxConns != 3 {
		t.Errorf("MaxConns = %d, want 3", pool.Config().MaxConns)
	}
	if jit := scalar[string](t, pool, "SELECT current_setting('jit')"); jit != "off" {
		t.Errorf("jit = %s, want off", jit)
	}
	if app := scalar[string](t, pool, "SELECT current_setting('application_name')"); app != "pfmea" {
		t.Errorf("application_name = %s, want pfmea", app)
	}
}
