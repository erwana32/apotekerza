// Command migrate applies goose migrations as a separate step before the
// backend starts, as required by deploy/AGENTS.md.
package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/erwana32/apotekerza/internal/platform/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	dir := flag.String("dir", "migrations", "directory containing goose SQL migrations")
	flag.Parse()

	if err := run(*dir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(dir string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	conn, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("migrate: open database: %w", err)
	}
	defer func() { _ = conn.Close() }()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("migrate: dialect: %w", err)
	}
	if err := goose.UpContext(context.Background(), conn, dir); err != nil {
		// No schema migrations exist yet (Phase 0). This is a valid no-op.
		if errors.Is(err, goose.ErrNoMigrationFiles) {
			return nil
		}
		return fmt.Errorf("migrate: up: %w", err)
	}
	return nil
}
