-- Query pengguna untuk sqlc (backend/internal/store/sqlc). Hash kata sandi hanya dibaca oleh
-- GetUserByUsername (dipakai login mulai M2), tidak oleh daftar pengguna.

-- name: GetUserByUsername :one
SELECT id, username, display_name, email, department, role, auth_source, password_hash,
       is_active, last_login_at, created_at, updated_at, version
FROM users
WHERE username = $1;

-- name: GetUserByID :one
SELECT id, username, display_name, email, department, role, auth_source,
       is_active, last_login_at, created_at, updated_at, version
FROM users
WHERE id = $1;

-- name: ListUsers :many
SELECT id, username, display_name, email, department, role, auth_source,
       is_active, last_login_at, version
FROM users
ORDER BY username;
