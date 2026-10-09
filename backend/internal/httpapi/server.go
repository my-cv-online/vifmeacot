// Package httpapi menyusun router HTTP server: endpoint operasional (/healthz, /readyz),
// wilayah API /api dan SPA yang tertanam. Mulai M2 paket ini juga berisi handler hasil
// oapi-codegen dan rantai middleware (docs/03-architecture.md §5).
package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"pfmea/backend/internal/i18n"
)

// defaultReadinessTimeout adalah batas waktu satu pemeriksaan kesiapan bila Options tidak
// menentukannya; cukup pendek supaya load balancer tidak menunggu lama.
const defaultReadinessTimeout = 2 * time.Second

// Options berisi dependensi router.
type Options struct {
	// Logger dipakai untuk mencatat pemeriksaan kesiapan yang gagal.
	Logger *slog.Logger
	// Readiness adalah pemeriksaan yang dijalankan /readyz (M0: kosong, M1: database).
	Readiness []ReadinessCheck
	// ReadinessTimeout adalah batas waktu per pemeriksaan; 0 berarti defaultReadinessTimeout.
	ReadinessTimeout time.Duration
	// UI menyajikan SPA untuk semua path non-API (paket webui).
	UI http.Handler
}

// NewHandler membuat router server. Pola "/" sengaja tanpa method: pola "GET /" bentrok dengan
// "/api/" di ServeMux Go dan membuat server panic saat start, jadi method SPA diperiksa oleh
// handler UI sendiri.
func NewHandler(opts Options) http.Handler {
	if opts.Logger == nil {
		opts.Logger = slog.New(slog.DiscardHandler)
	}
	if opts.ReadinessTimeout <= 0 {
		opts.ReadinessTimeout = defaultReadinessTimeout
	}
	if opts.UI == nil {
		opts.UI = http.NotFoundHandler()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealthz)
	mux.Handle("GET /readyz", readyzHandler(opts.Logger, opts.Readiness, opts.ReadinessTimeout))
	// /api didaftarkan dua kali supaya "/api" tanpa garis miring tidak di-redirect 301.
	// TODO(M2): rute hasil oapi-codegen dipasang di bawah /api/v1; path lain tetap 404.
	mux.HandleFunc("/api", handleAPINotFound)
	mux.HandleFunc("/api/", handleAPINotFound)
	mux.Handle("/", opts.UI)
	return mux
}

// problem adalah bentuk minimal RFC 9457 Problem Details yang dipakai sebelum paket apperr ada
// (docs/05-api.md §2).
// TODO(M2): ganti dengan apperr.Problem dan pemetaan error terpusat.
type problem struct {
	// Type adalah URN jenis masalah, urn:pfmea:problem:<code>.
	Type string `json:"type"`
	// Title adalah ringkasan bahasa Inggris dari i18n.
	Title string `json:"title"`
	// Status adalah kode status HTTP.
	Status int `json:"status"`
	// Code adalah salah satu ProblemCode di api/openapi.yaml.
	Code string `json:"code"`
}

// handleAPINotFound menjawab path /api yang belum punya rute dengan 404 Problem Details, supaya
// klien API tidak pernah menerima index.html.
func handleAPINotFound(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusNotFound, "application/problem+json", problem{
		Type:   "urn:pfmea:problem:not_found",
		Title:  i18n.ProblemNotFoundTitle,
		Status: http.StatusNotFound,
		Code:   "not_found",
	})
}

// writeJSON menulis body JSON dengan content type dan status yang diberikan. Respons
// operasional dan error tidak boleh di-cache.
func writeJSON(w http.ResponseWriter, status int, contentType string, v any) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	// Error encode tidak bisa dilaporkan lagi ke klien setelah header terkirim.
	_ = json.NewEncoder(w).Encode(v)
}
