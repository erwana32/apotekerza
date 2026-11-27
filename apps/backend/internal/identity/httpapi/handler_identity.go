package httpapi

import (
	"context"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/erwana32/apotekerza/internal/identity"
	"github.com/erwana32/apotekerza/internal/identity/app"
	"github.com/erwana32/apotekerza/internal/identity/domain"
	"github.com/erwana32/apotekerza/internal/platform/http/gen"
	"github.com/erwana32/apotekerza/internal/platform/pagination"
)

// Handler implements the generated strict server interface for identity
// operations. It is thin: read input, call the service, map the result.
type Handler struct {
	module *identity.Module
}

// NewHandler builds the identity HTTP handler.
func NewHandler(module *identity.Module) *Handler {
	return &Handler{module: module}
}

// Login is public (x-public: true). It sets the session cookie and returns the
// signed-in user.
func (h *Handler) Login(ctx context.Context, request gen.LoginRequestObject) (gen.LoginResponseObject, error) {
	if request.Body == nil {
		return nil, domain.ErrValidation("body", "is required")
	}
	password := ""
	if request.Body.Password != nil {
		password = *request.Body.Password
	}
	result, err := h.module.Service().Authenticate(ctx, request.Body.Username, password)
	if err != nil {
		return nil, err
	}

	IssueSessionCookie(ctx, result.Token, result.ExpiresAt)
	return gen.Login200JSONResponse(gen.Me{
		User:        toAPIUser(result.User),
		Permissions: nonNilStrings(result.Permissions),
	}), nil
}

// Logout revokes the session cookie. It is idempotent.
func (h *Handler) Logout(ctx context.Context, _ gen.LogoutRequestObject) (gen.LogoutResponseObject, error) {
	token, _ := SessionToken(ctx)
	actor, _ := SessionFrom(ctx)
	if err := h.module.Service().Logout(ctx, token, &actor); err != nil {
		return nil, err
	}
	ClearSessionCookie(ctx)
	return gen.Logout204Response{}, nil
}

// GetMe returns the signed-in user and its permissions.
func (h *Handler) GetMe(ctx context.Context, _ gen.GetMeRequestObject) (gen.GetMeResponseObject, error) {
	session, ok := SessionFrom(ctx)
	if !ok {
		return nil, domain.ErrSessionExpired
	}
	user, err := h.module.Service().GetUser(ctx, session.UserID)
	if err != nil {
		return nil, err
	}
	return gen.GetMe200JSONResponse(gen.Me{
		User:        toAPIUser(user),
		Permissions: nonNilStrings(session.Permissions),
	}), nil
}

// ListUsers returns one page of users.
func (h *Handler) ListUsers(ctx context.Context, request gen.ListUsersRequestObject) (gen.ListUsersResponseObject, error) {
	page, err := h.module.Service().ListUsers(ctx, app.ListUsersInput{
		Limit:  limitOrDefault(request.Params.Limit),
		Search: stringOrEmpty(request.Params.Search),
		Cursor: cursorOrEmpty(request.Params.Cursor),
	})
	if err != nil {
		return nil, err
	}
	items := make([]gen.User, 0, len(page.Items))
	for _, user := range page.Items {
		items = append(items, toAPIUser(user))
	}
	return gen.ListUsers200JSONResponse(gen.UserList{Items: items, NextCursor: page.NextCursor}), nil
}

// CreateUser creates a user.
func (h *Handler) CreateUser(ctx context.Context, request gen.CreateUserRequestObject) (gen.CreateUserResponseObject, error) {
	if request.Body == nil {
		return nil, domain.ErrValidation("body", "is required")
	}
	actor, _ := SessionFrom(ctx)
	password := ""
	if request.Body.Password != nil {
		password = *request.Body.Password
	}
	user, err := h.module.Service().CreateUser(ctx, &actor, app.CreateUserInput{
		CompanyID:   actor.CompanyID,
		Username:    request.Body.Username,
		DisplayName: request.Body.DisplayName,
		Password:    password,
		GroupIDs:    toUUIDs(request.Body.GroupIds),
	})
	if err != nil {
		return nil, err
	}
	return gen.CreateUser201JSONResponse(toAPIUser(user)), nil
}

// GetUser returns one user.
func (h *Handler) GetUser(ctx context.Context, request gen.GetUserRequestObject) (gen.GetUserResponseObject, error) {
	user, err := h.module.Service().GetUser(ctx, uuidOf(request.UserId))
	if err != nil {
		return nil, err
	}
	return gen.GetUser200JSONResponse(toAPIUser(user)), nil
}

// UpdateUser updates a user.
func (h *Handler) UpdateUser(ctx context.Context, request gen.UpdateUserRequestObject) (gen.UpdateUserResponseObject, error) {
	if request.Body == nil {
		return nil, domain.ErrValidation("body", "is required")
	}
	actor, _ := SessionFrom(ctx)
	input := app.UpdateUserInput{
		ID:          uuidOf(request.UserId),
		DisplayName: request.Body.DisplayName,
		IsActive:    request.Body.IsActive,
	}
	if request.Body.GroupIds != nil {
		groupIDs := toUUIDs(request.Body.GroupIds)
		input.GroupIDs = &groupIDs
	}
	user, err := h.module.Service().UpdateUser(ctx, &actor, input)
	if err != nil {
		return nil, err
	}
	return gen.UpdateUser200JSONResponse(toAPIUser(user)), nil
}

