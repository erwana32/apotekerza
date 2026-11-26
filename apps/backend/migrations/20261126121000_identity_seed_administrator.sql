-- +goose Up

-- Seeds the built-in "Administrator" access group (docs/modules/identity.md).
--
-- The group holds every x-permission in api/openapi.yaml. Permissions are data,
-- so they are seeded by migration (auditable) rather than synced at startup.
-- At this point the spec only declares platform.health.read; every later
-- endpoint's permission is added by the migration that introduces it.
--
-- No user is seeded: the first Administrator needs a real password, which is
-- created by the createadmin command, not by a default credential.

-- The group is global (no company), so the seed needs no company id.
INSERT INTO identity_access_groups (id, name, is_builtin)
VALUES ('0193b7a0-0000-7000-8000-000000000001', 'Administrator', true);

INSERT INTO identity_group_permissions (access_group_id, permission)
VALUES ('0193b7a0-0000-7000-8000-000000000001', 'platform.health.read');

-- +goose Down

DELETE FROM identity_group_permissions
WHERE access_group_id = '0193b7a0-0000-7000-8000-000000000001';

DELETE FROM identity_access_groups
WHERE id = '0193b7a0-0000-7000-8000-000000000001';