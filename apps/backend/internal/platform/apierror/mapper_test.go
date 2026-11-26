package apierror

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

type domainError struct{}

func (domainError) Error() string            { return "insufficient stock" }
func (domainError) Code() string             { return "INSUFFICIENT_STOCK" }
func (domainError) StatusCode() int          { return http.StatusConflict }
func (domainError) DeveloperMessage() string { return "stock is insufficient" }

func TestMap_DomainErrorAndWrappedError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantCode   string
		wantStatus int
	}{
		{"direct", domainError{}, "INSUFFICIENT_STOCK", http.StatusConflict},
		{"wrapped", fmt.Errorf("receive goods: %w", domainError{}), "INSUFFICIENT_STOCK", http.StatusConflict},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapped := Map(tt.err)
			if mapped.Code != tt.wantCode || mapped.Status != tt.wantStatus {
				t.Fatalf("Map() = %+v, want %s/%d", mapped, tt.wantCode, tt.wantStatus)
			}
			if mapped.Message != "stock is insufficient" {
				t.Fatalf("message = %q", mapped.Message)
			}
		})
	}
}

func TestMap_UnknownErrorIsGeneric(t *testing.T) {
	mapped := Map(errors.New("secret detail"))
	if mapped.Code != "INTERNAL_ERROR" || mapped.Status != http.StatusInternalServerError {
		t.Fatalf("Map() = %+v, want INTERNAL_ERROR/500", mapped)
	}
	if mapped.Message == "secret detail" {
		t.Fatal("Map leaked an internal error message")
	}
}
