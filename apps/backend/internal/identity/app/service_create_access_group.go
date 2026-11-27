package app

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/erwana32/apotekerza/internal/identity/domain"
	"github.com/erwana32/apotekerza/internal/identity/store"
	"github.com/erwana32/apotekerza/internal/identity/store/gen"
	"github.com/erwana32/apotekerza/internal/platform/db"
	"github.com/erwana32/apotekerza/internal/platform/ids"
)

// CreateAccessGroupInput is the validated input of CreateAccessGroup.
type CreateAccessGroupInput struct {
	Name        string
	Permissions []string
}

// CreateAccessGroup creates a group with its permissions in one transaction.
// Every permission must exist in the registry, so a typo cannot create a dead
// permission that no endpoint checks.
func (s *Service) CreateAccessGroup(ctx context.Context, actor *SessionContext, in CreateAccessGroupInput) (domain.AccessGroup, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return domain.AccessGroup{}, domain.ErrValidation("name", "must not be empty")
	}
	permissions, err := s.validatePermissions(in.Permissions)
	if err != nil {
		return domain.AccessGroup{}, err
	}

	var created domain.AccessGroup
	err = db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		queries := store.New(tx)
		groupID, err := ids.New()
		if err != nil {
			return err
		}
		group, err := queries.InsertAccessGroup(ctx, gen.InsertAccessGroupParams{
			ID:        groupID,
			Name:      name,
			IsBuiltin: false,
			IsActive:  true,
		})
		if errors.Is(err, store.ErrDuplicateGroupName) {
			return domain.ErrGroupNameTaken
		}
		if err != nil {
			return err
		}
		if err := queries.ReplaceGroupPermissions(ctx, group.ID, permissions); err != nil {
			return err
		}
		if err := recordAudit(ctx, queries, auditInput{
			companyID: actorCompanyID(actor),
			actorID:   actorUserID(actor),
			action:    domain.ActionCreate,
			entity:    domain.EntityAccessGroup,
			entityID:  &group.ID,
			changes: map[string]any{
				"name":        map[string]any{"from": nil, "to": group.Name},
				"permissions": map[string]any{"from": nil, "to": permissions},
			},
		}); err != nil {
			return err
		}
		created = toDomainGroup(group, permissions)
		return nil
	})
	if err != nil {
		return domain.AccessGroup{}, asDomainError(err)
	}
	return created, nil
}

// GetAccessGroup returns one group with its permissions.
func (s *Service) GetAccessGroup(ctx context.Context, id uuid.UUID) (domain.AccessGroup, error) {
	queries := store.New(s.pool)
	group, err := queries.GetAccessGroupByID(ctx, id)
	if errors.Is(err, store.ErrGroupNotFound) {
		return domain.AccessGroup{}, domain.ErrAccessGroupNotFound
	}
	if err != nil {
		return domain.AccessGroup{}, asDomainError(err)
	}
	permissions, err := queries.ListGroupPermissions(ctx, id)
	if err != nil {
		return domain.AccessGroup{}, asDomainError(err)
	}
	return toDomainGroup(group, permissions), nil
}

// AccessGroupPage is one page of groups plus the cursor for the next page.
type AccessGroupPage struct {
	Items      []domain.AccessGroup
	NextCursor *string
}

// ListAccessGroups returns one page of groups, keyset-paginated on name.
func (s *Service) ListAccessGroups(ctx context.Context, limit int, search, cursor string) (AccessGroupPage, error) {
	var afterName *string
	if cursor != "" {
		decoded, err := decodeCursor(cursor)
		if err != nil {
			return AccessGroupPage{}, domain.ErrValidation("cursor", "is not valid")
		}
		afterName = &decoded
	}
	var searchValue *string
	if strings.TrimSpace(search) != "" {
		value := strings.TrimSpace(search)
		searchValue = &value
	}

	rows, err := store.New(s.pool).ListAccessGroups(ctx, gen.ListAccessGroupsParams{
		Search:    searchValue,
		AfterName: afterName,
		PageSize:  int32(limit + 1),
	})
	if err != nil {
		return AccessGroupPage{}, asDomainError(err)
	}

	page := AccessGroupPage{}
	if len(rows) > limit {
		rows = rows[:limit]
		next := rows[len(rows)-1].Name
		encoded := encodeCursor(next)
		page.NextCursor = &encoded
	}

	queries := store.New(s.pool)
	for _, row := range rows {
		permissions, err := queries.ListGroupPermissions(ctx, row.ID)
		if err != nil {
			return AccessGroupPage{}, asDomainError(err)
		}
		page.Items = append(page.Items, toDomainGroup(row, permissions))
	}
	return page, nil
}

