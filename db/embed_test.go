// Test file migrasi dan seed yang tertanam (test case TC-M01-001,
// docs/test-cases/M01-database.md). Tidak butuh database.

package db

import (
	"io/fs"
	"testing"
)

// TestFiles_TC_M01_001 memastikan migrasi dan seed demo ikut tertanam di binary dan versi migrasi
// terakhir dibaca dari nama file.
func TestFiles_TC_M01_001(t *testing.T) {
	for _, name := range []string{"migrations/00001_init.sql", "seed/demo.sql"} {
		b, err := fs.ReadFile(Files, name)
		if err != nil {
			t.Fatalf("embedded file %s missing: %v", name, err)
		}
		if len(b) == 0 {
			t.Errorf("embedded file %s is empty", name)
		}
	}
	v, err := LatestVersion()
	if err != nil {
		t.Fatalf("LatestVersion: %v", err)
	}
	if v != 1 {
		t.Errorf("LatestVersion() = %d, want 1", v)
	}
}
