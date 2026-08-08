//go:build integration
// +build integration

package notificationsinfra_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/riverqueue/river"

	documentsdomain "metaldocs/internal/modules/documents/domain"
	notificationsinfra "metaldocs/internal/modules/notifications/infrastructure"
	"metaldocs/tests/integration/testdb"
)

func TestNotificationsFanoutWorker(t *testing.T) {
	db, _ := testdb.Open(t)
	ctx := context.Background()
	worker := notificationsinfra.NewNotificationsFanoutWorker(db)

	makeJob := func(args documentsdomain.LifecycleEventArgs) *river.Job[documentsdomain.LifecycleEventArgs] {
		return &river.Job[documentsdomain.LifecycleEventArgs]{Args: args}
	}

	countByEvent := func(t *testing.T, tenantID, sourceEventID string) int {
		t.Helper()
		var n int
		if err := db.QueryRowContext(ctx,
			`SELECT count(*) FROM metaldocs.notifications
			  WHERE tenant_id = $1::uuid AND source_event_id = $2::uuid`,
			tenantID, sourceEventID,
		).Scan(&n); err != nil {
			t.Fatalf("count notifications: %v", err)
		}
		return n
	}

	seedUserGrant := func(t *testing.T, tenantID, cdID, userID string) {
		t.Helper()
		if _, err := db.ExecContext(ctx,
			`INSERT INTO public.controlled_document_user_grants
			   (tenant_id, controlled_document_id, user_id)
			 VALUES ($1::uuid, $2::uuid, $3)`,
			tenantID, cdID, userID,
		); err != nil {
			t.Fatalf("seed user grant: %v", err)
		}
	}

	t.Run("published_to_obligated_readers", func(t *testing.T) {
		ten := testdb.NewTenant(t, db)
		cd := testdb.NewControlledDoc(t, db, testdb.WithTenant(ten.ID))
		r1 := testdb.NewUser(t, db, testdb.WithTenant(ten.ID))
		r2 := testdb.NewUser(t, db, testdb.WithTenant(ten.ID))
		seedUserGrant(t, ten.ID, cd.ID, r1.ID)
		seedUserGrant(t, ten.ID, cd.ID, r2.ID)

		eventID := uuid.NewString()
		args := documentsdomain.LifecycleEventArgs{
			EventID: eventID, TenantID: ten.ID,
			EventType:    documentsdomain.EventTypeDocumentPublished,
			ResourceType: "document", ResourceID: uuid.NewString(),
			ControlledDocumentID: cd.ID, OccurredAt: time.Now(),
		}
		if err := worker.Work(ctx, makeJob(args)); err != nil {
			t.Fatalf("Work: %v", err)
		}
		if n := countByEvent(t, ten.ID, eventID); n != 2 {
			t.Errorf("published: want 2 notifications, got %d", n)
		}
	})

	t.Run("superseded_to_obligated_readers", func(t *testing.T) {
		ten := testdb.NewTenant(t, db)
		cd := testdb.NewControlledDoc(t, db, testdb.WithTenant(ten.ID))
		r := testdb.NewUser(t, db, testdb.WithTenant(ten.ID))
		seedUserGrant(t, ten.ID, cd.ID, r.ID)

		eventID := uuid.NewString()
		args := documentsdomain.LifecycleEventArgs{
			EventID: eventID, TenantID: ten.ID,
			EventType:    documentsdomain.EventTypeDocumentSuperseded,
			ResourceType: "document", ResourceID: uuid.NewString(),
			ControlledDocumentID: cd.ID, OccurredAt: time.Now(),
		}
		if err := worker.Work(ctx, makeJob(args)); err != nil {
			t.Fatalf("Work: %v", err)
		}
		if n := countByEvent(t, ten.ID, eventID); n != 1 {
			t.Errorf("superseded: want 1 notification, got %d", n)
		}
	})

	t.Run("obsoleted_to_obligated_readers", func(t *testing.T) {
		ten := testdb.NewTenant(t, db)
		cd := testdb.NewControlledDoc(t, db, testdb.WithTenant(ten.ID))
		r := testdb.NewUser(t, db, testdb.WithTenant(ten.ID))
		seedUserGrant(t, ten.ID, cd.ID, r.ID)

		eventID := uuid.NewString()
		args := documentsdomain.LifecycleEventArgs{
			EventID: eventID, TenantID: ten.ID,
			EventType:    documentsdomain.EventTypeDocumentObsoleted,
			ResourceType: "document", ResourceID: uuid.NewString(),
			ControlledDocumentID: cd.ID, OccurredAt: time.Now(),
		}
		if err := worker.Work(ctx, makeJob(args)); err != nil {
			t.Fatalf("Work: %v", err)
		}
		if n := countByEvent(t, ten.ID, eventID); n != 1 {
			t.Errorf("obsoleted: want 1 notification, got %d", n)
		}
	})

	t.Run("approved_to_submitter", func(t *testing.T) {
		ten := testdb.NewTenant(t, db)
		submitter := testdb.NewUser(t, db, testdb.WithTenant(ten.ID))

		eventID := uuid.NewString()
		args := documentsdomain.LifecycleEventArgs{
			EventID: eventID, TenantID: ten.ID,
			EventType:    documentsdomain.EventTypeDocumentApproved,
			ResourceType: "approval_instance", ResourceID: uuid.NewString(),
			SubmittedBy: submitter.ID, OccurredAt: time.Now(),
		}
		if err := worker.Work(ctx, makeJob(args)); err != nil {
			t.Fatalf("Work: %v", err)
		}
		if n := countByEvent(t, ten.ID, eventID); n != 1 {
			t.Errorf("approved: want 1 notification, got %d", n)
		}
	})

	t.Run("rejected_to_submitter", func(t *testing.T) {
		ten := testdb.NewTenant(t, db)
		submitter := testdb.NewUser(t, db, testdb.WithTenant(ten.ID))

		eventID := uuid.NewString()
		args := documentsdomain.LifecycleEventArgs{
			EventID: eventID, TenantID: ten.ID,
			EventType:    documentsdomain.EventTypeDocumentRejected,
			ResourceType: "approval_instance", ResourceID: uuid.NewString(),
			SubmittedBy: submitter.ID, OccurredAt: time.Now(),
		}
		if err := worker.Work(ctx, makeJob(args)); err != nil {
			t.Fatalf("Work: %v", err)
		}
		if n := countByEvent(t, ten.ID, eventID); n != 1 {
			t.Errorf("rejected: want 1 notification, got %d", n)
		}
	})

	t.Run("redelivery_is_noop", func(t *testing.T) {
		ten := testdb.NewTenant(t, db)
		cd := testdb.NewControlledDoc(t, db, testdb.WithTenant(ten.ID))
		r := testdb.NewUser(t, db, testdb.WithTenant(ten.ID))
		seedUserGrant(t, ten.ID, cd.ID, r.ID)

		eventID := uuid.NewString()
		args := documentsdomain.LifecycleEventArgs{
			EventID: eventID, TenantID: ten.ID,
			EventType:    documentsdomain.EventTypeDocumentPublished,
			ResourceType: "document", ResourceID: uuid.NewString(),
			ControlledDocumentID: cd.ID, OccurredAt: time.Now(),
		}
		job := makeJob(args)
		if err := worker.Work(ctx, job); err != nil {
			t.Fatalf("Work first: %v", err)
		}
		if err := worker.Work(ctx, job); err != nil {
			t.Fatalf("Work second (redelivery): %v", err)
		}
		if n := countByEvent(t, ten.ID, eventID); n != 1 {
			t.Errorf("redelivery: want exactly 1 notification, got %d", n)
		}
	})

	t.Run("no_cd_link_skips_readers", func(t *testing.T) {
		ten := testdb.NewTenant(t, db)

		eventID := uuid.NewString()
		args := documentsdomain.LifecycleEventArgs{
			EventID: eventID, TenantID: ten.ID,
			EventType:            documentsdomain.EventTypeDocumentPublished,
			ResourceType:         "document",
			ResourceID:           uuid.NewString(),
			ControlledDocumentID: "",
			OccurredAt:           time.Now(),
		}
		if err := worker.Work(ctx, makeJob(args)); err != nil {
			t.Fatalf("Work: %v", err)
		}
		if n := countByEvent(t, ten.ID, eventID); n != 0 {
			t.Errorf("no-cd-link: want 0 notifications, got %d", n)
		}
	})

}
