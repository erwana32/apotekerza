// Package identity owns users, sessions, access groups, and the audit trail.
// Other modules import only this package; it is their only way in (ADR-0001).
package identity

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/erwana32/apotekerza/internal/identity/app"
	"github.com/erwana32/apotekerza/internal/identity/domain"
)

// Session is what the middleware knows about a signed-in request.
type Session = app.SessionContext

// AuditEvent is one audit entry written by another module.
type AuditEvent = app.RecordAuditEvent

// ErrSessionExpired is returned when the request has no live session.
var ErrSessionExpired = domain.ErrSessionExpired

// ErrPermissionDenied is returned when the user lacks the required permission.
var ErrPermissionDenied = domain.ErrPermissionDenied

// Module is the public interface of identity: authenticate, resolve the current
// user, check a permission, and record an audit event (docs/modules/identity.md,
// "Public interface").
type Module struct {
	service *app.Service
}

// New builds the identity module.
func New(service *app.Service) *Module {
	return &Module{service: service}
}

// Authenticate resolves a session cookie token to the signed-in user and its
// permissions. A missing, revoked, or expired session is ErrSessionExpired.
func (m *Module) Authenticate(ctx context.Context, token string) (Session, error) {
	session, err := m.service.CurrentSession(ctx, token)
	if err != nil {
		if errors.Is(err, domain.ErrSessionExpired) {
			return Session{}, ErrSessionExpired
		}
		return Session{}, err
	}
	return session, nil
}

// CurrentUser returns the signed-in user of a request. It is a convenience over
// Authenticate for handlers that already hold a Session.
func (m *Module) CurrentUser(session Session) (uuid.UUID, bool) {
	if session.UserID == uuid.Nil {
		return uuid.Nil, false
	}
	return session.UserID, true
}

// HasPermission reports whether the session holds permission. Permissions are
// the union of the user's active access groups (ADR-0007).
func (m *Module) HasPermission(session Session, permission string) bool {
	return domain.HasPermission(session.Permissions, permission)
}

// RecordAudit appends an audit event in its own transaction.
func (m *Module) RecordAudit(ctx context.Context, event AuditEvent) error {
	return m.service.RecordAudit(ctx, event)
}

// Service returns the underlying service for the module's own HTTP layer. It is
// not part of the cross-module contract: other modules must use the methods
// above.
func (m *Module) Service() *app.Service {
	return m.service
}
