package notificationsinfra

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/riverqueue/river"

	documentsdomain "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/iam/authz"
	platformdb "metaldocs/internal/platform/db"
)

var ptBRMessages = map[string][2]string{
	documentsdomain.EventTypeDocumentPublished:  {"Novo documento controlado publicado para leitura", "Um documento controlado que você deve ler foi publicado."},
	documentsdomain.EventTypeDocumentSuperseded: {"Documento substituído por nova revisão", "Um documento controlado que você deve ler foi substituído por nova revisão."},
	documentsdomain.EventTypeDocumentObsoleted:  {"Documento obsoletado", "Um documento controlado que você deve ler foi marcado como obsoleto."},
	documentsdomain.EventTypeDocumentApproved:   {"Seu documento foi aprovado", "O documento que você submeteu foi aprovado."},
	documentsdomain.EventTypeDocumentRejected:   {"Documento rejeitado — ajustes solicitados", "O documento que você submeteu foi rejeitado. Revise e resubmeta."},
}

// NotificationsFanoutWorker is a River worker that consumes LifecycleEventArgs
// jobs and inserts per-recipient rows into metaldocs.notifications.
// Runs after commit (H-PRE-1 compliant — no authz.Require/lock inside this
// worker; authz.SeedTxTenant is a SET LOCAL config write, not an
// authz-recording read, per M3 F3.2 validation-contract.md §2.2 site 4).
// Idempotent via partial unique index on (recipient_user_id, source_event_id).
type NotificationsFanoutWorker struct {
	river.WorkerDefaults[documentsdomain.LifecycleEventArgs]
	runner platformdb.TxRunner
}

// NewNotificationsFanoutWorker builds a ready worker. runner is required.
func NewNotificationsFanoutWorker(runner platformdb.TxRunner) *NotificationsFanoutWorker {
	if runner == nil {
		panic("notifications_fanout_worker: db is required")
	}
	return &NotificationsFanoutWorker{runner: runner}
}

// Work runs the whole fanout for one delivered event inside a single
// transaction seeded with the event's tenant (M3 F3.2 — validation-
// contract.md §2.2 site 4: "wrap the insert loop in a tx"). The event is
// single-tenant by construction (args.TenantID), so one seeded tx covers the
// obligated-readers read + every per-recipient insert without mixing tenants
// (ADR 0054 rule 2).
func (w *NotificationsFanoutWorker) Work(ctx context.Context, job *river.Job[documentsdomain.LifecycleEventArgs]) error {
	args := job.Args
	switch args.EventType {
	case documentsdomain.EventTypeDocumentPublished,
		documentsdomain.EventTypeDocumentSuperseded,
		documentsdomain.EventTypeDocumentObsoleted,
		documentsdomain.EventTypeDocumentApproved,
		documentsdomain.EventTypeDocumentRejected:
		// fall through to the seeded-tx path below.
	default:
		// Never silent. An unrecognised event type means the producer and this
		// consumer have diverged; returning nil dropped the event with no error
		// and no dead-letter, so the divergence was invisible. River now retries
		// and dead-letters it.
		return fmt.Errorf("fanout_worker: unknown event type %q", args.EventType)
	}

	return w.runner.Do(ctx, func(tx *sql.Tx) error {
		if err := authz.SeedTxTenant(ctx, tx, args.TenantID); err != nil {
			return fmt.Errorf("fanout_worker: seed tenant: %w", err)
		}
		switch args.EventType {
		case documentsdomain.EventTypeDocumentPublished,
			documentsdomain.EventTypeDocumentSuperseded,
			documentsdomain.EventTypeDocumentObsoleted:
			return w.fanoutToReaders(ctx, tx, args)
		case documentsdomain.EventTypeDocumentApproved,
			documentsdomain.EventTypeDocumentRejected:
			return w.fanoutToAuthor(ctx, tx, args)
		}
		return nil
	})
}

// fanoutToReaders inserts one notification per obligated reader with a single
// set-based INSERT...SELECT — the fanout is computed in Postgres (view
// v_cd_obligated_readers), so there are no per-recipient round-trips and no
// Go-side cursor/exec interleaving on the tx. Idempotent via the same partial
// unique ON CONFLICT (recipient_user_id, source_event_id) DO NOTHING.
func (w *NotificationsFanoutWorker) fanoutToReaders(ctx context.Context, tx *sql.Tx, args documentsdomain.LifecycleEventArgs) error {
	if args.ControlledDocumentID == "" {
		return nil
	}
	msgs := ptBRMessages[args.EventType]
	_, err := tx.ExecContext(ctx,
		`INSERT INTO metaldocs.notifications
		   (tenant_id, recipient_user_id, event_type, resource_type, resource_id, title, message, source_event_id)
		 SELECT $1::uuid, r.user_id, $3, $4, $5, $6, $7, $8::uuid
		   FROM metaldocs.v_cd_obligated_readers r
		  WHERE r.tenant_id = $1::uuid AND r.controlled_document_id = $2::uuid
		 ON CONFLICT (recipient_user_id, source_event_id) WHERE source_event_id IS NOT NULL
		 DO NOTHING`,
		args.TenantID, args.ControlledDocumentID, args.EventType, args.ResourceType, args.ResourceID,
		msgs[0], msgs[1], args.EventID,
	)
	if err != nil {
		return fmt.Errorf("fanout_worker: fanout insert for cd %s: %w", args.ControlledDocumentID, err)
	}
	return nil
}

func (w *NotificationsFanoutWorker) fanoutToAuthor(ctx context.Context, tx *sql.Tx, args documentsdomain.LifecycleEventArgs) error {
	if args.SubmittedBy == "" {
		return nil
	}
	msgs := ptBRMessages[args.EventType]
	return w.insertRow(ctx, tx, args, args.SubmittedBy, msgs[0], msgs[1])
}

func (w *NotificationsFanoutWorker) insertRow(ctx context.Context, tx *sql.Tx, args documentsdomain.LifecycleEventArgs, recipientUserID, title, message string) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO metaldocs.notifications
		   (tenant_id, recipient_user_id, event_type, resource_type, resource_id, title, message, source_event_id)
		 VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8::uuid)
		 ON CONFLICT (recipient_user_id, source_event_id) WHERE source_event_id IS NOT NULL
		 DO NOTHING`,
		args.TenantID, recipientUserID, args.EventType, args.ResourceType, args.ResourceID,
		title, message, args.EventID,
	)
	if err != nil {
		return fmt.Errorf("fanout_worker: insert notification for %s: %w", recipientUserID, err)
	}
	return nil
}
