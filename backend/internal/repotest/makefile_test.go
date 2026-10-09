// Test target Makefile (test case TC-M00-015 dan TC-M00-016, docs/test-cases/M00-scaffold.md).

package repotest

import (
	"regexp"
	"strings"
	"testing"
)

// TestTargets_TC_M00_015 memastikan target yang berjalan di M0 ada dan resepnya memanggil alat
// yang benar. `make -n` hanya mencetak perintah tanpa menjalankannya, jadi test ini cepat dan
// tidak butuh Docker, Node atau database.
func TestTargets_TC_M00_015(t *testing.T) {
	// Potongan perintah yang wajib muncul di resep setiap target (docs/03-architecture.md §3.3,
	// docs/10-milestones.md M0).
	want := map[string][]string{
		"tools": {"sqlc", "goose", "oapi-codegen", "golangci-lint", "npm ci"},
		"db":    {"docker compose", "up", "--wait", "db"},
		"dev":   {".env", "air", "npm run dev"},
		"test":  {"go test -race", "npm run test"},
		"lint":  {"golangci-lint", "npm run check", "npm run lint", "redocly", "lint api/openapi.yaml"},
		"build": {"npm run build", "backend/internal/webui/dist", "go build", "./backend/cmd/pfmea"},
		"check": {"gen", "lint", "test", "test-rules"},
	}
	for target, parts := range want {
		t.Run(target, func(t *testing.T) {
			out, code := run(t, "make", "--no-print-directory", "-n", target)
			if code != 0 {
				t.Fatalf("make -n %s: kode keluar %d\n%s", target, code, out)
			}
			// check memanggil gen dan test-rules yang di M0 masih placeholder, jadi hanya target
			// lain yang tidak boleh mencetak "available from".
			if target != "check" && strings.Contains(out, "available from") {
				t.Errorf("make %s seharusnya sudah berjalan di M0:\n%s", target, out)
			}
			for _, p := range parts {
				if !strings.Contains(out, p) {
					t.Errorf("resep make %s tidak memuat %q:\n%s", target, p, out)
				}
			}
		})
	}

	// check harus menjalankan gen, lint, test, test-rules berurutan (docs/11-testing.md §1).
	out, _ := run(t, "make", "--no-print-directory", "-n", "check")
	sub := regexp.MustCompile(`(?m)^\S*make --no-print-directory (\S+)$`)
	var order []string
	for _, m := range sub.FindAllStringSubmatch(out, -1) {
		order = append(order, m[1])
	}
	if strings.Join(order, " ") != "gen lint test test-rules" {
		t.Errorf("urutan make check = %q, ingin \"gen lint test test-rules\"\n%s", order, out)
	}
}

// TestPlaceholders_TC_M00_016 memastikan target milestone berikutnya hanya mencetak kapan
// tersedia dan keluar dengan kode 0, sehingga `make check` dan CI tetap hijau di M0.
func TestPlaceholders_TC_M00_016(t *testing.T) {
	want := map[string]string{
		"migrate":      "M1",
		"migrate-down": "M1",
		"seed":         "M1",
		"test-rules":   "M1",
		"gen":          "M1",
		"e2e":          "M2",
		"perf":         "M13",
	}
	for target, milestone := range want {
		t.Run(target, func(t *testing.T) {
			out, code := run(t, "make", "--no-print-directory", target)
			if code != 0 {
				t.Errorf("make %s: kode keluar %d, ingin 0\n%s", target, code, out)
			}
			if !strings.Contains(out, "available from "+milestone+"\n") {
				t.Errorf("make %s: output %q tidak memuat \"available from %s\"", target, out, milestone)
			}
		})
	}
}
