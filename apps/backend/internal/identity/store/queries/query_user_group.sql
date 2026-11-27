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