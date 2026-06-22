package domain

import (
	"context"
	"time"

	"metaldocs/internal/platform/db"
)

// LifecycleEventArgs is the River-job envelope for document lifecycle domain events.
// One kind, discriminated by EventType, keeps the fan-out worker simple (single worker).
// No `river` import here — domain stays infra-free; Kind() satisfies river.JobArgs via
// a plain string method, not a structural River dependency.
type LifecycleEventArgs struct {
	EventID              string    `json:"event_id"`               // uuid, minted at emit — idempotency key
	TenantID             string    `json:"tenant_id"`
	EventType            string    `json:"event_type"`             // one of the constants below
	ResourceType         string    `json:"resource_type"`
	ResourceID           string    `json:"resource_id"`
	ControlledDocumentID string    `json:"controlled_document_id"` // set for reader events; "" for author events
	SubmittedBy          string    `json:"submitted_by"`           // set for author events; "" for reader events
	OccurredAt           time.Time `json:"occurred_at"`
}

// Kind implements the river.JobArgs interface without importing River.
func (LifecycleEventArgs) Kind() string { return "notification_fanout" }

// Document lifecycle event types — the M3 bundle (ADR-0044 §3).
const (
	EventTypeDocumentPublished  = "document.published"
	EventTypeDocumentSuperseded = "document.superseded"
	EventTypeDocumentObsoleted  = "document.obsoleted"
	EventTypeDocumentApproved   = "document.approved"
	EventTypeDocumentRejected   = "document.rejected"
)

// LifecycleEventEnqueuer is the port the five emit sites depend on.
// Takes db.Tx (domain-clean) — the River infra adapter narrows to *sql.Tx internally.
type LifecycleEventEnqueuer interface {
	EnqueueLifecycleEventTx(ctx context.Context, tx db.Tx, args LifecycleEventArgs) error
}
