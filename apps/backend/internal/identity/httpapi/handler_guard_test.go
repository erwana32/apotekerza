package httpapi

import (
	"context"
	"encoding/json"
	nethttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/erwana32/apotekerza/internal/identity/app"
	"github.com/erwana32/apotekerza/internal/identity/domain"
	"github.com/erwana32/apotekerza/internal/platform/http/gen"
)

// rulesFor builds the permission table as the real registry does, so the test
// exercises the guard without a database: the session resolution is the only
// part the guard delegates to the service.
func rulesFor() app.PermissionRules {
	return app.PermissionRules{
		Permissions: map[string]string{
			"GetMe":           "identity.user.read",
			"CreateUser":      "identity.user.write",
			"ListPermissions": "identity.group.read",
		},
		Public: map[string]struct{}{"Login": {}},
	}
}

// runGuard drives one request through the guard and returns the recorded status
// plus body.
func runGuard(t *testing.T, middleware gen.StrictMiddlewareFunc, operationID string, cookie *nethttp.Cookie) (int, gen.Error) {
	t.Helper()

	handler := middleware(func(ctx context.Context, w nethttp.ResponseWriter, r *nethttp.Request, request any) (any, error) {
		return "ok", nil
	}, operationID)

	req := httptest.NewRequest(nethttp.MethodGet, "/api/v1/identity/me", nil)
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()

	_, err := handler(req.Context(), rec, req, nil)
	if err != nil {
		WriteDomainError(rec, req, err)
	}

	var body gen.Error
	if rec.Body.Len() > 0 {
		if decodeErr := json.Unmarshal(rec.Body.Bytes(), &body); decodeErr != nil {
			t.Fatalf("error body is not JSON: %v (%s)", decodeErr, rec.Body.String())
		}
	}
	return rec.Code, body
}

// The guard must run for the real handler set: a nil service is enough here
// because every case below is refused before the service is reached.
func newTestMiddleware(rules app.PermissionRules) gen.StrictMiddlewareFunc {
	return NewGuard(nil, rules)
}

func TestGuard_RefusesOperationWithoutDeclaredPermission(t *testing.T) {
	status, body := runGuard(t, newTestMiddleware(rulesFor()), "UndeclaredOperation", nil)
	if status != nethttp.StatusForbidden {
		t.Fatalf("status = %d, want 403", status)
	}
	if body.Code != "PERMISSION_DENIED" {
		t.Fatalf("code = %q, want PERMISSION_DENIED", body.Code)
	}
}

func TestGuard_RefusesRequestWithoutSession(t *testing.T) {
	status, body := runGuard(t, newTestMiddleware(rulesFor()), "GetMe", nil)
	if status != nethttp.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", status)
	}
	if body.Code != "SESSION_EXPIRED" {
		t.Fatalf("code = %q, want SESSION_EXPIRED", body.Code)
	}
}

func TestGuard_AllowsPublicOperationWithoutSession(t *testing.T) {
	var reached bool
	middleware := newTestMiddleware(rulesFor())
	handler := middleware(func(ctx context.Context, w nethttp.ResponseWriter, r *nethttp.Request, request any) (any, error) {
		reached = true
		return "ok", nil
	}, "Login")

	req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/identity/login", nil)
	rec := httptest.NewRecorder()
	if _, err := handler(req.Context(), rec, req, nil); err != nil {
		t.Fatalf("public operation returned %v", err)
	}
	if !reached {
		t.Fatal("the guard must not block a public operation")
	}
}

func TestGuard_EmptyPermissionValueIsNotTreatedAsPublic(t *testing.T) {
	rules := app.PermissionRules{
		Permissions: map[string]string{"GetMe": ""},
		Public:      map[string]struct{}{},
	}
	status, body := runGuard(t, newTestMiddleware(rules), "GetMe", nil)
	if status != nethttp.StatusForbidden || body.Code != "PERMISSION_DENIED" {
		t.Fatalf("status = %d code = %q, want 403 PERMISSION_DENIED", status, body.Code)
	}
}

func TestWriteError_MapsDomainErrorToStableShape(t *testing.T) {
	rec := httptest.NewRecorder()
	writeError(rec, domain.ErrPermissionDenied)

	if rec.Code != nethttp.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
	var body gen.Error
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Code != "PERMISSION_DENIED" || body.Message != "permission denied" {
		t.Fatalf("body = %+v", body)
	}
}

func TestWriteError_ValidationCarriesFieldDetail(t *testing.T) {
	rec := httptest.NewRecorder()
	writeError(rec, domain.ErrValidation("username", "must be 3-64 characters"))

	if rec.Code != nethttp.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", rec.Code)
	}
	var body gen.Error
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Code != "VALIDATION_FAILED" {
		t.Fatalf("code = %q, want VALIDATION_FAILED", body.Code)
	}
	if body.Details == nil {
		t.Fatal("details must carry the offending field")
	}
	if field, _ := (*body.Details)["field"].(string); field != "username" {
		t.Fatalf("details.field = %v, want username", (*body.Details)["field"])
	}
}
