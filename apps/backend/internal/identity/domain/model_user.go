// Package domain holds the identity entities and the rules that need no
// database and no HTTP.
package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// User is a login account. It is a master record, not a document: it is
// deactivated (IsActive false), never deleted (docs/modules/identity.md).
type User struct {
	ID           uuid.UUID
	CompanyID    uuid.UUID
	Username     string
	PasswordHash string
	DisplayName  string
	IsActive     bool
	GroupIDs     []uuid.UUID
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// AccessGroup is a named set of permissions. Groups are global: they carry no
// company_id (ADR-0007, Odoo style).
type AccessGroup struct {
	ID          uuid.UUID
	Name        string
	IsBuiltin   bool
	IsActive    bool
	Permissions []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// AuditEvent is one append-only audit row. Changes holds field references
// only, never personal-data contents (hard rule 10).
type AuditEvent struct {
	ID          uuid.UUID
	CompanyID   uuid.UUID
	ActorUserID *uuid.UUID
	Action      string
	Entity      string
	EntityID    *uuid.UUID
	Changes     map[string]any
	CreatedAt   time.Time
}

// Audit action verbs. Later modules add their document verbs (confirm, post,
// cancel) when they post documents (ADR-0010).
const (
	ActionCreate     = "create"
	ActionUpdate     = "update"
	ActionDeactivate = "deactivate"
	ActionLogin      = "login"
	ActionLogout     = "logout"
)

// Audit entity names.
const (
	EntityUser        = "user"
	EntityAccessGroup = "access_group"
	EntitySession     = "session"
)

// NormalizeUsername trims and lowercases a username, so login and creation
// agree on one spelling. The database enforces the lowercase form with a CHECK.
func NormalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

// HasPermission reports whether a permission set contains permission.
// Permissions are the union of the user's active groups.
func HasPermission(permissions []string, permission string) bool {
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// PermissionAllowed reports whether every permission in wanted belongs to
// known, the registry built from api/openapi.yaml. The Administrator may only
// grant permissions that exist, so a typo cannot create a dead permission.
func PermissionAllowed(known []string, wanted string) bool {
	return HasPermission(known, wanted)
}
