package application

import (
	"context"
	"encoding/json"
	"time"

	"metaldocs/internal/platform/db"
)

type EventType string

const (
	EventTypeApprovalInstanceCancelled EventType = "approval.instance_cancelled"
	EventTypeDocumentPublished         EventType = "document_published"
	EventTypePublishScheduled          EventType = "publish_scheduled"
	EventTypeSignoffRejected           EventType = "signoff.rejected"
	EventTypeSignoffRecorded           EventType = "signoff_recorded"

	EventTypeRouteConfigCreated     EventType = "route.config.created"
	EventTypeRouteConfigUpdated     EventType = "route.config.updated"
	EventTypeRouteConfigDeactivated EventType = "route.config.deactivated"
)

// GovernanceEvent mirrors the governance_events table columns.
type GovernanceEvent struct {
	TenantID     string
	EventType    EventType
	ActorUserID  string
	ResourceType string
	ResourceID   string
	Reason       string
	PayloadJSON  json.RawMessage
	OccurredAt   time.Time
}

// EventEmitter writes governance events within the caller's transaction.
// The tx must be the same transaction as the state-change write (outbox pattern).
type EventEmitter interface {
	Emit(ctx context.Context, tx db.Tx, event GovernanceEvent) error
}

// sqlEmitter is the default production implementation.
type sqlEmitter struct{}

// NewSQLEmitter returns the production event emitter.
func NewSQLEmitter() EventEmitter { return &sqlEmitter{} }

const insertEventSQL = `
INSERT INTO governance_events
  (tenant_id, event_type, actor_user_id, resource_type, resource_id, reason, payload_json, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, COALESCE($8::timestamptz, now()))`

func (e *sqlEmitter) Emit(ctx context.Context, tx db.Tx, ev GovernanceEvent) error {
	payload := ev.PayloadJSON
	if payload == nil {
		payload = json.RawMessage("{}")
	}
	var occurredAt any
	if !ev.OccurredAt.IsZero() {
		occurredAt = ev.OccurredAt.UTC()
	}
	_, err := tx.ExecContext(ctx, insertEventSQL,
		ev.TenantID, string(ev.EventType), ev.ActorUserID,
		ev.ResourceType, ev.ResourceID, ev.Reason, payload, occurredAt,
	)
	return err
}

// MemoryEmitter is an in-memory EventEmitter for tests.
type MemoryEmitter struct {
	Events []GovernanceEvent
}

func (m *MemoryEmitter) Emit(_ context.Context, _ db.Tx, ev GovernanceEvent) error {
	m.Events = append(m.Events, ev)
	return nil
}
