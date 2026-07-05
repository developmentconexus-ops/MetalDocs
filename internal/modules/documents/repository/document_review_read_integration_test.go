//go:build integration
// +build integration

package repository_test

// M6 F6.2 T6 — read-side integration proof for the review/expiry fields on
// the document detail + list endpoints (contract-first: api/openapi/v1/
// openapi.yaml DocumentSummary/DocumentDetailResponse + the review_due list
// filter on GET /documents, both required+nullable per ADR 0035).
//
// Proves:
//   1. GetDocument (detail read) returns effective_from/effective_to/
//      review_due_at/last_reviewed_at when set on the row (nil-safe scan,
//      values flow through unchanged).
//   2. ListDocumentsPaginated with ListOptions.ReviewDue=true returns only
//      documents currently due for periodic review (published, effective,
//      review_due_at <= now) — seeding one due doc and one not-due doc,
//      asserting exactly the due one comes back.

import (
	"context"
	"database/sql"
	"testing"
	"time"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/documents/repository"
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
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("setReviewColumns: begin tx: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	testdb.SetCapsOnTx(t, tx, `[{"cap":"document.edit"}]`)
	if _, err := tx.ExecContext(ctx,
		`UPDATE public.documents
		    SET effective_from = $1, effective_to = $2,
		        review_due_at = $3, last_reviewed_at = $4
		  WHERE id = $5::uuid`,
		effectiveFrom, effectiveTo, reviewDueAt, lastReviewedAt, docID,
	); err != nil {
		t.Fatalf("setReviewColumns: update: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("setReviewColumns: commit: %v", err)
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
	repo := repository.New(db, iamdomain.NoopUserDisplayNameReader{}, controlleddocumentsdomain.NoopCDFieldReader{}, taxonomydomain.NoopAreaCatalogReader{})

	now := time.Now().UTC().Truncate(time.Second)
	effectiveFrom := now.Add(-60 * 24 * time.Hour)
	effectiveTo := now.Add(300 * 24 * time.Hour)
	reviewDueAt := now.Add(30 * 24 * time.Hour)
	lastReviewedAt := now.Add(-10 * 24 * time.Hour)

	docWithReview := testdb.NewDocument(t, db, testdb.WithTenant(tnt.ID), testdb.WithStatus("published"))
	setReviewColumns(t, db, docWithReview.ID, &effectiveFrom, &effectiveTo, &reviewDueAt, &lastReviewedAt)

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

	// Legacy doc: all four columns left NULL at create time -> nil-safe scan,
	// no panic, all four fields nil on the returned domain.Document.
	docNoReview := testdb.NewDocument(t, db, testdb.WithTenant(tnt.ID), testdb.WithStatus("draft"))
	gotLegacy, err := repo.GetDocument(ctx, tnt.ID, docNoReview.ID)
	if err != nil {
		t.Fatalf("GetDocument (legacy): %v", err)
	}
	if gotLegacy.EffectiveFrom != nil || gotLegacy.EffectiveTo != nil || gotLegacy.ReviewDueAt != nil || gotLegacy.LastReviewedAt != nil {
		t.Errorf("legacy doc review/expiry fields = (%v,%v,%v,%v), want all nil",
			gotLegacy.EffectiveFrom, gotLegacy.EffectiveTo, gotLegacy.ReviewDueAt, gotLegacy.LastReviewedAt)
	}
}

// TestListDocumentsPaginated_ReviewDueFilter proves ListOptions.ReviewDue=true
// (wired from the GET /documents review_due=true query filter) returns only
// documents currently due for periodic review: published, effective, and
// review_due_at <= now. Seeds one due doc and one not-due (future review_due_at)
// doc; asserts exactly the due one is returned.
func TestListDocumentsPaginated_ReviewDueFilter(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tnt := testdb.NewTenant(t, db)
	repo := repository.New(db, iamdomain.NoopUserDisplayNameReader{}, controlleddocumentsdomain.NoopCDFieldReader{}, taxonomydomain.NoopAreaCatalogReader{})

	now := time.Now().UTC().Truncate(time.Second)
	effectiveFrom := now.Add(-60 * 24 * time.Hour)
	due := now.Add(-1 * time.Hour)
	future := now.Add(30 * 24 * time.Hour)

	dueDoc := testdb.NewDocument(t, db, testdb.WithTenant(tnt.ID), testdb.WithStatus("published"))
	setReviewColumns(t, db, dueDoc.ID, &effectiveFrom, nil, &due, nil)

	notDueDoc := testdb.NewDocument(t, db, testdb.WithTenant(tnt.ID), testdb.WithStatus("published"))
	setReviewColumns(t, db, notDueDoc.ID, &effectiveFrom, nil, &future, nil)

	// Also seed a due-but-draft doc (non-published) to prove the filter excludes it.
	draftDueDoc := testdb.NewDocument(t, db, testdb.WithTenant(tnt.ID), testdb.WithStatus("draft"))
	setReviewColumns(t, db, draftDueDoc.ID, &effectiveFrom, nil, &due, nil)

	items, total, _, err := repo.ListDocumentsPaginated(ctx, tnt.ID, repository.ListOptions{
		PageSize:  20,
		ReviewDue: true,
	})
	if err != nil {
		t.Fatalf("ListDocumentsPaginated: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d items, want 1 (only the due+published doc): %+v", len(items), items)
	}
	if items[0].ID != dueDoc.ID {
		t.Fatalf("got doc id %q, want the due doc %q", items[0].ID, dueDoc.ID)
	}
	if total != 1 {
		t.Fatalf("got total %d, want 1", total)
	}
}