// UpdateAccessGroupInput is the editable part of a group.
type UpdateAccessGroupInput struct {
	ID       uuid.UUID
	Name     string
	IsActive bool
	// Permissions replaces the set when non-nil.
	Permissions *[]string
}

// UpdateAccessGroup renames a group, changes its active flag, and optionally
// replaces its permissions. The builtin Administrator group is immutable: it
// must always hold every permission, so weakening it is refused rather than
// silently ignored.
func (s *Service) UpdateAccessGroup(ctx context.Context, actor *SessionContext, in UpdateAccessGroupInput) (domain.AccessGroup, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return domain.AccessGroup{}, domain.ErrValidation("name", "must not be empty")
	}
	var permissions []string
	if in.Permissions != nil {
		var err error
		permissions, err = s.validatePermissions(*in.Permissions)
		if err != nil {
			return domain.AccessGroup{}, err
		}
	}

	var updated domain.AccessGroup
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		queries := store.New(tx)

		before, err := queries.GetAccessGroupByID(ctx, in.ID)
		if errors.Is(err, store.ErrGroupNotFound) {
			return domain.ErrAccessGroupNotFound
		}
		if err != nil {
			return err
		}

		if before.IsBuiltin && (before.Name != name || !in.IsActive || in.Permissions != nil) {
			return domain.ErrBuiltinGroupImmutable
		}

		group, err := queries.UpdateAccessGroup(ctx, gen.UpdateAccessGroupParams{
			ID:       in.ID,
			Name:     name,
			IsActive: in.IsActive,
		})
		if errors.Is(err, store.ErrDuplicateGroupName) {
			return domain.ErrGroupNameTaken
		}
		if err != nil {
			return err
		}

		if in.Permissions != nil {
			if err := queries.ReplaceGroupPermissions(ctx, group.ID, permissions); err != nil {
				return err
			}
		}
		current, err := queries.ListGroupPermissions(ctx, group.ID)
		if err != nil {
			return err
		}

		changes := map[string]any{}
		if before.Name != group.Name {
			changes["name"] = map[string]any{"from": before.Name, "to": group.Name}
		}
		if before.IsActive != group.IsActive {
			changes["isActive"] = map[string]any{"from": before.IsActive, "to": group.IsActive}
		}
		action := domain.ActionUpdate
		if before.IsActive && !group.IsActive {
			action = domain.ActionDeactivate
		}
		if err := recordAudit(ctx, queries, auditInput{
			companyID: actorCompanyID(actor),
			actorID:   actorUserID(actor),
			action:    action,
			entity:    domain.EntityAccessGroup,
			entityID:  &group.ID,
			changes:   changes,
		}); err != nil {
			return err
		}

		updated = toDomainGroup(group, current)
		return nil
	})
	if err != nil {
		return domain.AccessGroup{}, asDomainError(err)
	}
	return updated, nil
}

// ListPermissions returns the permission registry, sorted.
func (s *Service) ListPermissions() []string {
	return s.registry.All()
}

// validatePermissions de-duplicates and sorts a permission set and rejects any
// value that is not in the registry.
func (s *Service) validatePermissions(wanted []string) ([]string, error) {
	seen := make(map[string]struct{}, len(wanted))
	result := make([]string, 0, len(wanted))
	for _, permission := range wanted {
		permission = strings.TrimSpace(permission)
		if permission == "" {
			continue
		}
		if !s.registry.Has(permission) {
			return nil, domain.ErrUnknownPermission.WithDetail("permission", permission)
		}
		if _, duplicate := seen[permission]; duplicate {
			continue
		}
		seen[permission] = struct{}{}
		result = append(result, permission)
	}
	sort.Strings(result)
	return result, nil
}

// toDomainGroup converts a stored row plus its permissions.
func toDomainGroup(group gen.IdentityAccessGroup, permissions []string) domain.AccessGroup {
	return domain.AccessGroup{
		ID:          group.ID,
		Name:        group.Name,
		IsBuiltin:   group.IsBuiltin,
		IsActive:    group.IsActive,
		Permissions: permissions,
		CreatedAt:   group.CreatedAt,
		UpdatedAt:   group.UpdatedAt,
	}
}

// actorCompanyID returns the actor's company, or uuid.Nil for a system action.
func actorCompanyID(actor *SessionContext) uuid.UUID {
	if actor == nil {
		return uuid.Nil
	}
	return actor.CompanyID
}
