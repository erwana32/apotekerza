-- name: InsertAuditEvent :one
INSERT INTO identity_audit_log (id, company_id, actor_user_id, action, entity, entity_id, changes)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, company_id, actor_user_id, action, entity, entity_id, changes, created_at;

-- name: ListAuditEvents :many
-- Keyset pagination on the primary key: timestamps can tie inside one
-- transaction, and UUIDv7 ids are time-ordered, so id is the stable tiebreaker.
-- entity and entity_id are optional filters (NULL means no filter).
SELECT id, company_id, actor_user_id, action, entity, entity_id, changes, created_at
FROM identity_audit_log
WHERE (sqlc.narg('entity')::text IS NULL OR entity = sqlc.narg('entity')::text)
  AND (sqlc.narg('entity_id')::uuid IS NULL OR entity_id = sqlc.narg('entity_id')::uuid)
  AND (sqlc.narg('after_id')::uuid IS NULL OR id < sqlc.narg('after_id')::uuid)
ORDER BY id DESC
LIMIT sqlc.arg('page_size')::int;

-- name: ListAuditEventsForEntity :many
-- The per-document history (ADR-0010): oldest first, as a history reads.
SELECT id, company_id, actor_user_id, action, entity, entity_id, changes, created_at
FROM identity_audit_log
WHERE entity = $1 AND entity_id = $2
ORDER BY created_at, id;