package app

import (
	"github.com/erwana32/apotekerza/internal/identity/domain"
	"github.com/erwana32/apotekerza/internal/platform/http/gen"
)

// PermissionRegistry is the resolved registry: every x-permission in
// api/openapi.yaml. Deriving it from the generated spec rather than a hand-kept
// list means a new endpoint cannot be deployed without its permission being
// enforceable and listable (docs/modules/identity.md).
type PermissionRegistry struct {
	domain *domain.PermissionRegistry
	rules  PermissionRules
}

// NewPermissionRegistry builds the registry and the guard rules from the
// embedded OpenAPI spec. It panics only when the embedded spec itself is
// malformed, which is a build problem detected at startup, never a request-time
// condition.
func NewPermissionRegistry() *PermissionRegistry {
	spec, err := gen.GetSpec()
	if err != nil {
		panic("identity: embedded OpenAPI spec is invalid: " + err.Error())
	}

	var permissions []string
	rules := PermissionRules{
		Permissions: make(map[string]string),
		Public:      make(map[string]struct{}),
	}
	for _, pathItem := range spec.Paths.Map() {
		for _, operation := range pathItem.Operations() {
			if operation.OperationID == "" {
				continue
			}
			if value, ok := operation.Extensions["x-public"]; ok {
				if flag, ok := value.(bool); ok && flag {
					rules.Public[operation.OperationID] = struct{}{}
				}
			}
			value, ok := operation.Extensions["x-permission"]
			if !ok {
				continue
			}
			permission, ok := value.(string)
			if !ok || permission == "" {
				continue
			}
			rules.Permissions[operation.OperationID] = permission
			permissions = append(permissions, permission)
		}
	}

	return &PermissionRegistry{
		domain: domain.NewRegistry(permissions),
		rules:  rules,
	}
}

// All returns every known permission, sorted.
func (r *PermissionRegistry) All() []string { return r.domain.All() }

// Has reports whether permission is in the registry.
func (r *PermissionRegistry) Has(permission string) bool { return r.domain.Has(permission) }

// Rules returns the per-operation permission table the guard uses.
func (r *PermissionRegistry) Rules() PermissionRules { return r.rules }

// PermissionRules is the permission table read from the spec: the x-permission
// of every operationId, and the operations marked x-public.
type PermissionRules struct {
	Permissions map[string]string
	Public      map[string]struct{}
}

// IsPublic reports whether the operation needs no session.
func (p PermissionRules) IsPublic(operationID string) bool {
	_, ok := p.Public[operationID]
	return ok
}

// Permission returns the permission an operation requires.
func (p PermissionRules) Permission(operationID string) (string, bool) {
	permission, ok := p.Permissions[operationID]
	return permission, ok
}
