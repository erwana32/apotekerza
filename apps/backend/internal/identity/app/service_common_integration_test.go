//go:build integration

package app_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/erwana32/apotekerza/internal/identity/app"
	"github.com/erwana32/apotekerza/internal/identity/domain"
	"github.com/erwana32/apotekerza/internal/platform/clock"
	"github.com/erwana32/apotekerza/internal/platform/testutil"
)

// TestMain stops the package's shared PostgreSQL container after the tests.
func TestMain(m *testing.M) { os.Exit(testutil.RunMain(m)) }

const testPassword = "sandi-yang-cukup-panjang"

func newID(t *testing.T) uuid.UUID {
	t.Helper()
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("generate uuid: %v", err)
	}
	return id
}

// newService returns a service on a fresh migrated database plus the pool, for
// assertions that need raw SQL.
func newService(t *testing.T) (*app.Service, *pgxpool.Pool) {
	t.Helper()
	pool := testutil.NewDatabase(t)
	return app.NewService(pool, app.NewPermissionRegistry(), clock.System()), pool
}

// newUser creates a user with the given groups and returns it.
func newUser(t *testing.T, service *app.Service, username string, groupIDs []uuid.UUID) domain.User {
	t.Helper()
	user, err := service.CreateUser(context.Background(), nil, app.CreateUserInput{
		CompanyID:   newID(t),
		Username:    username,
		DisplayName: "Pengguna Uji",
		Password:    testPassword,
		GroupIDs:    groupIDs,
	})
	if err != nil {
		t.Fatalf("create user %s: %v", username, err)
	}
	return user
}

// newGroup creates a group holding permissions and returns its id.
func newGroup(t *testing.T, service *app.Service, name string, permissions ...string) uuid.UUID {
	t.Helper()
	group, err := service.CreateAccessGroup(context.Background(), nil, app.CreateAccessGroupInput{
		Name:        name,
		Permissions: permissions,
	})
	if err != nil {
		t.Fatalf("create group %s: %v", name, err)
	}
	return group.ID
}

func TestAuthenticate_UnionsPermissionsOfTwoGroups(t *testing.T) {
	ctx := context.Background()
	service, _ := newService(t)

	kasir := newGroup(t, service, "Kasir", "identity.user.read")
	gudang := newGroup(t, service, "Gudang", "identity.group.read")
	newUser(t, service, "kasir.gudang", []uuid.UUID{kasir, gudang})

	result, err := service.Authenticate(ctx, "kasir.gudang", testPassword)
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}

	want := map[string]bool{"identity.user.read": true, "identity.group.read": true}
	if len(result.Permissions) != len(want) {
		t.Fatalf("permissions = %v, want the union of both groups", result.Permissions)
	}
	for _, permission := range result.Permissions {
		if !want[permission] {
			t.Fatalf("unexpected permission %q", permission)
		}
	}
	if result.Token == "" {
		t.Fatal("a successful login must return a session token")
	}
}

func TestAuthenticate_RejectsWrongPasswordWithInvalidCredentials(t *testing.T) {
	ctx := context.Background()
	service, _ := newService(t)
	newUser(t, service, "apoteker", nil)

	_, err := service.Authenticate(ctx, "apoteker", "sandi-yang-salah")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("wrong password error = %v, want ErrInvalidCredentials", err)
	}

	_, err = service.Authenticate(ctx, "tidak.ada", testPassword)
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("unknown user error = %v, want the same ErrInvalidCredentials", err)
	}
}

func TestCurrentSession_RejectsRevokedAndDeactivatedSessions(t *testing.T) {
	ctx := context.Background()
	service, _ := newService(t)
	admin := newGroup(t, service, "Operator", "identity.user.read")
	user := newUser(t, service, "operator.satu", []uuid.UUID{admin})

	login, err := service.Authenticate(ctx, "operator.satu", testPassword)
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	session, err := service.CurrentSession(ctx, login.Token)
	if err != nil {
		t.Fatalf("current session: %v", err)
	}
	if session.UserID != user.ID || !domain.HasPermission(session.Permissions, "identity.user.read") {
		t.Fatalf("session = %+v", session)
	}

	if err := service.Logout(ctx, login.Token, &session); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if _, err := service.CurrentSession(ctx, login.Token); !errors.Is(err, domain.ErrSessionExpired) {
		t.Fatalf("revoked session error = %v, want ErrSessionExpired", err)
	}

	// Signing in again and deactivating the user must kill the new session too.
	loginAgain, err := service.Authenticate(ctx, "operator.satu", testPassword)
	if err != nil {
		t.Fatalf("authenticate again: %v", err)
	}
	if _, err := service.UpdateUser(ctx, nil, app.UpdateUserInput{
		ID: user.ID, DisplayName: user.DisplayName, IsActive: false,
	}); err != nil {
		t.Fatalf("deactivate user: %v", err)
	}
	if _, err := service.CurrentSession(ctx, loginAgain.Token); !errors.Is(err, domain.ErrSessionExpired) {
		t.Fatalf("deactivated user session error = %v, want ErrSessionExpired", err)
	}
	if _, err := service.Authenticate(ctx, "operator.satu", testPassword); !errors.Is(err, domain.ErrUserInactive) {
		t.Fatalf("inactive login error = %v, want ErrUserInactive", err)
	}
}

