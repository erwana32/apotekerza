package http

import (
	"context"

	"github.com/erwana32/apotekerza/internal/platform/http/gen"
)

// StrictHandler implements the server interface generated from api/openapi.yaml.
type StrictHandler struct{}

// NewStrictHandler returns the generated-contract handler.
func NewStrictHandler() *StrictHandler {
	return &StrictHandler{}
}

// GetHealth is generated from operationId GetHealth. It is public (x-public:
// true in the spec) and carries no domain logic.
func (h *StrictHandler) GetHealth(_ context.Context, _ gen.GetHealthRequestObject) (gen.GetHealthResponseObject, error) {
	return gen.GetHealth200JSONResponse{Status: "ok"}, nil
}
