package domain

import (
	"context"
	"time"

	"metaldocs/internal/platform/db"
)

// Approval notification event types. Deliberately two and no more: something
// arrived for you, and something you were given is late. Anything richer is a
// product decision, not a plumbing one.
const (
	EventTypeApprovalPending = "approval.pending"
	EventTypeApprovalOverdue = "approval.overdue"
)

// ApprovalNotificationArgs is the approval module's OWN notification envelope,
// distinct from documents.LifecycleEventArgs by design.
//
// The difference that matters is RecipientUserIDs. The document envelope names
// a document and lets the consumer work out who cares (obligated readers, or
// the author); an approval has no document to name when its subject is a
// template, and the people who care are precisely the eligible actors this
// module already snapshotted on the stage. Carrying the resolved list keeps the
// recipient logic inside the module that owns it and leaves `notifications` as
// delivery only — invariant 6, and the boundary ADR 0082 established.
type ApprovalNotificationArgs struct {
	// EventID is a uuid minted at emit. It is the idempotency key: the
	// consumer's ON CONFLICT (recipient_user_id, source_event_id) makes a
	// redelivered job a no-op.
	EventID string `json:"event_id"`
	// TenantID scopes the whole event. Single-tenant by construction, so one
	// seeded tx covers every insert (ADR 0054 rule 2).
	TenantID string `json:"tenant_id"`
	// EventType is EventTypeApprovalPending or EventTypeApprovalOverdue.
	EventType string `json:"event_type"`
	// SubjectKind is "document" or "template" — the M3 kernel's subject
	// discrimination, carried so the consumer can address the reader correctly
	// without needing to know how approval stores subjects.
	SubjectKind string `json:"subject_kind"`
	// SubjectID is the subject's key (a document id, or a template version id).
	SubjectID string `json:"subject_id"`
	// RecipientUserIDs is the RESOLVED recipient list. Empty is legitimate and
	// means nothing to deliver (a livre route has no stage and therefore no
	// eligible actors) — the consumer treats it as a no-op, never an error.
	RecipientUserIDs []string `json:"recipient_user_ids"`
	// ActorUserID is who caused the event (the submitter). Empty for events with
	// no human cause, such as the SLA surfacer's overdue tick.
	ActorUserID string    `json:"actor_user_id"`
	OccurredAt  time.Time `json:"occurred_at"`
}

// Kind implements river.JobArgs. A DISTINCT kind from "notification_fanout":
// the two envelopes have different shapes and different consumers, and sharing
// a kind would force one worker to branch on which envelope it received.
func (ApprovalNotificationArgs) Kind() string { return "approval_notification" }

// ApprovalNotificationEnqueuer is the approval-owned published port for
// enqueuing a notification inside the caller's transaction (the outbox
// discipline: the state write and the promise to notify commit together, and
// the network-shaped work happens after commit).
type ApprovalNotificationEnqueuer interface {
	EnqueueApprovalNotificationTx(ctx context.Context, tx db.Tx, args ApprovalNotificationArgs) error
}

// NoopApprovalNotificationEnqueuer is the fail-closed default used where no
// River client is wired (unit tests, the API binary's read paths). It enqueues
// nothing and reports success.
type NoopApprovalNotificationEnqueuer struct{}

func (NoopApprovalNotificationEnqueuer) EnqueueApprovalNotificationTx(context.Context, db.Tx, ApprovalNotificationArgs) error {
	return nil
}
