-- Query paket untuk sqlc (backend/internal/store/sqlc). Daftar lengkap dengan filter dan
-- paginasi dibuat di M4.

-- name: GetPackageByID :one
SELECT id, kind, code, name, customer_id, part_id, pfmea_method, cp_format, status, revision,
       doc_seq, owner_id, plant, core_team, last_reviewed_at, content_version,
       created_at, updated_at, version
FROM packages
WHERE id = $1;

-- name: ListPackages :many
-- Template General (kind general) tampil lebih dulu, lalu paket model urut kode.
SELECT id, kind, code, name, customer_id, part_id, status, revision, owner_id,
       content_version, updated_at, version
FROM packages
ORDER BY kind, code;
