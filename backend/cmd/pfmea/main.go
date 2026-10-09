// Command pfmea adalah binary tunggal aplikasi (docs/03-architecture.md §3.2). Di M0 hanya
// subcommand serve yang tersedia; migrate, seed-demo, init, perf-gen, xlsx-compare dan perf
// ditambahkan oleh milestone masing-masing.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pfmea/backend/internal/config"
	"pfmea/backend/internal/httpapi"
	"pfmea/backend/internal/i18n"
	"pfmea/backend/internal/webui"
)

// Batas waktu server HTTP.
const (
	// readHeaderTimeout membatasi klien lambat yang menahan koneksi tanpa mengirim header.
	readHeaderTimeout = 10 * time.Second
	// idleTimeout menutup koneksi keep-alive yang menganggur.
	idleTimeout = 2 * time.Minute
	// shutdownTimeout adalah waktu yang diberikan ke request yang sedang berjalan saat server
	// dihentikan sebelum koneksinya diputus.
	shutdownTimeout = 15 * time.Second
)

// Kode keluar proses.
const (
	// exitOK berarti selesai normal.
	exitOK = 0
	// exitError berarti perintah gagal dijalankan (misalnya konfigurasi salah).
	exitError = 1
	// exitUsage berarti perintah tidak dikenal atau belum tersedia.
	exitUsage = 2
)

// main menghubungkan sinyal SIGINT/SIGTERM ke konteks supaya serve berhenti dengan rapi, lalu
// menjalankan run dan keluar dengan kode hasilnya.
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	code := run(ctx, os.Args[1:], os.Getenv, os.Stdout, os.Stderr)
	stop()
	os.Exit(code)
}

// run menjalankan satu subcommand. Semua dependensi proses (argumen, environment, output)
// diberikan sebagai parameter supaya bisa diuji tanpa menjalankan binary.
func run(ctx context.Context, args []string, getenv func(string) string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		_, _ = fmt.Fprint(stderr, i18n.CLIUsage)
		return exitUsage
	}
	switch cmd := args[0]; cmd {
	case "serve":
		return serve(ctx, getenv, stdout, stderr)
	case "help", "-h", "--help":
		_, _ = fmt.Fprint(stdout, i18n.CLIUsage)
		return exitOK
	default:
		if milestone := availableFrom(cmd); milestone != "" {
			_, _ = fmt.Fprintf(stderr, i18n.CLINotAvailable+"\n", cmd, milestone)
			return exitUsage
		}
		_, _ = fmt.Fprintf(stderr, i18n.CLIUnknownCommand+"\n\n", cmd)
		_, _ = fmt.Fprint(stderr, i18n.CLIUsage)
		return exitUsage
	}
}

// availableFrom mengembalikan milestone yang membuat subcommand (docs/03-architecture.md §3.2),
// atau string kosong bila subcommand tidak dikenal sama sekali.
func availableFrom(cmd string) string {
	switch cmd {
	case "migrate", "seed-demo":
		return "M1"
	case "init":
		return "M2"
	case "perf-gen":
		return "M6"
	case "xlsx-compare":
		return "M10"
	case "perf":
		return "M13"
	default:
		return ""
	}
}

// serve membaca konfigurasi, lalu menjalankan server HTTP sampai ctx dibatalkan dan
// menghentikannya dengan rapi. Log berformat JSON ditulis ke stdout; kesalahan konfigurasi
// ditulis ke stderr sebagai teks biasa supaya mudah dibaca operator.
func serve(ctx context.Context, getenv func(string) string, stdout, stderr io.Writer) int {
	cfg, err := config.Load(getenv)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "pfmea serve: %v\n", err)
		return exitError
	}

	logger := slog.New(slog.NewJSONHandler(stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	for _, w := range cfg.Warnings {
		logger.WarnContext(ctx, w)
	}
	// Config mengimplementasikan slog.LogValuer, jadi kata sandi database sudah disamarkan.
	logger.InfoContext(ctx, "starting", slog.Any("config", cfg))

	// TODO(M1): daftarkan pemeriksaan kesiapan "database reachable and migrations current".
	handler := httpapi.NewHandler(httpapi.Options{
		Logger: logger,
		UI:     webui.Handler(webui.Files()),
	})

	ln, err := new(net.ListenConfig).Listen(ctx, "tcp", cfg.HTTPAddr)
	if err != nil {
		logger.ErrorContext(ctx, "listen failed", slog.String("addr", cfg.HTTPAddr), slog.Any("error", err))
		return exitError
	}
	srv := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		IdleTimeout:       idleTimeout,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}
	// Alamat sebenarnya dicatat (penting bila port 0 dipakai untuk memilih port acak).
	logger.InfoContext(ctx, "listening", slog.String("addr", ln.Addr().String()))

	serveErr := make(chan error, 1)
	go func() { serveErr <- srv.Serve(ln) }()

	select {
	case err := <-serveErr:
		logger.ErrorContext(ctx, "server stopped unexpectedly", slog.Any("error", err))
		return exitError
	case <-ctx.Done():
	}

	// ctx sudah dibatalkan, jadi shutdown memakai konteks baru dengan batas waktunya sendiri.
	logger.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown failed", slog.Any("error", err))
		return exitError
	}
	if err := <-serveErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server stopped with error", slog.Any("error", err))
		return exitError
	}
	logger.Info("stopped")
	return exitOK
}
