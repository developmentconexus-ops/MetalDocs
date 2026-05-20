package jobs

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/riverqueue/river"

	approvalapp "metaldocs/internal/modules/documents/approval/application"
	approvalrepo "metaldocs/internal/modules/documents/approval/repository"
	"metaldocs/internal/testsupport/pgtest"
)

type fixedClock struct {
	t time.Time
}

func (c fixedClock) Now() time.Time {
	return c.t
}

func TestScheduledPublishWorker_PublishesWhenTruthMatches(t *testing.T) {
	db := pgtest.OpenAndMigrate(t)
	ctx := context.Background()
	effectiveAt := time.Date(2026, 5, 21, 14, 0, 0, 0, time.UTC)

	const (
		tenantID          = "11111111-aaaa-bbbb-cccc-111111111111"
		ownerUserID       = "22222222-aaaa-bbbb-cccc-222222222222"
		templateVersionID = "33333333-aaaa-bbbb-cccc-333333333333"
		controlledDocID   = "44444444-aaaa-bbbb-cccc-444444444444"
		documentID        = "55555555-aaaa-bbbb-cccc-555555555555"
	)

	seedScheduledDocument(t, ctx, db, scheduledDocumentSeed{
		tenantID:           tenantID,
		ownerUserID:        ownerUserID,
		templateVersionID:  templateVersionID,
		controlledDocID:    controlledDocID,
		documentID:         documentID,
		effectiveFrom:      effectiveAt,
		revisionVersion:    4,
		scheduleGeneration: 2,
	})

	emitter := &approvalapp.MemoryEmitter{}
	services := approvalapp.NewServices(approvalrepo.NewPostgresApprovalRepository(db), emitter, approvalapp.RealClock{})
	worker := &ScheduledPublishWorker{Service: services.Scheduler, DB: db}

	if err := worker.Work(ctx, &river.Job[ScheduledPublishArgs]{
		Args: ScheduledPublishArgs{
			TenantID:                tenantID,
			DocumentID:              documentID,
			ExpectedRevisionVersion: 4,
			ScheduledEffectiveAt:    effectiveAt,
			ScheduleGeneration:      2,
		},
	}); err != nil {
		t.Fatalf("Work: %v", err)
	}

	var status string
	var effectiveFrom sql.NullTime
	var revisionVersion int
	err := db.QueryRowContext(ctx, `
		SELECT status, effective_from, revision_version
		  FROM public.documents
		 WHERE tenant_id = $1
		   AND id = $2`,
		tenantID, documentID,
	).Scan(&status, &effectiveFrom, &revisionVersion)
	if err != nil {
		t.Fatalf("load document: %v", err)
	}
	if status != "published" {
		t.Fatalf("status = %q, want published", status)
	}
	if effectiveFrom.Valid {
		t.Fatalf("effective_from should be null after publish")
	}
	if revisionVersion != 5 {
		t.Fatalf("revision_version = %d, want 5", revisionVersion)
	}
	if len(emitter.Events) != 1 || emitter.Events[0].EventType != "document_published" {
		t.Fatalf("expected one document_published event, got %+v", emitter.Events)
	}
}

func TestScheduledPublishWorker_NoOpWhenGenerationIsStale(t *testing.T) {
	db := pgtest.OpenAndMigrate(t)
	ctx := context.Background()
	effectiveAt := time.Date(2026, 5, 21, 14, 0, 0, 0, time.UTC)

	const (
		tenantID          = "66666666-aaaa-bbbb-cccc-666666666666"
		ownerUserID       = "77777777-aaaa-bbbb-cccc-777777777777"
		templateVersionID = "88888888-aaaa-bbbb-cccc-888888888888"
		controlledDocID   = "99999999-aaaa-bbbb-cccc-999999999999"
		documentID        = "aaaaaaaa-aaaa-bbbb-cccc-aaaaaaaaaaaa"
	)

	seedScheduledDocument(t, ctx, db, scheduledDocumentSeed{
		tenantID:           tenantID,
		ownerUserID:        ownerUserID,
		templateVersionID:  templateVersionID,
		controlledDocID:    controlledDocID,
		documentID:         documentID,
		effectiveFrom:      effectiveAt,
		revisionVersion:    7,
		scheduleGeneration: 3,
	})

	emitter := &approvalapp.MemoryEmitter{}
	services := approvalapp.NewServices(approvalrepo.NewPostgresApprovalRepository(db), emitter, approvalapp.RealClock{})
	worker := &ScheduledPublishWorker{Service: services.Scheduler, DB: db}

	if err := worker.Work(ctx, &river.Job[ScheduledPublishArgs]{
		Args: ScheduledPublishArgs{
			TenantID:                tenantID,
			DocumentID:              documentID,
			ExpectedRevisionVersion: 7,
			ScheduledEffectiveAt:    effectiveAt,
			ScheduleGeneration:      2,
		},
	}); err != nil {
		t.Fatalf("Work: %v", err)
	}

	var status string
	var scheduleGeneration int64
	err := db.QueryRowContext(ctx, `
		SELECT status, schedule_generation
		  FROM public.documents
		 WHERE tenant_id = $1
		   AND id = $2`,
		tenantID, documentID,
	).Scan(&status, &scheduleGeneration)
	if err != nil {
		t.Fatalf("load document: %v", err)
	}
	if status != "scheduled" {
		t.Fatalf("status = %q, want scheduled", status)
	}
	if scheduleGeneration != 3 {
		t.Fatalf("schedule_generation = %d, want 3", scheduleGeneration)
	}
	if len(emitter.Events) != 0 {
		t.Fatalf("expected no events for stale job, got %+v", emitter.Events)
	}
}

