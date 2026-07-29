//go:build integration
// +build integration

package outbox_retention

// Real-DB acceptance proofs for the metaldocs.outbox_events terminal-row
// retention purge. Mirrors the testdb-factory convention used by
// internal/modules/render/fanout/retention/retention_integration_test.go: a
// template-cloned database per test, the real production seams for the system
// under test (outboxpg.Publisher to seed, outboxpg.Consumer to drive rows to
// their terminal state, outboxpg.Retention + this package's Worker to purge),
// and direct SQL only for test-only scaffolding a production seam cannot do
// (backdating a timestamp; MarkPublished always stamps now()).
//
// Coverage:
//  1. TestOutboxRetention_Integration_PurgesOnlyAgedTerminalRows — the
//     fail-closed core. Seeds five rows across every state the predicate must
//     discriminate and asserts exactly the two eligible ones are gone.
//  2. TestOutboxRetention_Integration_NeverPurgesUnpublishedHoweverOld — the
//     fail-closed invariant on its own, at an absurd age, so a future
//     predicate edit that drops the published_at/dead_lettered_at guard fails
//     here loudly rather than eating the live queue.
//  3. TestOutboxRetention_Integration_FreesIdempotencyKeyForRepublish — the
//     motivating class. Publisher.Publish dedupes on UNIQUE (idempotency_key)
//     with ON CONFLICT DO NOTHING returning nil, so a retained terminal row
//     silently swallows any later re-publish on that key. Proves the purge
//     releases the key.
//  4. TestOutboxRetention_Integration_BatchBounded_CapsAtBatchSizeTimesMaxIterations
//     — the delete loop is bounded by the batch/iteration product, not a
//     single unbounded statement.
//  5. TestOutboxRetention_Integration_RejectsNonPositiveBounds — non-positive
//     bounds error out instead of degrading to a silent no-op or an unbounded
//     delete.

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/riverqueue/river"

	"metaldocs/internal/platform/messaging"
	outboxpg "metaldocs/internal/platform/messaging/outbox/postgres"
	"metaldocs/tests/integration/testdb"
)

// testProducer tags every row these tests seed, so aggregate assertions can be
// scoped to this suite's own rows rather than to the whole table.
const testProducer = "outbox-retention-integration-test"

// openOutboxDB checks out a leased database from the testdb pool — the default
// factory path, not the OpenFreshDatabase clone: nothing here mutates DDL.
// metaldocs.outbox_events carries no tenant_id and no RLS policy, and every
// statement under test is schema-qualified, so unlike the staging-outbox suite
// this needs neither a pinned single connection nor a search_path.
func openOutboxDB(t *testing.T) *sql.DB {
	t.Helper()
	db, _ := testdb.Open(t)
	return db
}

// publishEvent inserts one unpublished outbox row through the real Publisher
// and returns its event id. The idempotency key is the caller's, so tests can
// re-publish the same key deliberately.
func publishEvent(t *testing.T, ctx context.Context, pub *outboxpg.Publisher, idempotencyKey string) messaging.EventID {
	t.Helper()
	eventID := messaging.EventID(uuid.NewString())
	revisionID := uuid.NewString()
	if err := pub.Publish(ctx, messaging.Event{
		EventID:        eventID,
		EventType:      messaging.EventTypePDFConvert,
		AggregateType:  messaging.AggregateType("document_revision"),
		AggregateID:    messaging.AggregateID(revisionID),
		Version:        1,
		IdempotencyKey: messaging.IdempotencyKey(idempotencyKey),
		Producer:       testProducer,
		TraceID:        messaging.TraceID(uuid.NewString()),
		Payload: messaging.PDFConvertPayload{
			RevisionID:     revisionID,
			FinalDocxS3Key: "tenants/test/" + revisionID + "/frozen.docx",
		},
	}); err != nil {
		t.Fatalf("publish outbox event %q: %v", idempotencyKey, err)
	}
	return eventID
}

// seedPublished publishes a row, marks it published through the real Consumer,
// then backdates published_at by age (test-only scaffolding: MarkPublished
// always stamps now(), and there is no production seam to set an arbitrary
// published_at).
func seedPublished(t *testing.T, ctx context.Context, db *sql.DB, pub *outboxpg.Publisher, con *outboxpg.Consumer, idempotencyKey string, age time.Duration) messaging.EventID {
	t.Helper()
	eventID := publishEvent(t, ctx, pub, idempotencyKey)
	if err := con.MarkPublished(ctx, []messaging.EventID{eventID}); err != nil {
		t.Fatalf("mark published %q: %v", idempotencyKey, err)
	}
	if age != 0 {
		backdate(t, ctx, db, "published_at", eventID, age)
	}
	return eventID
}

