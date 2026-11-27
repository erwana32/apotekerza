package app

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/erwana32/apotekerza/internal/identity/domain"
	"github.com/erwana32/apotekerza/internal/identity/store"
	"github.com/erwana32/apotekerza/internal/identity/store/gen"
	"github.com/erwana32/apotekerza/internal/platform/db"
)

// ListAuditEventsInput is the paged audit request.
type ListAuditEventsInput struct {
	Limit    int
	Cursor   string
	Entity   string
	EntityID *uuid.UUID
}

// AuditEventPage is one page of audit events.
type AuditEventPage struct {
	Items      []domain.AuditEvent
	NextCursor *string
}

// ListAuditEvents returns one page of the audit trail, newest first. The cursor
// is the last id of the previous page: timestamps can tie inside one
// transaction, and UUIDv7 ids are time-ordered, so id is the stable tiebreaker.
func (s *Service) ListAuditEvents(ctx context.Context, in ListAuditEventsInput) (AuditEventPage, error) {
	var afterID *uuid.UUID
	if in.Cursor != "" {
		decoded, err := decodeCursor(in.Cursor)
		if err != nil {
			return AuditEventPage{}, domain.ErrValidation("cursor", "is not valid")
		}
		parsed, err := uuid.Parse(decoded)
		if err != nil {
			return AuditEventPage{}, domain.ErrValidation("cursor", "is not valid")
		}
		afterID = &parsed
	}

	params := gen.ListAuditEventsParams{
		PageSize: int32(in.Limit + 1),
	}
	if in.Entity != "" {
		params.Entity = &in.Entity
	}
	if in.EntityID != nil {
		params.EntityID = toPgUUID(*in.EntityID)
	}
	if afterID != nil {
		params.AfterID = toPgUUID(*afterID)
	}

	rows, err := store.New(s.pool).ListAuditEvents(ctx, params)
	if err != nil {
		return AuditEventPage{}, asDomainError(err)
	}

	page := AuditEventPage{}
	if len(rows) > in.Limit {
		rows = rows[:in.Limit]
		next := encodeCursor(rows[len(rows)-1].ID.String())
		page.NextCursor = &next
	}
	for _, row := range rows {
		changes, err := store.DecodeChanges(row.Changes)
		if err != nil {
			return AuditEventPage{}, asDomainError(err)
		}
		page.Items = append(page.Items, domain.AuditEvent{
			ID:          row.ID,
			CompanyID:   row.CompanyID,
			ActorUserID: fromPgUUID(row.ActorUserID),
			Action:      row.Action,
			Entity:      row.Entity,
			EntityID:    fromPgUUID(row.EntityID),
			Changes:     changes,
			CreatedAt:   row.CreatedAt,
		})
	}
	return page, nil
}

// RecordAudit appends an audit event for another module (module.go's
// RecordAudit). Other modules call it inside their own transaction so the
// document change and its history commit together (ADR-0010).
type RecordAuditEvent struct {
	CompanyID   uuid.UUID
	ActorUserID *uuid.UUID
	Action      string
	Entity      string
	EntityID    *uuid.UUID
	Changes     map[string]any
}

// RecordAudit writes one audit event in its own transaction. Modules that
// already hold a transaction should call RecordAuditTx instead.
func (s *Service) RecordAudit(ctx context.Context, event RecordAuditEvent) error {
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		return s.RecordAuditTx(ctx, tx, event)
	})
	if err != nil {
		return asDomainError(err)
	}
	return nil
}

// RecordAuditTx writes one audit event inside the caller's transaction.
func (s *Service) RecordAuditTx(ctx context.Context, tx db.Executor, event RecordAuditEvent) error {
	err := recordAudit(ctx, store.New(tx), auditInput{
		companyID: event.CompanyID,
		actorID:   event.ActorUserID,
		action:    event.Action,
		entity:    event.Entity,
		entityID:  event.EntityID,
		changes:   event.Changes,
	})
	if err != nil {
		return asDomainError(err)
	}
	return nil
}

// EncodeChangesJSON is a small helper for callers that build jsonb by hand.
func EncodeChangesJSON(changes map[string]any) (json.RawMessage, error) {
	raw, err := json.Marshal(changes)
	if err != nil {
		return nil, err
	}
	return raw, nil
}
