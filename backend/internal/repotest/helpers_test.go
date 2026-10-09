// Fungsi bantu bersama untuk test tingkat repository.

package repotest

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// repoRoot mencari root repository dengan naik dari folder test sampai menemukan go.mod.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found in any parent folder")
		}
		dir = parent
	}
}

// cleanEnv menyalin environment proses tanpa variabel make. Test ini sendiri dijalankan oleh
// `make test`; tanpa pembersihan, flag make induk (misalnya -n atau -j) ikut ke make anak.
func cleanEnv() []string {
	var env []string
	for _, kv := range os.Environ() {
		name, _, _ := strings.Cut(kv, "=")
		switch name {
		case "MAKEFLAGS", "MAKELEVEL", "MFLAGS", "MAKEOVERRIDES":
			continue
		}
		env = append(env, kv)
	}
	return env
}

// run menjalankan perintah di root repository dan mengembalikan output gabungan dan kode
// keluarnya. Perintah yang tidak ada (misalnya make belum terpasang) menggagalkan test, bukan
// melewatinya, karena perkakas itu wajib (docs/build-guide.md §2.1).
func run(t *testing.T, name string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = repoRoot(t)
	cmd.Env = cleanEnv()
	out, err := cmd.CombinedOutput()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return string(out), exitErr.ExitCode()
		}
		t.Fatalf("%s %v could not be run: %v", name, args, err)
	}
	return string(out), 0
}
