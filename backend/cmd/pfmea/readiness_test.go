// Test pemeriksaan kesiapan "database" di `pfmea serve` (test case TC-M01-009,
// docs/test-cases/M01-database.md).

package main

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"pfmea/backend/internal/testdb"
)

// startServe menjalankan serve di port acak dengan DATABASE_URL yang diberikan dan mengembalikan
// alamatnya. Server dihentikan saat test selesai dan harus berhenti dengan kode 0.
func startServe(t *testing.T, databaseURL string) string {
	t.Helper()
	env := envOf(map[string]string{
		"DATABASE_URL": databaseURL,
		"APP_BASE_URL": "http://localhost:8080",
		"HTTP_ADDR":    "127.0.0.1:0",
	})
	ctx, cancel := context.WithCancel(context.Background())
	var stdout, stderr syncBuffer
	done := make(chan int, 1)
	go func() { done <- serveWith(ctx, env, &stdout, &stderr, serveOptions{shutdownTimeout: time.Second}) }()
	t.Cleanup(func() {
		cancel()
		select {
		case code := <-done:
			if code != 0 {
				t.Errorf("serve exit code = %d, want 0\nstdout: %s\nstderr: %s", code, stdout.String(), stderr.String())
			}
		case <-time.After(10 * time.Second):
			t.Error("serve did not stop")
		}
	})

	for deadline := time.Now().Add(10 * time.Second); time.Now().Before(deadline); {
		if addr := listeningAddr(stdout.String()); addr != "" {
			return addr
		}
		select {
		case code := <-done:
			t.Fatalf("serve stopped early with exit code %d\nstdout: %s\nstderr: %s", code, stdout.String(), stderr.String())
		case <-time.After(20 * time.Millisecond):
		}
	}
	t.Fatalf("log line \"listening\" did not appear\nstdout: %s", stdout.String())
	return ""
}

// readyBody adalah bagian body /readyz yang diperiksa test.
type readyBody struct {
	Status string `json:"status"`
	Checks []struct {
		Name   string `json:"name"`
		Status string `json:"status"`
	} `json:"checks"`
}

// TestReadyzDatabase_TC_M01_009 memastikan /readyz memeriksa bahwa database terjangkau dan sudah
// dimigrasi ke versi terbaru, tanpa membuat tabel goose di database kosong, dan menjawab cepat
// bila database tidak terjangkau; /healthz selalu 200.
func TestReadyzDatabase_TC_M01_009(t *testing.T) {
	migrated := testdb.New(t)
	empty := testdb.NewEmpty(t)
	cases := []struct {
		name       string
		url        string
		wantStatus int
		wantCheck  string
	}{
		{"migrated database", migrated.URL, http.StatusOK, "ok"},
		{"empty database", empty.URL, http.StatusServiceUnavailable, "failed"},
		{"unreachable database", "postgres://pfmea:secret@127.0.0.1:1/pfmea", http.StatusServiceUnavailable, "failed"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			addr := startServe(t, c.url)

			start := time.Now()
			resp, err := http.Get("http://" + addr + "/readyz")
			if err != nil {
				t.Fatalf("GET /readyz: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()
			if elapsed := time.Since(start); elapsed >= 3*time.Second {
				t.Errorf("/readyz took %v, want < 3s", elapsed)
			}
			var body readyBody
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				t.Fatalf("decode /readyz: %v", err)
			}
			if resp.StatusCode != c.wantStatus {
				t.Errorf("/readyz status = %d, want %d (body %+v)", resp.StatusCode, c.wantStatus, body)
			}
			found := false
			for _, ch := range body.Checks {
				if ch.Name == "database" {
					found = true
					if ch.Status != c.wantCheck {
						t.Errorf("check database = %q, want %q", ch.Status, c.wantCheck)
					}
				}
			}
			if !found {
				t.Errorf("/readyz has no check named database: %+v", body)
			}

			health, err := http.Get("http://" + addr + "/healthz")
			if err != nil {
				t.Fatalf("GET /healthz: %v", err)
			}
			_ = health.Body.Close()
			if health.StatusCode != http.StatusOK {
				t.Errorf("/healthz status = %d, want 200", health.StatusCode)
			}
		})
	}

	if n := count(t, empty.Pool, "SELECT count(*) FROM pg_tables WHERE schemaname = 'public'"); n != 0 {
		t.Errorf("the readiness check created %d tables in the empty database", n)
	}
}
