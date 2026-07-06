//go:build integration
// +build integration

package jobs

// Real-DB tests for the scheduled-publish worker, migrated off the drifted
// shared `pgtest` database onto the unified `testdb` factory (template-DB cloned
// per test, curated baseline). The scheduled document graph is seeded through
// the factory's Scenario.ScheduledRevision; the worker's own publish UPDATE
// satisfies the curated-baseline capability tripwire via authz.BypassSystem
// (metaldocs.bypass_authz = 'scheduler'), exactly as in production.

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/riverqueue/river"

	approvalapp "metaldocs/internal/modules/documents/approval/application"
	approvalrepo "metaldocs/internal/modules/documents/approval/infrastructure"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/tests/integration/testdb"
)

type fixedClock struct {
	t time.Time
}

func (c fixedClock) Now() time.Time {
	return c.t
}

// openSchedulerDB opens a template-cloned DB pinned to a single connection with
// the runtime search path, so the worker's unqualified `documents` reads resolve
// and the session setting survives into the TxRunner's transaction.
func openSchedulerDB(t *testing.T) *sql.DB {
	t.Helper()
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)
	if _, err := db.ExecContext(context.Background(), `SET search_path TO metaldocs, public`); err != nil {
		t.Fatalf("set search_path: %v", err)
	}
	return db
}

func TestScheduledPublishWorker_PublishesWhenTruthMatches(t *testing.T) {
	ctx := context.Background()
	db := openSchedulerDB(t)
	effectiveAt := time.Date(2026, 5, 21, 14, 0, 0, 0, time.UTC)

	doc := testdb.Scenario{}.ScheduledRevision(t, db, 2,
		testdb.WithEffectiveFrom(effectiveAt),
		testdb.WithRevisionVersion(4),
	)

	emitter := &approvalapp.MemoryEmitter{}
	services := approvalapp.NewServices(approvalrepo.NewPostgresApprovalRepository(db, iamdomain.NoopUserDisplayNameReader{}), emitter, approvalapp.RealClock{}, testdb.NewCDFieldReader(t))
	worker := NewScheduledPublishWorker(services.Scheduler, db)

	if err := worker.Work(ctx, &river.Job[ScheduledPublishArgs]{
		Args: ScheduledPublishArgs{
			TenantID:                doc.TenantID,
			DocumentID:              doc.ID,
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
		doc.TenantID, doc.ID,
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
	ctx := context.Background()
	db := openSchedulerDB(t)
	effectiveAt := time.Date(2026, 5, 21, 14, 0, 0, 0, time.UTC)

	doc := testdb.Scenario{}.ScheduledRevision(t, db, 3,
		testdb.WithEffectiveFrom(effectiveAt),
		testdb.WithRevisionVersion(7),
	)

	emitter := &approvalapp.MemoryEmitter{}
	services := approvalapp.NewServices(approvalrepo.NewPostgresApprovalRepository(db, iamdomain.NoopUserDisplayNameReader{}), emitter, approvalapp.RealClock{}, testdb.NewCDFieldReader(t))
	worker := NewScheduledPublishWorker(services.Scheduler, db)

	if err := worker.Work(ctx, &river.Job[ScheduledPublishArgs]{
		Args: ScheduledPublishArgs{
			TenantID:                doc.TenantID,
			DocumentID:              doc.ID,
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
		doc.TenantID, doc.ID,
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
	ctx := context.Background()
	db := openSchedulerDB(t)
	effectiveAt := time.Date(2026, 5, 21, 14, 0, 0, 0, time.UTC)

	doc := testdb.Scenario{}.ScheduledRevision(t, db, 2,
		testdb.WithEffectiveFrom(effectiveAt),
		testdb.WithRevisionVersion(4),
	)

	emitter := &approvalapp.MemoryEmitter{}
	services := approvalapp.NewServices(approvalrepo.NewPostgresApprovalRepository(db, iamdomain.NoopUserDisplayNameReader{}), emitter, fixedClock{t: effectiveAt.Add(-1 * time.Minute)}, testdb.NewCDFieldReader(t))
	worker := NewScheduledPublishWorker(services.Scheduler, db)

	if err := worker.Work(ctx, &river.Job[ScheduledPublishArgs]{
		Args: ScheduledPublishArgs{
			TenantID:                doc.TenantID,
			DocumentID:              doc.ID,
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
		doc.TenantID, doc.ID,
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
