package domain_test

import (
	"testing"

	"metaldocs/internal/modules/documents/domain"
)

// TestCanTransitionDocumentStatus_DBTriggerParity pins
// domain.CanTransitionDocumentStatus as a faithful mirror of the Postgres
// trigger public.enforce_document_transition (rewritten by migration
// db/migrations/0272_documents_remove_rejected.sql). It is the contract
// dual of TestCanTransitionDocumentStatus (state_test.go): that test proves
// exhaustiveness against the trigger from the app side; this test pins the
// literal expected-legal set as an independent transcription of the DB
// trigger's arcs, so the two together catch drift on either wording.
//
// Expected-legal set — transcribed from db/migrations/0272
// enforce_document_transition (the DB trigger's last-committed body).
// Keep this literal in sync with that migration's trigger body:
//
//	under_review → draft       (GUC-gated special case: only when
//	                             metaldocs.cancel_in_progress is set; the DB
//	                             trigger enforces the GUC gate, this function
//	                             only reports the arc as structurally legal)
//	draft         → under_review
//	under_review  → approved
//	approved      → published
//	approved      → scheduled
//	approved      → draft
//	scheduled     → published
//	scheduled     → draft
//	published     → superseded
//	published     → obsolete
//	superseded    → obsolete
//
// 11 arcs total. Note: the live cross-check against a real Postgres
// instance (actually running the 0272 trigger and comparing outcomes) is
// F4.2 / dedicated-integration scope — bounded out of this unit test.
func TestCanTransitionDocumentStatus_DBTriggerParity(t *testing.T) {
	// transcribed from db/migrations/0272 enforce_document_transition (the DB last line). Keep in sync.
	expectedLegal := [][2]domain.DocumentStatus{
		{domain.DocStatusUnderReview, domain.DocStatusDraft},
		{domain.DocStatusDraft, domain.DocStatusUnderReview},
		{domain.DocStatusUnderReview, domain.DocStatusApproved},
		{domain.DocStatusApproved, domain.DocStatusPublished},
		{domain.DocStatusApproved, domain.DocStatusScheduled},
		{domain.DocStatusApproved, domain.DocStatusDraft},
		{domain.DocStatusScheduled, domain.DocStatusPublished},
		{domain.DocStatusScheduled, domain.DocStatusDraft},
		{domain.DocStatusPublished, domain.DocStatusSuperseded},
		{domain.DocStatusPublished, domain.DocStatusObsolete},
		{domain.DocStatusSuperseded, domain.DocStatusObsolete},
	}
	if len(expectedLegal) != 11 {
		t.Fatalf("test setup error: expected 11 legal arcs transcribed from the DB trigger, got %d", len(expectedLegal))
	}

	legalSet := make(map[[2]domain.DocumentStatus]bool, len(expectedLegal))
	for _, arc := range expectedLegal {
		legalSet[arc] = true
	}

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

	for _, cur := range statuses {
		for _, next := range statuses {
			pair := [2]domain.DocumentStatus{cur, next}
			wantLegal := legalSet[pair]
			err := domain.CanTransitionDocumentStatus(cur, next)
			gotLegal := err == nil

			if gotLegal != wantLegal {
				t.Errorf("CanTransitionDocumentStatus(%s, %s) legal=%v, want legal=%v (DB-trigger-transcribed parity contract violated)",
					cur, next, gotLegal, wantLegal)
			}
		}
	}
}
