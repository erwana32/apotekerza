package app

import (
	"testing"

	"github.com/erwana32/apotekerza/internal/identity/domain"
)

func TestPermissionRegistry_ReadsEveryXPemissionFromSpec(t *testing.T) {
	registry := NewPermissionRegistry()

	// Every operation in api/openapi.yaml is covered either by a permission or
	// by x-public, and the registry lists the permissions only.
	specPermissions := []string{
		"identity.audit.read",
		"identity.group.read",
		"identity.group.write",
		"identity.session.write",
		"identity.user.read",
		"identity.user.write",
		"platform.health.read",
	}
	all := registry.All()
	if len(all) != len(specPermissions) {
		t.Fatalf("registry = %v, want %v", all, specPermissions)
	}
	for i, permission := range specPermissions {
		if all[i] != permission {
			t.Fatalf("registry[%d] = %q, want %q", i, all[i], permission)
		}
	}
}

func TestPermissionRegistry_RulesComeFromTheSpec(t *testing.T) {
	rules := NewPermissionRegistry().Rules()

	if !rules.IsPublic("Login") {
		t.Fatal("Login must be public: the spec marks it x-public")
	}
	if !rules.IsPublic("GetHealth") {
		t.Fatal("GetHealth must be public: the spec marks it x-public")
	}
	if rules.IsPublic("GetMe") {
		t.Fatal("GetMe must require a session")
	}

	required, ok := rules.Permission("CreateUser")
	if !ok || required != "identity.user.write" {
		t.Fatalf("CreateUser permission = %q (found %v), want identity.user.write", required, ok)
	}
}

func TestPermissionRegistry_RejectsUnknownPermission(t *testing.T) {
	registry := NewPermissionRegistry()
	if registry.Has("identity.user.delete") {
		t.Fatal("a permission that no operation declares must not be known")
	}
	if !registry.Has("identity.user.read") {
		t.Fatal("identity.user.read is declared by GetMe and must be known")
	}
}

func TestValidatePermissions_RejectsUnknownAndDeduplicates(t *testing.T) {
	service := NewService(nil, NewPermissionRegistry(), nil)

	_, err := service.validatePermissions([]string{"identity.user.read", "identity.user.delete"})
	if !isDomainCode(err, "UNKNOWN_PERMISSION") {
		t.Fatalf("unknown permission error = %v, want UNKNOWN_PERMISSION", err)
	}

	got, err := service.validatePermissions([]string{
		"identity.user.write",
		"identity.user.read",
		"identity.user.write",
		"",
	})
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	want := []string{"identity.user.read", "identity.user.write"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("permissions = %v, want sorted and de-duplicated %v", got, want)
	}
}

// isDomainCode reports whether err carries the stable API code.
func isDomainCode(err error, code string) bool {
	domainErr, ok := err.(*domain.Error)
	return ok && domainErr.Code() == code
}
