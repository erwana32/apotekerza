package app

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/erwana32/apotekerza/internal/identity/domain"
	"github.com/erwana32/apotekerza/internal/identity/store"
	"github.com/erwana32/apotekerza/internal/identity/store/gen"
	"github.com/erwana32/apotekerza/internal/platform/db"
	"github.com/erwana32/apotekerza/internal/platform/ids"
)

// usernamePattern mirrors the CHECK constraint in the migration, so validation
// fails with a clear field error before the database rejects it.
var usernamePattern = regexp.MustCompile(`^[a-z0-9._-]{3,64}$`)

// MinPasswordLength is the shortest accepted password.
const MinPasswordLength = 8

// CreateUserInput is the validated input of CreateUser.
type CreateUserInput struct {
	CompanyID   uuid.UUID
	Username    string
	DisplayName string
	Password    string
	GroupIDs    []uuid.UUID
}

// CreateUser creates a user, assigns its groups, and audits it in one
// transaction. The password is hashed with argon2id and never stored in plain.
func (s *Service) CreateUser(ctx context.Context, actor *SessionContext, in CreateUserInput) (domain.User, error) {
	username := domain.NormalizeUsername(in.Username)
	if !usernamePattern.MatchString(username) {
		return domain.User{}, domain.ErrValidation("username", "must be 3-64 characters of a-z, 0-9, dot, underscore or dash")
	}
	displayName := strings.TrimSpace(in.DisplayName)
	if displayName == "" {
		return domain.User{}, domain.ErrValidation("displayName", "must not be empty")
	}
	if len(in.Password) < MinPasswordLength {
		return domain.User{}, domain.ErrValidation("password", "must be at least 8 characters")
	}

	passwordHash, err := domain.HashPassword(in.Password)
	if err != nil {
		return domain.User{}, asDomainError(err)
	}

	var created domain.User
	err = db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		queries := store.New(tx)

		userID, err := ids.New()
		if err != nil {
			return err
		}
		user, err := queries.InsertUser(ctx, gen.InsertUserParams{
			ID:           userID,
			CompanyID:    in.CompanyID,
			Username:     username,
			PasswordHash: passwordHash,
			DisplayName:  displayName,
			IsActive:     true,
		})
		if errors.Is(err, store.ErrDuplicateUsername) {
			return domain.ErrUsernameTaken
		}
		if err != nil {
			return err
		}

		if err := s.assignGroups(ctx, queries, user.ID, in.GroupIDs); err != nil {
			return err
		}
		groupIDs, err := queries.ListGroupIDsForUser(ctx, user.ID)
		if err != nil {
			return err
		}

		// The audit row records the username and display name as references, not
		// the password: changes are for history, never for credentials.
		if err := recordAudit(ctx, queries, auditInput{
			companyID: user.CompanyID,
			actorID:   actorUserID(actor),
			action:    domain.ActionCreate,
			entity:    domain.EntityUser,
			entityID:  &user.ID,
			changes: map[string]any{
				"username":    map[string]any{"from": nil, "to": user.Username},
				"displayName": map[string]any{"from": nil, "to": user.DisplayName},
				"groupIds":    map[string]any{"from": nil, "to": groupIDs},
			},
		}); err != nil {
			return err
		}

		created = domain.User{
			ID:          user.ID,
			CompanyID:   user.CompanyID,
			Username:    user.Username,
			DisplayName: user.DisplayName,
			IsActive:    user.IsActive,
			GroupIDs:    groupIDs,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		}
		return nil
	})
	if err != nil {
		return domain.User{}, asDomainError(err)
	}
	return created, nil
}

// GetUser returns one user with its group ids.
func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (domain.User, error) {
	queries := store.New(s.pool)
	user, err := queries.GetUserByID(ctx, id)
	if errors.Is(err, store.ErrUserNotFound) {
		return domain.User{}, domain.ErrUserNotFound
	}
	if err != nil {
		return domain.User{}, asDomainError(err)
	}
	groupIDs, err := queries.ListGroupIDsForUser(ctx, id)
	if err != nil {
		return domain.User{}, asDomainError(err)
	}
	return toDomainUser(user, groupIDs), nil
}

// ListUsersInput is the paged list request.
type ListUsersInput struct {
	Limit  int
	Search string
	Cursor string
}

// UserPage is one page of users plus the cursor for the next page.
type UserPage struct {
	Items      []domain.User
	NextCursor *string
}

// ListUsers returns one page of users, keyset-paginated on username.
func (s *Service) ListUsers(ctx context.Context, in ListUsersInput) (UserPage, error) {
	var afterUsername *string
	if in.Cursor != "" {
		decoded, err := decodeCursor(in.Cursor)
		if err != nil {
			return UserPage{}, domain.ErrValidation("cursor", "is not valid")
		}
		afterUsername = &decoded
	}
	var search *string
	if strings.TrimSpace(in.Search) != "" {
		value := strings.TrimSpace(in.Search)
		search = &value
	}

	rows, err := store.New(s.pool).ListUsers(ctx, gen.ListUsersParams{
		Search:        search,
		AfterUsername: afterUsername,
		PageSize:      int32(in.Limit + 1),
	})
	if err != nil {
		return UserPage{}, asDomainError(err)
	}

	page := UserPage{}
	if len(rows) > in.Limit {
		rows = rows[:in.Limit]
		next := rows[len(rows)-1].Username
		cursor := encodeCursor(next)
		page.NextCursor = &cursor
	}

	userIDs := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		userIDs = append(userIDs, row.ID)
	}
	groupsByUser, err := store.New(s.pool).ListGroupIDsForUsers(ctx, userIDs)
	if err != nil {
		return UserPage{}, asDomainError(err)
	}
	for _, row := range rows {
		page.Items = append(page.Items, toDomainUser(row, groupsByUser[row.ID]))
	}
	return page, nil
}

