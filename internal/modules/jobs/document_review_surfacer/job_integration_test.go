//go:build integration
// +build integration

package document_review_surfacer

// M6 F6.2 T4 — integration proof for the River periodic review-due surfacer's
// core run() body against real Postgres via the canonical testdb factory
// (ADR 0034). Per validation-contract.md §4.3, proves:
//
//  1. cross-tenant coverage in one sweep: a due+published+effective document
//     in tenant A AND one in tenant B both get review_surfaced_at set by a
//     single run() call (the surfacer runs cross-tenant under the scheduler
//     bypass with no tenant GUC seeded — mirrors stuck_instance_watchdog);
//     a not-yet-due document in tenant A is left untouched;
//  2. idempotency: running run() twice surfaces a due document ONCE — the
//     second run flips 0 rows and review_surfaced_at is unchanged;
//  3. re-surfacing: after review_due_at is advanced past the stored
//     review_surfaced_at (simulating mark-reviewed), a subsequent run()
//     re-surfaces the document (new review_surfaced_at > the first).

import (
	"context"
	"database/sql"
	"testing"
	"time"

	documentsrepo "metaldocs/internal/modules/documents/repository"
	"metaldocs/internal/modules/iam/authz"
	"metaldocs/tests/integration/testdb"
)

// seedSurfacerDoc creates a base document row and drives a status-preserving
// UPDATE to set review_due_at, mirroring
// document_review_columns_integration_test.go's updateReviewColumns helper
// (the fixture factory's Opt surface has no WithReviewDueAt).
func seedSurfacerDoc(t *testing.T, db *sql.DB, tenantID string, status string, effectiveFrom time.Time, reviewDueAt *time.Time) (docID string) {
	t.Helper()

	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenantID))
	actor := testdb.NewUser(t, db, testdb.WithTenant(tenantID))
	cd := testdb.NewControlledDoc(t, db,
		testdb.WithTenant(tenantID),
		testdb.WithTaxonomy(tax),
		testdb.WithOwner(actor.ID),
	)
	doc := testdb.NewDocument(t, db,
		testdb.WithTenant(tenantID),
		testdb.WithControlledDoc(cd),
		testdb.WithOwner(actor.ID),
		testdb.WithStatus(status),
		testdb.WithEffectiveFrom(effectiveFrom),
	)

	if reviewDueAt != nil {
		setReviewDueAt(t, db, doc.ID, *reviewDueAt)
	}

	return doc.ID
}

// setReviewDueAt drives a status-preserving UPDATE to (re)set review_due_at,
// asserting document.edit to satisfy the documents/UPDATE tripwire — the same
// pattern seedReviewDueDoc uses in review_due_reader_integration_test.go.
func setReviewDueAt(t *testing.T, db *sql.DB, docID string, reviewDueAt time.Time) {
	t.Helper()
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("setReviewDueAt: begin tx: %v", err)
	}
	defer tx.Rollback()

	testdb.SetCapsOnTx(t, tx, `[{"cap":"document.edit"}]`)
	if _, err := tx.ExecContext(ctx,
		`UPDATE public.documents SET review_due_at = $1 WHERE id = $2::uuid`,
		reviewDueAt, docID,
	); err != nil {
		t.Fatalf("setReviewDueAt: set review_due_at: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("setReviewDueAt: commit: %v", err)
	}
}

func readReviewSurfacedAt(t *testing.T, db *sql.DB, docID string) *time.Time {
	t.Helper()
	var ts sql.NullTime
	if err := db.QueryRowContext(context.Background(),
		`SELECT review_surfaced_at FROM public.documents WHERE id = $1::uuid`, docID,
	).Scan(&ts); err != nil {
		t.Fatalf("readReviewSurfacedAt: %v", err)
	}
	if !ts.Valid {
		return nil
	}
	return &ts.Time
}

// TestIntegration_Surfacer_CrossTenant_MarksDueUntouchedNotDue proves one
// run() sweeps due+published+effective documents across BOTH tenant A and
// tenant B in a single tick (cross-tenant scheduler-bypass sweep, no
// per-tenant iteration), while a not-yet-due document in tenant A is left
// untouched.
func TestIntegration_Surfacer_CrossTenant_MarksDueUntouchedNotDue(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tntA := testdb.NewTenant(t, db)
	tntB := testdb.NewTenant(t, db)
	now := time.Now().UTC().Truncate(time.Second)
	effectiveFrom := now.Add(-30 * 24 * time.Hour)
	due := now.Add(-1 * time.Hour)
	future := now.Add(30 * 24 * time.Hour)

	dueDocA := seedSurfacerDoc(t, db, tntA.ID, "published", effectiveFrom, &due)
	notDueDocA := seedSurfacerDoc(t, db, tntA.ID, "published", effectiveFrom, &future)
	dueDocB := seedSurfacerDoc(t, db, tntB.ID, "published", effectiveFrom, &due)

	reader := documentsrepo.NewReviewDueReaderPG(db)
	writer := documentsrepo.NewReviewSurfaceWriterPG(db)
	ctx := authz.WithBackgroundBypass(context.Background())

	if err := run(ctx, db, reader, writer, now); err != nil {
		t.Fatalf("run: %v", err)
	}

	if got := readReviewSurfacedAt(t, db, dueDocA); got == nil || !got.Equal(now) {
		t.Fatalf("tenant A due doc review_surfaced_at = %v, want %v", got, now)
	}
	if got := readReviewSurfacedAt(t, db, dueDocB); got == nil || !got.Equal(now) {
		t.Fatalf("tenant B due doc review_surfaced_at = %v, want %v (cross-tenant sweep must cover both tenants)", got, now)
	}
	if got := readReviewSurfacedAt(t, db, notDueDocA); got != nil {
		t.Fatalf("tenant A not-yet-due doc review_surfaced_at = %v, want nil (untouched)", got)
	}
}

