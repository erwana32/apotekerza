// Package ids generates and parses UUIDv7 identifiers.
package ids

import (
	"fmt"

	"github.com/google/uuid"
)

// New returns a time-ordered UUIDv7.
func New() (uuid.UUID, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, fmt.Errorf("ids: generate v7: %w", err)
	}
	return id, nil
}

// Parse validates and parses a UUID string.
func Parse(value string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, fmt.Errorf("ids: parse %q: %w", value, err)
	}
	return id, nil
}
