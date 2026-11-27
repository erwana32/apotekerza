// Package app holds the identity use cases. One use case per service_*.go
// file; each owns the transaction boundary and calls the store layer.
package app

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/erwana32/apotekerza/internal/identity/domain"
)

// SessionTTL is how long a login session stays valid.
const SessionTTL = 12 * time.Hour

// TokenBytes is the size of the random session token given to the browser.
const TokenBytes = 32

// Service holds the dependencies of every identity use case. Handlers build
// one Service per request group; services open their own transactions.
type Service struct {
	pool     *pgxpool.Pool
	registry *PermissionRegistry
	clock    Clock
}

// Clock supplies the current time, so tests can control session expiry.
type Clock interface {
	Now() time.Time
}

// NewService builds the identity service.
func NewService(pool *pgxpool.Pool, registry *PermissionRegistry, clock Clock) *Service {
	return &Service{pool: pool, registry: registry, clock: clock}
}

// Registry exposes the permission registry for the list endpoint.
func (s *Service) Registry() *PermissionRegistry { return s.registry }

// NewSessionToken returns a random token and its SHA-256. Only the hash is
// stored, so a database read cannot be replayed as a cookie.
func NewSessionToken() (token string, tokenHash []byte, err error) {
	raw := make([]byte, TokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", nil, fmt.Errorf("identity: generate session token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), HashToken(raw), nil
}

// HashToken returns the SHA-256 of a token's raw bytes.
func HashToken(raw []byte) []byte {
	sum := sha256.Sum256(raw)
	return sum[:]
}

// HashEncodedToken hashes a token as it arrives in a cookie.
func HashEncodedToken(encoded string) ([]byte, error) {
	raw, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("identity: decode token: %w", err)
	}
	return HashToken(raw), nil
}

// asDomainError keeps an existing domain error and wraps anything else as an
// internal error, so a handler never leaks a raw database message.
func asDomainError(err error) error {
	var domainErr *domain.Error
	if errors.As(err, &domainErr) {
		return domainErr
	}
	return fmt.Errorf("%w: %w", domain.ErrInternal, err)
}
