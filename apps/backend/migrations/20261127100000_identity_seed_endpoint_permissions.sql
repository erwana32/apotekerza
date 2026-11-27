-- +goose Up

-- Grants the built-in Administrator group every permission that
-- api/openapi.yaml declares (docs/modules/identity.md, "Seeding").
-- Permissions are data, so each endpoint migration adds its permission here
-- rather than syncing at startup.
-- The Administrator group is global; this seed references it by its fixed id
-- from 20261126121000_identity_seed_administrator.sql.

INSERT INTO identity_group_permissions (access_group_id, permission)
VALUES
    ('0193b7a0-0000-7000-8000-000000000001', 'identity.session.write'),
    ('0193b7a0-0000-7000-8000-000000000001', 'identity.user.read'),
    ('0193b7a0-0000-7000-8000-000000000001', 'identity.user.write'),
    ('0193b7a0-0000-7000-8000-000000000001', 'identity.group.read'),
    ('0193b7a0-0000-7000-8000-000000000001', 'identity.group.write'),
    ('0193b7a0-0000-7000-8000-000000000001', 'identity.audit.read')
ON CONFLICT DO NOTHING;

-- +goose Down

DELETE FROM identity_group_permissions
WHERE access_group_id = '0193b7a0-0000-7000-8000-000000000001'
  AND permission LIKE 'identity.%';