//go:build integration
// +build integration

package application

// TestPublishRace proves the F4.2 (M4) invariant (validation-contract §2.2): a
// single approved document revision cannot be BOTH manually published
// (PublishService.PublishApproved) AND scheduler-published
// (SchedulerService.RunScheduledPublishJob) such that it publishes twice or
// lands in a wrong terminal state.
//
// The two write paths have MUTUALLY EXCLUSIVE status predicates in their OCC
// UPDATEs — manual requires status='approved', scheduler requires
// status='scheduled' — plus a revision_version CAS. On a single static row only
// ONE predicate can ever match, so at most one path publishes. This safety
// property (exactly-one-winner) is proven in two complementary halves, each run
// under a real concurrent start barrier and in both goroutine orderings:
//
//   - Scheduled-seed half (status='scheduled'): proves the PREDICATE-DIVERGENCE
//     guarantee. The manual UPDATE (WHERE status='approved') matches zero rows
//     regardless of interleaving and loses with repository.ErrStaleRevision; the
//     scheduler wins. There is no genuine lock contention here — manual never
//     touches the row because its predicate cannot match.
//
//   - Approved-seed half (status='approved'): the ONLY case with genuine
//     FOR-UPDATE-vs-UPDATE lock contention (scheduler takes FOR UPDATE, manual
//     issues a bare UPDATE against the same row). Manual wins and publishes; the
//     scheduler observes the divergence (status no longer 'scheduled', or
//     effective_from unset / revision_version bumped) and no-ops to a successful
//     nil (RunScheduledPublishJob maps its no-op sentinels to nil).
//
// Both halves assert exactly-one publish, exactly-one revision_version bump
// (N -> N+1), and exactly-one document_published governance event — the §2.2
// terminal-state and single-side-effect guarantees. No application-level choke
// point is required: the DB predicate + OCC CAS already serialize the outcome.

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

// seedState selects which of the two mutually-exclusive document statuses the
// race is seeded in, which in turn dictates which path is the expected winner.
type seedState int

const (
	// seedScheduled: document status='scheduled' — scheduler wins, manual loses
	// with repository.ErrStaleRevision (predicate-divergence half).
	seedScheduled seedState = iota
	// seedApproved: document status='approved' — manual wins, scheduler no-ops
	// to nil (lock-contention half).
	seedApproved
)

func TestPublishRace(t *testing.T) {
	t.Run("ScheduledSeed_ManualFirst", func(t *testing.T) {
		runPublishRaceInterleaving(t, seedScheduled, true)
	})
	t.Run("ScheduledSeed_SchedulerFirst", func(t *testing.T) {
		runPublishRaceInterleaving(t, seedScheduled, false)
	})
	t.Run("ApprovedSeed_ManualFirst", func(t *testing.T) {
		runPublishRaceInterleaving(t, seedApproved, true)
	})
	t.Run("ApprovedSeed_SchedulerFirst", func(t *testing.T) {
		runPublishRaceInterleaving(t, seedApproved, false)
	})
}

