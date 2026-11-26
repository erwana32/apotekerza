//go:build integration

package store_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/erwana32/apotekerza/internal/identity/store"
	"github.com/erwana32/apotekerza/internal/identity/store/gen"
	"github.com/erwana32/apotekerza/internal/platform/testutil"
)

// TestMain stops the package's shared PostgreSQL container after the tests.
func TestMain(m *testing.M) { os.Exit(testutil.RunMain(m)) }

// newMigratedPool returns a pool on a fresh database cloned from the migrated
// template, so these tests exercise the schema that ships and never share rows.
func newMigratedPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	return testutil.NewDatabase(t)
}

// newID returns a UUIDv7 for test rows; a failure to generate one is fatal.
func newID(t *testing.T) uuid.UUID {
	t.Helper()
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("generate uuid: %v", err)
	}
	return id
}

func TestIdentityAuditLog_RejectsUpdateAndDelete(t *testing.T) {
	ctx := context.Background()
	pool := newMigratedPool(t)

	entryID := newID(t)
	if _, err := pool.Exec(ctx,
		`INSERT INTO identity_audit_log (id, company_id, action, entity, entity_id)
		 VALUES ($1, $2, 'create', 'user', $3)`,
		entryID, newID(t), newID(t),
	); err != nil {
		t.Fatalf("insert audit row: %v", err)
	}

	tests := []struct {
		name string
		sql  string
	}{
		{"update", `UPDATE identity_audit_log SET action = 'update' WHERE id = $1`},
		{"delete", `DELETE FROM identity_audit_log WHERE id = $1`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := pool.Exec(ctx, tt.sql, entryID)
			var pgErr *pgconn.PgError
			if !errors.As(err, &pgErr) || pgErr.Code != "23001" {
				t.Fatalf("%s must fail with restrict_violation (23001), got %v", tt.name, err)
			}
		})
	}

	var count int
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM identity_audit_log WHERE id = $1`, entryID,
	).Scan(&count); err != nil {
		t.Fatalf("read back: %v", err)
	}
	if count != 1 {
		t.Fatalf("audit row must survive, want 1 row, got %d", count)
	}
}

func TestIdentitySeed_CreatesBuiltinAdministrator(t *testing.T) {
	ctx := context.Background()
	pool := newMigratedPool(t)

	var isBuiltin bool
	if err := pool.QueryRow(ctx,
		`SELECT is_builtin FROM identity_access_groups WHERE name = 'Administrator'`,
	).Scan(&isBuiltin); err != nil {
		t.Fatalf("read Administrator group: %v", err)
	}
	if !isBuiltin {
		t.Fatal("Administrator group must be builtin")
	}

	var permissions int
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM identity_group_permissions gp
		 JOIN identity_access_groups g ON g.id = gp.access_group_id
		 WHERE g.name = 'Administrator'`,
	).Scan(&permissions); err != nil {
		t.Fatalf("count Administrator permissions: %v", err)
	}
	if permissions == 0 {
		t.Fatal("Administrator must hold at least platform.health.read")
	}
}

func TestIdentityUsers_RejectsDuplicateAndUppercaseUsername(t *testing.T) {
	ctx := context.Background()
	pool := newMigratedPool(t)
	queries := store.New(pool)
	companyID := newID(t)

	insert := func(username string) error {
		_, err := queries.InsertUser(ctx, gen.InsertUserParams{
			ID: newID(t), CompanyID: companyID,
			Username: username, PasswordHash: "argon2id$v=19$m=65536,t=3,p=2$c2FsdA$aGFzaA",
			DisplayName: "Pengguna", IsActive: true,
		})
		return err
	}

	if err := insert("kasir.1"); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	if err := insert("kasir.1"); !errors.Is(err, store.ErrDuplicateUsername) {
		t.Fatalf("duplicate username: want ErrDuplicateUsername, got %v", err)
	}

	var pgErr *pgconn.PgError
	if err := insert("Kasir.2"); !errors.As(err, &pgErr) || pgErr.Code != "23514" {
		t.Fatalf("uppercase username: want SQLSTATE 23514, got %v", err)
	}
}

