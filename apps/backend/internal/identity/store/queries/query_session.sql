-- name: InsertSession :one
INSERT INTO identity_sessions (id, user_id, token_hash, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, token_hash, expires_at, revoked_at, created_at;

-- name: GetSessionByTokenHash :one
-- Returns the session joined with its user, so the middleware needs one query
-- per request. An expired or revoked session is filtered out here.
SELECT
    s.id AS session_id,
    s.user_id,
    s.expires_at,
    s.revoked_at,
    u.username,
    u.display_name,
    u.is_active AS user_is_active,
    u.company_id
FROM identity_sessions s
JOIN identity_users u ON u.id = s.user_id
WHERE s.token_hash = $1
  AND s.revoked_at IS NULL
  AND s.expires_at > now();

-- name: RevokeSession :exec
UPDATE identity_sessions
SET revoked_at = now()
WHERE token_hash = $1 AND revoked_at IS NULL;

-- name: RevokeSessionsForUser :exec
UPDATE identity_sessions
SET revoked_at = now()
WHERE user_id = $1 AND revoked_at IS NULL;