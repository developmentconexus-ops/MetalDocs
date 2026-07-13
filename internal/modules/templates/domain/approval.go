package domain

// SegregationRole identifies which lifecycle actor role is being checked for
// ISO segregation-of-duties conflicts.
type SegregationRole string

const (
	// SegregationRoleApprover identifies the approver role for segregation checks.
	SegregationRoleApprover SegregationRole = "approver"
)

// CheckSegregation enforces ISO segregation of duties.
// role = "approver"
// Rules:
//
//	role="approver": actorID != authorID AND (reviewerID == nil OR actorID != *reviewerID)
//
// Returns ErrISOSegregationViolation on conflict.
func CheckSegregation(role SegregationRole, actorID, authorID string, reviewerID *string) error {
	switch role {
	case SegregationRoleApprover:
		if actorID == authorID {
			return ErrISOSegregationViolation
		}
		if reviewerID != nil && actorID == *reviewerID {
			return ErrISOSegregationViolation
		}
		return nil
	default:
		return ErrForbidden
	}
}
