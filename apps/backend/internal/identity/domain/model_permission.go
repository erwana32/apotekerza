package domain

import "sort"

// PermissionRegistry is every permission the installation knows. It is the
// pure form: it holds a set and sorts it. The OpenAPI spec is read into this
// type by the app layer, so the domain stays free of I/O.
type PermissionRegistry struct {
	permissions map[string]struct{}
	sorted      []string
}

// NewRegistry builds a registry from a permission list.
func NewRegistry(permissions []string) *PermissionRegistry {
	registry := &PermissionRegistry{permissions: make(map[string]struct{}, len(permissions))}
	for _, permission := range permissions {
		registry.permissions[permission] = struct{}{}
	}
	registry.sorted = make([]string, 0, len(registry.permissions))
	for permission := range registry.permissions {
		registry.sorted = append(registry.sorted, permission)
	}
	sort.Strings(registry.sorted)
	return registry
}

// All returns every known permission, sorted. The caller must not modify it.
func (r *PermissionRegistry) All() []string { return r.sorted }

// Has reports whether permission is in the registry.
func (r *PermissionRegistry) Has(permission string) bool {
	_, ok := r.permissions[permission]
	return ok
}

// Len returns the number of known permissions.
func (r *PermissionRegistry) Len() int { return len(r.permissions) }
