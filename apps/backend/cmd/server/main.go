// Command server runs the HTTP server with graceful shutdown.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/erwana32/apotekerza/internal/platform/config"
	platformdb "github.com/erwana32/apotekerza/internal/platform/db"
	platformhttp "github.com/erwana32/apotekerza/internal/platform/http"
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
	// Operational liveness probe only. The versioned GET /api/v1/health is
	// added from api/openapi.yaml in Prompt 0.3.
	mux.Handle("GET /healthz", platformhttp.NewHealthHandler())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Info("starting http server", "addr", cfg.HTTPAddr)
	return platformhttp.NewServer(cfg.HTTPAddr, logger, platformhttp.RequestID(logger, mux)).Run(ctx)
}
