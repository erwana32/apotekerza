-- name: GetAccessGroupByID :one
SELECT id, name, is_builtin, is_active, created_at, updated_at
FROM identity_access_groups
WHERE id = $1;

-- name: ListAccessGroups :many
-- Keyset pagination on the unique name (groups are global): pass the last name
-- of the previous page as after_name (NULL for the first page). The caller asks
-- for one row more than the page size. search is NULL for no filter.
SELECT id, name, is_builtin, is_active, created_at, updated_at
FROM identity_access_groups
WHERE (sqlc.narg('search')::text IS NULL OR name ILIKE '%' || sqlc.narg('search')::text || '%')
  AND (sqlc.narg('after_name')::text IS NULL OR name > sqlc.narg('after_name')::text)
ORDER BY name
LIMIT sqlc.arg('page_size')::int;

-- name: InsertAccessGroup :one
INSERT INTO identity_access_groups (id, name, is_builtin, is_active)
VALUES ($1, $2, $3, $4)
RETURNING id, name, is_builtin, is_active, created_at, updated_at;

-- name: UpdateAccessGroup :one
UPDATE identity_access_groups
SET name = $2,
    is_active = $3,
    updated_at = now()
WHERE id = $1
RETURNING id, name, is_builtin, is_active, created_at, updated_at;

-- name: DeleteGroupPermissions :exec
DELETE FROM identity_group_permissions
WHERE access_group_id = $1;

-- name: InsertGroupPermission :exec
INSERT INTO identity_group_permissions (access_group_id, permission)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: ListGroupPermissions :many
SELECT permission
FROM identity_group_permissions
WHERE access_group_id = $1
ORDER BY permission;

-- name: ListPermissionsForUser :many
SELECT DISTINCT gp.permission
FROM identity_group_permissions gp
JOIN identity_user_groups ug ON ug.access_group_id = gp.access_group_id
JOIN identity_access_groups g ON g.id = ug.access_group_id
WHERE ug.user_id = $1 AND g.is_active
ORDER BY gp.permission;