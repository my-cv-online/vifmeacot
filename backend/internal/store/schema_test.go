// Test karakterisasi skema: trigger dan constraint di db/migrations/00001_init.sql yang menjaga
// nilai turunan, versi, audit dan integritas (test case TC-M01-021 sampai TC-M01-028,
// docs/test-cases/M01-database.md). Skema sudah ada di paket spesifikasi; test ini membuktikan
// perilakunya di PostgreSQL 18.

package store_test

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"pfmea/backend/internal/store"
)

// generalPkg adalah paket Template General di data demo.
var generalPkg = uuid.MustParse("07ea2e55-0ce4-5e2b-b9b0-681883ae41ce")

// exec menjalankan satu statement dan menggagalkan test bila error.
func exec(t *testing.T, pool *pgxpool.Pool, sql string, args ...any) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), sql, args...); err != nil {
		t.Fatalf("exec %q: %v", sql, err)
	}
}

// sqlState menjalankan statement yang diharapkan gagal dan mengembalikan SQLSTATE-nya.
func sqlState(t *testing.T, pool *pgxpool.Pool, sql string, args ...any) string {
	t.Helper()
	_, err := pool.Exec(context.Background(), sql, args...)
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		t.Fatalf("exec %q: want a PostgreSQL error, got %v", sql, err)
	}
	return pgErr.Code
}

// failureModeID mencari failure mode PS-07 berdasarkan teksnya.
func failureModeID(t *testing.T, pool *pgxpool.Pool, text string) uuid.UUID {
	t.Helper()
	return scalar[uuid.UUID](t, pool, "SELECT id FROM failure_modes WHERE package_id = $1 AND text = $2", ps07, text)
}

