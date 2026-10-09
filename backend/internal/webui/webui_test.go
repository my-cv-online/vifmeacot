// Test untuk penyajian SPA yang tertanam di binary (test case TC-M00-009, TC-M00-010 dan
// TC-M00-012, docs/test-cases/M00-scaffold.md).

package webui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

// indexHTML adalah isi index.html tiruan hasil build SvelteKit.
const indexHTML = "<!doctype html><html><body>INDEX</body></html>"

// builtFS meniru isi dist/ setelah `make build`: index.html, satu file statis biasa, satu aset
// ber-hash di _app/immutable dan penanda .keep.
func builtFS() fstest.MapFS {
	return fstest.MapFS{
		"index.html":                           {Data: []byte(indexHTML)},
		"robots.txt":                           {Data: []byte("User-agent: *\n")},
		"_app/immutable/entry/start.abc123.js": {Data: []byte("export const x = 1;\n")},
		".keep":                                {Data: nil},
	}
}

// get menjalankan satu request terhadap handler dan mengembalikan rekamannya.
func get(h http.Handler, method, target string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, target, nil))
	return rec
}

// TestHandler_TC_M00_009 memastikan binary yang dibangun tanpa hasil build web (dist/ hanya
// berisi .keep) menjawab "UI not built yet" untuk setiap halaman, bukan halaman kosong.
func TestHandler_TC_M00_009(t *testing.T) {
	h := Handler(fstest.MapFS{".keep": {Data: nil}})
	for _, target := range []string{"/", "/packages"} {
		rec := get(h, http.MethodGet, target)
		if rec.Code != http.StatusServiceUnavailable {
			t.Errorf("%s: status = %d, ingin 503", target, rec.Code)
		}
		if ct := rec.Header().Get("Content-Type"); ct != "text/plain; charset=utf-8" {
			t.Errorf("%s: Content-Type = %q", target, ct)
		}
		if !strings.Contains(rec.Body.String(), "UI not built yet") {
			t.Errorf("%s: body = %q", target, rec.Body.String())
		}
	}
}

// TestHandler_TC_M00_010 memastikan deep link SPA (path non-API yang tidak ada filenya) dijawab
// dengan index.html tanpa cache, /index.html tidak di-redirect, file statis yang ada disajikan
// apa adanya, dan method selain GET/HEAD ditolak.
func TestHandler_TC_M00_010(t *testing.T) {
	h := Handler(builtFS())

	for _, target := range []string{"/", "/packages/PS-07/pfmea", "/findings?rule=R04", "/index.html"} {
		rec := get(h, http.MethodGet, target)
		if rec.Code != http.StatusOK {
			t.Errorf("%s: status = %d, ingin 200 (Location: %q)", target, rec.Code, rec.Header().Get("Location"))
			continue
		}
		if rec.Body.String() != indexHTML {
			t.Errorf("%s: body = %q, ingin index.html", target, rec.Body.String())
		}
		if ct := rec.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
			t.Errorf("%s: Content-Type = %q", target, ct)
		}
		if cc := rec.Header().Get("Cache-Control"); cc != "no-cache" {
			t.Errorf("%s: Cache-Control = %q, ingin no-cache", target, cc)
		}
	}

	// HEAD untuk deep link juga dijawab (tanpa body).
	if rec := get(h, http.MethodHead, "/packages"); rec.Code != http.StatusOK {
		t.Errorf("HEAD /packages: status = %d", rec.Code)
	}

	// File statis yang ada disajikan apa adanya, tidak diganti index.html.
	rec := get(h, http.MethodGet, "/robots.txt")
	if rec.Code != http.StatusOK || rec.Body.String() != "User-agent: *\n" {
		t.Errorf("/robots.txt: %d %q", rec.Code, rec.Body.String())
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "no-cache" {
		t.Errorf("/robots.txt: Cache-Control = %q", cc)
	}

	// SPA hanya dibaca; method lain ditolak dengan 405 dan header Allow.
	rec = get(h, http.MethodPost, "/packages")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST /packages: status = %d, ingin 405", rec.Code)
	}
	if allow := rec.Header().Get("Allow"); allow != "GET, HEAD" {
		t.Errorf("POST /packages: Allow = %q", allow)
	}
}

// TestHandler_TC_M00_012 memastikan aset ber-hash mendapat cache permanen, aset atau folder yang
// tidak ada di bawah /_app/ dijawab 404 (bukan index.html yang akan merusak MIME type), dan
// dotfile seperti .keep tidak pernah disajikan sebagai file.
func TestHandler_TC_M00_012(t *testing.T) {
	h := Handler(builtFS())

	rec := get(h, http.MethodGet, "/_app/immutable/entry/start.abc123.js")
	if rec.Code != http.StatusOK || rec.Body.String() != "export const x = 1;\n" {
		t.Fatalf("aset: %d %q", rec.Code, rec.Body.String())
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "public, max-age=31536000, immutable" {
		t.Errorf("aset: Cache-Control = %q", cc)
	}
	// Tipe MIME .js bisa berbeda antar mesin (/etc/mime.types), cukup pastikan javascript.
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "javascript") {
		t.Errorf("aset: Content-Type = %q", ct)
	}

	for _, target := range []string{"/_app/immutable/entry/missing.js", "/_app/immutable/", "/_app/immutable/entry"} {
		rec := get(h, http.MethodGet, target)
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s: status = %d, ingin 404", target, rec.Code)
		}
		if strings.Contains(rec.Body.String(), "INDEX") {
			t.Errorf("%s: dijawab dengan index.html", target)
		}
		// Teks 404 berasal dari i18n (bahasa Inggris), bukan teks bawaan net/http.
		if got := strings.TrimSpace(rec.Body.String()); got != "Not found" {
			t.Errorf("%s: body = %q, ingin \"Not found\"", target, got)
		}
	}

	// Dotfile diperlakukan seperti path yang tidak dikenal: SPA, bukan isi file.
	rec = get(h, http.MethodGet, "/.keep")
	if rec.Code != http.StatusOK || rec.Body.String() != indexHTML {
		t.Errorf("/.keep: %d %q, ingin index.html", rec.Code, rec.Body.String())
	}
}
