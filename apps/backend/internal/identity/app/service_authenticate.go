package app

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/erwana32/apotekerza/internal/identity/domain"
	"github.com/erwana32/apotekerza/internal/identity/store"
	"github.com/erwana32/apotekerza/internal/identity/store/gen"
	"github.com/erwana32/apotekerza/internal/platform/db"
	"github.com/erwana32/apotekerza/internal/platform/ids"
)

// LoginResult is what a successful login returns to the handler.
type LoginResult struct {
	User        domain.User
	Permissions []string
	// Token is the raw session token for the cookie. It is never logged.
	Token     string
	ExpiresAt time.Time
}

// Authenticate checks a username and password, opens a session, and records the
// login in the audit trail. A wrong username and a wrong password both return
// ErrInvalidCredentials, so the response cannot probe which users exist.
func (s *Service) Authenticate(ctx context.Context, username, password string) (LoginResult, error) {
	username = domain.NormalizeUsername(username)

	token, tokenHash, err := NewSessionToken()
	if err != nil {
		return LoginResult{}, asDomainError(err)
	}

	var result LoginResult
	// One transaction: look up the user, verify, insert the session, audit it.
	err = db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		queries := store.New(tx)

		user, err := queries.GetUserByUsername(ctx, username)
		if errors.Is(err, store.ErrUserNotFound) {
			// Do the same argon2id work as for a real user, so timing does not
			// reveal which usernames exist.
			domain.BurnPasswordCheck(password)
			return domain.ErrInvalidCredentials
		}
		if err != nil {
			return err
		}

		if err := domain.VerifyPassword(user.PasswordHash, password); err != nil {
			if errors.Is(err, domain.ErrPasswordMismatch) {
				return domain.ErrInvalidCredentials
			}
			return err
		}
		if !user.IsActive {
			return domain.ErrUserInactive
		}

		expiresAt := s.clock.Now().Add(SessionTTL)
		sessionID, err := ids.New()
		if err != nil {
			return err
		}
		if _, err := queries.InsertSession(ctx, gen.InsertSessionParams{
			ID:        sessionID,
			UserID:    user.ID,
			TokenHash: tokenHash,
			ExpiresAt: expiresAt,
		}); err != nil {
			return err
		}

		permissions, err := queries.ListPermissionsForUser(ctx, user.ID)
		if err != nil {
			return err
		}
		groupIDs, err := queries.ListGroupIDsForUser(ctx, user.ID)
		if err != nil {
			return err
		}

		if err := recordAudit(ctx, queries, auditInput{
			companyID: user.CompanyID,
			actorID:   &user.ID,
			action:    domain.ActionLogin,
			entity:    domain.EntitySession,
			entityID:  &sessionID,
		}); err != nil {
			return err
		}

		result = LoginResult{
			User: domain.User{
				ID:           user.ID,
				CompanyID:    user.CompanyID,
				Username:     user.Username,
				PasswordHash: "", // never leaves the service
				DisplayName:  user.DisplayName,
				IsActive:     user.IsActive,
				GroupIDs:     groupIDs,
				CreatedAt:    user.CreatedAt,
				UpdatedAt:    user.UpdatedAt,
			},
			Permissions: permissions,
			Token:       token,
			ExpiresAt:   expiresAt,
		}
		return nil
	})
	if err != nil {
		return LoginResult{}, asDomainError(err)
	}
	return result, nil
}

// SessionContext is what the middleware knows about a signed-in request.
type SessionContext struct {
	UserID      uuid.UUID
	CompanyID   uuid.UUID
	Username    string
	DisplayName string
	Permissions []string
	SessionID   uuid.UUID
}

// CurrentSession resolves a cookie token to a live session and its
// permissions. An unknown, revoked, or expired token is ErrSessionExpired.
func (s *Service) CurrentSession(ctx context.Context, token string) (SessionContext, error) {
	tokenHash, err := HashEncodedToken(token)
	if err != nil {
		return SessionContext{}, domain.ErrSessionExpired
	}

	session, err := store.New(s.pool).GetSessionByTokenHash(ctx, tokenHash)
	if errors.Is(err, store.ErrSessionNotFound) {
		return SessionContext{}, domain.ErrSessionExpired
	}
	// Any other error is an infrastructure failure: report it as such instead of
	// treating the zero-value session as an inactive user (a 401 would log
	// everyone out when the database is down).
	if err != nil {
		return SessionContext{}, asDomainError(err)
	}
	if !session.UserIsActive {
		return SessionContext{}, domain.ErrSessionExpired
	}

	permissions, err := store.New(s.pool).ListPermissionsForUser(ctx, session.UserID)
	if err != nil {
		return SessionContext{}, asDomainError(err)
	}

	return SessionContext{
		UserID:      session.UserID,
		CompanyID:   session.CompanyID,
		Username:    session.Username,
		DisplayName: session.DisplayName,
		Permissions: permissions,
		SessionID:   session.SessionID,
	}, nil
}

// Logout revokes the session behind a token and records the logout. Logging out
// an unknown token succeeds, so the endpoint is idempotent.
func (s *Service) Logout(ctx context.Context, token string, actor *SessionContext) error {
	tokenHash, err := HashEncodedToken(token)
	if err != nil {
		return nil
	}
	err = db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		queries := store.New(tx)
		if err := queries.RevokeSession(ctx, tokenHash); err != nil {
			return err
		}
		if actor == nil {
			return nil
		}
		return recordAudit(ctx, queries, auditInput{
			companyID: actor.CompanyID,
			actorID:   &actor.UserID,
			action:    domain.ActionLogout,
			entity:    domain.EntitySession,
			entityID:  &actor.SessionID,
		})
	})
	if err != nil {
		return asDomainError(err)
	}
	return nil
}
