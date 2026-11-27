package http

import (
	"context"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/erwana32/apotekerza/internal/platform/http/gen"
)

// healthOnlyServer implements the generated strict interface for the health
// operation only. Since the identity endpoints joined the spec, the package's
// StrictHandler is composed by cmd/server; a test that exercises one route does
// not need the other operations.
type healthOnlyServer struct {
	*StrictHandler
}

func (healthOnlyServer) Login(context.Context, gen.LoginRequestObject) (gen.LoginResponseObject, error) {
	return nil, nil
}
func (healthOnlyServer) Logout(context.Context, gen.LogoutRequestObject) (gen.LogoutResponseObject, error) {
	return nil, nil
}
func (healthOnlyServer) GetMe(context.Context, gen.GetMeRequestObject) (gen.GetMeResponseObject, error) {
	return nil, nil
}
func (healthOnlyServer) ListUsers(context.Context, gen.ListUsersRequestObject) (gen.ListUsersResponseObject, error) {
	return nil, nil
}
func (healthOnlyServer) CreateUser(context.Context, gen.CreateUserRequestObject) (gen.CreateUserResponseObject, error) {
	return nil, nil
}
func (healthOnlyServer) GetUser(context.Context, gen.GetUserRequestObject) (gen.GetUserResponseObject, error) {
	return nil, nil
}
func (healthOnlyServer) UpdateUser(context.Context, gen.UpdateUserRequestObject) (gen.UpdateUserResponseObject, error) {
	return nil, nil
}
func (healthOnlyServer) ListAccessGroups(context.Context, gen.ListAccessGroupsRequestObject) (gen.ListAccessGroupsResponseObject, error) {
	return nil, nil
}
func (healthOnlyServer) CreateAccessGroup(context.Context, gen.CreateAccessGroupRequestObject) (gen.CreateAccessGroupResponseObject, error) {
	return nil, nil
}
func (healthOnlyServer) GetAccessGroup(context.Context, gen.GetAccessGroupRequestObject) (gen.GetAccessGroupResponseObject, error) {
	return nil, nil
}
func (healthOnlyServer) UpdateAccessGroup(context.Context, gen.UpdateAccessGroupRequestObject) (gen.UpdateAccessGroupResponseObject, error) {
	return nil, nil
}
func (healthOnlyServer) ListPermissions(context.Context, gen.ListPermissionsRequestObject) (gen.ListPermissionsResponseObject, error) {
	return nil, nil
}
func (healthOnlyServer) ListAuditEvents(context.Context, gen.ListAuditEventsRequestObject) (gen.ListAuditEventsResponseObject, error) {
	return nil, nil
}

func TestGetHealth_ReturnsOKStatus(t *testing.T) {
	response, err := NewStrictHandler().GetHealth(context.Background(), gen.GetHealthRequestObject{})
	if err != nil {
		t.Fatalf("GetHealth() error = %v", err)
	}
	health, ok := response.(gen.GetHealth200JSONResponse)
	if !ok {
		t.Fatalf("GetHealth() response = %T, want GetHealth200JSONResponse", response)
	}
	if health.Status != "ok" {
		t.Fatalf("status = %q, want ok", health.Status)
	}
}

func TestGetHealth_GeneratedRouteReturnsJSON(t *testing.T) {
	mux := nethttp.NewServeMux()
	handler := gen.HandlerFromMux(gen.NewStrictHandler(healthOnlyServer{NewStrictHandler()}, nil), mux)

	req := httptest.NewRequest(nethttp.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != nethttp.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"status":"ok"}` {
		t.Fatalf("body = %q, want %q", got, `{"status":"ok"}`)
	}
}
