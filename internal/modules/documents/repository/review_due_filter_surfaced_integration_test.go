//go:build integration
// +build integration

package repository_test

// F6.4 D2 — integration proof that GET /documents?review_due=true (via
// ListOptions.ReviewDue, buildDocumentFilter in repository.go) reads the
// review_surfaced_at WORKLIST marker rather than recomputing "due right now".
// Contract: docs/superpowers/milestones/global-maximum-remediation/
// milestone-6-eqms-review-reason/f6.4-surfacer-contract-and-consumer/spec.md
// §D2, validation gate #4.
//
// Flow proved (single test, three phases against one seeded document):
//  1. Published + effective + review_due_at <= now, but review_surfaced_at
//     IS NULL (the River surfacer has not run yet) -> the filter EXCLUDES it.
//     The worklist is surfacer-driven, not a live recompute of "due".
//  2. review_surfaced_at is set to now (>= review_due_at, i.e. the surfacer
//     just flagged it for this cycle) -> the filter now INCLUDES it.
//  3. review_due_at is advanced past review_surfaced_at (simulating
//     mark-reviewed's effect on the row, per MarkReviewedService set out in
//     M6 F6.2 T5 — this test drives the same column update directly rather
//     than invoking the full service, since only the filter's read of the
//     marker relationship is under test here) -> review_surfaced_at <
//     review_due_at again -> the filter auto-EXCLUDES (expels) it, with no
//     separate "un-surface" step.

import (
	"context"
	"testing"
	"time"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/internal/modules/documents/repository"
	iamdomain "metaldocs/internal/modules/iam/domain"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/tests/integration/testdb"
)

func TestIntegration_ReviewDueFilter_ReadsSurfaced(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tnt := testdb.NewTenant(t, db)
	repo := repository.New(db, iamdomain.NoopUserDisplayNameReader{}, controlleddocumentsdomain.NoopCDFieldReader{}, taxonomydomain.NoopAreaCatalogReader{})

	now := time.Now().UTC().Truncate(time.Second)
	effectiveFrom := now.Add(-60 * 24 * time.Hour)
	reviewDueAt := now.Add(-1 * time.Hour) // already due

	doc := testdb.NewDocument(t, db, testdb.WithTenant(tnt.ID), testdb.WithStatus("published"))

	// Phase 1: due, but never surfaced (review_surfaced_at NULL) -> excluded.
	setReviewColumns(t, db, doc.ID, &effectiveFrom, nil, &reviewDueAt, nil)

	items, _, _, err := repo.ListDocumentsPaginated(ctx, tnt.ID, repository.ListOptions{PageSize: 20, ReviewDue: true})
	if err != nil {
		t.Fatalf("phase 1 ListDocumentsPaginated: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("phase 1: got %d items, want 0 (due but not yet surfaced must be excluded): %+v", len(items), items)
	}

	// Phase 2: surfacer flags it for this cycle (review_surfaced_at >= review_due_at) -> included.
	surfacedAt := now
	setReviewColumnsWithSurfaced(t, db, doc.ID, &effectiveFrom, nil, &reviewDueAt, nil, &surfacedAt)

	items, _, _, err = repo.ListDocumentsPaginated(ctx, tnt.ID, repository.ListOptions{PageSize: 20, ReviewDue: true})
	if err != nil {
		t.Fatalf("phase 2 ListDocumentsPaginated: %v", err)
	}
	if len(items) != 1 || items[0].ID != doc.ID {
		t.Fatalf("phase 2: got %+v, want exactly the surfaced doc %q", items, doc.ID)
	}

	// Phase 3: mark-reviewed advances review_due_at past review_surfaced_at
	// (the same relationship MarkReviewedService's UPDATE produces) -> the
	// doc auto-expels from the worklist until the surfacer re-flags it on
	// its next due cycle.
	advancedReviewDueAt := surfacedAt.Add(1 * time.Hour)
	setReviewColumnsWithSurfaced(t, db, doc.ID, &effectiveFrom, nil, &advancedReviewDueAt, &now, &surfacedAt)

	items, _, _, err = repo.ListDocumentsPaginated(ctx, tnt.ID, repository.ListOptions{PageSize: 20, ReviewDue: true})
	if err != nil {
		t.Fatalf("phase 3 ListDocumentsPaginated: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("phase 3: got %d items, want 0 (mark-reviewed must auto-expel until re-surfaced): %+v", len(items), items)
	}
}
