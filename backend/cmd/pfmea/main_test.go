// Test untuk binary pfmea: penanganan subcommand dan siklus hidup `serve` (test case TC-M00-013
// dan TC-M00-014, docs/test-cases/M00-scaffold.md).

package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

// syncBuffer adalah bytes.Buffer yang aman dipakai bersamaan oleh server (menulis log) dan test
// (membaca log).
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

// Write menambahkan data ke buffer dengan kunci.
func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

// String mengembalikan salinan isi buffer dengan kunci.
func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// envOf membuat fungsi getenv tiruan dari map.
func envOf(vars map[string]string) func(string) string {
	return func(key string) string { return vars[key] }
}

// TestRun_TC_M00_013 memastikan binary menampilkan usage, menolak subcommand yang tidak dikenal,
// memberi tahu subcommand milestone berikutnya, dan menolak `serve` dengan konfigurasi salah
// sambil mencetak semua masalah sekaligus.
func TestRun_TC_M00_013(t *testing.T) {
	cases := []struct {
		name       string
		args       []string
		wantCode   int
		wantStdout []string
		wantStderr []string
	}{
		{"no arguments", nil, 2, nil, []string{"Usage: pfmea", "serve"}},
		{"help", []string{"help"}, 0, []string{"Usage: pfmea", "serve"}, nil},
		{"migrate", []string{"migrate", "up"}, 2, nil, []string{"pfmea migrate: available from M1"}},
		{"seed-demo", []string{"seed-demo"}, 2, nil, []string{"pfmea seed-demo: available from M1"}},
		{"init", []string{"init"}, 2, nil, []string{"pfmea init: available from M2"}},
		{"unknown command", []string{"unknown"}, 2, nil, []string{`unknown command "unknown"`, "Usage: pfmea"}},
		{"serve without configuration", []string{"serve"}, 1, nil, []string{"DATABASE_URL is required", "APP_BASE_URL is required"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var stdout, stderr syncBuffer
			code := run(context.Background(), c.args, envOf(nil), &stdout, &stderr)
			if code != c.wantCode {
				t.Errorf("exit code = %d, want %d\nstdout: %s\nstderr: %s", code, c.wantCode, stdout.String(), stderr.String())
			}
			for _, s := range c.wantStdout {
				if !strings.Contains(stdout.String(), s) {
					t.Errorf("stdout lacks %q: %s", s, stdout.String())
				}
			}
			for _, s := range c.wantStderr {
				if !strings.Contains(stderr.String(), s) {
					t.Errorf("stderr lacks %q: %s", s, stderr.String())
				}
			}
		})
	}
}

// listeningAddr mencari baris log JSON "listening" dan mengembalikan alamatnya.
func listeningAddr(log string) string {
	sc := bufio.NewScanner(strings.NewReader(log))
	for sc.Scan() {
		var line struct {
			Msg  string `json:"msg"`
			Addr string `json:"addr"`
		}
		if json.Unmarshal(sc.Bytes(), &line) == nil && line.Msg == "listening" {
			return line.Addr
		}
	}
	return ""
}

// TestServe_TC_M00_014 menjalankan `serve` sungguhan di port acak, memastikan /healthz menjawab
// 200, lalu menghentikannya (setara SIGINT/SIGTERM) dan memastikan proses selesai rapi dengan
// kode 0 tanpa mencatat kata sandi database ke log.
func TestServe_TC_M00_014(t *testing.T) {
	env := envOf(map[string]string{
		"DATABASE_URL": "postgres://pfmea:secret@localhost:5432/pfmea",
		"APP_BASE_URL": "http://localhost:8080",
		"HTTP_ADDR":    "127.0.0.1:0",
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var stdout, stderr syncBuffer
	done := make(chan int, 1)
	go func() { done <- run(ctx, []string{"serve"}, env, &stdout, &stderr) }()

	// Tunggu sampai server mencatat alamat listen-nya.
	var addr string
	deadline := time.Now().Add(10 * time.Second)
	for addr == "" && time.Now().Before(deadline) {
		select {
		case code := <-done:
			t.Fatalf("serve stopped early with exit code %d\nstdout: %s\nstderr: %s", code, stdout.String(), stderr.String())
		case <-time.After(20 * time.Millisecond):
		}
		addr = listeningAddr(stdout.String())
	}
	if addr == "" {
		t.Fatalf("log line \"listening\" did not appear\nstdout: %s", stdout.String())
	}

	resp, err := http.Get("http://" + addr + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("/healthz status = %d, want 200", resp.StatusCode)
	}

	cancel()
	select {
	case code := <-done:
		if code != 0 {
			t.Errorf("exit code = %d, want 0\nstderr: %s", code, stderr.String())
		}
	case <-time.After(15 * time.Second):
		t.Fatal("serve did not stop after the context was cancelled")
	}

	if strings.Contains(stdout.String()+stderr.String(), "secret") {
		t.Errorf("log leaks the database password:\n%s", stdout.String())
	}
}

// TestServeShutdown_TC_M00_014 memastikan klien yang mengirim header Content-Length tetapi tidak
// pernah mengirim body tidak bisa menahan shutdown: setelah batas waktu shutdown koneksinya
// diputus paksa dan proses tetap selesai dengan kode 0.
func TestServeShutdown_TC_M00_014(t *testing.T) {
	env := envOf(map[string]string{
		"DATABASE_URL": "postgres://pfmea:secret@localhost:5432/pfmea",
		"APP_BASE_URL": "http://localhost:8080",
		"HTTP_ADDR":    "127.0.0.1:0",
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var stdout, stderr syncBuffer
	done := make(chan int, 1)
	go func() {
		done <- serveWith(ctx, env, &stdout, &stderr, serveOptions{shutdownTimeout: 200 * time.Millisecond})
	}()

	var addr string
	for deadline := time.Now().Add(10 * time.Second); addr == "" && time.Now().Before(deadline); {
		time.Sleep(20 * time.Millisecond)
		addr = listeningAddr(stdout.String())
	}
	if addr == "" {
		t.Fatalf("log line \"listening\" did not appear\nstdout: %s", stdout.String())
	}

	// Klien lambat: header lengkap dengan body 10 byte yang tidak pernah dikirim.
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer func() { _ = conn.Close() }()
	if _, err := conn.Write([]byte("GET /packages HTTP/1.1\r\nHost: x\r\nContent-Length: 10\r\n\r\n")); err != nil {
		t.Fatalf("write: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	start := time.Now()
	cancel()
	select {
	case code := <-done:
		if code != 0 {
			t.Errorf("exit code = %d, want 0\nstdout: %s", code, stdout.String())
		}
	case <-time.After(5 * time.Second):
		t.Fatal("a slow client blocked the shutdown")
	}
	if elapsed := time.Since(start); elapsed > 3*time.Second {
		t.Errorf("shutdown took %v", elapsed)
	}
	if !strings.Contains(stdout.String(), "forcing close") {
		t.Errorf("log does not record the forced close:\n%s", stdout.String())
	}
}
