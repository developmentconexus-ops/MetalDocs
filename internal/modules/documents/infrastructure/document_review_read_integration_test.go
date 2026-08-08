//go:build integration
// +build integration

package infrastructure_test

// M6 F6.2 T6 + F6.4 D2 — read-side integration proof for the review/expiry
// fields on the document detail + list endpoints (contract-first: api/openapi/v1/
// openapi.yaml DocumentSummary/DocumentDetailResponse + the review_due list
// filter on GET /documents, both required+nullable per ADR 0035).
//
// Proves:
//   1. GetDocument (detail read) returns effective_from/effective_to/
//      review_due_at/last_reviewed_at when set on the row (nil-safe scan,
//      values flow through unchanged).
//   2. ListDocumentsPaginated with ListOptions.ReviewDue=true returns the
//      SURFACED worklist (F6.4 D2, superseding the F6.2 recompute predicate):
//      published, effective, and review_surfaced_at >= review_due_at — i.e.
//      the River surfacer has flagged the doc for its current cycle. A doc
//      that is merely "due" (review_due_at <= now) but not yet surfaced is
//      EXCLUDED; see TestIntegration_ReviewDueFilter_ReadsSurfaced for the
//      full surface -> include -> mark-reviewed -> auto-expel round trip.

import (
	"context"
	"database/sql"
	"testing"
	"time"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/internal/modules/documents/infrastructure"
	iamdomain "metaldocs/internal/modules/iam/domain"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/tests/integration/testdb"
)

// setReviewColumns drives a status-preserving UPDATE against public.documents
// to set the four review/expiry columns, mirroring
// document_review_columns_integration_test.go's updateReviewColumns /
// review_due_reader_integration_test.go's seedReviewDueDoc helpers (NewDocument's
// Opt surface has no direct setter for review_due_at/last_reviewed_at).
func setReviewColumns(t *testing.T, db *sql.DB, docID string, effectiveFrom, effectiveTo, reviewDueAt, lastReviewedAt *time.Time) {
	t.Helper()
	setReviewColumnsWithSurfaced(t, db, docID, effectiveFrom, effectiveTo, reviewDueAt, lastReviewedAt, nil)
}

