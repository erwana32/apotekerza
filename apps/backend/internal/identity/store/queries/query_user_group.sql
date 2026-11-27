-- name: InsertUserGroup :exec
INSERT INTO identity_user_groups (user_id, access_group_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: DeleteUserGroups :exec
DELETE FROM identity_user_groups
WHERE user_id = $1;

-- name: ListGroupIDsForUser :many
SELECT access_group_id
FROM identity_user_groups
WHERE user_id = $1
ORDER BY access_group_id;

-- name: ListGroupIDsForUsers :many
-- Batch form for ListUsers, so a page of users costs one extra query instead of
-- one per user.
SELECT user_id, access_group_id
FROM identity_user_groups
WHERE user_id = ANY($1::uuid[])
ORDER BY user_id, access_group_id;

-- name: LockBuiltinAccessGroups :exec
-- Serialises changes that could remove the last Administrator: two concurrent
-- transactions each deactivating a different admin would both see "one left".
SELECT id FROM identity_access_groups WHERE is_builtin FOR UPDATE;

-- name: CountActiveBuiltinMembers :one
-- Active users that hold an active builtin group (Administrator).
SELECT count(DISTINCT u.id)::bigint
FROM identity_users u
JOIN identity_user_groups ug ON ug.user_id = u.id
JOIN identity_access_groups g ON g.id = ug.access_group_id
WHERE u.is_active AND g.is_builtin AND g.is_active;
