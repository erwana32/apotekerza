package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/erwana32/apotekerza/internal/identity/store/gen"
)

// ListGroupIDsForUser returns the ids of the groups a user holds.
func (q *Queries) ListGroupIDsForUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	ids, err := q.Queries.ListGroupIDsForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("store: list group ids for user: %w", err)
	}
	return ids, nil
}

// ListGroupIDsForUsers returns the group ids of several users in one query, so
// a page of users costs one extra query instead of one per user.
func (q *Queries) ListGroupIDsForUsers(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID][]uuid.UUID, error) {
	result := make(map[uuid.UUID][]uuid.UUID, len(userIDs))
	rows, err := q.Queries.ListGroupIDsForUsers(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("store: list group ids for users: %w", err)
	}
	for _, row := range rows {
		result[row.UserID] = append(result[row.UserID], row.AccessGroupID)
	}
	return result, nil
}

// DeleteUserGroups removes every group of a user, so membership can be replaced.
func (q *Queries) DeleteUserGroups(ctx context.Context, userID uuid.UUID) error {
	if err := q.Queries.DeleteUserGroups(ctx, userID); err != nil {
		return fmt.Errorf("store: delete user groups: %w", err)
	}
	return nil
}

// InsertUserGroup adds the user to a group; an existing membership is a no-op.
func (q *Queries) InsertUserGroup(ctx context.Context, params gen.InsertUserGroupParams) error {
	if err := q.Queries.InsertUserGroup(ctx, params); err != nil {
		return fmt.Errorf("store: insert user group: %w", err)
	}
	return nil
}
