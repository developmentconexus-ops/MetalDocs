package domain

// CanTransitionDocument returns true iff a document can move from cur to next.
// draft -> under_review (submit for approval)
// under_review -> archived
// archived -> (terminal)
func CanTransitionDocument(cur, next DocumentStatus) bool {
	switch cur {
	case DocStatusDraft:
		return next == DocStatusFinalized || next == DocStatusArchived
	case DocStatusFinalized:
		return next == DocStatusArchived
	default:
		return false
	}
}
