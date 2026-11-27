package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/erwana32/apotekerza/internal/identity/store/gen"
)

// LoginSession is a session joined with its user, as the middleware needs it.
// The token itself is never carried: only its hash is stored.
type LoginSession struct {
	SessionID    uuid.UUID
	UserID       uuid.UUID
	CompanyID    uuid.UUID
	Username     string
	DisplayName  string
	UserIsActive bool
}

// InsertSession stores a new session. token_hash holds the SHA-256 of the
// cookie token (32 bytes), never the token.
func (q *Queries) InsertSession(ctx context.Context, params gen.InsertSessionParams) (gen.IdentitySession, error) {
	session, err := q.Queries.InsertSession(ctx, params)
	if err != nil {
		return gen.IdentitySession{}, fmt.Errorf("store: insert session: %w", err)
	}
	return session, nil
}

// GetSessionByTokenHash returns a live session and its user. A missing,
// revoked or expired session is ErrSessionNotFound.
func (q *Queries) GetSessionByTokenHash(ctx context.Context, tokenHash []byte) (LoginSession, error) {
	row, err := q.Queries.GetSessionByTokenHash(ctx, tokenHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return LoginSession{}, fmt.Errorf("store: session: %w", ErrSessionNotFound)
	}
	if err != nil {
		return LoginSession{}, fmt.Errorf("store: get session: %w", err)
	}
	return LoginSession{
		SessionID:    row.SessionID,
		UserID:       row.UserID,
		CompanyID:    row.CompanyID,
		Username:     row.Username,
		DisplayName:  row.DisplayName,
		UserIsActive: row.UserIsActive,
	}, nil
}

// RevokeSession ends the session of a token. Revoking an already revoked or
// unknown session is a no-op, so logout is idempotent.
func (q *Queries) RevokeSession(ctx context.Context, tokenHash []byte) error {
	if err := q.Queries.RevokeSession(ctx, tokenHash); err != nil {
		return fmt.Errorf("store: revoke session: %w", err)
	}
	return nil
}

// RevokeSessionsForUser ends every live session of a user, used when the user
// is deactivated so an old cookie cannot keep working.
func (q *Queries) RevokeSessionsForUser(ctx context.Context, userID uuid.UUID) error {
	if err := q.Queries.RevokeSessionsForUser(ctx, userID); err != nil {
		return fmt.Errorf("store: revoke sessions for user: %w", err)
	}
	return nil
}
