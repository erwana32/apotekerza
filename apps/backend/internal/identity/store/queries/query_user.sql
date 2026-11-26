-- name: GetUserByID :one
SELECT id, company_id, username, password_hash, display_name, is_active, created_at, updated_at
FROM identity_users
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT id, company_id, username, password_hash, display_name, is_active, created_at, updated_at
FROM identity_users
WHERE username = $1;

-- name: ListUsers :many
-- Keyset pagination on the unique username: pass the last username of the
-- previous page as after_username (NULL for the first page). The caller asks
-- for one row more than the page size to know whether another page exists.
-- search is NULL for no filter.
SELECT id, company_id, username, password_hash, display_name, is_active, created_at, updated_at
FROM identity_users
WHERE (sqlc.narg('search')::text IS NULL OR username ILIKE '%' || sqlc.narg('search')::text || '%')
  AND (sqlc.narg('after_username')::text IS NULL OR username > sqlc.narg('after_username')::text)
ORDER BY username
LIMIT sqlc.arg('page_size')::int;

-- name: InsertUser :one
INSERT INTO identity_users (id, company_id, username, password_hash, display_name, is_active)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, company_id, username, password_hash, display_name, is_active, created_at, updated_at;

-- name: UpdateUser :one
UPDATE identity_users
SET display_name = $2,
    is_active = $3,
    updated_at = now()
WHERE id = $1
RETURNING id, company_id, username, password_hash, display_name, is_active, created_at, updated_at;

-- name: UpdateUserPasswordHash :one
UPDATE identity_users
SET password_hash = $2,
    updated_at = now()
WHERE id = $1
RETURNING id, company_id, username, password_hash, display_name, is_active, created_at, updated_at;
