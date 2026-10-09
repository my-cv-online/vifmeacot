// Test untuk pembacaan dan validasi konfigurasi (test case TC-M00-005 sampai TC-M00-008,
// docs/test-cases/M00-scaffold.md).

package config

import (
	"errors"
	"fmt"
	"log/slog"
	"net/netip"
	"strings"
	"testing"
	"time"
)

// envOf membuat fungsi getenv tiruan dari map supaya test tidak bergantung pada environment
// proses yang menjalankan test.
func envOf(vars map[string]string) func(string) string {
	return func(key string) string { return vars[key] }
}

// requiredOnly berisi dua variabel wajib dengan nilai benar; test lain menambahkan variabel
// di atasnya.
func requiredOnly() map[string]string {
	return map[string]string{
		"DATABASE_URL": "postgres://pfmea:secret@localhost:5432/pfmea?sslmode=disable",
		"APP_BASE_URL": "http://localhost:5173",
	}
}

// problemsOf mengambil daftar masalah dari error validasi; test gagal bila error bukan
// *ValidationError.
func problemsOf(t *testing.T, err error) []string {
	t.Helper()
	var verr *ValidationError
	if !errors.As(err, &verr) {
		t.Fatalf("error bukan *ValidationError: %v", err)
	}
	return verr.Problems
}

// TestLoad_TC_M00_005 memastikan semua nilai bawaan docs/03-architecture.md §8 terpasang bila
// hanya variabel wajib yang di-set, dan kata sandi database tidak muncul di log.
func TestLoad_TC_M00_005(t *testing.T) {
	cfg, err := Load(envOf(requiredOnly()))
	if err != nil {
		t.Fatalf("Load gagal: %v", err)
	}

	checks := []struct {
		name string
		got  any
		want any
	}{
		{"DatabaseURL", cfg.DatabaseURL, "postgres://pfmea:secret@localhost:5432/pfmea?sslmode=disable"},
		{"AppBaseURL", cfg.AppBaseURL, "http://localhost:5173"},
		{"HTTPAddr", cfg.HTTPAddr, ":8080"},
		{"SessionTTL", cfg.SessionTTL, 12 * time.Hour},
		{"ExportDir", cfg.ExportDir, "/data/exports"},
		{"ExportTemplateDir", cfg.ExportTemplateDir, "/data/templates"},
		{"DBMaxConns", cfg.DBMaxConns, 20},
		{"RiverWorkers", cfg.RiverWorkers, 10},
		{"RulePoolSize", cfg.RulePoolSize, 8},
		{"RuleParallelism", cfg.RuleParallelism, 4},
		{"LogLevel", cfg.LogLevel, slog.LevelInfo},
		{"MetricsAllow", fmt.Sprint(cfg.MetricsAllow), fmt.Sprint([]netip.Prefix{netip.MustParsePrefix("127.0.0.1/32")})},
		{"DevMode", cfg.DevMode, false},
		{"DevFakeToday", cfg.DevFakeToday, ""},
		{"Warnings", len(cfg.Warnings), 0},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %v, ingin %v", c.name, c.got, c.want)
		}
	}

	// Konfigurasi dicatat ke log saat start; kata sandi di DATABASE_URL harus disamarkan.
	logged := cfg.LogValue().String()
	if strings.Contains(logged, "secret") {
		t.Errorf("LogValue membocorkan kata sandi: %s", logged)
	}
	if !strings.Contains(logged, "localhost:5432") {
		t.Errorf("LogValue tidak memuat host database: %s", logged)
	}

	// Kata sandi juga bisa dikirim lewat parameter query yang dibaca pgx.
	t.Run("password in query", func(t *testing.T) {
		env := requiredOnly()
		env["DATABASE_URL"] = "postgres://pfmea@localhost:5432/pfmea?password=topsecret&sslpassword=keypass&sslmode=require"
		cfg, err := Load(envOf(env))
		if err != nil {
			t.Fatalf("Load gagal: %v", err)
		}
		logged := cfg.LogValue().String()
		for _, secret := range []string{"topsecret", "keypass"} {
			if strings.Contains(logged, secret) {
				t.Errorf("LogValue membocorkan %q: %s", secret, logged)
			}
		}
		if !strings.Contains(logged, "sslmode=require") {
			t.Errorf("parameter lain seharusnya tetap tampil: %s", logged)
		}
	})

	// URL socket Unix (host kosong, socket di parameter host) adalah bentuk pgx yang sah.
	t.Run("unix socket", func(t *testing.T) {
		env := requiredOnly()
		env["DATABASE_URL"] = "postgres:///pfmea?host=/var/run/postgresql"
		if _, err := Load(envOf(env)); err != nil {
			t.Errorf("URL socket Unix ditolak: %v", err)
		}
	})
}

