package domain

// ApprovalConfig defines the role-based approval workflow for a template:
// the required approver role, and an optional reviewer role for a two-step
// review-then-approve flow.
type ApprovalConfig struct {
	TemplateID   string
	ReviewerRole *string
	ApproverRole string
}

// NewApprovalConfig creates a new ApprovalConfig for templateID, requiring a
// non-empty approverRole. reviewerRole is optional (nil disables the review
// step). Returns ErrInvalidApprovalConfig when templateID or approverRole is
// empty.
func NewApprovalConfig(templateID, approverRole string, reviewerRole *string) (ApprovalConfig, error) {
	if templateID == "" || approverRole == "" {
		return ApprovalConfig{}, ErrInvalidApprovalConfig
	}
	return ApprovalConfig{
		TemplateID:   templateID,
		ApproverRole: approverRole,
		ReviewerRole: reviewerRole,
	}, nil
}

// HasReviewer reports whether this config has a non-empty reviewer role
// configured, i.e. whether the two-step review-then-approve flow applies.
func (c ApprovalConfig) HasReviewer() bool { return c.ReviewerRole != nil && *c.ReviewerRole != "" }

// SegregationRole identifies which lifecycle actor role is being checked for
// ISO segregation-of-duties conflicts (reviewer or approver).
type SegregationRole string

const (
	// SegregationRoleReviewer identifies the reviewer role for segregation checks.
	SegregationRoleReviewer SegregationRole = "reviewer"
	// SegregationRoleApprover identifies the approver role for segregation checks.
	SegregationRoleApprover SegregationRole = "approver"
)

// CheckSegregation enforces ISO segregation of duties.
// role = "reviewer" | "approver"
// Rules:
//
//	role="reviewer": actorID != authorID
//	role="approver": actorID != authorID AND (reviewerID == nil OR actorID != *reviewerID)
//
// Returns ErrISOSegregationViolation on conflict.
func CheckSegregation(role SegregationRole, actorID, authorID string, reviewerID *string) error {
	switch role {
	case SegregationRoleReviewer:
		if actorID == authorID {
			return ErrISOSegregationViolation
		}
		return nil
	case SegregationRoleApprover:
		if actorID == authorID {
			return ErrISOSegregationViolation
		}
		if reviewerID != nil && actorID == *reviewerID {
			return ErrISOSegregationViolation
		}
		return nil
	default:
		return ErrForbiddenRole
	}
}
