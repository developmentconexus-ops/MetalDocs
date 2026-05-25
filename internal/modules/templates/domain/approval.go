package domain

type ApprovalConfig struct {
	TemplateID   string
	ReviewerRole *string
	ApproverRole string
}

func NewApprovalConfig(templateID, approverRole string, reviewerRole *string) (ApprovalConfig, error) {
	if templateID == "" || approverRole == "" {
		return ApprovalConfig{}, ErrInvalidApprovalConfig
	}
	return ApprovalConfig{
		TemplateID:   templateID,
		ReviewerRole: reviewerRole,
		ApproverRole: approverRole,
	}, nil
}

func (c ApprovalConfig) HasReviewer() bool { return c.ReviewerRole != nil && *c.ReviewerRole != "" }

type SegregationRole string

const (
	SegregationRoleReviewer SegregationRole = "reviewer"
	SegregationRoleApprover SegregationRole = "approver"
)

// CheckSegregation enforces ISO segregation of duties.
// role = SegregationRoleReviewer | SegregationRoleApprover
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
