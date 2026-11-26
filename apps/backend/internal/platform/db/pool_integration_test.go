//go:build integration

package db

import (
	"context"
	"errors"
	"testing"

	"github.com/erwana32/apotekerza/internal/platform/testutil"
	"github.com/jackc/pgx/v5"
)

func TestWithTx_CommitsAndRollsBack(t *testing.T) {
	ctx := context.Background()
	pool, cleanup, err := testutil.StartPostgres(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	if _, err := pool.Exec(ctx, "CREATE TABLE tx_probe (id text PRIMARY KEY)"); err != nil {
		t.Fatal(err)
	}

	err = WithTx(ctx, pool, func(ctx context.Context, tx pgx.Tx) error {
		_, err := tx.Exec(ctx, "INSERT INTO tx_probe (id) VALUES ($1)", "committed")
		return err
	})
	if err != nil {
		t.Fatalf("committing transaction: %v", err)
	}

	sentinel := errors.New("abort")
	err = WithTx(ctx, pool, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, "INSERT INTO tx_probe (id) VALUES ($1)", "rolled-back"); err != nil {
			return err
		}
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("WithTx() error = %v, want %v", err, sentinel)
	}

	var count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM tx_probe").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("rows = %d, want 1 (only the committed row)", count)
	}
}
