//go:build integration
// +build integration

package infrastructure_test

// M6 F6.2 T3 — integration proof for the ReviewDueReader.ListDueForReview
// published read-port (documents module), consumed by the jobs module's River
// periodic surfacer (T4). Per the M6 validation contract §4.1, this is the
// ONLY documents-owned surface the jobs module may call for review-due
// documents — never raw SQL against public.documents.
//
// Proves (contract §4.3, adapted to the read-port itself rather than the full
// surfacer job which is T4's scope):
//   1. only published + currently-effective docs with review_due_at <= now
//      are returned (seed one due, one future-due, one NULL review_due_at,
//      one non-published due doc -> assert exactly the one due+published row);
//   2. tenant isolation: a due doc seeded in tenant B is invisible to a call
//      scoped to tenant A via the explicit tenant_id predicate (primary),
//      independent of RLS (the tx's metaldocs.tenant_id GUC is also seeded,
//      backstopping via RLS FORCE policy on public.documents, migration
//      baseline :4840, but is inert under the superuser/BYPASSRLS role used
//      in dev/CI — the explicit predicate is what actually proves isolation
//      here);
//   3. limit is respected and rows are ordered by review_due_at ASC.

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"metaldocs/internal/modules/documents/infrastructure"
	"metaldocs/tests/integration/testdb"
)

// seedReviewDueDoc creates a base document row (status + effective_from set at
// create time via the fixture factory, non-draft snapshot columns stubbed by
// NewDocument itself), then drives a status-preserving UPDATE to set
// review_due_at — mirroring document_review_columns_integration_test.go's
// updateReviewColumns helper, since NewDocument's Opt surface has no
// WithReviewDueAt.
func seedReviewDueDoc(t *testing.T, db *sql.DB, tenantID string, status string, effectiveFrom time.Time, reviewDueAt *time.Time) (docID string) {
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
		ctx := context.Background()
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("seedReviewDueDoc: begin tx: %v", err)
		}
		defer tx.Rollback()

		testdb.SetCapsOnTx(t, tx, `[{"cap":"document.edit"}]`)
		if _, err := tx.ExecContext(ctx,
			`UPDATE public.documents SET review_due_at = $1 WHERE id = $2::uuid`,
			*reviewDueAt, doc.ID,
		); err != nil {
			t.Fatalf("seedReviewDueDoc: set review_due_at: %v", err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatalf("seedReviewDueDoc: commit: %v", err)
		}
	}

	return doc.ID
}

// openTenantScopedTx opens a fresh tx and sets the metaldocs.tenant_id GUC
// tx-locally, exactly as the production request lifecycle seeds tenant
// identity before a tenant-scoped read (M3 RLS backstop).
func openTenantScopedTx(t *testing.T, db *sql.DB, tenantID string) *sql.Tx {
	t.Helper()
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("openTenantScopedTx: begin tx: %v", err)
	}
	if _, err := tx.ExecContext(ctx,
		`SELECT set_config('metaldocs.tenant_id', $1, true)`, tenantID,
	); err != nil {
		t.Fatalf("openTenantScopedTx: set tenant_id GUC: %v", err)
	}
	return tx
}

func newReviewDueReader(db *sql.DB) *infrastructure.ReviewDueReaderPG {
	return infrastructure.NewReviewDueReaderPG(db)
}