func TestListPermissionsForUser_UnionsTwoGroups(t *testing.T) {
	ctx := context.Background()
	pool := newMigratedPool(t)
	queries := store.New(pool)
	companyID, userID := newID(t), newID(t)

	if _, err := queries.InsertUser(ctx, gen.InsertUserParams{
		ID: userID, CompanyID: companyID,
		Username: "apoteker.1", PasswordHash: "argon2id$v=19$m=65536,t=3,p=2$c2FsdA$aGFzaA",
		DisplayName: "Apoteker Satu", IsActive: true,
	}); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	groups := []struct{ name, permission string }{
		{"Kasir", "pos.sale.create"},
		{"Gudang", "inventory.stock.read"},
	}
	for _, g := range groups {
		groupID := newID(t)
		if _, err := queries.InsertAccessGroup(ctx, gen.InsertAccessGroupParams{
			ID: groupID, Name: g.name, IsBuiltin: false, IsActive: true,
		}); err != nil {
			t.Fatalf("insert group %s: %v", g.name, err)
		}
		if err := queries.InsertGroupPermission(ctx, gen.InsertGroupPermissionParams{
			AccessGroupID: groupID, Permission: g.permission,
		}); err != nil {
			t.Fatalf("insert permission %s: %v", g.permission, err)
		}
		if _, err := pool.Exec(ctx,
			`INSERT INTO identity_user_groups (user_id, access_group_id) VALUES ($1, $2)`,
			userID, groupID,
		); err != nil {
			t.Fatalf("assign group %s: %v", g.name, err)
		}
	}

	permissions, err := queries.ListPermissionsForUser(ctx, userID)
	if err != nil {
		t.Fatalf("list permissions: %v", err)
	}
	want := map[string]bool{"pos.sale.create": true, "inventory.stock.read": true}
	if len(permissions) != len(want) {
		t.Fatalf("want the union of both groups, got %v", permissions)
	}
	for _, p := range permissions {
		if !want[p] {
			t.Fatalf("unexpected permission %q", p)
		}
	}
}

func TestIdentityUsers_UsernameIsUniqueAcrossCompanies(t *testing.T) {
	ctx := context.Background()
	pool := newMigratedPool(t)
	queries := store.New(pool)

	insert := func(companyID uuid.UUID) error {
		_, err := queries.InsertUser(ctx, gen.InsertUserParams{
			ID: newID(t), CompanyID: companyID,
			Username: "owner", PasswordHash: "argon2id$v=19$m=65536,t=3,p=2$c2FsdA$aGFzaA",
			DisplayName: "Pemilik", IsActive: true,
		})
		return err
	}
	if err := insert(newID(t)); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if err := insert(newID(t)); !errors.Is(err, store.ErrDuplicateUsername) {
		t.Fatalf("the same username in another company: want ErrDuplicateUsername, got %v", err)
	}

	user, err := queries.GetUserByUsername(ctx, "owner")
	if err != nil {
		t.Fatalf("get user by username: %v", err)
	}
	if user.Username != "owner" {
		t.Fatalf("got user %q", user.Username)
	}
	if _, err := queries.GetUserByUsername(ctx, "nobody"); !errors.Is(err, store.ErrUserNotFound) {
		t.Fatalf("unknown username: want ErrUserNotFound, got %v", err)
	}
}

func TestListUsers_KeysetPaginationAndSearch(t *testing.T) {
	ctx := context.Background()
	pool := newMigratedPool(t)
	queries := store.New(pool)
	companyID := newID(t)

	for _, username := range []string{"adi", "budi", "citra", "dewi", "eka"} {
		if _, err := queries.InsertUser(ctx, gen.InsertUserParams{
			ID: newID(t), CompanyID: companyID, Username: username,
			PasswordHash: "argon2id$v=19$m=65536,t=3,p=2$c2FsdA$aGFzaA",
			DisplayName:  username, IsActive: true,
		}); err != nil {
			t.Fatalf("insert %s: %v", username, err)
		}
	}

	// Page 1: ask for size+1 rows to learn that another page exists.
	first, err := queries.ListUsers(ctx, gen.ListUsersParams{PageSize: 3})
	if err != nil {
		t.Fatalf("list page 1: %v", err)
	}
	if len(first) != 3 || first[0].Username != "adi" || first[2].Username != "citra" {
		t.Fatalf("page 1 (3 rows) = %v", usernames(first))
	}

	// Page 2 continues after the last username of page 1.
	after := first[2].Username
	second, err := queries.ListUsers(ctx, gen.ListUsersParams{AfterUsername: &after, PageSize: 3})
	if err != nil {
		t.Fatalf("list page 2: %v", err)
	}
	if got := usernames(second); len(got) != 2 || got[0] != "dewi" || got[1] != "eka" {
		t.Fatalf("page 2 = %v, want [dewi eka]", got)
	}

	// A NULL search means no filter; a value filters (the old $1::text IS NULL
	// branch could never be NULL, now it can).
	search := "ud"
	found, err := queries.ListUsers(ctx, gen.ListUsersParams{Search: &search, PageSize: 10})
	if err != nil {
		t.Fatalf("list with search: %v", err)
	}
	if got := usernames(found); len(got) != 1 || got[0] != "budi" {
		t.Fatalf("search %q = %v, want [budi]", search, got)
	}
}

func usernames(users []gen.IdentityUser) []string {
	names := make([]string, 0, len(users))
	for _, u := range users {
		names = append(names, u.Username)
	}
	return names
}
