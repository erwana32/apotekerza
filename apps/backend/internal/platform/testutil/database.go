package testutil

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	testDBUser     = "apotekerza"
	testDBPassword = "test-only"
	adminDatabase  = "apotekerza"
	templateDBName = "apotekerza_template"
)

// shared holds the one PostgreSQL container of a test binary. It starts on the
// first NewDatabase call and stops in RunMain, so a package pays for one
// container however many tests it has.
var shared struct {
	once      sync.Once
	err       error
	container testcontainers.Container
	host      string
	port      string
	admin     *pgxpool.Pool
}

// RunMain runs a package's tests and then stops the shared container. Use it
// from TestMain in every package whose integration tests call NewDatabase:
//
//	func TestMain(m *testing.M) { os.Exit(testutil.RunMain(m)) }
func RunMain(m *testing.M) int {
	code := m.Run()
	shutdown()
	return code
}

func shutdown() {
	if shared.admin != nil {
		shared.admin.Close()
	}
	if shared.container != nil {
		_ = shared.container.Terminate(context.Background())
	}
}

// NewDatabase returns a pool on a fresh, fully migrated database that belongs to
// this test alone. The real goose migrations run once per test binary into a
// template database; each test then clones the template, which is fast and
// isolates tests from one another without TRUNCATE (the ledger-style tables are
// append-only and reject it row by row). The database is dropped when the test
// ends. Never a mocked database (hard rule 8).
func NewDatabase(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	shared.once.Do(func() { shared.err = startShared(ctx) })
	if shared.err != nil {
		t.Fatalf("testutil: shared postgres: %v", shared.err)
	}

	name := "t_" + uuid.NewString()[:8] + "_" + uuid.NewString()[:8]
	if _, err := shared.admin.Exec(ctx, fmt.Sprintf(`CREATE DATABASE %q TEMPLATE %q`, name, templateDBName)); err != nil {
		t.Fatalf("testutil: clone template database: %v", err)
	}
	pool, err := pgxpool.New(ctx, dsnFor(name))
	if err != nil {
		t.Fatalf("testutil: connect to %s: %v", name, err)
	}
	t.Cleanup(func() {
		pool.Close()
		if _, err := shared.admin.Exec(context.Background(), fmt.Sprintf(`DROP DATABASE IF EXISTS %q WITH (FORCE)`, name)); err != nil {
			t.Logf("testutil: drop %s: %v", name, err)
		}
	})
	return pool
}

func dsnFor(database string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		testDBUser, testDBPassword, shared.host, shared.port, database)
}

// startShared starts the container and builds the migrated template database.
func startShared(ctx context.Context) error {
	container, err := postgres.Run(ctx,
		"postgres:17-alpine",
		postgres.WithDatabase(adminDatabase),
		postgres.WithUsername(testDBUser),
		postgres.WithPassword(testDBPassword),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		return fmt.Errorf("run postgres: %w", err)
	}
	shared.container = container

	if shared.host, err = container.Host(ctx); err != nil {
		return fmt.Errorf("container host: %w", err)
	}
	mapped, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		return fmt.Errorf("container port: %w", err)
	}
	shared.port = mapped.Port()

	if shared.admin, err = pgxpool.New(ctx, dsnFor(adminDatabase)); err != nil {
		return fmt.Errorf("admin pool: %w", err)
	}
	if _, err = shared.admin.Exec(ctx, fmt.Sprintf(`CREATE DATABASE %q`, templateDBName)); err != nil {
		return fmt.Errorf("create template database: %w", err)
	}
	if err = MigrateUp(ctx, dsnFor(templateDBName)); err != nil {
		return err
	}
	return nil
}

// OpenAPIPath returns the absolute path of api/openapi.yaml, resolved from the
// repository layout, so tests can compare the contract with the database.
func OpenAPIPath() string {
	return repoPath("api", "openapi.yaml")
}
