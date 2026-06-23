package domain

// CanWriteWorkingContent decides whether an actor may write the document's
// working-content revision chain.
//   - draft:        only the owner/author may write.
//   - under_review: only an eligible approver of the active stage may write.
//   - any other status: nobody writes working content.
func CanWriteWorkingContent(status string, isOwner, isEligibleApprover bool) bool {
	switch status {
	case string(DocStatusDraft):
		return isOwner
	case string(DocStatusUnderReview):
		return isEligibleApprover
	default:
		return false
	}
}