// chainSeverities mengembalikan s semua chain failure mode (NULL menjadi -1).
func chainSeverities(t *testing.T, pool *pgxpool.Pool, fm uuid.UUID) []int {
	t.Helper()
	rows, err := pool.Query(context.Background(),
		"SELECT coalesce(s, -1) FROM failure_chains WHERE failure_mode_id = $1 ORDER BY id", fm)
	if err != nil {
		t.Fatal(err)
	}
	s, err := pgx.CollectRows(rows, pgx.RowTo[int])
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// addChain menambah cause dan chain baru (o = 3, d = 4) pada failure mode, mengembalikan id chain.
func addChain(t *testing.T, pool *pgxpool.Pool, fm uuid.UUID, cause string) uuid.UUID {
	t.Helper()
	causeID := scalar[uuid.UUID](t, pool,
		"INSERT INTO failure_causes (package_id, failure_mode_id, text) VALUES ($1, $2, $3) RETURNING id", ps07, fm, cause)
	return scalar[uuid.UUID](t, pool,
		"INSERT INTO failure_chains (package_id, failure_mode_id, failure_cause_id, o, d) VALUES ($1, $2, $3, 3, 4) RETURNING id",
		ps07, fm, causeID)
}

// TestChainSeverity_TC_M01_021 memastikan failure_chains.s selalu S effect tertinggi dari failure
// mode-nya: saat chain dibuat, saat effect ditambah, diubah, dihapus dan dipindah, dan NULL bila
// failure mode tidak punya effect.
func TestChainSeverity_TC_M01_021(t *testing.T) {
	db, _ := newStore(t)
	pool := db.Pool
	fm70 := failureModeID(t, pool, "Defect escapes AOI") // dua chain, satu effect S 7
	fm90 := failureModeID(t, pool, "Cosmetic defect not detected")

	// check membandingkan s semua chain failure mode dengan nilai yang diharapkan (-1 = NULL).
	check := func(fm uuid.UUID, want ...int) {
		t.Helper()
		if got := chainSeverities(t, pool, fm); !slices.Equal(got, want) {
			t.Errorf("chain s = %v, want %v", got, want)
		}
	}
	check(fm70, 7, 7)

	high := scalar[uuid.UUID](t, pool,
		"INSERT INTO failure_effects (package_id, failure_mode_id, level, text, s) VALUES ($1, $2, 'end_user', 'Safety issue', 9) RETURNING id", ps07, fm70)
	check(fm70, 9, 9)
	exec(t, pool, "UPDATE failure_effects SET s = 6 WHERE id = $1", high)
	check(fm70, 7, 7)
	exec(t, pool, "UPDATE failure_effects SET s = 10 WHERE id = $1", high)
	check(fm70, 10, 10)
	exec(t, pool, "DELETE FROM failure_effects WHERE id = $1", high)
	check(fm70, 7, 7)

	// Chain baru langsung mendapat s dari effect yang ada (trigger BEFORE INSERT).
	addChain(t, pool, fm70, "Lighting changed")
	check(fm70, 7, 7, 7)

	// Effect S 7 dipindah ke failure mode step 90 (S 5): fm70 tidak punya effect lagi.
	exec(t, pool, "UPDATE failure_effects SET failure_mode_id = $1 WHERE failure_mode_id = $2", fm90, fm70)
	check(fm70, -1, -1, -1)
	check(fm90, 7)

	// Failure mode baru tanpa effect: chain pertamanya mendapat s NULL saat dibuat.
	step := stepID(t, pool, ps07, "90")
	fresh := scalar[uuid.UUID](t, pool,
		`INSERT INTO failure_modes (package_id, step_id, characteristic_id, text)
		SELECT $1, $2, id, 'Label missing' FROM characteristics WHERE package_id = $1 AND step_id = $2 ORDER BY char_no LIMIT 1
		RETURNING id`, ps07, step)
	addChain(t, pool, fresh, "Printer out of labels")
	check(fresh, -1)
}

// TestRPN_TC_M01_022 memastikan rpn dan new_rpn dihitung database (S × O × D) dan NULL bila ada
// faktor NULL.
func TestRPN_TC_M01_022(t *testing.T) {
	db, _ := newStore(t)
	pool := db.Pool
	fm := failureModeID(t, pool, "Missing component") // S 7, O 4, D 4
	chain := scalar[uuid.UUID](t, pool, "SELECT id FROM failure_chains WHERE failure_mode_id = $1", fm)
	// rpn membaca rpn chain yang diuji.
	rpn := func() *int {
		return scalar[*int](t, pool, "SELECT rpn FROM failure_chains WHERE id = $1", chain)
	}
	if r := rpn(); r == nil || *r != 112 {
		t.Errorf("seed rpn = %v, want 112", r)
	}
	exec(t, pool, "UPDATE failure_chains SET o = 2, d = 3 WHERE id = $1", chain)
	if r := rpn(); r == nil || *r != 42 {
		t.Errorf("rpn = %v, want 42", r)
	}
	exec(t, pool, "UPDATE failure_chains SET d = NULL WHERE id = $1", chain)
	if r := rpn(); r != nil {
		t.Errorf("rpn = %d, want NULL when D is empty", *r)
	}

	action := scalar[uuid.UUID](t, pool,
		"INSERT INTO actions (package_id, failure_chain_id, text, new_s, new_o, new_d) VALUES ($1, $2, 'Add AOI', 7, 2, 3) RETURNING id", ps07, chain)
	// newRPN membaca new_rpn aksi yang diuji.
	newRPN := func() *int {
		return scalar[*int](t, pool, "SELECT new_rpn FROM actions WHERE id = $1", action)
	}
	if r := newRPN(); r == nil || *r != 42 {
		t.Errorf("new_rpn = %v, want 42", r)
	}
	exec(t, pool, "UPDATE actions SET new_o = NULL WHERE id = $1", action)
	if r := newRPN(); r != nil {
		t.Errorf("new_rpn = %d, want NULL when new O is empty", *r)
	}
}

// TestVersion_TC_M01_023 memastikan version hanya naik untuk perubahan yang bermakna bagi
// pengguna: tidak untuk update tanpa perubahan dan tidak untuk perubahan s/rpn oleh trigger.
func TestVersion_TC_M01_023(t *testing.T) {
	db, _ := newStore(t)
	pool := db.Pool
	fm := failureModeID(t, pool, "Missing component")
	chain := scalar[uuid.UUID](t, pool, "SELECT id FROM failure_chains WHERE failure_mode_id = $1", fm)
	effect := scalar[uuid.UUID](t, pool, "SELECT id FROM failure_effects WHERE failure_mode_id = $1", fm)
	// version membaca kolom version satu baris.
	version := func(table string, id uuid.UUID) int {
		return scalar[int](t, pool, "SELECT version FROM "+pgx.Identifier{table}.Sanitize()+" WHERE id = $1", id)
	}

	v := version("failure_chains", chain)
	exec(t, pool, "UPDATE failure_chains SET o = 5 WHERE id = $1", chain)
	if got := version("failure_chains", chain); got != v+1 {
		t.Errorf("version after changing O = %d, want %d", got, v+1)
	}
	exec(t, pool, "UPDATE failure_chains SET o = o WHERE id = $1", chain)
	if got := version("failure_chains", chain); got != v+1 {
		t.Errorf("version after a no-op update = %d, want %d", got, v+1)
	}

	ev := version("failure_effects", effect)
	exec(t, pool, "UPDATE failure_effects SET s = 8 WHERE id = $1", effect)
	if s := scalar[int](t, pool, "SELECT s FROM failure_chains WHERE id = $1", chain); s != 8 {
		t.Fatalf("chain s = %d, want 8", s)
	}
	if got := version("failure_chains", chain); got != v+1 {
		t.Errorf("chain version after the trigger changed s/rpn = %d, want %d", got, v+1)
	}
	if got := version("failure_effects", effect); got != ev+1 {
		t.Errorf("effect version = %d, want %d", got, ev+1)
	}
}

// TestContentVersion_TC_M01_024 memastikan content_version naik satu per statement untuk setiap
// paket yang tersentuh, tidak naik untuk update tanpa perubahan, dan tidak naik untuk tabel hasil
// cek (findings, check_runs).
func TestContentVersion_TC_M01_024(t *testing.T) {
	db, _ := newStore(t)
	pool := db.Pool
	// cv membaca content_version paket.
	cv := func(pkg uuid.UUID) int64 {
		return scalar[int64](t, pool, "SELECT content_version FROM packages WHERE id = $1", pkg)
	}
	steps := []string{"50", "60", "70"}

	cases := []struct {
		name              string
		sql               string
		args              []any
		wantPS07, wantGen int64
	}{
		{"one row", "UPDATE process_steps SET department = 'Line 1' WHERE package_id = $1 AND op_no = '50'", []any{ps07}, 1, 0},
		{"three rows in one statement", "UPDATE process_steps SET department = 'Line 2' WHERE package_id = $1 AND op_no = ANY($2)", []any{ps07, steps}, 1, 0},
		{"two packages", "UPDATE process_steps SET department = 'Warehouse 2' WHERE op_no = '10'", nil, 1, 1},
		{"no-op", "UPDATE process_steps SET department = department WHERE package_id = $1", []any{ps07}, 0, 0},
		{"effect S change", "UPDATE failure_effects SET s = 8 WHERE package_id = $1 AND text = 'Board fails function test; rework'", []any{ps07}, 1, 0},
		{"check run", "INSERT INTO check_runs (package_id, trigger, rule_codes) VALUES ($1, 'manual', '{K01}')", []any{ps07}, 0, 0},
		{"finding", `INSERT INTO findings (package_id, rule_code, level, object_type, message, fingerprint)
			VALUES ($1, 'K01', 'error', 'process_steps', 'Test finding', 'fp-cv-test')`, []any{ps07}, 0, 0},
	}
	for _, c := range cases {
		before07, beforeGen := cv(ps07), cv(generalPkg)
		exec(t, pool, c.sql, c.args...)
		if got := cv(ps07) - before07; got != c.wantPS07 {
			t.Errorf("%s: PS-07 content_version +%d, want +%d", c.name, got, c.wantPS07)
		}
		if got := cv(generalPkg) - beforeGen; got != c.wantGen {
			t.Errorf("%s: GENERAL content_version +%d, want +%d", c.name, got, c.wantGen)
		}
	}
}

// auditRow adalah satu baris audit_log yang dibaca test.
type auditRow struct {
	Action string
	Old    map[string]any
	New    map[string]any
}

// auditRows membaca baris audit sesudah id tertentu untuk satu tabel.
func auditRows(t *testing.T, pool *pgxpool.Pool, table string, afterID int64) []auditRow {
	t.Helper()
	rows, err := pool.Query(context.Background(),
		"SELECT action, coalesce(old_row, '{}'), coalesce(new_row, '{}') FROM audit_log WHERE table_name = $1 AND id > $2 ORDER BY id",
		table, afterID)
	if err != nil {
		t.Fatal(err)
	}
	// Ditutup juga saat t.Fatal di dalam loop, supaya koneksi kembali ke pool dan cleanup tidak
	// menggantung.
	defer rows.Close()
	var out []auditRow
	for rows.Next() {
		var r auditRow
		var oldJSON, newJSON []byte
		if err := rows.Scan(&r.Action, &oldJSON, &newJSON); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(oldJSON, &r.Old); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(newJSON, &r.New); err != nil {
			t.Fatal(err)
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

// TestAudit_TC_M01_025 memastikan audit_log hanya menyimpan kolom yang berubah, tidak menulis
// baris untuk update tanpa perubahan atau perubahan turunan (s chain, content_version), dan
// menyimpan baris lengkap untuk insert dan delete.
func TestAudit_TC_M01_025(t *testing.T) {
	ctx := context.Background()
	db, st := newStore(t)
	pool := db.Pool
	step := stepID(t, pool, ps07, "50")
	actor := store.Actor{UserID: apratama, RequestID: "req-audit", Source: store.SourceAPI}
	// in menjalankan satu statement lewat WithTx dengan actor audit di atas.
	in := func(sql string, args ...any) {
		t.Helper()
		err := st.WithTx(ctx, actor, func(ctx context.Context, tx pgx.Tx) error {
			_, err := tx.Exec(ctx, sql, args...)
			return err
		})
		if err != nil {
			t.Fatalf("exec %q: %v", sql, err)
		}
	}
	// mark mengembalikan id audit terakhir sebagai batas baris audit berikutnya.
	mark := func() int64 { return scalar[int64](t, pool, "SELECT coalesce(max(id), 0) FROM audit_log") }

	m := mark()
	in("UPDATE process_steps SET name = 'Reflow (audit)' WHERE id = $1", step)
	rows := auditRows(t, pool, "process_steps", m)
	if len(rows) != 1 || rows[0].Action != "update" ||
		!slices.Equal(slices.Sorted(maps.Keys(rows[0].Old)), []string{"name"}) ||
		!slices.Equal(slices.Sorted(maps.Keys(rows[0].New)), []string{"name"}) ||
		rows[0].New["name"] != "Reflow (audit)" {
		t.Errorf("update audit = %+v, want one row with only the name column", rows)
	}

	m = mark()
	in("UPDATE process_steps SET name = name WHERE id = $1", step)
	if rows := auditRows(t, pool, "process_steps", m); len(rows) != 0 {
		t.Errorf("no-op update wrote audit rows: %+v", rows)
	}

	m = mark()
	in("UPDATE failure_effects SET s = 8 WHERE package_id = $1 AND text = 'Board fails function test; rework'", ps07)
	if rows := auditRows(t, pool, "failure_effects", m); len(rows) != 1 {
		t.Errorf("effect audit rows = %d, want 1", len(rows))
	}
	for _, table := range []string{"failure_chains", "packages"} {
		if rows := auditRows(t, pool, table, m); len(rows) != 0 {
			t.Errorf("derived change wrote %s audit rows: %+v", table, rows)
		}
	}

	m = mark()
	in("INSERT INTO process_steps (package_id, op_no, seq, name) VALUES ($1, '95', 950, 'Conformal coating')", ps07)
	in("DELETE FROM process_steps WHERE package_id = $1 AND op_no = '95'", ps07)
	rows = auditRows(t, pool, "process_steps", m)
	if len(rows) != 2 || rows[0].Action != "insert" || rows[0].New["name"] != "Conformal coating" || rows[0].New["op_no"] != "95" ||
		rows[1].Action != "delete" || rows[1].Old["name"] != "Conformal coating" || len(rows[1].New) != 0 {
		t.Errorf("insert/delete audit = %+v, want full rows", rows)
	}
}

// TestRestrict_TC_M01_026 memastikan step yang punya failure mode dan karakteristik yang dipakai
// baris CP tidak bisa dihapus (RESTRICT), sedangkan step bebas terhapus beserta turunannya.
func TestRestrict_TC_M01_026(t *testing.T) {
	db, _ := newStore(t)
	pool := db.Pool
	step50 := stepID(t, pool, ps07, "50")
	// count menjalankan query hitung.
	count := func(sql string, args ...any) int { return scalar[int](t, pool, sql, args...) }
	fmBefore := count("SELECT count(*) FROM failure_modes WHERE step_id = $1", step50)

	// ON DELETE RESTRICT dilaporkan PostgreSQL sebagai 23001 (restrict_violation), bukan 23503.
	if code := sqlState(t, pool, "DELETE FROM process_steps WHERE id = $1", step50); code != "23001" {
		t.Errorf("delete step 50: SQLSTATE %s, want 23001", code)
	}
	if count("SELECT count(*) FROM process_steps WHERE id = $1", step50) != 1 ||
		count("SELECT count(*) FROM failure_modes WHERE step_id = $1", step50) != fmBefore {
		t.Errorf("rows were deleted although the delete failed")
	}
	if code := sqlState(t, pool,
		"DELETE FROM characteristics WHERE package_id = $1 AND char_no = '60-01'", ps07); code != "23001" {
		t.Errorf("delete characteristic used by a CP line: SQLSTATE %s, want 23001", code)
	}

	free := scalar[uuid.UUID](t, pool,
		"INSERT INTO process_steps (package_id, op_no, seq, name, symbol) VALUES ($1, '95', 950, 'Conformal coating', 'operation') RETURNING id", ps07)
	exec(t, pool, "INSERT INTO characteristics (package_id, step_id, char_no, kind, name) VALUES ($1, $2, '95-01', 'product', 'Coating thickness')", ps07, free)
	exec(t, pool, "INSERT INTO step_flows (package_id, from_step_id, kind, disposition) VALUES ($1, $2, 'scrap', 'Scrap')", ps07, free)
	exec(t, pool, "DELETE FROM process_steps WHERE id = $1", free)
	if n := count("SELECT count(*) FROM characteristics WHERE step_id = $1", free) +
		count("SELECT count(*) FROM step_flows WHERE from_step_id = $1", free); n != 0 {
		t.Errorf("%d child rows left after deleting a free step", n)
	}
}

// TestCrossPackage_TC_M01_027 memastikan foreign key gabungan (package_id, x_id) menolak
// referensi ke baris paket lain.
func TestCrossPackage_TC_M01_027(t *testing.T) {
	db, _ := newStore(t)
	pool := db.Pool
	generalStep := stepID(t, pool, generalPkg, "10")
	generalChar := scalar[uuid.UUID](t, pool,
		"SELECT id FROM characteristics WHERE package_id = $1 ORDER BY char_no LIMIT 1", generalPkg)
	ps07Step := stepID(t, pool, ps07, "50")

	if code := sqlState(t, pool,
		"INSERT INTO characteristics (package_id, step_id, char_no, kind, name) VALUES ($1, $2, '50-09', 'product', 'Cross')", ps07, generalStep); code != "23503" {
		t.Errorf("characteristic referencing a GENERAL step: SQLSTATE %s, want 23503", code)
	}
	if code := sqlState(t, pool,
		"INSERT INTO failure_modes (package_id, step_id, characteristic_id, text) VALUES ($1, $2, $3, 'Cross')", ps07, ps07Step, generalChar); code != "23503" {
		t.Errorf("failure mode referencing a GENERAL characteristic: SQLSTATE %s, want 23503", code)
	}
}

// TestSnakeToCamel_TC_M01_028 memastikan snake_to_camel mengubah nama kolom database menjadi
// nama field API.
func TestSnakeToCamel_TC_M01_028(t *testing.T) {
	db, _ := newStore(t)
	cases := map[string]string{
		"sample_freq":    "sampleFreq",
		"ep_verify_freq": "epVerifyFreq",
		"d":              "d",
		"package_id":     "packageId",
	}
	for in, want := range cases {
		if got := scalar[string](t, db.Pool, "SELECT snake_to_camel($1)", in); got != want {
			t.Errorf("snake_to_camel(%q) = %q, want %q", in, got, want)
		}
	}
	if got := scalar[*string](t, db.Pool, "SELECT snake_to_camel(NULL)"); got != nil {
		t.Errorf("snake_to_camel(NULL) = %q, want NULL", *got)
	}
}
