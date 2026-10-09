// Test file aturan Tahap 1: kelengkapan file (TC-M01-030), fixture setiap aturan
// (TC-RULE-<CODE>-1) dan baseline data demo (TC-RULE-BASE-1/2), lihat docs/06-rules.md §5 dan §7
// serta docs/test-cases/rules.md. `make test` melewati TestRuleFixtures dan TestBaseline;
// keduanya dijalankan `make test-rules`.

package rules

import (
	"context"
	"encoding/json"
	"io/fs"
	"maps"
	"os"
	"path"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"pfmea/backend/internal/testdb"
)

// phase1Codes adalah 32 kode aturan Tahap 1 sesuai docs/06-rules.md §4.
var phase1Codes = []string{
	"C01", "F01", "F02", "F03", "F07", "F08", "F09", "F10", "F11", "F12",
	"K01", "K02", "K03", "K05", "K06", "R01", "R02", "R03", "R04", "R05", "R06",
	"S01", "S02", "S03", "S04", "T01", "T02", "T03", "T04", "W01", "W02", "W04",
}

// Paket demo dan tanggal "hari ini" demo (db/seed/demo.sql, CLAUDE.md).
var (
	// ps07 adalah paket model PS-07.
	ps07 = uuid.MustParse("3c91e2ed-8941-5f73-a84f-8be64945878f")
	// general adalah paket Template General.
	general = uuid.MustParse("07ea2e55-0ce4-5e2b-b9b0-681883ae41ce")
	// demoToday adalah @today untuk semua fixture dan baseline.
	demoToday = time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC)
)

// ruleHeaders adalah awalan empat baris pertama file aturan (docs/06-rules.md §1).
var ruleHeaders = []string{"-- rule: ", "-- reads: ", "-- message: ", "-- returns: "}

// TestMain menjalankan test paket lalu mematikan container database test.
func TestMain(m *testing.M) { testdb.Main(m) }

// ruleSQL membaca file aturan sql/<code>.sql dari file yang tertanam.
func ruleSQL(t *testing.T, code string) string {
	t.Helper()
	b, err := fs.ReadFile(Files, "sql/"+code+".sql")
	if err != nil {
		t.Fatalf("read rule %s: %v", code, err)
	}
	return string(b)
}

// codesIn mengembalikan kode (nama file tanpa .sql) dari daftar file .sql, terurut.
func codesIn(entries []fs.DirEntry) []string {
	var codes []string
	for _, e := range entries {
		if !e.IsDir() && path.Ext(e.Name()) == ".sql" {
			codes = append(codes, strings.TrimSuffix(e.Name(), ".sql"))
		}
	}
	slices.Sort(codes)
	return codes
}

