package main

import (
	"context"

	"github.com/erwana32/apotekerza/internal/identity"
	"github.com/erwana32/apotekerza/internal/identity/app"
	identityhttp "github.com/erwana32/apotekerza/internal/identity/httpapi"
	platformhttp "github.com/erwana32/apotekerza/internal/platform/http"
	"github.com/erwana32/apotekerza/internal/platform/http/gen"
)

// apiHandler implements the generated strict server interface: the platform
// health operation plus everything in the identity module.
type apiHandler struct {
	*platformhttp.StrictHandler
	identity *identityhttp.Handler
}

func newAPIHandler(module *identity.Module) *apiHandler {
	return &apiHandler{
		StrictHandler: platformhttp.NewStrictHandler(),
		identity:      identityhttp.NewHandler(module),
	}
}

// The generated interface requires every operation on one receiver, so the
// identity handler's methods are forwarded here.

func (h *apiHandler) Login(ctx context.Context, request gen.LoginRequestObject) (gen.LoginResponseObject, error) {
	return h.identity.Login(ctx, request)
}

func (h *apiHandler) Logout(ctx context.Context, request gen.LogoutRequestObject) (gen.LogoutResponseObject, error) {
	return h.identity.Logout(ctx, request)
}

func (h *apiHandler) GetMe(ctx context.Context, request gen.GetMeRequestObject) (gen.GetMeResponseObject, error) {
	return h.identity.GetMe(ctx, request)
}

func (h *apiHandler) ListUsers(ctx context.Context, request gen.ListUsersRequestObject) (gen.ListUsersResponseObject, error) {
	return h.identity.ListUsers(ctx, request)
}

func (h *apiHandler) CreateUser(ctx context.Context, request gen.CreateUserRequestObject) (gen.CreateUserResponseObject, error) {
	return h.identity.CreateUser(ctx, request)
}

func (h *apiHandler) GetUser(ctx context.Context, request gen.GetUserRequestObject) (gen.GetUserResponseObject, error) {
	return h.identity.GetUser(ctx, request)
}

func (h *apiHandler) UpdateUser(ctx context.Context, request gen.UpdateUserRequestObject) (gen.UpdateUserResponseObject, error) {
	return h.identity.UpdateUser(ctx, request)
}

func (h *apiHandler) ListAccessGroups(ctx context.Context, request gen.ListAccessGroupsRequestObject) (gen.ListAccessGroupsResponseObject, error) {
	return h.identity.ListAccessGroups(ctx, request)
}

func (h *apiHandler) CreateAccessGroup(ctx context.Context, request gen.CreateAccessGroupRequestObject) (gen.CreateAccessGroupResponseObject, error) {
	return h.identity.CreateAccessGroup(ctx, request)
}

func (h *apiHandler) GetAccessGroup(ctx context.Context, request gen.GetAccessGroupRequestObject) (gen.GetAccessGroupResponseObject, error) {
	return h.identity.GetAccessGroup(ctx, request)
}

func (h *apiHandler) UpdateAccessGroup(ctx context.Context, request gen.UpdateAccessGroupRequestObject) (gen.UpdateAccessGroupResponseObject, error) {
	return h.identity.UpdateAccessGroup(ctx, request)
}

func (h *apiHandler) ListPermissions(ctx context.Context, request gen.ListPermissionsRequestObject) (gen.ListPermissionsResponseObject, error) {
	return h.identity.ListPermissions(ctx, request)
}

func (h *apiHandler) ListAuditEvents(ctx context.Context, request gen.ListAuditEventsRequestObject) (gen.ListAuditEventsResponseObject, error) {
	return h.identity.ListAuditEvents(ctx, request)
}

// newStrictAPI builds the strict server with the identity guard installed as a
// middleware, so the permission check derived from the spec runs before every
// operation and cannot be forgotten for a new endpoint.
func newStrictAPI(module *identity.Module) gen.ServerInterface {
	return gen.NewStrictHandlerWithOptions(newAPIHandler(module),
		[]gen.StrictMiddlewareFunc{identityhttp.NewGuard(module.Service(), app.NewPermissionRegistry().Rules())},
		gen.StrictHTTPServerOptions{
			RequestErrorHandlerFunc:  identityhttp.WriteRequestError,
			ResponseErrorHandlerFunc: identityhttp.WriteDomainError,
		})
}
