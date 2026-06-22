package notificationsinfra

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/riverqueue/river"

	documentsdomain "metaldocs/internal/modules/documents/domain"
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
// Runs after commit (H-PRE-1 compliant — no authz reads inside this worker).
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

func (w *NotificationsFanoutWorker) Work(ctx context.Context, job *river.Job[documentsdomain.LifecycleEventArgs]) error {
	args := job.Args
	switch args.EventType {
	case documentsdomain.EventTypeDocumentPublished,
		documentsdomain.EventTypeDocumentSuperseded,
		documentsdomain.EventTypeDocumentObsoleted:
		return w.fanoutToReaders(ctx, args)
	case documentsdomain.EventTypeDocumentApproved,
		documentsdomain.EventTypeDocumentRejected:
		return w.fanoutToAuthor(ctx, args)
	default:
		return nil
	}
}

func (w *NotificationsFanoutWorker) fanoutToReaders(ctx context.Context, args documentsdomain.LifecycleEventArgs) error {
	if args.ControlledDocumentID == "" {
		return nil
	}
	rows, err := w.db.QueryContext(ctx,
		`SELECT user_id FROM metaldocs.v_cd_obligated_readers
		  WHERE tenant_id = $1::uuid AND controlled_document_id = $2::uuid`,
		args.TenantID, args.ControlledDocumentID,
	)
	if err != nil {
		return fmt.Errorf("fanout_worker: query obligated readers: %w", err)
	}
	defer rows.Close()

	msgs := ptBRMessages[args.EventType]
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return fmt.Errorf("fanout_worker: scan user_id: %w", err)
		}
		if err := w.insertRow(ctx, args, userID, msgs[0], msgs[1]); err != nil {
			return err
		}
	}
	return rows.Err()
}

func (w *NotificationsFanoutWorker) fanoutToAuthor(ctx context.Context, args documentsdomain.LifecycleEventArgs) error {
	if args.SubmittedBy == "" {
		return nil
	}
	msgs := ptBRMessages[args.EventType]
	return w.insertRow(ctx, args, args.SubmittedBy, msgs[0], msgs[1])
}

func (w *NotificationsFanoutWorker) insertRow(ctx context.Context, args documentsdomain.LifecycleEventArgs, recipientUserID, title, message string) error {
	_, err := w.db.ExecContext(ctx,
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