// TestListDueForReview_FiltersPublishedAndDue proves only published,
// currently-due documents are returned: a due published doc; a not-yet-due
// published doc; a published doc with NULL review_due_at (no cycle set); and
// a due-but-draft doc (non-published) — only the first is returned.
func TestListDueForReview_FiltersPublishedAndDue(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tnt := testdb.NewTenant(t, db)
	now := time.Now().UTC().Truncate(time.Second)
	effectiveFrom := now.Add(-30 * 24 * time.Hour)
	due := now.Add(-1 * time.Hour)
	future := now.Add(30 * 24 * time.Hour)

	dueDocID := seedReviewDueDoc(t, db, tnt.ID, "published", effectiveFrom, &due)
	_ = seedReviewDueDoc(t, db, tnt.ID, "published", effectiveFrom, &future)
	_ = seedReviewDueDoc(t, db, tnt.ID, "published", effectiveFrom, nil)
	_ = seedReviewDueDoc(t, db, tnt.ID, "draft", effectiveFrom, &due)

	reader := newReviewDueReader(db)

	tx := openTenantScopedTx(t, db, tnt.ID)
	defer tx.Rollback()
	testdb.SetCapsOnTx(t, tx, `[{"cap":"document.view"}]`)

	rows, err := reader.ListDueForReview(context.Background(), tx, tnt.ID, now, 10)
	if err != nil {
		t.Fatalf("ListDueForReview: %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("got %d rows, want 1: %+v", len(rows), rows)
	}
	if rows[0].ID != dueDocID {
		t.Fatalf("got doc id %q, want the due published doc %q", rows[0].ID, dueDocID)
	}
	if rows[0].TenantID != tnt.ID {
		t.Fatalf("got tenant id %q, want %q", rows[0].TenantID, tnt.ID)
	}
	if rows[0].Status != "published" {
		t.Fatalf("got status %q, want published", rows[0].Status)
	}
	if rows[0].ReviewDueAt == nil || !rows[0].ReviewDueAt.Equal(due) {
		t.Fatalf("got review_due_at %v, want %v", rows[0].ReviewDueAt, due)
	}
}

// TestListDueForReview_TenantIsolation proves a due+published document seeded
// in tenant B is invisible when ListDueForReview is called with tenant A's
// id. This is now proven by the EXPLICIT tenant_id predicate ListDueForReview
// carries (`tenant_id = $2::uuid`, review_due_reader.go) — not by RLS. The
// tx's metaldocs.tenant_id GUC is still seeded (openTenantScopedTx) as the
// RLS backstop, but the dev/CI DB role (metaldocs_app) is superuser +
// BYPASSRLS, so RLS is inert here; without the explicit predicate this test
// would be RED (false-negative isolation) under that role.
func TestListDueForReview_TenantIsolation(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tntA := testdb.NewTenant(t, db)
	tntB := testdb.NewTenant(t, db)
	now := time.Now().UTC().Truncate(time.Second)
	effectiveFrom := now.Add(-30 * 24 * time.Hour)
	due := now.Add(-1 * time.Hour)

	_ = seedReviewDueDoc(t, db, tntB.ID, "published", effectiveFrom, &due)

	reader := newReviewDueReader(db)

	tx := openTenantScopedTx(t, db, tntA.ID)
	defer tx.Rollback()
	testdb.SetCapsOnTx(t, tx, `[{"cap":"document.view"}]`)

	rows, err := reader.ListDueForReview(context.Background(), tx, tntA.ID, now, 10)
	if err != nil {
		t.Fatalf("ListDueForReview: %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("got %d rows under tenant A identity, want 0 (tenant B doc must be invisible): %+v", len(rows), rows)
	}
}

// TestListDueForReview_LimitAndOrder proves limit is respected and rows are
// ordered by review_due_at ascending (earliest-due first).
func TestListDueForReview_LimitAndOrder(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tnt := testdb.NewTenant(t, db)
	now := time.Now().UTC().Truncate(time.Second)
	effectiveFrom := now.Add(-60 * 24 * time.Hour)

	// Seed 3 due docs at distinct review_due_at instants (all <= now), request
	// limit=2 -> expect exactly the two earliest, in ascending order.
	due1 := now.Add(-72 * time.Hour) // earliest
	due2 := now.Add(-48 * time.Hour)
	due3 := now.Add(-24 * time.Hour) // latest (still <= now)

	id1 := seedReviewDueDoc(t, db, tnt.ID, "published", effectiveFrom, &due1)
	id2 := seedReviewDueDoc(t, db, tnt.ID, "published", effectiveFrom, &due2)
	_ = seedReviewDueDoc(t, db, tnt.ID, "published", effectiveFrom, &due3)

	reader := newReviewDueReader(db)

	tx := openTenantScopedTx(t, db, tnt.ID)
	defer tx.Rollback()
	testdb.SetCapsOnTx(t, tx, `[{"cap":"document.view"}]`)

	rows, err := reader.ListDueForReview(context.Background(), tx, tnt.ID, now, 2)
	if err != nil {
		t.Fatalf("ListDueForReview: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("got %d rows, want limit=2: %+v", len(rows), rows)
	}
	if rows[0].ID != id1 || rows[1].ID != id2 {
		t.Fatalf("got order [%s, %s], want [%s, %s] (ascending review_due_at)", rows[0].ID, rows[1].ID, id1, id2)
	}
	if !rows[0].ReviewDueAt.Before(*rows[1].ReviewDueAt) {
		t.Fatalf("rows not ordered ascending by review_due_at: %v, %v", rows[0].ReviewDueAt, rows[1].ReviewDueAt)
	}
}
