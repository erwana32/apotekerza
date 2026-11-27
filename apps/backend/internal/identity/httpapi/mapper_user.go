package httpapi

import (
	"github.com/erwana32/apotekerza/internal/identity/domain"
	"github.com/erwana32/apotekerza/internal/platform/http/gen"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// toAPIUser maps a domain user to the API shape.
func toAPIUser(user domain.User) gen.User {
	groupIDs := make([]openapi_types.UUID, 0, len(user.GroupIDs))
	for _, id := range user.GroupIDs {
		groupIDs = append(groupIDs, openapi_types.UUID(id))
	}
	return gen.User{
		Id:          openapi_types.UUID(user.ID),
		CompanyId:   openapi_types.UUID(user.CompanyID),
		Username:    user.Username,
		DisplayName: user.DisplayName,
		IsActive:    user.IsActive,
		GroupIds:    groupIDs,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}

// toAPIGroup maps a domain access group to the API shape.
func toAPIGroup(group domain.AccessGroup) gen.AccessGroup {
	permissions := group.Permissions
	if permissions == nil {
		permissions = []string{}
	}
	return gen.AccessGroup{
		Id:          openapi_types.UUID(group.ID),
		Name:        group.Name,
		IsBuiltin:   group.IsBuiltin,
		IsActive:    group.IsActive,
		Permissions: permissions,
		CreatedAt:   group.CreatedAt,
		UpdatedAt:   group.UpdatedAt,
	}
}

// toAPIAuditEvent maps a domain audit event to the API shape.
func toAPIAuditEvent(event domain.AuditEvent) gen.AuditEvent {
	changes := event.Changes
	if changes == nil {
		changes = map[string]any{}
	}
	result := gen.AuditEvent{
		Id:        openapi_types.UUID(event.ID),
		Action:    event.Action,
		Entity:    event.Entity,
		Changes:   changes,
		CreatedAt: event.CreatedAt,
	}
	if event.ActorUserID != nil {
		actor := openapi_types.UUID(*event.ActorUserID)
		result.ActorUserId = &actor
	}
	if event.EntityID != nil {
		entityID := openapi_types.UUID(*event.EntityID)
		result.EntityId = &entityID
	}
	return result
}
