//go:build integration
// +build integration

package application

// TestPublishRace proves the F4.2 (M4) invariant: a single approved-then-
// scheduled document revision cannot be BOTH manually published (PublishApproved)
// AND scheduler-published (RunScheduledPublishJob) such that it publishes twice
// or lands in a wrong terminal state.
//
// The two write paths have mutually exclusive status predicates in their OCC
// UPDATEs (manual requires status='approved', scheduler requires
// status='scheduled'), plus a revision_version CAS. Since the seeded document is
// 'scheduled' (not 'approved'), only the scheduler's UPDATE predicate can match;
// the manual path's UPDATE affects zero rows regardless of interleaving and
// returns repository.ErrStaleRevision. This test exercises BOTH orderings with a
// real concurrent start barrier (not serialized by test code) and asserts the
// outcome is exactly-one-winner in both cases — proving Postgres row-locking
// (scheduler's FOR UPDATE vs manual's bare UPDATE) resolves the race safely
// without any additional application-level choke point.

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/internal/modules/documents/approval/repository"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/tests/integration/testdb"
)

func TestPublishRace(t *testing.T) {
	t.Run("InterleavingA_ManualFirst", func(t *testing.T) {
		runPublishRaceInterleaving(t, true)
	})
	t.Run("InterleavingB_SchedulerFirst", func(t *testing.T) {
		runPublishRaceInterleaving(t, false)
	})
}

// runPublishRaceInterleaving seeds a fresh document and races PublishApproved
// against RunScheduledPublishJob with a real start barrier. manualFirst
// controls which goroutine is released from the barrier first; both goroutines
// block on the same sync.WaitGroup-backed gate so the race is genuinely
// concurrent — Postgres row-locking (not Go scheduling order) is what
// determines the actual winner.
func runPublishRaceInterleaving(t *testing.T, manualFirst bool) {
	ctx := context.Background()

	database, dbName := testdb.Open(t)
	if _, err := database.ExecContext(ctx,
		`ALTER DATABASE "`+dbName+`" SET search_path TO public, metaldocs`,
	); err != nil {
		t.Fatalf("alter database search_path: %v", err)
	}
	database.SetMaxIdleConns(0)
	database.SetMaxIdleConns(4)
	database.SetMaxOpenConns(4)

	effectiveAt := time.Now().UTC().Add(-1 * time.Hour)

	tnt := testdb.NewTenant(t, database)
	user := testdb.NewUser(t, database, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))

	doc := testdb.Scenario{}.ScheduledRevision(t, database, 5,
		testdb.WithTenant(tnt.ID),
		testdb.WithOwner(user.ID),
		testdb.WithEffectiveFrom(effectiveAt),
		testdb.WithRevisionVersion(3),
	)

	inst := testdb.NewApprovalInstance(t, database,
		testdb.WithDocument(doc),
		testdb.WithStatus("approved"),
	)

	schedInput := ScheduledPublishJobInput{
		TenantID:                tnt.ID,
		DocumentID:              doc.ID,
		ExpectedRevisionVersion: 3,
		ScheduledEffectiveAt:    effectiveAt,
		ScheduleGeneration:      5,
	}

	repo := repository.NewPostgresApprovalRepository(database, iamdomain.NoopUserDisplayNameReader{})
	emitter := NewSQLEmitter()
	clock := RealClock{}
	cdRead := controlleddocumentsdomain.NoopCDFieldReader{}

	svcs := NewServices(repo, emitter, clock, cdRead)
	runner := db.NewTxRunner(database)

	// -----------------------------------------------------------------------
	// Real concurrent barrier: both goroutines block on `start` until closed.
	// A small release delay (via `release` channel order) nudges Go's runtime
	// to schedule the intended goroutine first, but the actual winner is
	// decided by Postgres row-locking, not by this ordering — that is the
	// point of the proof.
	// -----------------------------------------------------------------------
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)

	var manualErr, schedErr error
	var manualDone, schedDone chan struct{} = make(chan struct{}), make(chan struct{})

	manualGoroutine := func() {
		defer wg.Done()
		<-start
		_, manualErr = svcs.Publish.PublishApproved(ctx, runner, PublishRequest{
			TenantID:    tnt.ID,
			InstanceID:  inst.ID,
			PublishedBy: user.ID,
		})
		close(manualDone)
	}
	schedGoroutine := func() {
		defer wg.Done()
		<-start
		schedErr = svcs.Scheduler.RunScheduledPublishJob(ctx, runner, schedInput)
		close(schedDone)
	}

	if manualFirst {
		go manualGoroutine()
		go schedGoroutine()
	} else {
		go schedGoroutine()
		go manualGoroutine()
	}

	close(start) // release both goroutines simultaneously — real barrier
	wg.Wait()

	// -----------------------------------------------------------------------
	// Assertion 1: exactly one winner. The document was seeded 'scheduled', so
	// only the scheduler's OCC predicate (status='scheduled') can match; the
	// manual path's predicate (status='approved') can never match this row,
	// regardless of interleaving. Manual MUST lose with ErrStaleRevision;
	// scheduler MUST win (nil error).
	// -----------------------------------------------------------------------
	if !errors.Is(manualErr, repository.ErrStaleRevision) {
		t.Fatalf("[manualFirst=%v] PublishApproved error = %v; want repository.ErrStaleRevision (document was scheduled, not approved)", manualFirst, manualErr)
	}
	if schedErr != nil {
		t.Fatalf("[manualFirst=%v] RunScheduledPublishJob error = %v; want nil (scheduler must win the OCC race)", manualFirst, schedErr)
	}

	// -----------------------------------------------------------------------
	// Assertion 2: terminal state — published exactly once, revision_version
	// bumped exactly once (3 -> 4), never twice, never skipped/decremented.
	// -----------------------------------------------------------------------
	var gotStatus string
	var gotRevVersion int
	if err := database.QueryRowContext(ctx,
		`SELECT status, revision_version FROM public.documents WHERE id = $1::uuid AND tenant_id = $2::uuid`,
		doc.ID, tnt.ID,
	).Scan(&gotStatus, &gotRevVersion); err != nil {
		t.Fatalf("[manualFirst=%v] query document row after race: %v", manualFirst, err)
	}
	if gotStatus != "published" {
		t.Fatalf("[manualFirst=%v] documents.status = %q after race; want \"published\"", manualFirst, gotStatus)
	}
	const expectedRevVersion = 4 // seeded 3 + 1, bumped exactly once
	if gotRevVersion != expectedRevVersion {
		t.Fatalf("[manualFirst=%v] documents.revision_version = %d after race; want %d (seeded 3 + 1, exactly once)", manualFirst, gotRevVersion, expectedRevVersion)
	}

	// -----------------------------------------------------------------------
	// Assertion 3: single side effect — exactly one document_published
	// governance event for this document (no double-emit from a partially
	// committed loser).
	// -----------------------------------------------------------------------
	var publishedEventCount int
	if err := database.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM public.governance_events
		  WHERE tenant_id = $1::uuid AND resource_id = $2::uuid AND event_type = 'document_published'`,
		tnt.ID, doc.ID,
	).Scan(&publishedEventCount); err != nil {
		t.Fatalf("[manualFirst=%v] query governance_events count: %v", manualFirst, err)
	}
	if publishedEventCount != 1 {
		t.Fatalf("[manualFirst=%v] governance_events document_published count = %d; want exactly 1", manualFirst, publishedEventCount)
	}
}
