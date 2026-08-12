package notificationsinfra

import (
	"context"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/riverqueue/river"

	documentsdomain "metaldocs/internal/modules/documents/domain"
	platformdb "metaldocs/internal/platform/db"
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
	defer func() { _ = db.Close() }()

	worker := NewNotificationsFanoutWorker(platformdb.NewTxRunner(db))

	mock.MatchExpectationsInOrder(true)
	mock.ExpectBegin()
	mock.ExpectExec(`SELECT set_config\('metaldocs\.tenant_id', \$1, true\)`).
		WithArgs("tenant-fanout-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO metaldocs\.notifications[\s\S]*SELECT[\s\S]*FROM metaldocs\.v_cd_obligated_readers`).
		WithArgs("tenant-fanout-1", "cd-1", documentsdomain.EventTypeDocumentPublished, "document", "doc-1",
			sqlmock.AnyArg(), sqlmock.AnyArg(), "evt-1").
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
	defer func() { _ = db.Close() }()

	worker := NewNotificationsFanoutWorker(platformdb.NewTxRunner(db))

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

// TestNotificationsFanoutWorker_Work_MultipleReaders_AllInserted — formerly a
// regression test for the "driver: bad connection" cursor/exec interleave bug.
// The fanout is now a single set-based INSERT...SELECT targeting the
// v_cd_obligated_readers view, so multi-reader cardinality is enforced by
// Postgres, not a Go loop: one exec, RowsAffected = number of readers. Real
// multi-row coverage lives in the integration race test
// (fanout_worker_race_integration_test.go).
func TestNotificationsFanoutWorker_Work_MultipleReaders_AllInserted(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer func() { _ = db.Close() }()

	worker := NewNotificationsFanoutWorker(platformdb.NewTxRunner(db))

	mock.MatchExpectationsInOrder(true)
	mock.ExpectBegin()
	mock.ExpectExec(`SELECT set_config\('metaldocs\.tenant_id', \$1, true\)`).
		WithArgs("tenant-fanout-4").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO metaldocs\.notifications[\s\S]*SELECT[\s\S]*FROM metaldocs\.v_cd_obligated_readers`).
		WithArgs("tenant-fanout-4", "cd-4", documentsdomain.EventTypeDocumentPublished, "document", "doc-4",
			sqlmock.AnyArg(), sqlmock.AnyArg(), "evt-4").
		WillReturnResult(sqlmock.NewResult(0, 3))
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

// TestNotificationsFanoutWorker_Work_UnhandledEventType_NoTxOpened_DELETED:
// this test used to assert that an unrecognised event type made Work return
// nil with no DB interaction. That encoded the exact defect Task 5 removes —
// a silently dropped event is a live regression risk with no error and no
// dead-letter trail. See TestNotificationsFanoutWorker_Work_UnhandledEventType_Errors
// below for the replacement behavior.
func TestNotificationsFanoutWorker_Work_UnhandledEventType_Errors(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer func() { _ = db.Close() }()

	worker := NewNotificationsFanoutWorker(platformdb.NewTxRunner(db))

	job := &river.Job[documentsdomain.LifecycleEventArgs]{Args: documentsdomain.LifecycleEventArgs{
		EventID:    "evt-3",
		TenantID:   "tenant-fanout-3",
		EventType:  "some.unhandled.event",
		OccurredAt: time.Now(),
	}}

	if err := worker.Work(context.Background(), job); err == nil {
		t.Fatal("Work: want error for unrecognised event type, got nil (a dropped event must never be silent)")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB interaction for unhandled event type: %v", err)
	}
}