// ListAccessGroups returns one page of access groups.
func (h *Handler) ListAccessGroups(ctx context.Context, request gen.ListAccessGroupsRequestObject) (gen.ListAccessGroupsResponseObject, error) {
	page, err := h.module.Service().ListAccessGroups(ctx,
		limitOrDefault(request.Params.Limit),
		stringOrEmpty(request.Params.Search),
		cursorOrEmpty(request.Params.Cursor),
	)
	if err != nil {
		return nil, err
	}
	items := make([]gen.AccessGroup, 0, len(page.Items))
	for _, group := range page.Items {
		items = append(items, toAPIGroup(group))
	}
	return gen.ListAccessGroups200JSONResponse(gen.AccessGroupList{Items: items, NextCursor: page.NextCursor}), nil
}

// CreateAccessGroup creates an access group.
func (h *Handler) CreateAccessGroup(ctx context.Context, request gen.CreateAccessGroupRequestObject) (gen.CreateAccessGroupResponseObject, error) {
	if request.Body == nil {
		return nil, domain.ErrValidation("body", "is required")
	}
	actor, _ := SessionFrom(ctx)
	group, err := h.module.Service().CreateAccessGroup(ctx, &actor, app.CreateAccessGroupInput{
		Name:        request.Body.Name,
		Permissions: stringsOrEmpty(request.Body.Permissions),
	})
	if err != nil {
		return nil, err
	}
	return gen.CreateAccessGroup201JSONResponse(toAPIGroup(group)), nil
}

// GetAccessGroup returns one access group.
func (h *Handler) GetAccessGroup(ctx context.Context, request gen.GetAccessGroupRequestObject) (gen.GetAccessGroupResponseObject, error) {
	group, err := h.module.Service().GetAccessGroup(ctx, uuidOf(request.AccessGroupId))
	if err != nil {
		return nil, err
	}
	return gen.GetAccessGroup200JSONResponse(toAPIGroup(group)), nil
}

// UpdateAccessGroup updates an access group and optionally its permissions.
func (h *Handler) UpdateAccessGroup(ctx context.Context, request gen.UpdateAccessGroupRequestObject) (gen.UpdateAccessGroupResponseObject, error) {
	if request.Body == nil {
		return nil, domain.ErrValidation("body", "is required")
	}
	actor, _ := SessionFrom(ctx)
	input := app.UpdateAccessGroupInput{
		ID:       uuidOf(request.AccessGroupId),
		Name:     request.Body.Name,
		IsActive: request.Body.IsActive,
	}
	if request.Body.Permissions != nil {
		permissions := *request.Body.Permissions
		input.Permissions = &permissions
	}
	group, err := h.module.Service().UpdateAccessGroup(ctx, &actor, input)
	if err != nil {
		return nil, err
	}
	return gen.UpdateAccessGroup200JSONResponse(toAPIGroup(group)), nil
}

// ListPermissions returns the permission registry.
func (h *Handler) ListPermissions(ctx context.Context, _ gen.ListPermissionsRequestObject) (gen.ListPermissionsResponseObject, error) {
	return gen.ListPermissions200JSONResponse(gen.PermissionList{
		Items: nonNilStrings(h.module.Service().ListPermissions()),
	}), nil
}

// ListAuditEvents returns one page of the audit trail.
func (h *Handler) ListAuditEvents(ctx context.Context, request gen.ListAuditEventsRequestObject) (gen.ListAuditEventsResponseObject, error) {
	input := app.ListAuditEventsInput{
		Limit:  limitOrDefault(request.Params.Limit),
		Cursor: cursorOrEmpty(request.Params.Cursor),
	}
	if request.Params.Entity != nil {
		input.Entity = *request.Params.Entity
	}
	if request.Params.EntityId != nil {
		id := uuidOf(*request.Params.EntityId)
		input.EntityID = &id
	}
	page, err := h.module.Service().ListAuditEvents(ctx, input)
	if err != nil {
		return nil, err
	}
	items := make([]gen.AuditEvent, 0, len(page.Items))
	for _, event := range page.Items {
		items = append(items, toAPIAuditEvent(event))
	}
	return gen.ListAuditEvents200JSONResponse(gen.AuditEventList{Items: items, NextCursor: page.NextCursor}), nil
}

// --- small helpers -----------------------------------------------------------

func limitOrDefault(limit *gen.Limit) int {
	if limit == nil {
		return pagination.DefaultLimit
	}
	value := int(*limit)
	if value < 1 || value > pagination.MaxLimit {
		return pagination.DefaultLimit
	}
	return value
}

func cursorOrEmpty(cursor *gen.Cursor) string {
	if cursor == nil {
		return ""
	}
	return string(*cursor)
}

func stringOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func stringsOrEmpty(value *[]string) []string {
	if value == nil {
		return nil
	}
	return *value
}

func nonNilStrings(value []string) []string {
	if value == nil {
		return []string{}
	}
	return value
}

func toUUIDs(ids *[]openapi_types.UUID) []uuid.UUID {
	if ids == nil {
		return nil
	}
	result := make([]uuid.UUID, 0, len(*ids))
	for _, id := range *ids {
		result = append(result, uuidOf(id))
	}
	return result
}

// uuidOf converts the API uuid type to google/uuid without a string round trip.
func uuidOf(id openapi_types.UUID) uuid.UUID {
	return uuid.UUID(id)
}