// runRule menjalankan satu query aturan di tx dengan @package_id dan @today, lalu membaca
// semua hasilnya.
func runRule(ctx context.Context, tx pgx.Tx, src string, pkg uuid.UUID) ([]result, error) {
	rows, err := tx.Query(ctx, src, pgx.NamedArgs{"package_id": pkg, "today": demoToday})
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []result
	for rows.Next() {
		var r result
		var id uuid.UUID
		var params []byte
		if err := rows.Scan(&r.ObjectType, &id, &r.Field, &r.Key, &params); err != nil {
			return nil, err
		}
		r.ObjectID = id.String()
		if err := json.Unmarshal(params, &r.Params); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// TestRuleFiles_TC_M01_030 memastikan setiap aturan Tahap 1 punya tepat satu file SQL dan satu
// fixture dengan nama sama, header file aturan lengkap dengan kode yang sama dengan nama file,
// dan header fixture menyebut kode yang sama.
func TestRuleFiles_TC_M01_030(t *testing.T) {
	sqlEntries, err := fs.ReadDir(Files, "sql")
	if err != nil {
		t.Fatal(err)
	}
	if got := codesIn(sqlEntries); !slices.Equal(got, phase1Codes) {
		t.Errorf("rule files = %v, want the 32 phase 1 codes %v", got, phase1Codes)
	}
	fixtureEntries, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}
	if got := codesIn(fixtureEntries); !slices.Equal(got, phase1Codes) {
		t.Errorf("fixtures = %v, want the 32 phase 1 codes %v", got, phase1Codes)
	}

	for _, code := range phase1Codes {
		t.Run(code, func(t *testing.T) {
			lines := strings.SplitN(ruleSQL(t, code), "\n", len(ruleHeaders)+1)
			for i, prefix := range ruleHeaders {
				if i >= len(lines) || !strings.HasPrefix(lines[i], prefix) {
					t.Errorf("line %d of sql/%s.sql must start with %q", i+1, code, prefix)
				}
			}
			if lines[0] != "-- rule: "+code {
				t.Errorf("sql/%s.sql header %q, want \"-- rule: %s\"", code, lines[0], code)
			}
			src, err := os.ReadFile(path.Join("testdata", code+".sql"))
			if err != nil {
				t.Fatal(err)
			}
			f, err := parseFixture(string(src))
			if err != nil {
				t.Fatalf("testdata/%s.sql: %v", code, err)
			}
			if f.Code != code {
				t.Errorf("testdata/%s.sql is a fixture for %q", code, f.Code)
			}
		})
	}
}

// TestRuleFixtures menjalankan fixture setiap aturan di atas data demo dalam transaksi yang
// di-rollback dan membandingkan seluruh hasil aturan untuk PS-07 dengan baris `-- expect:`
// (subtest TC-RULE-<CODE>-1, docs/test-cases/rules.md).
func TestRuleFixtures(t *testing.T) {
	ctx := context.Background()
	d := testdb.New(t)
	for _, code := range phase1Codes {
		t.Run("TC-RULE-"+code+"-1", func(t *testing.T) {
			src, err := os.ReadFile(path.Join("testdata", code+".sql"))
			if err != nil {
				t.Fatal(err)
			}
			f, err := parseFixture(string(src))
			if err != nil {
				t.Fatalf("testdata/%s.sql: %v", code, err)
			}
			tx, err := d.Pool.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			// Tanpa argumen pgx memakai simple protocol, jadi fixture boleh berisi banyak statement.
			if _, err := tx.Exec(ctx, f.SQL); err != nil {
				t.Fatalf("fixture %s: %v", code, err)
			}
			got, err := runRule(ctx, tx, ruleSQL(t, code), ps07)
			if err != nil {
				t.Fatalf("rule %s: %v", code, err)
			}
			if err := matchMultiset(f.Expect, got); err != nil {
				t.Errorf("rule %s (%s): %v", code, f.Scenario, err)
			}
		})
	}
}

// baselineFinding adalah satu temuan baseline PS-07 di docs/06-rules.md §5.
type baselineFinding struct {
	// Rule adalah kode aturan.
	Rule string
	// ObjectType dan Field adalah objek dan field temuan.
	ObjectType, Field string
}

// ps07Baseline adalah 14 temuan baseline PS-07 (docs/06-rules.md §5). Pesan yang dirender
// dibandingkan mulai M8, saat mesin aturan merender template pesan.
var ps07Baseline = []baselineFinding{
	{"F11", "actions", "targetDate"},
	{"F12", "failure_causes", "text"},
	{"K02", "characteristics", "name"},
	{"K05", "cp_lines", "machines"},
	{"K06", "documents", "header"},
	{"R01", "controls", "text"},
	{"R02", "controls", "text"},
	{"R03", "cp_lines", "controlId"},
	{"R04", "failure_chains", "d"},
	{"R06", "cp_lines", "epVerifyFreq"},
	{"S02", "characteristics", "scSymbolId"},
	{"S04", "characteristics", "scSymbolId"},
	{"W01", "process_steps", "name"},
	{"W04", "packages", "lastReviewedAt"},
}

// runAll menjalankan semua aturan Tahap 1 untuk satu paket dalam satu transaksi baca dan
// mengembalikan temuannya (kode aturan, objek, field) terurut.
func runAll(t *testing.T, d *testdb.DB, pkg uuid.UUID) []baselineFinding {
	t.Helper()
	ctx := context.Background()
	tx, err := d.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var out []baselineFinding
	for _, code := range phase1Codes {
		results, err := runRule(ctx, tx, ruleSQL(t, code), pkg)
		if err != nil {
			t.Fatalf("rule %s: %v", code, err)
		}
		for _, r := range results {
			out = append(out, baselineFinding{code, r.ObjectType, r.Field})
		}
	}
	slices.SortFunc(out, func(a, b baselineFinding) int {
		return strings.Compare(a.Rule+a.ObjectType+a.Field, b.Rule+b.ObjectType+b.Field)
	})
	return out
}

// TestBaseline memastikan baseline data demo (docs/06-rules.md §5): PS-07 menghasilkan tepat
// 14 temuan (5 error, 7 warning, 2 info) dan GENERAL tidak menghasilkan temuan
// (subtest TC-RULE-BASE-1 dan TC-RULE-BASE-2, docs/test-cases/rules.md).
func TestBaseline(t *testing.T) {
	d := testdb.New(t)

	t.Run("TC-RULE-BASE-1", func(t *testing.T) {
		got := runAll(t, d, ps07)
		if !slices.Equal(got, ps07Baseline) {
			t.Errorf("PS-07 findings:\n got %v\nwant %v", got, ps07Baseline)
		}
		// Level bawaan setiap aturan berasal dari katalog rules yang di-seed migrasi.
		levels := map[string]string{}
		rows, err := d.Pool.Query(context.Background(), "SELECT code, default_level::text FROM rules WHERE phase = 1")
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			var code, level string
			if err := rows.Scan(&code, &level); err != nil {
				t.Fatal(err)
			}
			levels[code] = level
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		if codes := slices.Sorted(maps.Keys(levels)); !slices.Equal(codes, phase1Codes) {
			t.Errorf("phase 1 rules in the catalog = %v, want %v", codes, phase1Codes)
		}
		count := map[string]int{}
		for _, f := range got {
			count[levels[f.Rule]]++
		}
		if want := map[string]int{"error": 5, "warning": 7, "info": 2}; !maps.Equal(count, want) {
			t.Errorf("PS-07 findings per level = %v, want %v", count, want)
		}
	})

	t.Run("TC-RULE-BASE-2", func(t *testing.T) {
		if got := runAll(t, d, general); len(got) != 0 {
			t.Errorf("GENERAL findings = %v, want none", got)
		}
	})
}
