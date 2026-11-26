package http

import (
	"context"
	"errors"
	"log/slog"
	nethttp "net/http"
	"time"
)

const shutdownTimeout = 5 * time.Second

// Server wraps net/http.Server with graceful shutdown.
type Server struct {
	httpServer *nethttp.Server
	logger     *slog.Logger
}

// NewServer builds a server for addr using handler.
func NewServer(addr string, logger *slog.Logger, handler nethttp.Handler) *Server {
	return &Server{
		httpServer: &nethttp.Server{Addr: addr, Handler: handler},
		logger:     logger,
	}
}

// Run serves until ctx is cancelled, then shuts down gracefully. A clean
// shutdown (http.ErrServerClosed) is reported as nil.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.httpServer.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, nethttp.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		s.logger.Info("shutting down http server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return nil
	}
}
