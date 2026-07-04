package notificationsinfra

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/riverqueue/river"

	documentsdomain "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/iam/authz"
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
	db *sql.DB
}

// NewNotificationsFanoutWorker builds a ready worker. db is required.
func NewNotificationsFanoutWorker(db *sql.DB) *NotificationsFanoutWorker {
	if db == nil {
		panic("notifications_fanout_worker: db is required")
	}
	return &NotificationsFanoutWorker{db: db}
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
		return nil
	}

	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("fanout_worker: begin tx: %w", err)
	}
	if err := authz.SeedTxTenant(ctx, tx, args.TenantID); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("fanout_worker: seed tenant: %w", err)
	}

	switch args.EventType {
	case documentsdomain.EventTypeDocumentPublished,
		documentsdomain.EventTypeDocumentSuperseded,
		documentsdomain.EventTypeDocumentObsoleted:
		err = w.fanoutToReaders(ctx, tx, args)
	case documentsdomain.EventTypeDocumentApproved,
		documentsdomain.EventTypeDocumentRejected:
		err = w.fanoutToAuthor(ctx, tx, args)
	}
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("fanout_worker: commit: %w", err)
	}
	return nil
}

func (w *NotificationsFanoutWorker) fanoutToReaders(ctx context.Context, tx *sql.Tx, args documentsdomain.LifecycleEventArgs) error {
	if args.ControlledDocumentID == "" {
		return nil
	}
	rows, err := tx.QueryContext(ctx,
		`SELECT user_id FROM metaldocs.v_cd_obligated_readers
		  WHERE tenant_id = $1::uuid AND controlled_document_id = $2::uuid`,
		args.TenantID, args.ControlledDocumentID,
	)
	if err != nil {
		return fmt.Errorf("fanout_worker: query obligated readers: %w", err)
	}
	defer rows.Close()

	// Buffer all reader IDs before issuing any Exec on the same tx: a
	// *sql.Tx has a single underlying connection, and interleaving
	// ExecContext with an open Query result set on that connection
	// surfaces as "driver: bad connection" (confirmed live, River job
	// kind=notification_fanout). Collect first, insert after rows is
	// drained.
	var userIDs []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return fmt.Errorf("fanout_worker: scan user_id: %w", err)
		}
		userIDs = append(userIDs, userID)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("fanout_worker: iterate obligated readers: %w", err)
	}

	msgs := ptBRMessages[args.EventType]
	for _, userID := range userIDs {
		if err := w.insertRow(ctx, tx, args, userID, msgs[0], msgs[1]); err != nil {
			return err
		}
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
