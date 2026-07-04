package domain

// CanTransitionDocumentStatus reports whether moving a document from cur to
// next is a legal lifecycle transition, mirroring the DB trigger
// enforce_document_transition (db/baseline/0001_current_schema.sql) exactly.
// Returns nil for a legal transition, ErrInvalidStateTransition otherwise.
//
// "archived" is NOT part of this lifecycle: it is a soft-archive timestamp
// flag (ADR 0010, documents.archived_at) set by MarkArchived without changing
// status, so every arc into or out of "archived" is illegal here.
//
// "rejected" still exists as a DocumentStatus constant but has no arcs in
// this function: the app runtime never writes it (reject collapses
// under_review back to "draft"), matching the trigger's app-reachable
// subset. The under_review→draft arc mirrors both the reject rollback and
// the cancel rollback; the DB trigger additionally gates that specific arc
// on the metaldocs.cancel_in_progress GUC — this function only reports the
// arc as legal, it does not enforce the GUC (that remains a DB-only concern).
func CanTransitionDocumentStatus(cur, next DocumentStatus) error {
	switch cur {
	case DocStatusDraft:
		if next == DocStatusUnderReview {
			return nil
		}
	case DocStatusUnderReview:
		if next == DocStatusApproved || next == DocStatusDraft {
			return nil
		}
	case DocStatusApproved:
		if next == DocStatusPublished || next == DocStatusScheduled || next == DocStatusDraft {
			return nil
		}
	case DocStatusScheduled:
		if next == DocStatusPublished || next == DocStatusDraft {
			return nil
		}
	case DocStatusPublished:
		if next == DocStatusSuperseded || next == DocStatusObsolete {
			return nil
		}
	case DocStatusSuperseded:
		if next == DocStatusObsolete {
			return nil
		}
	}
	return ErrInvalidStateTransition
}
