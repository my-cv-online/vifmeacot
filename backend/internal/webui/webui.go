// Package webui menyajikan SPA SvelteKit yang tertanam di binary. `make build` menyalin
// web/build ke folder dist/ paket ini sebelum `go build`; di repository dist/ hanya berisi
// penanda .keep, sehingga binary tanpa build web menjawab "UI not built yet".
package webui

import (
	"bytes"
	"embed"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"

	"pfmea/backend/internal/i18n"
)

// embedded berisi hasil build web yang disalin oleh `make build`. Prefiks all: ikut menanam
// file berawalan titik (.keep) supaya pola embed tetap valid saat dist/ belum berisi build.
//
//go:embed all:dist
var embedded embed.FS

// Prefiks path di dalam hasil build SvelteKit.
const (
	// appDir adalah folder aset yang dibuat SvelteKit; path di bawahnya tidak pernah
	// dijawab dengan index.html.
	appDir = "_app/"
	// immutableDir berisi aset ber-hash yang isinya tidak pernah berubah untuk nama yang sama.
	immutableDir = "_app/immutable/"
	// cacheImmutable boleh di-cache browser selamanya karena nama file berubah bila isinya
	// berubah.
	cacheImmutable = "public, max-age=31536000, immutable"
	// cacheRevalidate memaksa browser memeriksa ulang ke server, dipakai untuk index.html
	// supaya versi baru aplikasi langsung terpakai setelah deploy.
	cacheRevalidate = "no-cache"
)

// Files mengembalikan file system hasil build yang tertanam (isi folder dist/).
func Files() fs.FS {
	sub, err := fs.Sub(embedded, "dist")
	if err != nil {
		// Tidak mungkin terjadi: "dist" adalah nama folder yang valid dan selalu tertanam.
		panic(err)
	}
	return sub
}

// Handler membuat handler SPA dari file system hasil build. Bila index.html tidak ada (web
// belum di-build), setiap request dijawab 503 "UI not built yet".
func Handler(fsys fs.FS) http.Handler {
	index, err := fs.ReadFile(fsys, "index.html")
	if err != nil {
		return http.HandlerFunc(serveNotBuilt)
	}
	return &spa{fsys: fsys, index: index}
}

// serveNotBuilt menjawab request saat binary dibangun tanpa hasil build web.
func serveNotBuilt(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusServiceUnavailable)
	_, _ = io.WriteString(w, i18n.UINotBuilt+"\n")
}

// spa menyajikan file hasil build dan jatuh ke index.html untuk deep link, sehingga router
// SvelteKit di browser yang menentukan halaman.
type spa struct {
	// fsys adalah isi hasil build.
	fsys fs.FS
	// index adalah isi index.html, dibaca sekali saat start.
	index []byte
}

// ServeHTTP memilih antara file statis, 404 untuk aset yang tidak ada, atau index.html.
func (s *spa) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, i18n.MethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	// path.Clean membuang ".." dan garis miring ganda supaya nama tidak keluar dari fsys.
	name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
	if name == "" || name == "index.html" {
		s.serveIndex(w, r)
		return
	}
	if !hidden(name) && s.serveFile(w, r, name) {
		return
	}
	// Aset SvelteKit yang tidak ada harus 404: menjawabnya dengan index.html membuat browser
	// menerima HTML sebagai JavaScript dan error-nya sulit dilacak.
	if strings.HasPrefix(name+"/", appDir) {
		http.NotFound(w, r)
		return
	}
	s.serveIndex(w, r)
}

// serveIndex mengirim index.html tanpa redirect (http.ServeFile me-redirect /index.html ke /).
func (s *spa) serveIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", cacheRevalidate)
	http.ServeContent(w, r, "index.html", time.Time{}, bytes.NewReader(s.index))
}

// serveFile mengirim file biasa dari hasil build dan melaporkan apakah file itu ada. Folder
// tidak pernah disajikan (tidak ada daftar isi folder).
func (s *spa) serveFile(w http.ResponseWriter, r *http.Request, name string) bool {
	info, err := fs.Stat(s.fsys, name)
	if err != nil || !info.Mode().IsRegular() {
		return false
	}
	data, err := fs.ReadFile(s.fsys, name)
	if err != nil {
		return false
	}
	if strings.HasPrefix(name, immutableDir) {
		w.Header().Set("Cache-Control", cacheImmutable)
	} else {
		w.Header().Set("Cache-Control", cacheRevalidate)
	}
	// ServeContent menentukan Content-Type dari ekstensi dan menangani HEAD serta Range.
	http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(data))
	return true
}

// hidden melaporkan apakah salah satu bagian path diawali titik (misalnya .keep); file seperti
// itu bukan bagian aplikasi dan tidak disajikan.
func hidden(name string) bool {
	for _, part := range strings.Split(name, "/") {
		if strings.HasPrefix(part, ".") {
			return true
		}
	}
	return false
}
