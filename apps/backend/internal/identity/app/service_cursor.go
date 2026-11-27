package app

import (
	"encoding/base64"
	"fmt"
)

// encodeCursor makes an opaque cursor from a keyset value (base64url, so a
// username or name never travels as a visible query value).
func encodeCursor(key string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(key))
}

// decodeCursor reverses encodeCursor.
func decodeCursor(cursor string) (string, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return "", fmt.Errorf("identity: decode cursor: %w", err)
	}
	return string(decoded), nil
}