// setReviewColumnsWithSurfaced is setReviewColumns plus review_surfaced_at
// (F6.4 D2 worklist marker — nil leaves the column NULL, i.e. "not yet
// surfaced").
func setReviewColumnsWithSurfaced(t *testing.T, db *sql.DB, docID string, effectiveFrom, effectiveTo, reviewDueAt, lastReviewedAt, reviewSurfacedAt *time.Time) {
	t.Helper()
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("setReviewColumnsWithSurfaced: begin tx: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	testdb.SetCapsOnTx(t, tx, `[{"cap":"document.edit"}]`)
	if _, err := tx.ExecContext(ctx,
		`UPDATE public.documents
		    SET effective_from = $1, effective_to = $2,
		        review_due_at = $3, last_reviewed_at = $4,
		        review_surfaced_at = $5
		  WHERE id = $6::uuid`,
		effectiveFrom, effectiveTo, reviewDueAt, lastReviewedAt, reviewSurfacedAt, docID,
	); err != nil {
		t.Fatalf("setReviewColumnsWithSurfaced: update: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("setReviewColumnsWithSurfaced: commit: %v", err)
	}
}

// TestGetDocument_ReturnsReviewAndExpiryFields proves the detail read
// (GetDocument, consumed by GET /documents/{id} -> toDocumentDetailResponse)
// surfaces the four review/expiry columns nil-safe, both when set and when
// left NULL (legacy row / no cycle configured).
func TestGetDocument_ReturnsReviewAndExpiryFields(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tnt := testdb.NewTenant(t, db)
	repo := infrastructure.New(db, iamdomain.NoopUserDisplayNameReader{}, controlleddocumentsdomain.NoopCDFieldReader{}, taxonomydomain.NoopAreaCatalogReader{})

	now := time.Now().UTC().Truncate(time.Second)
	effectiveFrom := now.Add(-60 * 24 * time.Hour)
	effectiveTo := now.Add(300 * 24 * time.Hour)
	reviewDueAt := now.Add(30 * 24 * time.Hour)
	lastReviewedAt := now.Add(-10 * 24 * time.Hour)
	reviewSurfacedAt := now.Add(-5 * 24 * time.Hour)

	docWithReview := testdb.NewDocument(t, db, testdb.WithTenant(tnt.ID), testdb.WithStatus("published"))
	setReviewColumnsWithSurfaced(t, db, docWithReview.ID, &effectiveFrom, &effectiveTo, &reviewDueAt, &lastReviewedAt, &reviewSurfacedAt)

	got, err := repo.GetDocument(ctx, tnt.ID, docWithReview.ID)
	if err != nil {
		t.Fatalf("GetDocument: %v", err)
	}
	if got.EffectiveFrom == nil || !got.EffectiveFrom.Equal(effectiveFrom) {
		t.Errorf("EffectiveFrom = %v, want %v", got.EffectiveFrom, effectiveFrom)
	}
	if got.EffectiveTo == nil || !got.EffectiveTo.Equal(effectiveTo) {
		t.Errorf("EffectiveTo = %v, want %v", got.EffectiveTo, effectiveTo)
	}
	if got.ReviewDueAt == nil || !got.ReviewDueAt.Equal(reviewDueAt) {
		t.Errorf("ReviewDueAt = %v, want %v", got.ReviewDueAt, reviewDueAt)
	}
	if got.LastReviewedAt == nil || !got.LastReviewedAt.Equal(lastReviewedAt) {
		t.Errorf("LastReviewedAt = %v, want %v", got.LastReviewedAt, lastReviewedAt)
	}
	if got.ReviewSurfacedAt == nil || !got.ReviewSurfacedAt.Equal(reviewSurfacedAt) {
		t.Errorf("ReviewSurfacedAt = %v, want %v", got.ReviewSurfacedAt, reviewSurfacedAt)
	}

	// Legacy doc: all five columns left NULL at create time -> nil-safe scan,
	// no panic, all five fields nil on the returned domain.Document.
	docNoReview := testdb.NewDocument(t, db, testdb.WithTenant(tnt.ID), testdb.WithStatus("draft"))
	gotLegacy, err := repo.GetDocument(ctx, tnt.ID, docNoReview.ID)
	if err != nil {
		t.Fatalf("GetDocument (legacy): %v", err)
	}
	if gotLegacy.EffectiveFrom != nil || gotLegacy.EffectiveTo != nil || gotLegacy.ReviewDueAt != nil || gotLegacy.LastReviewedAt != nil || gotLegacy.ReviewSurfacedAt != nil {
		t.Errorf("legacy doc review/expiry fields = (%v,%v,%v,%v,%v), want all nil",
			gotLegacy.EffectiveFrom, gotLegacy.EffectiveTo, gotLegacy.ReviewDueAt, gotLegacy.LastReviewedAt, gotLegacy.ReviewSurfacedAt)
	}
}

// TestListDocumentsPaginated_ReviewDueFilter proves ListOptions.ReviewDue=true
// (wired from the GET /documents review_due=true query filter) returns the
// SURFACED worklist (F6.4 D2): published, effective, and review_surfaced_at
// >= review_due_at — i.e. the River periodic surfacer has flagged the doc for
// its current cycle. Seeds one surfaced+due doc, one due-but-NOT-yet-surfaced
// doc (must be excluded — worklist is surfacer-driven, not a live recompute),
// one surfaced+not-due doc (future review_due_at; must be excluded), and one
// surfaced-but-draft doc (non-published; must be excluded).
func TestListDocumentsPaginated_ReviewDueFilter(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tnt := testdb.NewTenant(t, db)
	repo := infrastructure.New(db, iamdomain.NoopUserDisplayNameReader{}, controlleddocumentsdomain.NoopCDFieldReader{}, taxonomydomain.NoopAreaCatalogReader{})

	now := time.Now().UTC().Truncate(time.Second)
	effectiveFrom := now.Add(-60 * 24 * time.Hour)
	due := now.Add(-1 * time.Hour)
	future := now.Add(30 * 24 * time.Hour)
	surfacedAt := now.Add(-30 * time.Minute) // >= due, i.e. surfaced for this cycle

	surfacedDoc := testdb.NewDocument(t, db, testdb.WithTenant(tnt.ID), testdb.WithStatus("published"))
	setReviewColumnsWithSurfaced(t, db, surfacedDoc.ID, &effectiveFrom, nil, &due, nil, &surfacedAt)

	// Due but never surfaced: review_surfaced_at stays NULL -> excluded.
	dueNotSurfacedDoc := testdb.NewDocument(t, db, testdb.WithTenant(tnt.ID), testdb.WithStatus("published"))
	setReviewColumns(t, db, dueNotSurfacedDoc.ID, &effectiveFrom, nil, &due, nil)

	// Surfaced, but not due yet (future review_due_at) -> excluded.
	notDueDoc := testdb.NewDocument(t, db, testdb.WithTenant(tnt.ID), testdb.WithStatus("published"))
	setReviewColumnsWithSurfaced(t, db, notDueDoc.ID, &effectiveFrom, nil, &future, nil, &surfacedAt)

	// Surfaced + due, but draft (non-published) -> excluded.
	draftDueDoc := testdb.NewDocument(t, db, testdb.WithTenant(tnt.ID), testdb.WithStatus("draft"))
	setReviewColumnsWithSurfaced(t, db, draftDueDoc.ID, &effectiveFrom, nil, &due, nil, &surfacedAt)

	items, total, _, err := repo.ListDocumentsPaginated(ctx, tnt.ID, infrastructure.ListOptions{
		PageSize:  20,
		ReviewDue: true,
	})
	if err != nil {
		t.Fatalf("ListDocumentsPaginated: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d items, want 1 (only the surfaced+due+published doc): %+v", len(items), items)
	}
	if items[0].ID != surfacedDoc.ID {
		t.Fatalf("got doc id %q, want the surfaced doc %q", items[0].ID, surfacedDoc.ID)
	}
	if total != 1 {
		t.Fatalf("got total %d, want 1", total)
	}
}