// TestLoad_TC_M00_006 memastikan dua variabel wajib yang kosong dilaporkan sekaligus dalam satu
// error (bukan berhenti di masalah pertama). Nilai berisi spasi saja dianggap kosong.
func TestLoad_TC_M00_006(t *testing.T) {
	for name, env := range map[string]map[string]string{
		"not set":         {},
		"whitespace only": {"DATABASE_URL": "  ", "APP_BASE_URL": "\t"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := Load(envOf(env))
			if err == nil {
				t.Fatal("Load seharusnya gagal")
			}
			problems := problemsOf(t, err)
			want := []string{"DATABASE_URL is required", "APP_BASE_URL is required"}
			if strings.Join(problems, "|") != strings.Join(want, "|") {
				t.Errorf("masalah = %q, ingin %q", problems, want)
			}
			// Teks error dipakai apa adanya saat start gagal: satu baris per masalah.
			for _, p := range want {
				if !strings.Contains(err.Error(), "\n  - "+p) {
					t.Errorf("teks error tidak memuat baris %q:\n%s", p, err.Error())
				}
			}
		})
	}
}

// TestLoad_TC_M00_007 memastikan setiap nilai salah menghasilkan satu masalah yang menyebut
// nama variabelnya, dan semuanya dilaporkan dalam satu error.
func TestLoad_TC_M00_007(t *testing.T) {
	_, err := Load(envOf(map[string]string{
		"DATABASE_URL":     "mysql://x",
		"APP_BASE_URL":     "/relative",
		"HTTP_ADDR":        "8080",
		"SESSION_TTL":      "abc",
		"DB_MAX_CONNS":     "0",
		"RIVER_WORKERS":    "-1",
		"RULE_POOL_SIZE":   "2",
		"RULE_PARALLELISM": "3",
		"LOG_LEVEL":        "verbose",
		"METRICS_ALLOW":    "10.0.0.0/33",
		"DEV_MODE":         "maybe",
	}))
	if err == nil {
		t.Fatal("Load seharusnya gagal")
	}
	problems := problemsOf(t, err)

	// Setiap variabel yang salah harus disebut tepat oleh satu masalah.
	vars := []string{
		"DATABASE_URL", "APP_BASE_URL", "HTTP_ADDR", "SESSION_TTL", "DB_MAX_CONNS",
		"RIVER_WORKERS", "RULE_PARALLELISM", "LOG_LEVEL", "METRICS_ALLOW", "DEV_MODE",
	}
	if len(problems) != len(vars) {
		t.Errorf("jumlah masalah = %d, ingin %d: %q", len(problems), len(vars), problems)
	}
	for _, v := range vars {
		found := 0
		for _, p := range problems {
			if strings.HasPrefix(p, v+" ") {
				found++
			}
		}
		if found != 1 {
			t.Errorf("variabel %s disebut %d kali, ingin 1: %q", v, found, problems)
		}
	}
	// RULE_POOL_SIZE sendiri bernilai benar; masalahnya ada di RULE_PARALLELISM yang lebih besar.
	for _, p := range problems {
		if strings.HasPrefix(p, "RULE_PARALLELISM ") && !strings.Contains(p, "RULE_POOL_SIZE") {
			t.Errorf("pesan RULE_PARALLELISM tidak menyebut RULE_POOL_SIZE: %q", p)
		}
	}

	// APP_BASE_URL dibandingkan dengan header Origin mulai M2, jadi harus berupa origin murni:
	// skema + host (+ port), tanpa path, query atau fragment.
	for _, bad := range []string{"http://:8080", "http://localhost:5173/app", "http://localhost:5173?x=1", "http://localhost:5173#top"} {
		t.Run("APP_BASE_URL "+bad, func(t *testing.T) {
			env := requiredOnly()
			env["APP_BASE_URL"] = bad
			_, err := Load(envOf(env))
			if err == nil {
				t.Fatalf("APP_BASE_URL=%q seharusnya ditolak", bad)
			}
			if p := problemsOf(t, err); len(p) != 1 || !strings.HasPrefix(p[0], "APP_BASE_URL ") {
				t.Errorf("masalah = %q", p)
			}
		})
	}
	t.Run("APP_BASE_URL trailing slash", func(t *testing.T) {
		env := requiredOnly()
		env["APP_BASE_URL"] = "http://localhost:5173/"
		cfg, err := Load(envOf(env))
		if err != nil || cfg.AppBaseURL != "http://localhost:5173" {
			t.Errorf("AppBaseURL = %q, err = %v", cfg.AppBaseURL, err)
		}
	})
}

