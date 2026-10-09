// Endpoint operasional /healthz (proses hidup) dan /readyz (siap melayani request),
// docs/03-architecture.md §8 Observability.

package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// Nilai status di body /healthz dan /readyz; dibaca mesin (load balancer, Docker), bukan
// ditampilkan ke pengguna.
const (
	// statusOK menandai proses hidup atau satu pemeriksaan yang lulus.
	statusOK = "ok"
	// statusFailed menandai pemeriksaan yang error atau melewati batas waktu.
	statusFailed = "failed"
	// statusReady menandai semua pemeriksaan kesiapan lulus.
	statusReady = "ready"
	// statusNotReady menandai minimal satu pemeriksaan kesiapan gagal.
	statusNotReady = "not_ready"
)

// ReadinessCheck adalah satu pemeriksaan yang harus lulus sebelum server dianggap siap,
// misalnya "database reachable and migrations current" mulai M1.
type ReadinessCheck struct {
	// Name adalah nama pemeriksaan yang tampil di body /readyz.
	Name string
	// Check mengembalikan nil bila lulus; harus berhenti saat ctx habis.
	Check func(ctx context.Context) error
}

// checkResult adalah hasil satu pemeriksaan di body /readyz. Detail error sengaja tidak
// disertakan karena bisa memuat informasi internal; detailnya hanya dicatat ke log.
type checkResult struct {
	// Name adalah nama pemeriksaan.
	Name string `json:"name"`
	// Status adalah statusOK atau statusFailed.
	Status string `json:"status"`
}

// readyResponse adalah body /readyz.
type readyResponse struct {
	// Status adalah statusReady atau statusNotReady.
	Status string `json:"status"`
	// Checks berisi hasil setiap pemeriksaan sesuai urutan pendaftaran.
	Checks []checkResult `json:"checks"`
}

// handleHealthz menjawab 200 selama proses berjalan, tanpa menyentuh dependensi apa pun.
func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, "application/json", map[string]string{"status": statusOK})
}

// readyzHandler menjalankan semua pemeriksaan kesiapan secara paralel, masing-masing dengan
// batas waktu, lalu menjawab 200 bila semuanya lulus dan 503 bila ada yang gagal.
func readyzHandler(logger *slog.Logger, checks []ReadinessCheck, timeout time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		results := make([]checkResult, len(checks))
		var wg sync.WaitGroup
		for i, c := range checks {
			wg.Go(func() {
				ctx, cancel := context.WithTimeout(r.Context(), timeout)
				defer cancel()
				results[i] = checkResult{Name: c.Name, Status: statusOK}
				if err := runCheck(ctx, c.Check); err != nil {
					results[i].Status = statusFailed
					logger.WarnContext(r.Context(), "readiness check failed", slog.String("check", c.Name), slog.Any("error", err))
				}
			})
		}
		wg.Wait()

		resp := readyResponse{Status: statusReady, Checks: results}
		status := http.StatusOK
		for _, res := range results {
			if res.Status != statusOK {
				resp.Status = statusNotReady
				status = http.StatusServiceUnavailable
				break
			}
		}
		writeJSON(w, status, "application/json", resp)
	})
}

// runCheck menjalankan satu pemeriksaan dan tidak menunggu lebih lama dari konteksnya,
// walaupun fungsi pemeriksaan sendiri mengabaikan konteks.
func runCheck(ctx context.Context, check func(context.Context) error) error {
	done := make(chan error, 1)
	go func() { done <- check(ctx) }()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
