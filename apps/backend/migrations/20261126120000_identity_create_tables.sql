-- +goose Up

-- Identity owns users, sessions, access groups, and the audit trail.
-- These are master records, not documents (ADR-0010: no lifecycle status and no
-- document number; records are deactivated, never deleted).
-- One schema, tables prefixed by module, and company_id as the only scope column
-- (ADR-0001). identity_users and identity_audit_log carry company_id; access
-- groups are global (as in Odoo) and link/session tables inherit their scope
-- through foreign keys. The foreign key from company_id to masterdata_companies
-- is added by the masterdata migration, once that table exists.

CREATE TABLE identity_users (
    id            uuid PRIMARY KEY,
    company_id    uuid NOT NULL,
    username      text NOT NULL,
    password_hash text NOT NULL,
    display_name  text NOT NULL,
    is_active     boolean NOT NULL DEFAULT true,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT identity_users_username_lowercase CHECK (username = lower(username)),
    CONSTRAINT identity_users_username_format CHECK (username ~ '^[a-z0-9._-]{3,64}$')
);

-- Login looks a user up by username alone, before any company is known, so the
-- username is unique across the installation (as Odoo's login). The service
-- lowercases the input.
CREATE UNIQUE INDEX identity_users_username_key ON identity_users (username);

CREATE TABLE identity_access_groups (
    id         uuid PRIMARY KEY,
    name       text NOT NULL,
    is_builtin boolean NOT NULL DEFAULT false,
    is_active  boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- Access groups are global: the name is unique across the installation.
CREATE UNIQUE INDEX identity_access_groups_name_key ON identity_access_groups (name);

CREATE TABLE identity_group_permissions (
    access_group_id uuid NOT NULL REFERENCES identity_access_groups (id) ON DELETE RESTRICT,
    permission      text NOT NULL,
    created_at      timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (access_group_id, permission)
);

-- The registry check (is this permission known?) reads by permission.
CREATE INDEX identity_group_permissions_permission_idx ON identity_group_permissions (permission);

CREATE TABLE identity_user_groups (
    user_id         uuid NOT NULL REFERENCES identity_users (id) ON DELETE RESTRICT,
    access_group_id uuid NOT NULL REFERENCES identity_access_groups (id) ON DELETE RESTRICT,
    created_at      timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, access_group_id)
);

-- "Who holds this group?" is answered per group.
CREATE INDEX identity_user_groups_access_group_id_idx ON identity_user_groups (access_group_id);

CREATE TABLE identity_sessions (
    id         uuid PRIMARY KEY,
    user_id    uuid NOT NULL REFERENCES identity_users (id) ON DELETE RESTRICT,
    token_hash bytea NOT NULL,
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT identity_sessions_token_hash_sha256 CHECK (octet_length(token_hash) = 32)
);

-- The cookie carries the token; only its SHA-256 is stored, so a database read
-- cannot be replayed as a session.
CREATE UNIQUE INDEX identity_sessions_token_hash_key ON identity_sessions (token_hash);
CREATE INDEX identity_sessions_user_id_idx ON identity_sessions (user_id);
CREATE INDEX identity_sessions_expires_at_idx ON identity_sessions (expires_at);

-- The audit trail is append-only (ADR-0004, ADR-0010). It is also the source of
-- the per-document history. It stores references and field changes only; never
-- personal-data contents.
CREATE TABLE identity_audit_log (
    id            uuid PRIMARY KEY,
    company_id    uuid NOT NULL,
    actor_user_id uuid REFERENCES identity_users (id) ON DELETE RESTRICT,
    action        text NOT NULL,
    entity        text NOT NULL,
    entity_id     uuid,
    changes       jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at    timestamptz NOT NULL DEFAULT now()
);

-- Document history: "what happened to this record, newest last".
CREATE INDEX identity_audit_log_entity_id_created_at_idx ON identity_audit_log (entity, entity_id, created_at);
CREATE INDEX identity_audit_log_actor_user_id_idx ON identity_audit_log (actor_user_id);
CREATE INDEX identity_audit_log_created_at_idx ON identity_audit_log (created_at);

-- +goose StatementBegin
CREATE FUNCTION identity_audit_log_append_only() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'identity_audit_log is append-only: % is not allowed', TG_OP
        USING ERRCODE = 'restrict_violation';
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER identity_audit_log_no_update_delete
    BEFORE UPDATE OR DELETE ON identity_audit_log
    FOR EACH ROW EXECUTE FUNCTION identity_audit_log_append_only();

-- +goose Down

DROP TRIGGER IF EXISTS identity_audit_log_no_update_delete ON identity_audit_log;
DROP FUNCTION IF EXISTS identity_audit_log_append_only();
DROP TABLE IF EXISTS identity_audit_log;
DROP TABLE IF EXISTS identity_sessions;
DROP TABLE IF EXISTS identity_user_groups;
DROP TABLE IF EXISTS identity_group_permissions;
DROP TABLE IF EXISTS identity_access_groups;
DROP TABLE IF EXISTS identity_users;