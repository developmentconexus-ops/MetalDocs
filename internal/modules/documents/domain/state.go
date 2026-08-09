package domain

// validDocumentTransitions is the legal-arc table CanTransitionDocumentStatus
// dispatches against. DocStatusObsolete and DocStatusArchived are
// deliberately absent (no entry == empty set == no outbound arcs): the former
// is terminal, the latter isn't a real lifecycle status (see
// CanTransitionDocumentStatus's doc comment below).
var validDocumentTransitions = map[DocumentStatus]map[DocumentStatus]bool{
	DocStatusDraft: {
		DocStatusUnderReview: true,
	},
	DocStatusUnderReview: {
		DocStatusApproved: true,
		DocStatusDraft:    true,
	},
	DocStatusApproved: {
		DocStatusPublished: true,
		DocStatusScheduled: true,
		DocStatusDraft:     true,
	},
	DocStatusScheduled: {
		DocStatusPublished: true,
		DocStatusDraft:     true,
	},
	DocStatusPublished: {
		DocStatusSuperseded: true,
		DocStatusObsolete:   true,
	},
	DocStatusSuperseded: {
		DocStatusObsolete: true,
	},
}

// CanTransitionDocumentStatus reports whether moving a document from cur to
// next is a legal lifecycle transition, mirroring the DB trigger
// enforce_document_transition (db/baseline/0001_current_schema.sql) exactly.
// Returns nil for a legal transition, ErrInvalidStateTransition otherwise.
//
// "archived" is NOT part of this lifecycle: it is a soft-archive timestamp
// flag (ADR 0010, documents.archived_at) set by MarkArchived without changing
// status, so every arc into or out of "archived" is illegal here.
//
// "rejected" was removed as a DocumentStatus entirely (migration 0272 +
// model.go), because the app runtime never wrote it: reject collapses
// under_review back to "draft". This function and the DB trigger dropped the
// under_review→rejected and rejected→draft arcs together (parity, see
// state_parity_test.go). The surviving under_review→draft arc mirrors both the
// reject rollback and the cancel rollback; the DB trigger additionally gates
// that specific arc on the metaldocs.cancel_in_progress GUC — this function
// only reports the arc as legal, it does not enforce the GUC (that remains a
// DB-only concern).
func CanTransitionDocumentStatus(cur, next DocumentStatus) error {
	if validDocumentTransitions[cur][next] {
		return nil
	}
	return ErrInvalidStateTransition
}
