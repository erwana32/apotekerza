package http

import (
	"context"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/erwana32/apotekerza/internal/platform/http/gen"
)

func TestGetHealth_ReturnsOKStatus(t *testing.T) {
	response, err := NewStrictHandler().GetHealth(context.Background(), gen.GetHealthRequestObject{})
	if err != nil {
		t.Fatalf("GetHealth() error = %v", err)
	}
	health, ok := response.(gen.GetHealth200JSONResponse)
	if !ok {
		t.Fatalf("GetHealth() response = %T, want GetHealth200JSONResponse", response)
	}
	if health.Status != "ok" {
		t.Fatalf("status = %q, want ok", health.Status)
	}
}

func TestGetHealth_GeneratedRouteReturnsJSON(t *testing.T) {
	mux := nethttp.NewServeMux()
	handler := gen.HandlerFromMux(gen.NewStrictHandler(NewStrictHandler(), nil), mux)

	req := httptest.NewRequest(nethttp.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != nethttp.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"status":"ok"}` {
		t.Fatalf("body = %q, want %q", got, `{"status":"ok"}`)
	}
}
