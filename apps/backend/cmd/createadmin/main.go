// Command createadmin creates the first Administrator user: the builtin
// Administrator group holds every permission, but no user is seeded because the
// first account needs a real password (docs/modules/identity.md, "Seeding").
//
// Usage:
//
//	CREATEADMIN_PASSWORD=... createadmin -company <uuid> -username admin
//
// The password is read from the environment so it never appears in the process
// list or in shell history.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/erwana32/apotekerza/internal/identity/app"
	"github.com/erwana32/apotekerza/internal/identity/domain"
	"github.com/erwana32/apotekerza/internal/identity/store"
	"github.com/erwana32/apotekerza/internal/platform/clock"
	"github.com/erwana32/apotekerza/internal/platform/config"
	platformdb "github.com/erwana32/apotekerza/internal/platform/db"
)

// administratorGroupName is the builtin group the first user joins. It is
// seeded by migration and marked is_builtin.
const administratorGroupName = "Administrator"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	companyID := flag.String("company", "", "company id (uuid) the administrator belongs to")
	username := flag.String("username", "", "login name, lowercase")
	displayName := flag.String("display-name", "Administrator", "name shown in the UI")
	flag.Parse()

	if *companyID == "" || *username == "" {
		return errors.New("createadmin: -company and -username are required")
	}
	parsedCompany, err := uuid.Parse(*companyID)
	if err != nil {
		return fmt.Errorf("createadmin: invalid company id: %w", err)
	}
	password := os.Getenv("CREATEADMIN_PASSWORD")
	if len(password) < app.MinPasswordLength {
		return fmt.Errorf("createadmin: set CREATEADMIN_PASSWORD to at least %d characters", app.MinPasswordLength)
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx := context.Background()
	pool, err := platformdb.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	adminGroupID, err := findAdministratorGroup(ctx, pool)
	if err != nil {
		return err
	}

	service := app.NewService(pool, app.NewPermissionRegistry(), clock.System())
	user, err := service.CreateUser(ctx, nil, app.CreateUserInput{
		CompanyID:   parsedCompany,
		Username:    *username,
		DisplayName: *displayName,
		Password:    password,
		GroupIDs:    []uuid.UUID{adminGroupID},
	})
	if err != nil {
		var domainErr *domain.Error
		if errors.As(err, &domainErr) {
			return fmt.Errorf("createadmin: %w", domainErr)
		}
		return fmt.Errorf("createadmin: %w", err)
	}

	fmt.Printf("created administrator %q in group %q\n", user.Username, administratorGroupName)
	return nil
}

// findAdministratorGroup returns the id of the seeded Administrator group.
func findAdministratorGroup(ctx context.Context, pool *pgxpool.Pool) (uuid.UUID, error) {
	groups, err := store.New(pool).GetBuiltinAccessGroups(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("createadmin: %w", err)
	}
	for _, group := range groups {
		if group.Name == administratorGroupName {
			return group.ID, nil
		}
	}
	return uuid.Nil, fmt.Errorf("createadmin: builtin group %q not found; run migrations first", administratorGroupName)
}
