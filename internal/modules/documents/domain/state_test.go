package domain_test

import (
	"errors"
	"testing"

	"metaldocs/internal/modules/documents/domain"
)

// TestCanTransitionDocumentStatus is the exhaustiveness proof for
// domain.CanTransitionDocumentStatus. It enumerates every ordered pair over
// the 8 non-"rejected" DocumentStatus values and asserts nil for exactly the
// 11 legal arcs mirrored from the DB trigger enforce_document_transition
// (db/baseline/0001_current_schema.sql), and domain.ErrInvalidStateTransition
// for every other pair, including all self-edges and every arc touching
// "archived" (archived is a soft-archive timestamp flag per ADR 0010, not a
// status transition).
func TestCanTransitionDocumentStatus(t *testing.T) {
	statuses := []domain.DocumentStatus{
		domain.DocStatusDraft,
		domain.DocStatusUnderReview,
		domain.DocStatusApproved,
		domain.DocStatusScheduled,
		domain.DocStatusPublished,
		domain.DocStatusSuperseded,
		domain.DocStatusObsolete,
		domain.DocStatusArchived,
	}

	legal := map[[2]domain.DocumentStatus]bool{
		{domain.DocStatusDraft, domain.DocStatusUnderReview}:    true,
		{domain.DocStatusUnderReview, domain.DocStatusApproved}: true,
		{domain.DocStatusUnderReview, domain.DocStatusDraft}:    true,
		{domain.DocStatusApproved, domain.DocStatusPublished}:   true,
		{domain.DocStatusApproved, domain.DocStatusScheduled}:   true,
		{domain.DocStatusApproved, domain.DocStatusDraft}:       true,
		{domain.DocStatusScheduled, domain.DocStatusPublished}:  true,
		{domain.DocStatusScheduled, domain.DocStatusDraft}:      true,
		{domain.DocStatusPublished, domain.DocStatusSuperseded}: true,
		{domain.DocStatusPublished, domain.DocStatusObsolete}:   true,
		{domain.DocStatusSuperseded, domain.DocStatusObsolete}:  true,
	}
	if len(legal) != 11 {
		t.Fatalf("test setup error: expected 11 legal arcs, got %d", len(legal))
	}

	for _, cur := range statuses {
		for _, next := range statuses {
			wantNil := legal[[2]domain.DocumentStatus{cur, next}]
			err := domain.CanTransitionDocumentStatus(cur, next)
			if wantNil {
				if err != nil {
					t.Errorf("CanTransitionDocumentStatus(%s, %s) = %v, want nil", cur, next, err)
				}
				continue
			}
			if err == nil {
				t.Errorf("CanTransitionDocumentStatus(%s, %s) = nil, want %v", cur, next, domain.ErrInvalidStateTransition)
				continue
			}
			if !errors.Is(err, domain.ErrInvalidStateTransition) {
				t.Errorf("CanTransitionDocumentStatus(%s, %s) = %v, want errors.Is(err, ErrInvalidStateTransition)", cur, next, err)
			}
		}
	}
}
