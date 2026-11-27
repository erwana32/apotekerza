// Command server runs the HTTP server with graceful shutdown.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/erwana32/apotekerza/internal/identity"
	"github.com/erwana32/apotekerza/internal/identity/app"
	"github.com/erwana32/apotekerza/internal/platform/clock"
	"github.com/erwana32/apotekerza/internal/platform/config"
	platformdb "github.com/erwana32/apotekerza/internal/platform/db"
	platformhttp "github.com/erwana32/apotekerza/internal/platform/http"
	"github.com/erwana32/apotekerza/internal/platform/http/gen"
	platformlog "github.com/erwana32/apotekerza/internal/platform/log"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	logger, err := platformlog.New(cfg.LogLevel)
	if err != nil {
		return err
	}

	pool, err := platformdb.NewPool(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	mux := http.NewServeMux()
	// Operational liveness probe, outside the OpenAPI contract.
	mux.Handle("GET /healthz", platformhttp.NewHealthHandler())

	// Identity: permission registry from the spec, argon2id, cookie sessions.
	identityModule := identity.New(app.NewService(pool, app.NewPermissionRegistry(), clock.System()))

	// Versioned API routes come from the generated strict server, with the
	// identity permission guard installed as a strict middleware.
	strict := newStrictAPI(identityModule)
	handler := gen.HandlerFromMux(strict, mux)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Info("starting http server", "addr", cfg.HTTPAddr)
	return platformhttp.NewServer(cfg.HTTPAddr, logger, platformhttp.RequestID(logger, handler)).Run(ctx)
}