// runPublishRaceInterleaving seeds a fresh document in the given seedState and
// races PublishApproved against RunScheduledPublishJob with a real start
// barrier. manualFirst controls which goroutine is launched first; both block
// on the same `start` channel, closed once, so the race is genuinely concurrent
// — the actual winner is decided by the DB status predicate (and, in the
// approved-seed half, Postgres row-locking), not by Go scheduling order.
func runPublishRaceInterleaving(t *testing.T, seed seedState, manualFirst bool) {
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

	// effective_from is only meaningful for the scheduled seed; in the past so
	// the scheduler's pre-effective-date gate passes. For the approved seed the
	// document carries no effective_from (NULL), which is exactly what makes the
	// scheduler no-op there — but the job input still references this timestamp
	// so the scheduler genuinely attempts (loads state under FOR UPDATE) before
	// diverging.
	effectiveAt := time.Now().UTC().Add(-1 * time.Hour)

	tnt := testdb.NewTenant(t, database)
	user := testdb.NewUser(t, database, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))

	var doc testdb.Document
	switch seed {
	case seedScheduled:
		doc = testdb.Scenario{}.ScheduledRevision(t, database, 5,
			testdb.WithTenant(tnt.ID),
			testdb.WithOwner(user.ID),
			testdb.WithEffectiveFrom(effectiveAt),
			testdb.WithRevisionVersion(3),
		)
	case seedApproved:
		// status='approved', no effective_from (NULL), schedule_generation at
		// its default. revision_version=3 to mirror the scheduled seed's CAS base.
		doc = testdb.NewDocument(t, database,
			testdb.WithTenant(tnt.ID),
			testdb.WithOwner(user.ID),
			testdb.WithStatus("approved"),
			testdb.WithRevisionVersion(3),
		)
	}

	inst := testdb.NewApprovalInstance(t, database,
		testdb.WithDocument(doc),
		testdb.WithStatus("approved"),
	)

	// The scheduled-publish job input matches the scheduled seed exactly. For the
	// approved seed we still register a well-formed input (matching
	// ScheduledEffectiveAt, and ScheduleGeneration=5) so RunScheduledPublishJob
	// actually runs and then no-ops on the state divergence rather than being
	// trivially skipped for a malformed input.
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
	// Real concurrent barrier: both goroutines block on `start` until it is
	// closed once. Launch order (manualFirst) only nudges Go's runtime; the
	// actual winner is decided by the DB, which is the point of the proof.
	// -----------------------------------------------------------------------
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)

	var manualErr, schedErr error

	manualGoroutine := func() {
		defer wg.Done()
		<-start
		_, manualErr = svcs.Publish.PublishApproved(ctx, runner, PublishRequest{
			TenantID:    tnt.ID,
			InstanceID:  inst.ID,
			PublishedBy: user.ID,
		})
	}
	schedGoroutine := func() {
		defer wg.Done()
		<-start
		schedErr = svcs.Scheduler.RunScheduledPublishJob(ctx, runner, schedInput)
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

	label := seedLabel(seed, manualFirst)

	// -----------------------------------------------------------------------
	// Assertion 1: exactly one winner, with each loser reporting its correct
	// sentinel.
	// -----------------------------------------------------------------------
	switch seed {
	case seedScheduled:
		// Predicate-divergence half: only the scheduler's predicate
		// (status='scheduled') matches. Manual loses with ErrStaleRevision
		// (its UPDATE ... WHERE status='approved' affects 0 rows); scheduler wins.
		if !errors.Is(manualErr, repository.ErrStaleRevision) {
			t.Fatalf("%s: PublishApproved error = %v; want repository.ErrStaleRevision (doc was scheduled, not approved)", label, manualErr)
		}
		if schedErr != nil {
			t.Fatalf("%s: RunScheduledPublishJob error = %v; want nil (scheduler must win)", label, schedErr)
		}
	case seedApproved:
		// Lock-contention half: only the manual predicate (status='approved')
		// matches. Manual wins (nil). The scheduler observes the state
		// divergence — status is 'approved' (effective_from NULL) if it loads
		// first, or revision_version already bumped if it loads after manual —
		// and no-ops. RunScheduledPublishJob maps BOTH no-op paths
		// (scheduledJobMatchesState==false and the errScheduledPublishNoOp OCC
		// sentinel) to a successful nil, so nil is the correct assertion; the
		// "no double-bump / single event" checks below prove the scheduler did
		// not additionally mutate the row.
		if manualErr != nil {
			t.Fatalf("%s: PublishApproved error = %v; want nil (manual must win)", label, manualErr)
		}
		if schedErr != nil {
			t.Fatalf("%s: RunScheduledPublishJob error = %v; want nil (scheduler must no-op to success)", label, schedErr)
		}
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
		t.Fatalf("%s: query document row after race: %v", label, err)
	}
	if gotStatus != "published" {
		t.Fatalf("%s: documents.status = %q after race; want \"published\"", label, gotStatus)
	}
	const expectedRevVersion = 4 // seeded 3 + 1, bumped exactly once
	if gotRevVersion != expectedRevVersion {
		t.Fatalf("%s: documents.revision_version = %d after race; want %d (seeded 3 + 1, exactly once)", label, gotRevVersion, expectedRevVersion)
	}

	// -----------------------------------------------------------------------
	// Assertion 3: single side effect — exactly one document_published
	// governance event for this document (no double-emit from a loser that
	// partially committed).
	// -----------------------------------------------------------------------
	var publishedEventCount int
	if err := database.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM public.governance_events
		  WHERE tenant_id = $1::uuid AND resource_id = $2::uuid AND event_type = 'document_published'`,
		tnt.ID, doc.ID,
	).Scan(&publishedEventCount); err != nil {
		t.Fatalf("%s: query governance_events count: %v", label, err)
	}
	if publishedEventCount != 1 {
		t.Fatalf("%s: governance_events document_published count = %d; want exactly 1", label, publishedEventCount)
	}
}

func seedLabel(seed seedState, manualFirst bool) string {
	s := "scheduledSeed"
	if seed == seedApproved {
		s = "approvedSeed"
	}
	order := "schedulerFirst"
	if manualFirst {
		order = "manualFirst"
	}
	return "[" + s + "/" + order + "]"
}
