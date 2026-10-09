# Makefile: satu pintu untuk semua perintah pengembangan (docs/03-architecture.md §3.3).
# Target yang milestone-nya belum tiba mencetak "available from M<n>" dan keluar dengan kode 0.
# Jalankan `make help` untuk daftar target.

# Bash dengan -e dan pipefail supaya langkah yang gagal di tengah pipeline menghentikan resep.
SHELL := /bin/bash
.SHELLFLAGS := -ec -o pipefail
.DEFAULT_GOAL := help

# Alat Go yang versinya dipatok dipasang di bin/ milik project (bukan alat global di komputer).
BIN := $(CURDIR)/bin
# Alat Node (redocly) dipasang oleh `npm ci` di web/node_modules.
NODE_BIN := $(CURDIR)/web/node_modules/.bin

# Paket Go yang diuji dan di-lint. ./db/... baru ikut setelah db/embed.go ada (M1). Pola
# eksplisit dipakai supaya web/node_modules tidak pernah dipindai.
GO_PKGS := ./backend/... $(if $(wildcard db/*.go),./db/...)

# Versi alat yang dipatok; ubah di sini lalu jalankan `make tools`.
SQLC_VERSION := v1.31.1
GOOSE_VERSION := v3.28.0
OAPI_CODEGEN_VERSION := v2.8.0
GOLANGCI_LINT_VERSION := v2.14.0
AIR_VERSION := v1.67.4

# Versi Go project (dari go.mod). Alat dibangun minimal dengan versi ini: golangci-lint yang
# dibangun dengan Go lebih lama menolak me-lint kode Go 1.27.
GO_TOOLCHAIN := $(shell go env GOVERSION)

# Tag build goose: hanya driver PostgreSQL yang dibutuhkan, driver lain dibuang supaya
# instalasi lebih cepat.
GOOSE_TAGS := no_clickhouse no_libsql no_mssql no_mysql no_sqlite3 no_vertica no_ydb

# go-install memasang satu alat Go ke bin/ bila versi yang dipatok belum terpasang, sehingga
# `make tools` cepat bila dijalankan ulang. File penanda bin/.<nama>-<versi>-<versi Go>
# mencatat apa yang terpasang. GOTOOLCHAIN=<versi project>+auto memakai Go project, atau versi
# yang lebih baru bila alatnya memintanya.
# Argumen: 1 = nama binary, 2 = path paket, 3 = versi, 4 = flag build tambahan.
define go-install
	@if [ -x "$(BIN)/$(1)" ] && [ -f "$(BIN)/.$(1)-$(3)-$(GO_TOOLCHAIN)" ]; then \
		echo "$(1) $(3) already installed"; \
	else \
		echo "installing $(1) $(3)"; \
		GOBIN="$(BIN)" GOTOOLCHAIN="$(GO_TOOLCHAIN)+auto" go install $(4) $(2)@$(3); \
		rm -f "$(BIN)/.$(1)-"*; touch "$(BIN)/.$(1)-$(3)-$(GO_TOOLCHAIN)"; \
	fi
endef

# db dan web juga nama folder, jadi semua target ditandai .PHONY supaya selalu dijalankan.
.PHONY: help tools db dev test lint build check gen migrate migrate-down seed test-rules e2e perf

## help: tampilkan daftar target
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed -e 's/^## /  /'

## tools: pasang alat Go yang versinya dipatok ke bin/ dan dependensi web (npm ci)
# TODO(M2): pasang juga Playwright Chromium (npx playwright install --with-deps chromium).
tools:
	@mkdir -p "$(BIN)"
	$(call go-install,sqlc,github.com/sqlc-dev/sqlc/cmd/sqlc,$(SQLC_VERSION))
	$(call go-install,goose,github.com/pressly/goose/v3/cmd/goose,$(GOOSE_VERSION),-tags '$(GOOSE_TAGS)')
	$(call go-install,oapi-codegen,github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen,$(OAPI_CODEGEN_VERSION))
	$(call go-install,golangci-lint,github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))
	$(call go-install,air,github.com/air-verse/air,$(AIR_VERSION))
	cd web && npm ci --no-audit --no-fund

## db: jalankan PostgreSQL 18 di Docker dan tunggu sampai sehat
db:
	docker compose up -d --wait db

## dev: server Go dengan live reload (air, :8080) + Vite (:5173) dengan proxy /api
# .env dimuat ke environment supaya server Go dan proxy Vite memakai HTTP_ADDR yang sama.
# Bila salah satu proses berhenti, yang lain ikut dihentikan; Ctrl+C menghentikan keduanya.
dev:
	@test -f .env || { echo "missing .env: run cp .env.example .env"; exit 1; }
	set -a; . ./.env; set +a; \
	"$(BIN)/air" -c .air.toml & air_pid=$$!; \
	(cd web && exec npm run dev) & vite_pid=$$!; \
	trap 'kill $$air_pid $$vite_pid 2>/dev/null || true' EXIT INT TERM; \
	wait -n

## test: test Go (dengan -race) + Vitest
test:
	go test -race $(GO_PKGS)
	cd web && npm run test

## lint: golangci-lint, svelte-check, eslint, redocly
# Telemetri dan cek versi redocly dimatikan karena jaringan pabrik bisa offline.
lint:
	"$(BIN)/golangci-lint" run $(GO_PKGS)
	cd web && npm run check
	cd web && npm run lint
	REDOCLY_TELEMETRY=off REDOCLY_SUPPRESS_UPDATE_NOTICE=true \
		"$(NODE_BIN)/redocly" lint api/openapi.yaml --config api/redocly.yaml

## build: build web → salin ke backend/internal/webui/dist → go build bin/pfmea
# Isi dist/ lama dihapus dulu (kecuali .keep) supaya aset build sebelumnya tidak ikut tertanam.
# Binary dibangun statis (CGO_ENABLED=0) untuk image distroless di M13.
build:
	cd web && npm run build
	find backend/internal/webui/dist -mindepth 1 -maxdepth 1 ! -name .keep -exec rm -rf {} +
	cp -R web/build/. backend/internal/webui/dist/
	CGO_ENABLED=0 go build -trimpath -o "$(BIN)/pfmea" ./backend/cmd/pfmea

## check: gen + lint + test + test-rules (wajib hijau sebelum setiap commit dan push)
# Dijalankan berurutan lewat sub-make, bukan sebagai prasyarat, supaya `make -j` tidak
# menjalankan test sebelum gen selesai.
check:
	$(MAKE) --no-print-directory gen
	$(MAKE) --no-print-directory lint
	$(MAKE) --no-print-directory test
	$(MAKE) --no-print-directory test-rules

# Target milestone berikutnya: hanya mencetak kapan tersedia, keluar dengan kode 0.

## migrate, migrate-down, seed, test-rules, gen: tersedia mulai M1 (gen lengkap mulai M2)
migrate migrate-down seed test-rules gen:
	@echo "make $@: available from M1"

## e2e: tersedia mulai M2
e2e:
	@echo "make $@: available from M2"

## perf: tersedia mulai M13
perf:
	@echo "make $@: available from M13"
