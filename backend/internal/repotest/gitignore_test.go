// Test aturan .gitignore (test case TC-M00-017, docs/test-cases/M00-scaffold.md).

package repotest

import "testing"

// TestGitignore_TC_M00_017 memastikan rahasia lokal, hasil build dan laporan tidak pernah
// ter-commit, sedangkan penanda dist/.keep dan .env.example tetap ikut repository.
// --no-index dipakai supaya file yang sudah di-track (dist/.keep) dinilai dari aturan, bukan
// dari indeks git.
func TestGitignore_TC_M00_017(t *testing.T) {
	ignored := []string{
		".env",
		".env.local",
		"bin/pfmea",
		"web/build/index.html",
		"web/node_modules/x",
		"web/.svelte-kit/x",
		"backend/internal/webui/dist/index.html",
		"backend/internal/webui/dist/_app/immutable/x.js",
		"tmp/pfmea",
		"perf-report.json",
	}
	for _, p := range ignored {
		if out, code := run(t, "git", "check-ignore", "-q", "--no-index", p); code != 0 {
			t.Errorf("%s seharusnya diabaikan git (kode %d) %s", p, code, out)
		}
	}

	kept := []string{"backend/internal/webui/dist/.keep", ".env.example"}
	for _, p := range kept {
		if _, code := run(t, "git", "check-ignore", "-q", "--no-index", p); code != 1 {
			t.Errorf("%s tidak boleh diabaikan git (kode %d)", p, code)
		}
	}
}
