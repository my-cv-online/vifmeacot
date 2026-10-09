// Package rules berisi 32 aturan konsistensi Tahap 1: satu query SQL per aturan di sql/<KODE>.sql
// dan satu fixture per aturan di testdata/<KODE>.sql (docs/06-rules.md). M1 menguji query dan
// fixture-nya terhadap data demo; mesin aturan (loader, eksekusi dalam satu snapshot, penulisan
// findings) dibuat di M8.
package rules

import "embed"

// Files berisi query aturan sql/*.sql yang tertanam di binary; mulai M8 dibaca loader mesin
// aturan saat start. Fixture di testdata/ hanya dipakai test dan tidak ikut tertanam.
//
//go:embed sql/*.sql
var Files embed.FS

// TODO(M8): loader (header rule/reads/message), konfigurasi efektif per paket, eksekusi paralel
// dalam satu snapshot, fingerprint, dan penulisan findings (docs/06-rules.md §2).
