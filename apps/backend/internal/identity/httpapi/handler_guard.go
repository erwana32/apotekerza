package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	nethttp "net/http"
	"time"

	"github.com/erwana32/apotekerza/internal/identity/app"
	"github.com/erwana32/apotekerza/internal/identity/domain"
	"github.com/erwana32/apotekerza/internal/platform/apierror"
	"github.com/erwana32/apotekerza/internal/platform/http/gen"
)

// SessionCookieName is the session cookie's name.
const SessionCookieName = "apotekerza_session"

type sessionKey struct{}
type tokenKey struct{}
type intentKey struct{}

// cookieIntent is what a handler asks the transport to do with the session
// cookie. The generated strict server hands a handler the request context only,
// so the handler records the intent and the guard writes the header, which is
// the one place that holds the ResponseWriter.
type cookieIntent struct {
	issueToken string
	expiresAt  time.Time
	hasIssue   bool
	clear      bool
}

// WithCookieIntent installs the per-request intent holder.
func WithCookieIntent(ctx context.Context) context.Context {
	return context.WithValue(ctx, intentKey{}, &cookieIntent{})
}

// IssueSessionCookie asks the transport to set the session cookie after the
// handler returns successfully.
func IssueSessionCookie(ctx context.Context, token string, expiresAt time.Time) {
	if intent, ok := ctx.Value(intentKey{}).(*cookieIntent); ok {
		intent.issueToken, intent.expiresAt, intent.hasIssue = token, expiresAt, true
	}
}

// ClearSessionCookie asks the transport to expire the session cookie.
func ClearSessionCookie(ctx context.Context) {
	if intent, ok := ctx.Value(intentKey{}).(*cookieIntent); ok {
		intent.clear = true
	}
}

// SessionFrom returns the signed-in session stored by the guard.
func SessionFrom(ctx context.Context) (app.SessionContext, bool) {
	session, ok := ctx.Value(sessionKey{}).(app.SessionContext)
	return session, ok
}

// SessionToken returns the raw session cookie token, for logout. It must never
// be logged or stored.
func SessionToken(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(tokenKey{}).(string)
	return token, ok && token != ""
}

// Guard is the strict middleware that enforces x-permission on every operation
// except those the spec marks x-public. It also writes the cookie a handler
// asked for.
//
// A StrictMiddlewareFunc sees each operationId and runs before the handler, so
// an endpoint added to the spec cannot be served unguarded.
type Guard struct {
	service *app.Service
	rules   app.PermissionRules
}

// NewGuard builds the guard from the permission table.
func NewGuard(service *app.Service, rules app.PermissionRules) gen.StrictMiddlewareFunc {
	guard := &Guard{service: service, rules: rules}
	return guard.Middleware
}

// Middleware implements gen.StrictMiddlewareFunc.
func (g *Guard) Middleware(f gen.StrictHandlerFunc, operationID string) gen.StrictHandlerFunc {
	isPublic := g.rules.IsPublic(operationID)
	required, guarded := g.rules.Permission(operationID)

	return func(ctx context.Context, w nethttp.ResponseWriter, r *nethttp.Request, request any) (any, error) {
		ctx = WithCookieIntent(ctx)

		token := ""
		if cookie, err := r.Cookie(SessionCookieName); err == nil {
			token = cookie.Value
			ctx = context.WithValue(ctx, tokenKey{}, token)
		}

		if !isPublic {
			// A guarded operation without a declared permission is a contract bug;
			// deny rather than serve it open.
			if !guarded || required == "" {
				return nil, domain.ErrPermissionDenied
			}
			if token == "" {
				return nil, domain.ErrSessionExpired
			}
			session, err := g.service.CurrentSession(ctx, token)
			if err != nil {
				if errors.Is(err, domain.ErrSessionExpired) {
					return nil, domain.ErrSessionExpired
				}
				return nil, err
			}
			if !domain.HasPermission(session.Permissions, required) {
				return nil, domain.ErrPermissionDenied
			}
			ctx = context.WithValue(ctx, sessionKey{}, session)
		}

		response, err := f(ctx, w, r, request)
		if err != nil {
			return nil, err
		}
		applyCookieIntent(w, ctx)
		return response, nil
	}
}

func applyCookieIntent(w nethttp.ResponseWriter, ctx context.Context) {
	intent, ok := ctx.Value(intentKey{}).(*cookieIntent)
	if !ok {
		return
	}
	if intent.clear {
		nethttp.SetCookie(w, &nethttp.Cookie{
			Name: SessionCookieName, Value: "", Path: "/", MaxAge: -1,
			HttpOnly: true, Secure: true, SameSite: nethttp.SameSiteLaxMode,
		})
	}
	if intent.hasIssue {
		nethttp.SetCookie(w, &nethttp.Cookie{
			Name: SessionCookieName, Value: intent.issueToken, Path: "/",
			Expires: intent.expiresAt, MaxAge: int(time.Until(intent.expiresAt).Seconds()),
			HttpOnly: true, Secure: true, SameSite: nethttp.SameSiteLaxMode,
		})
	}
}

// WriteDomainError writes the stable error shape for a failed operation.
func WriteDomainError(w nethttp.ResponseWriter, _ *nethttp.Request, err error) {
	writeError(w, err)
}

// WriteRequestError writes a request-shape failure (bad body, bad query) as a
// validation error rather than the generated plain-text 400.
func WriteRequestError(w nethttp.ResponseWriter, _ *nethttp.Request, err error) {
	writeError(w, domain.ErrValidation("request", err.Error()))
}

func writeError(w nethttp.ResponseWriter, err error) {
	apiErr := apierror.Map(err)
	status := apiErr.Status
	if status == 0 {
		status = nethttp.StatusInternalServerError
	}
	body := gen.Error{Code: apiErr.Code, Message: apiErr.Message}
	var domainErr *domain.Error
	if errors.As(err, &domainErr) && len(domainErr.Detail()) > 0 {
		detail := domainErr.Detail()
		body.Details = &detail
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
