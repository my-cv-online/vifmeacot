// Parser fixture aturan dan pencocokan hasil aturan secara multiset (docs/06-rules.md §7),
// beserta test-nya (test case TC-M01-029, docs/test-cases/M01-database.md).

package rules

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// Awalan baris header fixture.
const (
	// fixturePrefix membuka baris pertama fixture: "-- fixture: <CODE> di atas …".
	fixturePrefix = "-- fixture:"
	// scenarioPrefix adalah baris keterangan skenario (bahasa Indonesia).
	scenarioPrefix = "-- scenario:"
	// expectPrefix adalah baris JSON berisi seluruh hasil aturan yang diharapkan.
	expectPrefix = "-- expect:"
)

// expectItem adalah satu hasil yang diharapkan di baris `-- expect:`.
type expectItem struct {
	// ObjectType adalah nama tabel objek temuan.
	ObjectType string `json:"object_type"`
	// Field adalah nama field API (camelCase) atau "" untuk seluruh baris.
	Field string `json:"field"`
	// Params berisi kunci dan nilai yang harus ada di params hasil (boleh sebagian).
	Params map[string]any `json:"params"`
}

// fixture adalah isi satu file testdata/<CODE>.sql.
type fixture struct {
	// Code adalah kode aturan dari header `-- fixture:`.
	Code string
	// Scenario adalah keterangan skenario.
	Scenario string
	// Expect adalah seluruh hasil aturan yang diharapkan setelah fixture dijalankan.
	Expect []expectItem
	// SQL adalah isi file lengkap; baris komentar ikut terkirim dan diabaikan PostgreSQL.
	SQL string
}

// result adalah satu baris hasil query aturan (kolom sesuai docs/06-rules.md §1).
type result struct {
	// ObjectType adalah nama tabel objek temuan.
	ObjectType string
	// ObjectID adalah id objek.
	ObjectID string
	// Field adalah nama field API.
	Field string
	// Key membedakan beberapa temuan pada objek dan field yang sama.
	Key string
	// Params adalah isi kolom params yang sudah di-decode dari JSON.
	Params map[string]any
}

// String menampilkan hasil secara ringkas untuk pesan test.
func (r result) String() string {
	b, _ := json.Marshal(r.Params)
	return fmt.Sprintf("%s.%s key=%q params=%s", r.ObjectType, r.Field, r.Key, b)
}

// parseFixture membaca header fixture. Header `-- fixture:` dan `-- expect:` wajib ada tepat
// satu kali; JSON expect harus berupa array (boleh kosong).
func parseFixture(src string) (fixture, error) {
	f := fixture{SQL: src}
	var expects int
	for i, line := range strings.Split(src, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case i == 0 && strings.HasPrefix(line, fixturePrefix):
			fields := strings.Fields(strings.TrimPrefix(line, fixturePrefix))
			if len(fields) == 0 {
				return fixture{}, errors.New("fixture header has no rule code")
			}
			f.Code = fields[0]
		case strings.HasPrefix(line, scenarioPrefix):
			f.Scenario = strings.TrimSpace(strings.TrimPrefix(line, scenarioPrefix))
		case strings.HasPrefix(line, expectPrefix):
			expects++
			raw := strings.TrimSpace(strings.TrimPrefix(line, expectPrefix))
			dec := json.NewDecoder(strings.NewReader(raw))
			dec.DisallowUnknownFields()
			if err := dec.Decode(&f.Expect); err != nil {
				return fixture{}, fmt.Errorf("expect line is not a JSON array of results: %w", err)
			}
			if f.Expect == nil {
				return fixture{}, errors.New("expect line must be a JSON array, use [] for no results")
			}
		}
	}
	switch {
	case f.Code == "":
		return fixture{}, errors.New("first line must be \"-- fixture: <CODE> …\"")
	case expects == 0:
		return fixture{}, errors.New("missing \"-- expect:\" line")
	case expects > 1:
		return fixture{}, errors.New("more than one \"-- expect:\" line")
	}
	return f, nil
}

// matches melaporkan apakah hasil r memenuhi item yang diharapkan e: object_type dan field sama,
// dan setiap kunci params yang diharapkan ada dengan nilai yang sama.
func (e expectItem) matches(r result) bool {
	if e.ObjectType != r.ObjectType || e.Field != r.Field {
		return false
	}
	for k, want := range e.Params {
		got, ok := r.Params[k]
		if !ok || !reflect.DeepEqual(normalizeJSON(want), normalizeJSON(got)) {
			return false
		}
	}
	return true
}

// normalizeJSON menyamakan representasi nilai JSON (angka selalu float64) dengan encode lalu
// decode ulang, supaya nilai dari fixture dan dari database bisa dibandingkan langsung.
func normalizeJSON(v any) any {
	b, err := json.Marshal(v)
	if err != nil {
		return v
	}
	var out any
	if err := json.Unmarshal(b, &out); err != nil {
		return v
	}
	return out
}

// matchMultiset mencocokkan hasil aturan dengan item yang diharapkan sebagai multiset: jumlah
// harus sama dan setiap item harus dipasangkan dengan hasil yang berbeda. Pasangan dicari
// dengan augmenting path (bipartite matching), jadi tidak tertipu urutan seperti pencocokan
// rakus. Error menjelaskan item yang tidak mendapat pasangan.
func matchMultiset(want []expectItem, got []result) error {
	if len(want) != len(got) {
		return fmt.Errorf("got %d results, want %d\n%s", len(got), len(want), describe(want, got))
	}
	// owner[j] adalah indeks item yang sedang memakai hasil j, -1 bila belum dipakai.
	owner := make([]int, len(got))
	for j := range owner {
		owner[j] = -1
	}
	// augment mencoba memberi item i satu hasil, memindahkan item lain bila perlu.
	var augment func(i int, seen []bool) bool
	augment = func(i int, seen []bool) bool {
		for j, r := range got {
			if seen[j] || !want[i].matches(r) {
				continue
			}
			seen[j] = true
			if owner[j] < 0 || augment(owner[j], seen) {
				owner[j] = i
				return true
			}
		}
		return false
	}
	for i := range want {
		if !augment(i, make([]bool, len(got))) {
			b, _ := json.Marshal(want[i])
			return fmt.Errorf("no distinct result matches expected item %s\n%s", b, describe(want, got))
		}
	}
	return nil
}

