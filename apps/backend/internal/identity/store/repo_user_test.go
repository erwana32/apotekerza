package store

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestIsUniqueViolation_MatchesOnlyTheNamedConstraint(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		constraint string
		want       bool
	}{
		{
			name:       "named constraint",
			err:        &pgconn.PgError{Code: "23505", ConstraintName: "identity_users_username_key"},
			constraint: "identity_users_username_key",
			want:       true,
		},
		{
			name:       "different constraint",
			err:        &pgconn.PgError{Code: "23505", ConstraintName: "identity_sessions_token_hash_key"},
			constraint: "identity_users_username_key",
			want:       false,
		},
		{
			name:       "different SQLSTATE",
			err:        &pgconn.PgError{Code: "23514", ConstraintName: "identity_users_username_key"},
			constraint: "identity_users_username_key",
			want:       false,
		},
		{
			name:       "not a postgres error",
			err:        errors.New("boom"),
			constraint: "identity_users_username_key",
			want:       false,
		},
		{
			name:       "nil error",
			err:        nil,
			constraint: "identity_users_username_key",
			want:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isUniqueViolation(tt.err, tt.constraint); got != tt.want {
				t.Fatalf("isUniqueViolation() = %v, want %v", got, tt.want)
			}
		})
	}
}
