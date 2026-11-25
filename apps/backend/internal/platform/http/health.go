package http

import (
	"fmt"
	nethttp "net/http"
)

// NewHealthHandler returns a minimal operational liveness probe for the
// process and the container orchestrator. It returns 200 with a plain body and
// runs no domain logic. It is NOT part of the OpenAPI contract: the versioned
// GET /api/v1/health is added spec-first in Prompt 0.3.
func NewHealthHandler() nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		if r.Method != nethttp.MethodGet {
			w.Header().Set("Allow", nethttp.MethodGet)
			nethttp.Error(w, "method not allowed", nethttp.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(nethttp.StatusOK)
		fmt.Fprint(w, "ok\n")
	})
}
