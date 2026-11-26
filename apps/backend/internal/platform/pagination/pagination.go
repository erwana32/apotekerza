// Package pagination parses the shared limit/cursor query parameters.
package pagination

import (
	"encoding/base64"
	"fmt"
	"strconv"
)

const (
	// DefaultLimit is used when the client omits limit.
	DefaultLimit = 20
	// MaxLimit matches the maximum in api/openapi.yaml.
	MaxLimit = 100
)

// Params is the parsed page request. Cursor is empty on the first page.
type Params struct {
	Limit  int
	Cursor string
}

// Parse validates limit (default 20, range 1..100) and keeps cursor opaque.
func Parse(limit, cursor string) (Params, error) {
	params := Params{Limit: DefaultLimit, Cursor: cursor}
	if limit == "" {
		return params, nil
	}
	value, err := strconv.Atoi(limit)
	if err != nil {
		return Params{}, fmt.Errorf("pagination: invalid limit %q: %w", limit, err)
	}
	if value < 1 || value > MaxLimit {
		return Params{}, fmt.Errorf("pagination: limit %d out of range 1..%d", value, MaxLimit)
	}
	params.Limit = value
	return params, nil
}

// EncodeCursor makes an opaque, URL-safe cursor from a keyset value.
func EncodeCursor(key string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(key))
}

// DecodeCursor reverses EncodeCursor.
func DecodeCursor(cursor string) (string, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return "", fmt.Errorf("pagination: invalid cursor: %w", err)
	}
	return string(decoded), nil
}
