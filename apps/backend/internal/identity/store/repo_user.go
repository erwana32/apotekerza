// Package store is the persistence layer of module identity.
//
// repo_*.go files wrap the sqlc-generated code in gen/, which is never edited
// by hand. Queries live in queries/query_<entity>.sql and are regenerated with
// `make gen`.
package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/erwana32/apotekerza/internal/identity/store/gen"
)

var (
	// ErrUserNotFound is returned when no user matches the lookup.
	ErrUserNotFound = errors.New("identity: user not found")
	// ErrGroupNotFound is returned when no access group matches the lookup.
	ErrGroupNotFound = errors.New("identity: access group not found")
	// ErrDuplicateUsername is returned when the username already exists. Usernames
	// are unique across the installation.
	ErrDuplicateUsername = errors.New("identity: duplicate username")
	// ErrDuplicateGroupName is returned when the access group name already exists.
	ErrDuplicateGroupName = errors.New("identity: duplicate access group name")
	// ErrSessionNotFound is returned when the session cookie is unknown,
	// revoked, or expired.
	ErrSessionNotFound = errors.New("identity: session not found")
)

// Queries exposes the identity queries. A service passes either the pool or a
// transaction, so one use case can span several modules (ADR-0001).
type Queries struct {
	*gen.Queries
}

// New wraps a pgx connection (pool or transaction) for identity queries.
func New(db gen.DBTX) *Queries {
	return &Queries{Queries: gen.New(db)}
}

// GetUserByUsername looks a user up by username. A missing user is
// ErrUserNotFound rather than an empty value the caller might misread.
func (q *Queries) GetUserByUsername(ctx context.Context, username string) (gen.IdentityUser, error) {
	user, err := q.Queries.GetUserByUsername(ctx, username)
	if errors.Is(err, pgx.ErrNoRows) {
		return gen.IdentityUser{}, fmt.Errorf("store: user %q: %w", username, ErrUserNotFound)
	}
	if err != nil {
		return gen.IdentityUser{}, fmt.Errorf("store: get user by username: %w", err)
	}
	return user, nil
}

// GetUserByID fetches a user by its id, mapping a missing row to ErrUserNotFound.
func (q *Queries) GetUserByID(ctx context.Context, id uuid.UUID) (gen.IdentityUser, error) {
	user, err := q.Queries.GetUserByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return gen.IdentityUser{}, fmt.Errorf("store: user %s: %w", id, ErrUserNotFound)
	}
	if err != nil {
		return gen.IdentityUser{}, fmt.Errorf("store: get user by id: %w", err)
	}
	return user, nil
}

// InsertUser creates a user and maps a unique violation to ErrDuplicateUsername.
func (q *Queries) InsertUser(ctx context.Context, params gen.InsertUserParams) (gen.IdentityUser, error) {
	user, err := q.Queries.InsertUser(ctx, params)
	if isUniqueViolation(err, "identity_users_username_key") {
		return gen.IdentityUser{}, fmt.Errorf("store: %w", ErrDuplicateUsername)
	}
	if err != nil {
		return gen.IdentityUser{}, fmt.Errorf("store: insert user: %w", err)
	}
	return user, nil
}

// ListUsers returns one page of users ordered by username. The caller passes
// PageSize one larger than the page it shows, to learn whether another page
// exists, and AfterUsername (the last username of the previous page) for the
// next page.
func (q *Queries) ListUsers(ctx context.Context, params gen.ListUsersParams) ([]gen.IdentityUser, error) {
	users, err := q.Queries.ListUsers(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("store: list users: %w", err)
	}
	return users, nil
}

// ListPermissionsForUser returns the union of the permissions of the user's
// active groups.
func (q *Queries) ListPermissionsForUser(ctx context.Context, userID uuid.UUID) ([]string, error) {
	permissions, err := q.Queries.ListPermissionsForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("store: list permissions for user: %w", err)
	}
	return permissions, nil
}

// isUniqueViolation reports whether err violates the named unique constraint.
func isUniqueViolation(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == constraint
}
