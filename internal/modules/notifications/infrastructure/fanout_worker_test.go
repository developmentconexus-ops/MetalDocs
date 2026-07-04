package notificationsinfra

import (
	"context"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/riverqueue/river"

	documentsdomain "metaldocs/internal/modules/documents/domain"
)

// M3 F3.2 PG-2 (validation-contract.md §2.2 site 4) — the notifications-fanout
// worker must wrap its per-event work in a transaction seeded with
// authz.SeedTxTenant BEFORE the obligated-readers query / notification
// inserts, engaging the FORCE RLS backstop for this single-tenant tx.

func TestNotificationsFanoutWorker_Work_SeedsTenantBeforeReaderInsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	worker := NewNotificationsFanoutWorker(db)

	mock.MatchExpectationsInOrder(true)
	mock.ExpectBegin()
	mock.ExpectExec(`SELECT set_config\('metaldocs\.tenant_id', \$1, true\)`).
		WithArgs("tenant-fanout-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT user_id FROM metaldocs\.v_cd_obligated_readers`).
		WithArgs("tenant-fanout-1", "cd-1").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow("user-1"))
	mock.ExpectExec(`INSERT INTO metaldocs\.notifications`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	job := &river.Job[documentsdomain.LifecycleEventArgs]{Args: documentsdomain.LifecycleEventArgs{
		EventID:              "evt-1",
		TenantID:             "tenant-fanout-1",
		EventType:            documentsdomain.EventTypeDocumentPublished,
		ResourceType:         "document",
		ResourceID:           "doc-1",
		ControlledDocumentID: "cd-1",
		OccurredAt:           time.Now(),
	}}

	if err := worker.Work(context.Background(), job); err != nil {
		t.Fatalf("Work: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations (seed-before-write ordering violated?): %v", err)
	}
}

func TestNotificationsFanoutWorker_Work_SeedsTenantBeforeAuthorInsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	worker := NewNotificationsFanoutWorker(db)

	mock.MatchExpectationsInOrder(true)
	mock.ExpectBegin()
	mock.ExpectExec(`SELECT set_config\('metaldocs\.tenant_id', \$1, true\)`).
		WithArgs("tenant-fanout-2").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO metaldocs\.notifications`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	job := &river.Job[documentsdomain.LifecycleEventArgs]{Args: documentsdomain.LifecycleEventArgs{
		EventID:      "evt-2",
		TenantID:     "tenant-fanout-2",
		EventType:    documentsdomain.EventTypeDocumentApproved,
		ResourceType: "approval_instance",
		ResourceID:   "ai-1",
		SubmittedBy:  "user-2",
		OccurredAt:   time.Now(),
	}}

	if err := worker.Work(context.Background(), job); err != nil {
		t.Fatalf("Work: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations (seed-before-write ordering violated?): %v", err)
	}
}

// TestNotificationsFanoutWorker_Work_MultipleReaders_AllInserted is a
// regression test for the live bug (River job kind=notification_fanout,
// error "insert notification for admin: driver: bad connection"): issuing
// tx.ExecContext for each reader while the obligated-readers rows cursor
// from tx.QueryContext was still open on the same *sql.Tx broke the single
// underlying connection. The fix buffers all reader IDs, closes/drains the
// query, then inserts. This asserts every buffered reader still gets its
// own INSERT once the cursor is fully drained first.
func TestNotificationsFanoutWorker_Work_MultipleReaders_AllInserted(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	worker := NewNotificationsFanoutWorker(db)

	mock.MatchExpectationsInOrder(true)
	mock.ExpectBegin()
	mock.ExpectExec(`SELECT set_config\('metaldocs\.tenant_id', \$1, true\)`).
		WithArgs("tenant-fanout-4").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT user_id FROM metaldocs\.v_cd_obligated_readers`).
		WithArgs("tenant-fanout-4", "cd-4").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).
			AddRow("user-1").
			AddRow("user-2").
			AddRow("admin"))
	mock.ExpectExec(`INSERT INTO metaldocs\.notifications`).
		WithArgs("tenant-fanout-4", "user-1", documentsdomain.EventTypeDocumentPublished, "document", "doc-4",
			sqlmock.AnyArg(), sqlmock.AnyArg(), "evt-4").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO metaldocs\.notifications`).
		WithArgs("tenant-fanout-4", "user-2", documentsdomain.EventTypeDocumentPublished, "document", "doc-4",
			sqlmock.AnyArg(), sqlmock.AnyArg(), "evt-4").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO metaldocs\.notifications`).
		WithArgs("tenant-fanout-4", "admin", documentsdomain.EventTypeDocumentPublished, "document", "doc-4",
			sqlmock.AnyArg(), sqlmock.AnyArg(), "evt-4").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	job := &river.Job[documentsdomain.LifecycleEventArgs]{Args: documentsdomain.LifecycleEventArgs{
		EventID:              "evt-4",
		TenantID:             "tenant-fanout-4",
		EventType:            documentsdomain.EventTypeDocumentPublished,
		ResourceType:         "document",
		ResourceID:           "doc-4",
		ControlledDocumentID: "cd-4",
		OccurredAt:           time.Now(),
	}}

	if err := worker.Work(context.Background(), job); err != nil {
		t.Fatalf("Work: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations (multi-reader fanout regression): %v", err)
	}
}

// TestNotificationsFanoutWorker_Work_UnhandledEventType_NoTxOpened proves the
// default no-op path does not open a tx at all (no seed needed for an event
// type the worker does not act on).
func TestNotificationsFanoutWorker_Work_UnhandledEventType_NoTxOpened(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	worker := NewNotificationsFanoutWorker(db)

	job := &river.Job[documentsdomain.LifecycleEventArgs]{Args: documentsdomain.LifecycleEventArgs{
		EventID:    "evt-3",
		TenantID:   "tenant-fanout-3",
		EventType:  "some.unhandled.event",
		OccurredAt: time.Now(),
	}}

	if err := worker.Work(context.Background(), job); err != nil {
		t.Fatalf("Work: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB interaction for unhandled event type: %v", err)
	}
}