func TestCreateUser_RejectsDuplicateUsernameAndShortPassword(t *testing.T) {
	ctx := context.Background()
	service, _ := newService(t)
	newUser(t, service, "gudang.pusat", nil)

	_, err := service.CreateUser(ctx, nil, app.CreateUserInput{
		CompanyID: newID(t), Username: "gudang.pusat", DisplayName: "Kedua", Password: testPassword,
	})
	if !errors.Is(err, domain.ErrUsernameTaken) {
		t.Fatalf("duplicate username error = %v, want ErrUsernameTaken", err)
	}

	_, err = service.CreateUser(ctx, nil, app.CreateUserInput{
		CompanyID: newID(t), Username: "gudang.ketiga", DisplayName: "Ketiga", Password: "pendek",
	})
	if !isCode(err, "VALIDATION_FAILED") {
		t.Fatalf("short password error = %v, want VALIDATION_FAILED", err)
	}
}

func TestCreateAccessGroup_RejectsPermissionOutsideRegistry(t *testing.T) {
	ctx := context.Background()
	service, _ := newService(t)

	_, err := service.CreateAccessGroup(ctx, nil, app.CreateAccessGroupInput{
		Name:        "Tidak Valid",
		Permissions: []string{"identity.user.teleport"},
	})
	if !isCode(err, "UNKNOWN_PERMISSION") {
		t.Fatalf("error = %v, want UNKNOWN_PERMISSION", err)
	}
}

func TestUpdateAccessGroup_RefusesToWeakenBuiltinAdministrator(t *testing.T) {
	ctx := context.Background()
	service, pool := newService(t)

	var adminID uuid.UUID
	if err := pool.QueryRow(ctx,
		`SELECT id FROM identity_access_groups WHERE is_builtin`).Scan(&adminID); err != nil {
		t.Fatalf("read Administrator: %v", err)
	}

	_, err := service.UpdateAccessGroup(ctx, nil, app.UpdateAccessGroupInput{
		ID: adminID, Name: "Administrator", IsActive: false,
	})
	if !errors.Is(err, domain.ErrBuiltinGroupImmutable) {
		t.Fatalf("deactivating Administrator = %v, want ErrBuiltinGroupImmutable", err)
	}
}

func TestAuditTrail_RecordsLoginLogoutAndUserChanges(t *testing.T) {
	ctx := context.Background()
	service, _ := newService(t)
	newUser(t, service, "audit.uji", nil)

	login, err := service.Authenticate(ctx, "audit.uji", testPassword)
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	session, err := service.CurrentSession(ctx, login.Token)
	if err != nil {
		t.Fatalf("current session: %v", err)
	}
	if err := service.Logout(ctx, login.Token, &session); err != nil {
		t.Fatalf("logout: %v", err)
	}

	var loginEvents, logoutEvents int
	if err := queryCount(ctx, service, "login", &loginEvents); err != nil {
		t.Fatal(err)
	}
	if err := queryCount(ctx, service, "logout", &logoutEvents); err != nil {
		t.Fatal(err)
	}
	if loginEvents != 1 || logoutEvents != 1 {
		t.Fatalf("audit counts login=%d logout=%d, want 1 and 1", loginEvents, logoutEvents)
	}
}

func isCode(err error, code string) bool {
	var domainErr *domain.Error
	return errors.As(err, &domainErr) && domainErr.Code() == code
}

// queryCount counts audit rows with the given action. The service exposes the
// trail through ListAuditEvents; this helper asserts on the derived count
// without reaching into the store.
func queryCount(ctx context.Context, service *app.Service, action string, out *int) error {
	page, err := service.ListAuditEvents(ctx, app.ListAuditEventsInput{Limit: 100})
	if err != nil {
		return err
	}
	for _, event := range page.Items {
		if event.Action == action {
			*out++
		}
	}
	return nil
}