// seedDeadLettered publishes a row, dead-letters it through the real Consumer,
// then backdates dead_lettered_at by age.
func seedDeadLettered(t *testing.T, ctx context.Context, db *sql.DB, pub *outboxpg.Publisher, con *outboxpg.Consumer, idempotencyKey string, age time.Duration) messaging.EventID {
	t.Helper()
	eventID := publishEvent(t, ctx, pub, idempotencyKey)
	deadAt := time.Now().UTC()
	if err := con.MarkFailed(ctx, messaging.FailedEvent{
		EventID:        eventID,
		LastError:      "boom",
		DeadLetteredAt: &deadAt,
	}); err != nil {
		t.Fatalf("dead-letter %q: %v", idempotencyKey, err)
	}
	if age != 0 {
		backdate(t, ctx, db, "dead_lettered_at", eventID, age)
	}
	return eventID
}

func backdate(t *testing.T, ctx context.Context, db *sql.DB, column string, eventID messaging.EventID, age time.Duration) {
	t.Helper()
	// column is a test-local literal from the two call sites above, never
	// caller-supplied input.
	if _, err := db.ExecContext(ctx,
		`UPDATE metaldocs.outbox_events SET `+column+` = now() - $1::interval WHERE event_id = $2`,
		age.String(), string(eventID),
	); err != nil {
		t.Fatalf("backdate %s for %s: %v", column, eventID, err)
	}
}

func eventExists(t *testing.T, ctx context.Context, db *sql.DB, eventID messaging.EventID) bool {
	t.Helper()
	var n int
	if err := db.QueryRowContext(ctx,
		`SELECT count(*) FROM metaldocs.outbox_events WHERE event_id = $1`, string(eventID),
	).Scan(&n); err != nil {
		t.Fatalf("count outbox row %s: %v", eventID, err)
	}
	return n > 0
}

func keyExists(t *testing.T, ctx context.Context, db *sql.DB, idempotencyKey string) bool {
	t.Helper()
	var n int
	if err := db.QueryRowContext(ctx,
		`SELECT count(*) FROM metaldocs.outbox_events WHERE idempotency_key = $1`, idempotencyKey,
	).Scan(&n); err != nil {
		t.Fatalf("count outbox rows for key %q: %v", idempotencyKey, err)
	}
	return n > 0
}

// TestOutboxRetention_Integration_PurgesOnlyAgedTerminalRows drives the real
// Worker over five rows covering every state the predicate discriminates:
//
//	(a) published 8 days ago            -> purged (past PublishedRetention)
//	(b) published just now              -> survives
//	(c) never published, never dead-    -> survives, whatever its age
//	    lettered (still claimable)
//	(d) dead-lettered 8 days ago        -> survives (inside DeadLetterRetention,
//	                                       proving the two windows are separate
//	                                       and the published window does not
//	                                       reach the DLQ)
//	(e) dead-lettered 100 days ago      -> purged (past DeadLetterRetention)
func TestOutboxRetention_Integration_PurgesOnlyAgedTerminalRows(t *testing.T) {
	ctx := context.Background()
	db := openOutboxDB(t)

	pub := outboxpg.NewPublisher(db)
	con := outboxpg.NewConsumer(db, 30*time.Second)

	oldPublished := seedPublished(t, ctx, db, pub, con, "retention-old-published", 8*24*time.Hour)
	recentPublished := seedPublished(t, ctx, db, pub, con, "retention-recent-published", 0)
	unpublished := publishEvent(t, ctx, pub, "retention-unpublished")
	recentDeadLettered := seedDeadLettered(t, ctx, db, pub, con, "retention-recent-dead-lettered", 8*24*time.Hour)
	ancientDeadLettered := seedDeadLettered(t, ctx, db, pub, con, "retention-ancient-dead-lettered", 100*24*time.Hour)

	worker := NewWorker(outboxpg.NewRetention(db))
	if err := worker.Work(ctx, &river.Job[Args]{}); err != nil {
		t.Fatalf("Worker.Work: %v", err)
	}

	if eventExists(t, ctx, db, oldPublished) {
		t.Error("published-8d row still exists after purge, want gone")
	}
	if !eventExists(t, ctx, db, recentPublished) {
		t.Error("published-now row was purged, want survive")
	}
	if !eventExists(t, ctx, db, unpublished) {
		t.Error("unpublished row was purged, want survive (fail-closed: it is still claimable work)")
	}
	if !eventExists(t, ctx, db, recentDeadLettered) {
		t.Error("dead-lettered-8d row was purged, want survive (the 7d published window must not reach the DLQ)")
	}
	if eventExists(t, ctx, db, ancientDeadLettered) {
		t.Error("dead-lettered-100d row still exists after purge, want gone")
	}

	// The surviving unpublished row must still be claimable — the purge must
	// not have touched its delivery state on the way past.
	claimed, err := con.ClaimUnpublished(ctx, 10)
	if err != nil {
		t.Fatalf("ClaimUnpublished after purge: %v", err)
	}
	var found bool
	for _, event := range claimed {
		if event.EventID == unpublished {
			found = true
		}
	}
	if !found {
		t.Errorf("unpublished row %s is no longer claimable after the purge tick", unpublished)
	}
}