func TestScheduledPublishWorker_NoOpWhenDeliveredBeforeEffectiveTime(t *testing.T) {
	db := pgtest.OpenAndMigrate(t)
	ctx := context.Background()
	effectiveAt := time.Date(2026, 5, 21, 14, 0, 0, 0, time.UTC)

	const (
		tenantID          = "12121212-aaaa-bbbb-cccc-121212121212"
		ownerUserID       = "34343434-aaaa-bbbb-cccc-343434343434"
		templateVersionID = "56565656-aaaa-bbbb-cccc-565656565656"
		controlledDocID   = "78787878-aaaa-bbbb-cccc-787878787878"
		documentID        = "90909090-aaaa-bbbb-cccc-909090909090"
	)

	seedScheduledDocument(t, ctx, db, scheduledDocumentSeed{
		tenantID:           tenantID,
		ownerUserID:        ownerUserID,
		templateVersionID:  templateVersionID,
		controlledDocID:    controlledDocID,
		documentID:         documentID,
		effectiveFrom:      effectiveAt,
		revisionVersion:    4,
		scheduleGeneration: 2,
	})

	emitter := &approvalapp.MemoryEmitter{}
	services := approvalapp.NewServices(approvalrepo.NewPostgresApprovalRepository(db), emitter, fixedClock{t: effectiveAt.Add(-1 * time.Minute)})
	worker := &ScheduledPublishWorker{Service: services.Scheduler, DB: db}

	if err := worker.Work(ctx, &river.Job[ScheduledPublishArgs]{
		Args: ScheduledPublishArgs{
			TenantID:                tenantID,
			DocumentID:              documentID,
			ExpectedRevisionVersion: 4,
			ScheduledEffectiveAt:    effectiveAt,
			ScheduleGeneration:      2,
		},
	}); err != nil {
		t.Fatalf("Work: %v", err)
	}

	var status string
	var revisionVersion int
	err := db.QueryRowContext(ctx, `
		SELECT status, revision_version
		  FROM public.documents
		 WHERE tenant_id = $1
		   AND id = $2`,
		tenantID, documentID,
	).Scan(&status, &revisionVersion)
	if err != nil {
		t.Fatalf("load document: %v", err)
	}
	if status != "scheduled" {
		t.Fatalf("status = %q, want scheduled", status)
	}
	if revisionVersion != 4 {
		t.Fatalf("revision_version = %d, want 4", revisionVersion)
	}
	if len(emitter.Events) != 0 {
		t.Fatalf("expected no events for early delivery, got %+v", emitter.Events)
	}
}

type scheduledDocumentSeed struct {
	tenantID           string
	ownerUserID        string
	templateVersionID  string
	controlledDocID    string
	documentID         string
	effectiveFrom      time.Time
	revisionVersion    int
	scheduleGeneration int64
}

func seedScheduledDocument(t *testing.T, ctx context.Context, db *sql.DB, seed scheduledDocumentSeed) {
	t.Helper()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	for _, stmt := range []struct {
		query string
		args  []any
	}{
		{
			query: `
				INSERT INTO public.controlled_documents
					(id, tenant_id, profile_code, process_area_code, code, title, owner_user_id, status)
				VALUES
					($1::uuid, $2::uuid, 'po', 'quality', 'PO-JOB-001', 'Scheduled Publish Worker', $3, 'active')`,
			args: []any{seed.controlledDocID, seed.tenantID, seed.ownerUserID},
		},
		{
			query: `
				INSERT INTO public.documents
					(id, tenant_id, template_version_id, name, status, form_data_json, created_by, controlled_document_id, revision_number, revision_version, effective_from, schedule_generation)
				VALUES
					($1::uuid, $2::uuid, $3::uuid, 'Scheduled Revision', 'scheduled', '{}'::jsonb, $4, $5::uuid, 1, $6, $7, $8)`,
			args: []any{
				seed.documentID,
				seed.tenantID,
				seed.templateVersionID,
				seed.ownerUserID,
				seed.controlledDocID,
				seed.revisionVersion,
				seed.effectiveFrom,
				seed.scheduleGeneration,
			},
		},
	} {
		if _, err := tx.ExecContext(ctx, stmt.query, stmt.args...); err != nil {
			t.Fatalf("seed rows: %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("commit seed: %v", err)
	}
}
