// Test untuk pembagian rute server: path /api yang tidak dikenal tidak boleh jatuh ke SPA
// (test case TC-M00-011, docs/test-cases/M00-scaffold.md).

package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// TestRouter_TC_M00_011 memastikan /api dan semua path di bawahnya yang belum punya rute dijawab
// 404 Problem Details (kode not_found, docs/05-api.md §2), bukan index.html dan bukan redirect,
// sementara path non-API tetap diteruskan ke SPA.
func TestRouter_TC_M00_011(t *testing.T) {
	h := newTestHandler(nil, 0)

	cases := []struct{ method, target string }{
		{http.MethodGet, "/api"},
		{http.MethodGet, "/api/"},
		{http.MethodGet, "/api/v1/nothing"},
		{http.MethodPost, "/api/v1/packages"},
	}
	for _, c := range cases {
		t.Run(c.method+" "+c.target, func(t *testing.T) {
			rec := serve(h, c.method, c.target)
			if rec.Code != http.StatusNotFound {
				t.Fatalf("status = %d, ingin 404 (Location: %q)", rec.Code, rec.Header().Get("Location"))
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
				t.Errorf("Content-Type = %q", ct)
			}
			if strings.Contains(rec.Body.String(), "INDEX") {
				t.Errorf("path API dijawab dengan SPA: %s", rec.Body.String())
			}
			var p map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
				t.Fatalf("body bukan JSON: %v", err)
			}
			want := map[string]any{
				"type":   "urn:pfmea:problem:not_found",
				"title":  "Not found",
				"status": float64(404),
				"code":   "not_found",
			}
			for k, v := range want {
				if p[k] != v {
					t.Errorf("%s = %v, ingin %v", k, p[k], v)
				}
			}
		})
	}

	// Path non-API tetap sampai ke SPA (index.html untuk deep link).
	if rec := serve(h, http.MethodGet, "/packages/PS-07/pfmea"); rec.Body.String() != "INDEX" {
		t.Errorf("path non-API tidak diteruskan ke SPA: %d %s", rec.Code, rec.Body.String())
	}
}
