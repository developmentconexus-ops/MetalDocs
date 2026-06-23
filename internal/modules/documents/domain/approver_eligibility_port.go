package domain

import "context"

// ApproverEligibilityReader reports whether an actor is an eligible approver of
// the document's currently-open approval stage. Backed by
// approval_stage_instances.eligible_actor_ids (populated at submit time).
type ApproverEligibilityReader interface {
	IsEligibleApprover(ctx context.Context, tenantID, documentID, actorUserID string) (bool, error)
}