// TestIntegration_Surfacer_Idempotent_SecondRunNoOp proves running run() twice
// surfaces a due document ONCE: the second run flips 0 rows for that document
// and review_surfaced_at is unchanged from the first run.
func TestIntegration_Surfacer_Idempotent_SecondRunNoOp(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tnt := testdb.NewTenant(t, db)
	now := time.Now().UTC().Truncate(time.Second)
	effectiveFrom := now.Add(-30 * 24 * time.Hour)
	due := now.Add(-1 * time.Hour)

	docID := seedSurfacerDoc(t, db, tnt.ID, "published", effectiveFrom, &due)

	reader := documentsrepo.NewReviewDueReaderPG(db)
	writer := documentsrepo.NewReviewSurfaceWriterPG(db)
	ctx := authz.WithBackgroundBypass(context.Background())

	if err := run(ctx, db, reader, writer, now); err != nil {
		t.Fatalf("first run: %v", err)
	}
	firstSurfacedAt := readReviewSurfacedAt(t, db, docID)
	if firstSurfacedAt == nil || !firstSurfacedAt.Equal(now) {
		t.Fatalf("after first run review_surfaced_at = %v, want %v", firstSurfacedAt, now)
	}

	// Second run at a LATER wall-clock instant: if MarkSurfaced were not
	// idempotent it would overwrite review_surfaced_at with the new `now`,
	// which this assertion would catch.
	secondNow := now.Add(10 * time.Minute)
	if err := run(ctx, db, reader, writer, secondNow); err != nil {
		t.Fatalf("second run: %v", err)
	}
	secondSurfacedAt := readReviewSurfacedAt(t, db, docID)
	if secondSurfacedAt == nil || !secondSurfacedAt.Equal(*firstSurfacedAt) {
		t.Fatalf("after second run review_surfaced_at = %v, want unchanged %v (idempotency guard must flip 0 rows)", secondSurfacedAt, firstSurfacedAt)
	}
}

// TestIntegration_Surfacer_ReSurfacesAfterReviewDueAdvances proves that after
// review_due_at is advanced past the stored review_surfaced_at (simulating a
// mark-reviewed workflow setting a new future cycle, then that cycle itself
// becoming due), a subsequent run() re-surfaces the document with a fresh,
// later review_surfaced_at.
func TestIntegration_Surfacer_ReSurfacesAfterReviewDueAdvances(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tnt := testdb.NewTenant(t, db)
	now := time.Now().UTC().Truncate(time.Second)
	effectiveFrom := now.Add(-90 * 24 * time.Hour)
	due := now.Add(-1 * time.Hour)

	docID := seedSurfacerDoc(t, db, tnt.ID, "published", effectiveFrom, &due)

	reader := documentsrepo.NewReviewDueReaderPG(db)
	writer := documentsrepo.NewReviewSurfaceWriterPG(db)
	ctx := authz.WithBackgroundBypass(context.Background())

	if err := run(ctx, db, reader, writer, now); err != nil {
		t.Fatalf("first run: %v", err)
	}
	firstSurfacedAt := readReviewSurfacedAt(t, db, docID)
	if firstSurfacedAt == nil {
		t.Fatalf("after first run review_surfaced_at is nil, want set")
	}

	// Simulate mark-reviewed: advance review_due_at to a new cycle that is
	// itself now due (<= a later `now`), while review_surfaced_at still holds
	// the OLD cycle's timestamp — review_surfaced_at < review_due_at once
	// again, which is the re-eligibility condition.
	newDue := now.Add(5 * time.Minute)
	setReviewDueAt(t, db, docID, newDue)

	secondNow := now.Add(10 * time.Minute)
	if err := run(ctx, db, reader, writer, secondNow); err != nil {
		t.Fatalf("second run: %v", err)
	}
	secondSurfacedAt := readReviewSurfacedAt(t, db, docID)
	if secondSurfacedAt == nil || !secondSurfacedAt.Equal(secondNow) {
		t.Fatalf("after review_due_at advance + second run, review_surfaced_at = %v, want %v (re-surfaced)", secondSurfacedAt, secondNow)
	}
	if !secondSurfacedAt.After(*firstSurfacedAt) {
		t.Fatalf("second review_surfaced_at %v is not after first %v, want re-surfacing to advance the marker", secondSurfacedAt, firstSurfacedAt)
	}
}
