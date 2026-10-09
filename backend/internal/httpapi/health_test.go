// Test untuk /healthz dan /readyz (test case TC-M00-001 sampai TC-M00-004,
// docs/test-cases/M00-scaffold.md).
package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// readyBody adalah bentuk JSON jawaban /readyz yang diperiksa test.
type readyBody struct {
	Status string `json:"status"`
	Checks []struct {
		Name   string `json:"name"`
		Status string `json:"status"`
	} `json:"checks"`
}

// newTestHandler membuat handler server dengan logger yang dibuang dan UI tiruan.
func newTestHandler(checks []ReadinessCheck, timeout time.Duration) http.Handler {
	return NewHandler(Options{
		Logger:           slog.New(slog.NewTextHandler(io.Discard, nil)),
		Readiness:        checks,
		ReadinessTimeout: timeout,
		UI: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = io.WriteString(w, "INDEX")
		}),
	})
}

// serve menjalankan satu request terhadap handler dan mengembalikan rekamannya.
func serve(h http.Handler, method, target string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, target, nil))
	return rec
}

// decodeReady membaca body /readyz; test gagal bila bukan JSON yang benar.
func decodeReady(t *testing.T, rec *httptest.ResponseRecorder) readyBody {
	t.Helper()
	var body readyBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body /readyz bukan JSON: %v\n%s", err, rec.Body.String())
	}
	return body
}

// TestHealthz_TC_M00_001 memastikan /healthz menjawab 200 untuk GET dan HEAD selama proses
// berjalan, tanpa pemeriksaan apa pun dan tanpa boleh di-cache.
func TestHealthz_TC_M00_001(t *testing.T) {
	h := newTestHandler(nil, 0)

	rec := serve(h, http.MethodGet, "/healthz")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin 200", rec.Code)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"status":"ok"}` {
		t.Errorf("body = %s", got)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q", ct)
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
		t.Errorf("Cache-Control = %q", cc)
	}

	if rec := serve(h, http.MethodHead, "/healthz"); rec.Code != http.StatusOK {
		t.Errorf("HEAD status = %d, ingin 200", rec.Code)
	}
}

// TestReadyz_TC_M00_002 memastikan /readyz menjawab 200 dengan daftar pemeriksaan kosong bila
// tidak ada pemeriksaan terdaftar (kondisi M0; M1 menambahkan pemeriksaan database).
func TestReadyz_TC_M00_002(t *testing.T) {
	rec := serve(newTestHandler(nil, 0), http.MethodGet, "/readyz")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin 200", rec.Code)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"status":"ready","checks":[]}` {
		t.Errorf("body = %s", got)
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
		t.Errorf("Cache-Control = %q", cc)
	}
}

// TestReadyz_TC_M00_003 memastikan /readyz menjawab 200 dan melaporkan setiap pemeriksaan
// sebagai ok (urutan pendaftaran) bila semuanya lulus.
func TestReadyz_TC_M00_003(t *testing.T) {
	pass := func(context.Context) error { return nil }
	rec := serve(newTestHandler([]ReadinessCheck{
		{Name: "alpha", Check: pass},
		{Name: "beta", Check: pass},
	}, time.Second), http.MethodGet, "/readyz")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin 200", rec.Code)
	}
	body := decodeReady(t, rec)
	if body.Status != "ready" || len(body.Checks) != 2 {
		t.Fatalf("body = %+v", body)
	}
	for i, name := range []string{"alpha", "beta"} {
		if body.Checks[i].Name != name || body.Checks[i].Status != "ok" {
			t.Errorf("checks[%d] = %+v, ingin %s ok", i, body.Checks[i], name)
		}
	}
}

// TestReadyz_TC_M00_004 memastikan /readyz menjawab 503 bila satu pemeriksaan error dan satu
// melewati batas waktu, tanpa membocorkan teks error ke body, dan tanpa menunggu lama karena
// pemeriksaan dijalankan paralel dengan batas waktu.
func TestReadyz_TC_M00_004(t *testing.T) {
	const timeout = 50 * time.Millisecond
	h := newTestHandler([]ReadinessCheck{
		{Name: "alpha", Check: func(context.Context) error { return nil }},
		{Name: "beta", Check: func(context.Context) error { return errors.New("secret connection detail") }},
		{Name: "gamma", Check: func(ctx context.Context) error {
			// Pemeriksaan yang macet: hanya selesai bila konteksnya habis.
			<-ctx.Done()
			return ctx.Err()
		}},
	}, timeout)

	start := time.Now()
	rec := serve(h, http.MethodGet, "/readyz")
	elapsed := time.Since(start)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, ingin 503", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "secret") {
		t.Errorf("body membocorkan error internal: %s", rec.Body.String())
	}
	body := decodeReady(t, rec)
	want := map[string]string{"alpha": "ok", "beta": "failed", "gamma": "failed"}
	if body.Status != "not_ready" || len(body.Checks) != len(want) {
		t.Fatalf("body = %+v", body)
	}
	for _, c := range body.Checks {
		if want[c.Name] != c.Status {
			t.Errorf("pemeriksaan %s = %s, ingin %s", c.Name, c.Status, want[c.Name])
		}
	}
	// Batas longgar supaya test tidak rapuh di CI, tetapi tetap jauh di bawah "menggantung".
	if elapsed > 2*time.Second {
		t.Errorf("/readyz butuh %v; pemeriksaan macet seharusnya dihentikan setelah %v", elapsed, timeout)
	}
}
