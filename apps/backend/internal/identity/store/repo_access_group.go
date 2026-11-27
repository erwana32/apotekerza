package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/erwana32/apotekerza/internal/identity/store/gen"
)

// GetAccessGroupByID fetches a group with its permissions, mapping a missing
// row to ErrGroupNotFound.
func (q *Queries) GetAccessGroupByID(ctx context.Context, id uuid.UUID) (gen.IdentityAccessGroup, error) {
	group, err := q.Queries.GetAccessGroupByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return gen.IdentityAccessGroup{}, fmt.Errorf("store: group %s: %w", id, ErrGroupNotFound)
	}
	if err != nil {
		return gen.IdentityAccessGroup{}, fmt.Errorf("store: get group by id: %w", err)
	}
	return group, nil
}

// ListAccessGroups returns one page of groups ordered by name.
func (q *Queries) ListAccessGroups(ctx context.Context, params gen.ListAccessGroupsParams) ([]gen.IdentityAccessGroup, error) {
	groups, err := q.Queries.ListAccessGroups(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("store: list access groups: %w", err)
	}
	return groups, nil
}

// InsertAccessGroup creates a group and maps a duplicate name to
// ErrDuplicateGroupName.
func (q *Queries) InsertAccessGroup(ctx context.Context, params gen.InsertAccessGroupParams) (gen.IdentityAccessGroup, error) {
	group, err := q.Queries.InsertAccessGroup(ctx, params)
	if isUniqueViolation(err, "identity_access_groups_name_key") {
		return gen.IdentityAccessGroup{}, fmt.Errorf("store: %w", ErrDuplicateGroupName)
	}
	if err != nil {
		return gen.IdentityAccessGroup{}, fmt.Errorf("store: insert access group: %w", err)
	}
	return group, nil
}

// ReplaceGroupPermissions sets the group's permission set in one transaction
// step: it deletes the old rows and inserts the new ones. Only a draft concept
// like a group is mutable; the audit trail records the change.
func (q *Queries) ReplaceGroupPermissions(ctx context.Context, groupID uuid.UUID, permissions []string) error {
	if err := q.DeleteGroupPermissions(ctx, groupID); err != nil {
		return fmt.Errorf("store: clear group permissions: %w", err)
	}
	for _, permission := range permissions {
		if err := q.InsertGroupPermission(ctx, gen.InsertGroupPermissionParams{
			AccessGroupID: groupID,
			Permission:    permission,
		}); err != nil {
			return fmt.Errorf("store: insert group permission %q: %w", permission, err)
		}
	}
	return nil
}

// ListGroupPermissions returns the permissions of one group, sorted.
func (q *Queries) ListGroupPermissions(ctx context.Context, groupID uuid.UUID) ([]string, error) {
	permissions, err := q.Queries.ListGroupPermissions(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("store: list group permissions: %w", err)
	}
	return permissions, nil
}

// ListGroupPermissionsForGroups returns the permissions of several groups in
// one query, so a page of groups costs one extra query instead of one per row.
func (q *Queries) ListGroupPermissionsForGroups(ctx context.Context, groupIDs []uuid.UUID) (map[uuid.UUID][]string, error) {
	result := make(map[uuid.UUID][]string, len(groupIDs))
	for _, groupID := range groupIDs {
		permissions, err := q.Queries.ListGroupPermissions(ctx, groupID)
		if err != nil {
			return nil, fmt.Errorf("store: list group permissions: %w", err)
		}
		result[groupID] = permissions
	}
	return result, nil
}

// GetBuiltinAccessGroups returns the seeded builtin groups, such as
// Administrator. It is used by the createadmin command and by tests.
func (q *Queries) GetBuiltinAccessGroups(ctx context.Context) ([]gen.IdentityAccessGroup, error) {
	groups, err := q.Queries.GetBuiltinAccessGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("store: list builtin access groups: %w", err)
	}
	return groups, nil
}
