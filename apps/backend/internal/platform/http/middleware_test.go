package http

import (
	"bytes"
	"encoding/hex"
	"log/slog"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestID_GeneratesAndLogsID(t *testing.T) {
	var buf bytes.Buffer
	base := slog.New(slog.NewJSONHandler(&buf, nil))
	var seen *slog.Logger

	handler := RequestID(base, nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		seen = LoggerFrom(r.Context(), base)
		w.WriteHeader(nethttp.StatusOK)
	}))

	req := httptest.NewRequest(nethttp.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	id := rec.Header().Get(requestIDHeader)
	if len(id) != 32 {
		t.Fatalf("request ID = %q, want 32 hex characters", id)
	}
	if _, err := hex.DecodeString(id); err != nil {
		t.Fatalf("request ID %q is not hex: %v", id, err)
	}
	seen.Info("probe")
	if !strings.Contains(buf.String(), `"request_id":"`+id+`"`) {
		t.Fatalf("log output %q has no request_id %q", buf.String(), id)
	}
}

func TestRequestID_PreservesInboundID(t *testing.T) {
	base := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	handler := RequestID(base, nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {}))

	req := httptest.NewRequest(nethttp.MethodGet, "/healthz", nil)
	req.Header.Set(requestIDHeader, "inbound-id")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get(requestIDHeader); got != "inbound-id" {
		t.Fatalf("request ID = %q, want inbound-id", got)
	}
}
