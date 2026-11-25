// Package http hosts the HTTP server, middleware, and graceful shutdown.
package http

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	nethttp "net/http"
)

const requestIDHeader = "X-Request-ID"

type contextKey int

const loggerKey contextKey = iota

// RequestID makes sure every request has an ID: an inbound X-Request-ID is
// kept, otherwise a random 128-bit ID is generated. The ID is echoed in the
// response and attached to the request-scoped logger.
func RequestID(logger *slog.Logger, next nethttp.Handler) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		id := r.Header.Get(requestIDHeader)
		if id == "" {
			id = newRequestID()
		}
		w.Header().Set(requestIDHeader, id)
		logger := logger.With("request_id", id)
		ctx := context.WithValue(r.Context(), loggerKey, logger)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoggerFrom returns the request-scoped logger, or fallback when the request
// did not pass through RequestID.
func LoggerFrom(ctx context.Context, fallback *slog.Logger) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok && logger != nil {
		return logger
	}
	return fallback
}

func newRequestID() string {
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		// crypto/rand failing is exceptional; never send an empty ID.
		return "unavailable"
	}
	return hex.EncodeToString(id[:])
}
