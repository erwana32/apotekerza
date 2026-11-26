//go:build integration

package store_test

import (
	"context"
	"os"
	"regexp"
	"sort"
	"testing"

	"github.com/erwana32/apotekerza/internal/identity/store"
	"github.com/erwana32/apotekerza/internal/identity/store/gen"
	"github.com/erwana32/apotekerza/internal/platform/testutil"
)

func TestListAccessGroups_KeysetPaginationAndSearch(t *testing.T) {
	ctx := context.Background()
	pool := newMigratedPool(t)
	queries := store.New(pool)

	// The seeded "Administrator" group already exists; add three more.
	for _, name := range []string{"Gudang", "Kasir", "Pembelian"} {
		if _, err := queries.InsertAccessGroup(ctx, gen.InsertAccessGroupParams{
			ID: newID(t), Name: name, IsBuiltin: false, IsActive: true,
		}); err != nil {
			t.Fatalf("insert group %s: %v", name, err)
		}
	}

	first, err := queries.ListAccessGroups(ctx, gen.ListAccessGroupsParams{PageSize: 2})
	if err != nil {
		t.Fatalf("list page 1: %v", err)
	}
	if len(first) != 2 || first[0].Name != "Administrator" || first[1].Name != "Gudang" {
		t.Fatalf("page 1 = %v", groupNames(first))
	}

	after := first[1].Name
	second, err := queries.ListAccessGroups(ctx, gen.ListAccessGroupsParams{AfterName: &after, PageSize: 5})
	if err != nil {
		t.Fatalf("list page 2: %v", err)
	}
	if got := groupNames(second); len(got) != 2 || got[0] != "Kasir" || got[1] != "Pembelian" {
		t.Fatalf("page 2 = %v, want [Kasir Pembelian]", got)
	}

	search := "mbel"
	found, err := queries.ListAccessGroups(ctx, gen.ListAccessGroupsParams{Search: &search, PageSize: 5})
	if err != nil {
		t.Fatalf("list with search: %v", err)
	}
	if got := groupNames(found); len(got) != 1 || got[0] != "Pembelian" {
		t.Fatalf("search %q = %v, want [Pembelian]", search, got)
	}
}

// TestAdministrator_HoldsEveryDeclaredPermission guards the rule that every
// x-permission in api/openapi.yaml is granted to the built-in Administrator by
// a migration (docs/modules/identity.md). A new endpoint whose migration forgets
// the grant fails here, and so does a stale grant for a permission that is no
// longer declared.
func TestAdministrator_HoldsEveryDeclaredPermission(t *testing.T) {
	ctx := context.Background()
	pool := newMigratedPool(t)

	raw, err := os.ReadFile(testutil.OpenAPIPath())
	if err != nil {
		t.Fatalf("read openapi.yaml: %v", err)
	}
	declared := map[string]bool{}
	for _, m := range regexp.MustCompile(`(?m)^\s*x-permission:\s*([a-z_]+\.[a-z_]+\.[a-z_]+)\s*$`).FindAllStringSubmatch(string(raw), -1) {
		declared[m[1]] = true
	}
	if len(declared) == 0 {
		t.Fatal("no x-permission found in api/openapi.yaml: the check would pass for nothing")
	}

	rows, err := pool.Query(ctx,
		`SELECT gp.permission FROM identity_group_permissions gp
		 JOIN identity_access_groups g ON g.id = gp.access_group_id
		 WHERE g.name = 'Administrator'`)
	if err != nil {
		t.Fatalf("read Administrator permissions: %v", err)
	}
	defer rows.Close()
	granted := map[string]bool{}
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			t.Fatalf("scan: %v", err)
		}
		granted[p] = true
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows: %v", err)
	}

	var missing, stale []string
	for p := range declared {
		if !granted[p] {
			missing = append(missing, p)
		}
	}
	for p := range granted {
		if !declared[p] {
			stale = append(stale, p)
		}
	}
	sort.Strings(missing)
	sort.Strings(stale)
	if len(missing) > 0 {
		t.Errorf("declared in openapi.yaml but not granted to Administrator: %v", missing)
	}
	if len(stale) > 0 {
		t.Errorf("granted to Administrator but not declared in openapi.yaml: %v", stale)
	}
}

func groupNames(groups []gen.IdentityAccessGroup) []string {
	names := make([]string, 0, len(groups))
	for _, g := range groups {
		names = append(names, g.Name)
	}
	return names
}