// TestLoad_TC_M00_008 memastikan DEV_FAKE_TODAY hanya dipakai bila DEV_MODE=true, diabaikan
// dengan peringatan bila tidak, dan format yang salah ditolak.
func TestLoad_TC_M00_008(t *testing.T) {
	// with menambahkan variabel ke konfigurasi wajib.
	with := func(extra map[string]string) map[string]string {
		env := requiredOnly()
		for k, v := range extra {
			env[k] = v
		}
		return env
	}

	t.Run("dev mode on", func(t *testing.T) {
		cfg, err := Load(envOf(with(map[string]string{"DEV_MODE": "true", "DEV_FAKE_TODAY": "2026-10-08"})))
		if err != nil {
			t.Fatalf("Load gagal: %v", err)
		}
		if cfg.DevFakeToday != "2026-10-08" || !cfg.DevMode {
			t.Errorf("DevFakeToday = %q, DevMode = %v", cfg.DevFakeToday, cfg.DevMode)
		}
		if len(cfg.Warnings) != 0 {
			t.Errorf("tidak boleh ada peringatan: %q", cfg.Warnings)
		}
	})

	t.Run("dev mode off", func(t *testing.T) {
		cfg, err := Load(envOf(with(map[string]string{"DEV_MODE": "false", "DEV_FAKE_TODAY": "2026-10-08"})))
		if err != nil {
			t.Fatalf("Load gagal: %v", err)
		}
		if cfg.DevFakeToday != "" {
			t.Errorf("DevFakeToday = %q, seharusnya diabaikan", cfg.DevFakeToday)
		}
		if len(cfg.Warnings) != 1 || !strings.Contains(cfg.Warnings[0], "DEV_FAKE_TODAY") {
			t.Errorf("peringatan = %q, ingin satu peringatan tentang DEV_FAKE_TODAY", cfg.Warnings)
		}
	})

	t.Run("wrong format", func(t *testing.T) {
		_, err := Load(envOf(with(map[string]string{"DEV_MODE": "true", "DEV_FAKE_TODAY": "08-10-2026"})))
		if err == nil {
			t.Fatal("Load seharusnya gagal")
		}
		problems := problemsOf(t, err)
		if len(problems) != 1 || !strings.HasPrefix(problems[0], "DEV_FAKE_TODAY ") || !strings.Contains(problems[0], "YYYY-MM-DD") {
			t.Errorf("masalah = %q", problems)
		}
	})
}
