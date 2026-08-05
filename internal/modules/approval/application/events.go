package application

import (
	"context"
	"encoding/json"
	"time"

	"metaldocs/internal/platform/db"
)

// EventType identifies the kind of governance event recorded via EventEmitter.
type EventType string

// EventType values written to governance_events by the approval subsystem.
const (
	EventTypeApprovalInstanceCancelled EventType = "approval.instance_cancelled"
	EventTypeDocumentPublished         EventType = "document_published"
	EventTypeDocumentSuperseded        EventType = "document_superseded"
	EventTypePublishScheduled          EventType = "publish_scheduled"
	EventTypeSignoffRejected           EventType = "signoff.rejected"
	EventTypeSignoffRecorded           EventType = "signoff_recorded"
	EventTypeDocumentReviewed          EventType = "document_reviewed"

	// EventTypeApprovalAutoApproved (ADR 0087) records the instant approval of
	// an instance bound to a ZERO-stage route (livre profile). It is the audit
	// record of the fact that no human satisfied any stage — the payload names
	// the route and the route version that authorized it, so "why is this
	// approved with an empty signoff set?" is answerable from the trail alone.
	EventTypeApprovalAutoApproved EventType = "approval_auto_approved"

	// EventTypeReviewVerdictRecorded / EventTypeReviewChangesRequested (F4):
	// review-stage runtime verdicts. Ready mirrors EventTypeSignoffRecorded's
	// role (a non-terminal, quorum-counted vote); ChangesRequested mirrors
	// EventTypeSignoffRejected's role (an instance-collapsing verdict).
	EventTypeReviewVerdictRecorded  EventType = "review_verdict.recorded"
	EventTypeReviewChangesRequested EventType = "review_verdict.changes_requested"

	EventTypeRouteConfigCreated     EventType = "route.config.created"
	EventTypeRouteConfigUpdated     EventType = "route.config.updated"
	EventTypeRouteConfigDeactivated EventType = "route.config.deactivated"

	// EventTypeDelegationGranted / EventTypeDelegationRevoked (F9/ADR 0077):
	// approval delegation grant/revoke audit trail. "Use" of a delegation is
	// NOT a separate event type — the existing EventTypeSignoffRecorded /
	// EventTypeReviewVerdictRecorded events already carry the full audit
	// trail for the action and are widened with an "on_behalf_of" payload
	// field (mirrors F7's precedent of extending an existing event's payload
	// rather than inventing a new event type).
	EventTypeDelegationGranted EventType = "delegation.granted"
	EventTypeDelegationRevoked EventType = "delegation.revoked"

	// EventTypeApprovalSLAExtended (approval accountability loop, Task 9) is
	// the audit record of a governed exception: one instance's active-stage
	// deadline was postponed. It IS the extension history — no parallel
	// history table exists, so a lost event is a lost record.
	EventTypeApprovalSLAExtended EventType = "approval_sla_extended"
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

// Emit inserts ev into governance_events using tx, so the write shares the
// caller's transaction (outbox pattern — never a separate connection).
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

// Emit appends ev to Events. Test-only implementation; ignores ctx and tx.
func (m *MemoryEmitter) Emit(_ context.Context, _ db.Tx, ev GovernanceEvent) error {
	m.Events = append(m.Events, ev)
	return nil
}