// UpdateUserInput is the editable part of a user.
type UpdateUserInput struct {
	ID          uuid.UUID
	DisplayName string
	IsActive    bool
	// GroupIDs replaces the membership when non-nil.
	GroupIDs *[]uuid.UUID
}

// UpdateUser changes the display name, the active flag, and optionally the
// group membership. Deactivating a user also revokes its sessions, so an old
// cookie stops working immediately.
func (s *Service) UpdateUser(ctx context.Context, actor *SessionContext, in UpdateUserInput) (domain.User, error) {
	displayName := strings.TrimSpace(in.DisplayName)
	if displayName == "" {
		return domain.User{}, domain.ErrValidation("displayName", "must not be empty")
	}

	var updated domain.User
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		queries := store.New(tx)

		// Lock first, so two concurrent updates cannot both remove an Administrator.
		if err := queries.LockBuiltinAccessGroups(ctx); err != nil {
			return err
		}
		administratorsBefore, err := queries.CountActiveBuiltinMembers(ctx)
		if err != nil {
			return err
		}

		before, err := queries.GetUserByID(ctx, in.ID)
		if errors.Is(err, store.ErrUserNotFound) {
			return domain.ErrUserNotFound
		}
		if err != nil {
			return err
		}

		user, err := queries.UpdateUser(ctx, gen.UpdateUserParams{
			ID:          in.ID,
			DisplayName: displayName,
			IsActive:    in.IsActive,
		})
		if err != nil {
			return err
		}

		if in.GroupIDs != nil {
			if err := s.assignGroups(ctx, queries, in.ID, *in.GroupIDs); err != nil {
				return err
			}
		}
		if before.IsActive && !in.IsActive {
			if err := queries.RevokeSessionsForUser(ctx, in.ID); err != nil {
				return err
			}
		}

		// The update must not leave the installation without an active
		// Administrator: nobody could repair it through the UI. When there were
		// none before (no admin created yet), there is nothing to protect.
		if administratorsBefore > 0 {
			administratorsAfter, err := queries.CountActiveBuiltinMembers(ctx)
			if err != nil {
				return err
			}
			if administratorsAfter == 0 {
				return domain.ErrLastAdministrator
			}
		}

		groupIDs, err := queries.ListGroupIDsForUser(ctx, in.ID)
		if err != nil {
			return err
		}

		changes := map[string]any{}
		if before.DisplayName != user.DisplayName {
			changes["displayName"] = map[string]any{"from": before.DisplayName, "to": user.DisplayName}
		}
		if before.IsActive != user.IsActive {
			changes["isActive"] = map[string]any{"from": before.IsActive, "to": user.IsActive}
		}
		action := domain.ActionUpdate
		if before.IsActive && !user.IsActive {
			action = domain.ActionDeactivate
		}

		if err := recordAudit(ctx, queries, auditInput{
			companyID: user.CompanyID,
			actorID:   actorUserID(actor),
			action:    action,
			entity:    domain.EntityUser,
			entityID:  &user.ID,
			changes:   changes,
		}); err != nil {
			return err
		}

		updated = toDomainUser(user, groupIDs)
		return nil
	})
	if err != nil {
		return domain.User{}, asDomainError(err)
	}
	return updated, nil
}

// assignGroups replaces a user's membership, rejecting unknown group ids so a
// typo cannot silently grant nothing.
func (s *Service) assignGroups(ctx context.Context, queries *store.Queries, userID uuid.UUID, groupIDs []uuid.UUID) error {
	for _, groupID := range groupIDs {
		if _, err := queries.GetAccessGroupByID(ctx, groupID); errors.Is(err, store.ErrGroupNotFound) {
			return domain.ErrValidation("groupIds", "unknown access group "+groupID.String())
		} else if err != nil {
			return err
		}
	}
	if err := queries.DeleteUserGroups(ctx, userID); err != nil {
		return err
	}
	for _, groupID := range groupIDs {
		if err := queries.InsertUserGroup(ctx, gen.InsertUserGroupParams{
			UserID:        userID,
			AccessGroupID: groupID,
		}); err != nil {
			return err
		}
	}
	return nil
}

// toDomainUser converts a stored row plus its groups to the domain type. The
// password hash is deliberately not copied.
func toDomainUser(user gen.IdentityUser, groupIDs []uuid.UUID) domain.User {
	return domain.User{
		ID:          user.ID,
		CompanyID:   user.CompanyID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		IsActive:    user.IsActive,
		GroupIDs:    groupIDs,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}

// actorUserID returns the actor's id, or nil for a system action.
func actorUserID(actor *SessionContext) *uuid.UUID {
	if actor == nil {
		return nil
	}
	return &actor.UserID
}
