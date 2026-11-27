package app

import (
	"context"

	"github.com/google/uuid"

	"github.com/erwana32/apotekerza/internal/identity/store"
	"github.com/erwana32/apotekerza/internal/identity/store/gen"
	"github.com/erwana32/apotekerza/internal/platform/ids"
)

// auditInput is one audit event. Changes carries field references only, never
// personal-data contents (hard rule 10).
type auditInput struct {
	companyID uuid.UUID
	actorID   *uuid.UUID
	action    string
	entity    string
	entityID  *uuid.UUID
	changes   map[string]any
}

// recordAudit appends one audit row inside the caller's transaction, so the
// audit entry and the change it describes commit or roll back together.
func recordAudit(ctx context.Context, queries *store.Queries, in auditInput) error {
	eventID, err := ids.New()
	if err != nil {
		return err
	}
	changes, err := store.EncodeChanges(in.changes)
	if err != nil {
		return err
	}
	params := gen.InsertAuditEventParams{
		ID:        eventID,
		CompanyID: in.companyID,
		Action:    in.action,
		Entity:    in.entity,
		Changes:   changes,
	}
	if in.actorID != nil {
		params.ActorUserID = toPgUUID(*in.actorID)
	}
	if in.entityID != nil {
		params.EntityID = toPgUUID(*in.entityID)
	}
	_, err = queries.InsertAuditEvent(ctx, params)
	return err
}
