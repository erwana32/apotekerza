package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/erwana32/apotekerza/internal/identity/store/gen"
)

// InsertAuditEvent appends one audit row. The table is append-only (trigger),
// so there is deliberately no update or delete in this file.
func (q *Queries) InsertAuditEvent(ctx context.Context, params gen.InsertAuditEventParams) (gen.IdentityAuditLog, error) {
	event, err := q.Queries.InsertAuditEvent(ctx, params)
	if err != nil {
		return gen.IdentityAuditLog{}, fmt.Errorf("store: insert audit event: %w", err)
	}
	return event, nil
}

// ListAuditEvents returns one page of audit rows, newest first.
func (q *Queries) ListAuditEvents(ctx context.Context, params gen.ListAuditEventsParams) ([]gen.IdentityAuditLog, error) {
	events, err := q.Queries.ListAuditEvents(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("store: list audit events: %w", err)
	}
	return events, nil
}

// ListAuditEventsForEntity returns the history of one record, oldest first.
func (q *Queries) ListAuditEventsForEntity(ctx context.Context, entity string, entityID uuid.UUID) ([]gen.IdentityAuditLog, error) {
	events, err := q.Queries.ListAuditEventsForEntity(ctx, gen.ListAuditEventsForEntityParams{
		Entity:   entity,
		EntityID: pgUUID(entityID),
	})
	if err != nil {
		return nil, fmt.Errorf("store: list audit events for entity: %w", err)
	}
	return events, nil
}

// EncodeChanges marshals a change map for the jsonb column.
func EncodeChanges(changes map[string]any) ([]byte, error) {
	if changes == nil {
		return []byte("{}"), nil
	}
	raw, err := json.Marshal(changes)
	if err != nil {
		return nil, fmt.Errorf("store: encode audit changes: %w", err)
	}
	return raw, nil
}

// DecodeChanges parses a jsonb column into a map.
func DecodeChanges(raw []byte) (map[string]any, error) {
	if len(raw) == 0 {
		return map[string]any{}, nil
	}
	var changes map[string]any
	if err := json.Unmarshal(raw, &changes); err != nil {
		return nil, fmt.Errorf("store: decode audit changes: %w", err)
	}
	return changes, nil
}

// pgUUID wraps a uuid.UUID in the pgtype form sqlc uses for nullable uuid.
func pgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

// UUIDFromPg converts a nullable pgtype.UUID to a *uuid.UUID.
func UUIDFromPg(value pgtype.UUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}
	id := uuid.UUID(value.Bytes)
	return &id
}
