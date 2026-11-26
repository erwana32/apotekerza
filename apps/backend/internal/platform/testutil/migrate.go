package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// MigrationsDir returns the absolute path of apps/backend/migrations, resolved
// from this source file, so tests work from any working directory.
func MigrationsDir() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "migrations"
	}
	return filepath.Join(filepath.Dir(file), "..", "..", "..", "migrations")
}

// repoPath joins elements onto the repository root (three levels above
// apps/backend/internal/platform/testutil is apps/backend; the root is two more).
func repoPath(elem ...string) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return filepath.Join(elem...)
	}
	root := filepath.Join(filepath.Dir(file), "..", "..", "..", "..", "..")
	return filepath.Join(append([]string{root}, elem...)...)
}

// MigrateUp runs every pending goose migration from apps/backend/migrations
// against dsn. Integration tests use the real migrations, never a hand-written
// copy of the schema.
func MigrateUp(ctx context.Context, dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("testutil: open database: %w", err)
	}
	defer func() { _ = db.Close() }()

	provider, err := goose.NewProvider(goose.DialectPostgres, db, os.DirFS(MigrationsDir()))
	if err != nil {
		return fmt.Errorf("testutil: migration provider: %w", err)
	}
	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("testutil: migrate up: %w", err)
	}
	return nil
}