// TestOutboxRetention_Integration_NeverPurgesUnpublishedHoweverOld pins the
// fail-closed invariant on its own: a row that was never published and never
// dead-lettered is ineligible at ANY age, including one far past both windows.
// A predicate edit that drops the terminal-timestamp guard would eat the live
// queue in production; it fails here instead.
func TestOutboxRetention_Integration_NeverPurgesUnpublishedHoweverOld(t *testing.T) {
	ctx := context.Background()
	db := openOutboxDB(t)

	pub := outboxpg.NewPublisher(db)
	eventID := publishEvent(t, ctx, pub, "retention-ancient-unpublished")

	// Age everything that could plausibly be mistaken for a retention clock.
	if _, err := db.ExecContext(ctx, `
UPDATE metaldocs.outbox_events
   SET occurred_at      = now() - interval '1000 days',
       last_attempt_at  = now() - interval '1000 days',
       next_attempt_at  = now() - interval '1000 days'
 WHERE event_id = $1`, string(eventID)); err != nil {
		t.Fatalf("age unpublished row: %v", err)
	}

	worker := NewWorker(outboxpg.NewRetention(db))
	if err := worker.Work(ctx, &river.Job[Args]{}); err != nil {
		t.Fatalf("Worker.Work: %v", err)
	}

	if !eventExists(t, ctx, db, eventID) {
		t.Fatal("1000-day-old unpublished row was purged; retention must be gated on published_at/dead_lettered_at, never on age alone")
	}
}

// TestOutboxRetention_Integration_FreesIdempotencyKeyForRepublish is the
// motivating class. Publisher.Publish is
// `INSERT ... ON CONFLICT (idempotency_key) DO NOTHING` and returns nil either
// way, so while a terminal row is retained, any later legitimate re-publish on
// the same key is silently swallowed. This proves the purge releases the key —
// and, first, that the swallow is real before the purge runs.
func TestOutboxRetention_Integration_FreesIdempotencyKeyForRepublish(t *testing.T) {
	ctx := context.Background()
	db := openOutboxDB(t)

	pub := outboxpg.NewPublisher(db)
	con := outboxpg.NewConsumer(db, 30*time.Second)

	const key = "docgen_v2_pdf:tenant:revision"
	original := seedPublished(t, ctx, db, pub, con, key, 8*24*time.Hour)

	// Before the purge: re-publishing the key reports success but writes
	// nothing. This is the ba24c4f2 expurgo failure mode, pinned as a fact.
	swallowed := publishEvent(t, ctx, pub, key)
	if eventExists(t, ctx, db, swallowed) {
		t.Fatal("pre-purge re-publish inserted a row; the ON CONFLICT dedup no longer holds and this test no longer proves what it claims")
	}
	if !eventExists(t, ctx, db, original) {
		t.Fatal("original row vanished before the purge ran")
	}

	worker := NewWorker(outboxpg.NewRetention(db))
	if err := worker.Work(ctx, &river.Job[Args]{}); err != nil {
		t.Fatalf("Worker.Work: %v", err)
	}
	if keyExists(t, ctx, db, key) {
		t.Fatal("aged published row holding the idempotency key survived the purge")
	}

	// After the purge the key is free: the same key re-publishes into a real,
	// claimable row.
	republished := publishEvent(t, ctx, pub, key)
	if !eventExists(t, ctx, db, republished) {
		t.Fatal("re-publish after purge inserted nothing; the idempotency key was not released")
	}
	claimed, err := con.ClaimUnpublished(ctx, 10)
	if err != nil {
		t.Fatalf("ClaimUnpublished: %v", err)
	}
	if len(claimed) != 1 || claimed[0].EventID != republished {
		t.Fatalf("claimed = %+v, want exactly the re-published event %s", claimed, republished)
	}
}