// describe menyusun daftar item yang diharapkan dan hasil yang didapat untuk pesan test.
func describe(want []expectItem, got []result) string {
	var b strings.Builder
	b.WriteString("want:\n")
	for _, w := range want {
		j, _ := json.Marshal(w)
		fmt.Fprintf(&b, "  %s\n", j)
	}
	b.WriteString("got:\n")
	for _, g := range got {
		fmt.Fprintf(&b, "  %s\n", g)
	}
	return b.String()
}

// TestFixtureMatcher_TC_M01_029 memastikan parser fixture menolak fixture yang salah dengan
// pesan jelas, dan pencocokan multiset menerima/menolak sesuai docs/06-rules.md §7, termasuk
// kasus yang gagal bila dicocokkan secara rakus tetapi benar dengan backtracking.
func TestFixtureMatcher_TC_M01_029(t *testing.T) {
	t.Run("parse", func(t *testing.T) {
		bad := map[string]string{
			"no header":      "-- expect: []\nSELECT 1;",
			"no expect":      "-- fixture: K01 di atas demo\n-- scenario: x\nSELECT 1;",
			"invalid JSON":   "-- fixture: K01 di atas demo\n-- expect: [{\"object_type\": }]\n",
			"not an array":   "-- fixture: K01 di atas demo\n-- expect: {\"object_type\": \"x\"}\n",
			"null":           "-- fixture: K01 di atas demo\n-- expect: null\n",
			"unknown key":    "-- fixture: K01 di atas demo\n-- expect: [{\"objectType\": \"x\"}]\n",
			"two expects":    "-- fixture: K01 di atas demo\n-- expect: []\n-- expect: []\n",
			"header no code": "-- fixture:\n-- expect: []\n",
		}
		for name, src := range bad {
			if _, err := parseFixture(src); err == nil {
				t.Errorf("%s: parseFixture should fail", name)
			}
		}
		f, err := parseFixture("-- fixture: K01 di atas db/seed/demo.sql\n-- scenario: Step baru.\n" +
			"-- expect: [{\"object_type\": \"process_steps\", \"field\": \"name\", \"params\": {\"opNo\": \"95\"}}]\nSELECT 1;")
		if err != nil {
			t.Fatalf("valid fixture: %v", err)
		}
		if f.Code != "K01" || f.Scenario != "Step baru." || len(f.Expect) != 1 ||
			f.Expect[0].ObjectType != "process_steps" || f.Expect[0].Params["opNo"] != "95" {
			t.Errorf("parsed fixture = %+v", f)
		}
		empty, err := parseFixture("-- fixture: W04 di atas demo\n-- expect: []\n")
		if err != nil || empty.Expect == nil || len(empty.Expect) != 0 {
			t.Errorf("empty expect: %+v, %v", empty, err)
		}
	})

	step := func(opNo string, extra map[string]any) result {
		p := map[string]any{"opNo": opNo}
		for k, v := range extra {
			p[k] = v
		}
		return result{ObjectType: "process_steps", Field: "name", Params: p}
	}
	item := func(params map[string]any) expectItem {
		return expectItem{ObjectType: "process_steps", Field: "name", Params: params}
	}
	cases := []struct {
		name string
		want []expectItem
		got  []result
		ok   bool
	}{
		{"empty", []expectItem{}, nil, true},
		{"count differs", []expectItem{item(nil)}, []result{step("10", nil), step("20", nil)}, false},
		{"params subset", []expectItem{item(map[string]any{"opNo": "10"})}, []result{step("10", map[string]any{"name": "x"})}, true},
		{"param value differs", []expectItem{item(map[string]any{"opNo": "10"})}, []result{step("20", nil)}, false},
		{"param missing", []expectItem{item(map[string]any{"daysLate": 13})}, []result{step("10", nil)}, false},
		{"number types", []expectItem{item(map[string]any{"daysLate": 13})}, []result{step("10", map[string]any{"daysLate": 13.0})}, true},
		{"field differs", []expectItem{{ObjectType: "process_steps", Field: "ngFlow"}}, []result{step("10", nil)}, false},
		{"duplicate items need distinct results", []expectItem{item(map[string]any{"opNo": "10"}), item(map[string]any{"opNo": "10"})},
			[]result{step("10", nil), step("20", nil)}, false},
		{"duplicates matched", []expectItem{item(map[string]any{"opNo": "10"}), item(map[string]any{"opNo": "10"})},
			[]result{step("10", nil), step("10", nil)}, true},
		// Item pertama cocok dengan kedua hasil; rakus memberinya hasil pertama sehingga item
		// kedua (yang hanya cocok dengan hasil pertama) kehabisan pasangan.
		{"needs backtracking", []expectItem{item(nil), item(map[string]any{"opNo": "10"})},
			[]result{step("10", nil), step("20", nil)}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := matchMultiset(c.want, c.got)
			if (err == nil) != c.ok {
				t.Errorf("matchMultiset error = %v, want ok = %v", err, c.ok)
			}
		})
	}
}
