package notificationsinfra

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	"github.com/riverqueue/river"

	approvaldomain "metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/iam/authz"
	platformdb "metaldocs/internal/platform/db"
)

// approvalMessages is this worker's own pt-BR catalogue, keyed by event type and
// then by subject kind, because "o documento" and "o modelo" are not the same
// sentence. Missing keys are impossible by construction: the Work switch rejects
// any event type not listed here before a lookup happens.
var approvalMessages = map[string]map[string][2]string{
	approvaldomain.EventTypeApprovalPending: {
		"document": {"Aprovação pendente com você", "Um documento foi submetido e aguarda sua aprovação."},
		"template": {"Aprovação pendente com você", "Um modelo foi submetido e aguarda sua aprovação."},
	},
	approvaldomain.EventTypeApprovalOverdue: {
		"document": {"Aprovação vencida", "Um documento que aguarda sua aprovação passou do prazo."},
		"template": {"Aprovação vencida", "Um modelo que aguarda sua aprovação passou do prazo."},
	},
}

// ApprovalNotifyWorker consumes ApprovalNotificationArgs and inserts one
// notification row per recipient.
//
// It is DELIVERY ONLY. It contains no approval SQL and knows nothing about
// stages, routes or eligibility: the recipients were resolved by the approval
// module and travel in the envelope. That is what keeps the module boundary
// (invariant 6) intact — see ADR 0082.
//
// Runs after commit, so it is H-PRE-1 safe: no authz.Require and no lock is
// taken here (authz.SeedTxTenant is a SET LOCAL config write, not an
// authz-recording read).
type ApprovalNotifyWorker struct {
	river.WorkerDefaults[approvaldomain.ApprovalNotificationArgs]
	runner platformdb.TxRunner
}

// NewApprovalNotifyWorker builds a ready worker. runner is required.
func NewApprovalNotifyWorker(runner platformdb.TxRunner) *ApprovalNotifyWorker {
	if runner == nil {
		panic("approval_notify_worker: db is required")
	}
	return &ApprovalNotifyWorker{runner: runner}
}

// Work delivers one event to every recipient inside a single tenant-seeded
// transaction. The event is single-tenant by construction (args.TenantID), so
// one seeded tx covers every insert without mixing tenants (ADR 0054 rule 2).
func (w *ApprovalNotifyWorker) Work(ctx context.Context, job *river.Job[approvaldomain.ApprovalNotificationArgs]) error {
	args := job.Args

	bySubject, ok := approvalMessages[args.EventType]
	if !ok {
		// NOT a silent drop. An unrecognised event type means producer and
		// consumer have diverged, and River must retry and then dead-letter so
		// the divergence is visible.
		return fmt.Errorf("approval_notify_worker: unknown event type %q", args.EventType)
	}

	// An empty recipient list is legitimate, not an error: a livre route
	// (ADR 0087) has zero stages and therefore no eligible actors. Checked
	// before the subject-kind lookup so a livre route's empty-recipient event
	// (which may carry no meaningful subject kind either) is never rejected on
	// that account.
	if len(args.RecipientUserIDs) == 0 {
		return nil
	}

	msgs, ok := bySubject[args.SubjectKind]
	if !ok {
		return fmt.Errorf("approval_notify_worker: unknown subject kind %q for event %s", args.SubjectKind, args.EventType)
	}

	return w.runner.Do(ctx, func(tx *sql.Tx) error {
		if err := authz.SeedTxTenant(ctx, tx, args.TenantID); err != nil {
			return fmt.Errorf("approval_notify_worker: seed tenant: %w", err)
		}

		// One set-based insert over the recipient array — no per-recipient round
		// trips, and idempotent on redelivery via the same partial unique index the
		// lifecycle fanout uses. pq.Array is the repo's existing text[]/uuid[]
		// binding helper (see internal/modules/approval/infrastructure/postgres_approval_repository.go
		// and internal/modules/documents/infrastructure/repository.go) — no new
		// driver dependency added.
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO metaldocs.notifications
			   (tenant_id, recipient_user_id, event_type, resource_type, resource_id, title, message, source_event_id)
			 SELECT $1::uuid, r, $2, $3, $4, $5, $6, $7::uuid
			   FROM unnest($8::text[]) AS r
			 ON CONFLICT (recipient_user_id, source_event_id) WHERE source_event_id IS NOT NULL
			 DO NOTHING`,
			args.TenantID, args.EventType, args.SubjectKind, args.SubjectID,
			msgs[0], msgs[1], args.EventID, pq.Array(args.RecipientUserIDs),
		); err != nil {
			return fmt.Errorf("approval_notify_worker: insert notifications for %s: %w", args.SubjectID, err)
		}
		return nil
	})
}