// TestOutboxRetention_Integration_BatchBounded_CapsAtBatchSizeTimesMaxIterations
// seeds 10 eligible published rows and calls PurgePublished with batchSize=3,
// maxIterations=2 (a cap of 6, below the eligible set), proving the delete loop
// is bounded by the batch/iteration product rather than one unbounded
// statement. Production always passes the real BatchSize (5000) /
// MaxIterations (10); this exercises the same code path with smaller numbers,
// which is equivalent proof of the mechanism without a 5000-row fixture.
func TestOutboxRetention_Integration_BatchBounded_CapsAtBatchSizeTimesMaxIterations(t *testing.T) {
	ctx := context.Background()
	db := openOutboxDB(t)

	pub := outboxpg.NewPublisher(db)
	con := outboxpg.NewConsumer(db, 30*time.Second)

	const seededEligibleRows = 10
	for i := 0; i < seededEligibleRows; i++ {
		seedPublished(t, ctx, db, pub, con, "batch-bound-"+string(rune('a'+i)), 8*24*time.Hour)
	}

	const testBatchSize = 3
	const testMaxIterations = 2
	wantDeleted := testBatchSize * testMaxIterations // 6

	deleted, err := outboxpg.NewRetention(db).PurgePublished(
		ctx, time.Now().Add(-PublishedRetention), testBatchSize, testMaxIterations)
	if err != nil {
		t.Fatalf("PurgePublished: %v", err)
	}
	if deleted != wantDeleted {
		t.Fatalf("deleted = %d, want %d (batchSize=%d x maxIterations=%d, capped below the %d eligible rows)",
			deleted, wantDeleted, testBatchSize, testMaxIterations, seededEligibleRows)
	}

	var survivors int
	if err := db.QueryRowContext(ctx,
		`SELECT count(*) FROM metaldocs.outbox_events WHERE published_at IS NOT NULL AND producer = $1`,
		testProducer,
	).Scan(&survivors); err != nil {
		t.Fatalf("count survivors: %v", err)
	}
	if want := seededEligibleRows - wantDeleted; survivors != want {
		t.Fatalf("survivors = %d, want %d (proves the loop stopped at the batch/iteration cap, not an unbounded delete)", survivors, want)
	}
}

// TestOutboxRetention_Integration_RejectsNonPositiveBounds proves a caller bug
// surfaces as an error rather than degrading into a silent no-op (rows never
// purged) or an unbounded delete (whole table in one statement) — neither of
// which is safe to pick on the caller's behalf.
func TestOutboxRetention_Integration_RejectsNonPositiveBounds(t *testing.T) {
	ctx := context.Background()
	db := openOutboxDB(t)

	pub := outboxpg.NewPublisher(db)
	con := outboxpg.NewConsumer(db, 30*time.Second)
	eventID := seedPublished(t, ctx, db, pub, con, "bounds-guard", 8*24*time.Hour)

	store := outboxpg.NewRetention(db)
	cutoff := time.Now().Add(-PublishedRetention)

	for _, tc := range []struct {
		name                     string
		batchSize, maxIterations int
	}{
		{"zero batch size", 0, MaxIterations},
		{"negative batch size", -1, MaxIterations},
		{"zero max iterations", BatchSize, 0},
		{"negative max iterations", BatchSize, -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := store.PurgePublished(ctx, cutoff, tc.batchSize, tc.maxIterations); err == nil {
				t.Error("PurgePublished returned nil error, want a bounds error")
			}
			if _, err := store.PurgeDeadLettered(ctx, cutoff, tc.batchSize, tc.maxIterations); err == nil {
				t.Error("PurgeDeadLettered returned nil error, want a bounds error")
			}
		})
	}

	if !eventExists(t, ctx, db, eventID) {
		t.Error("eligible row was deleted by a rejected call; the bounds guard must reject before touching the table")
	}
}
