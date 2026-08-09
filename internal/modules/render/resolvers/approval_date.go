package resolvers

import (
	"context"
	"errors"
	"time"
)

// ApprovalDateResolver resolves the "approval_date" placeholder to the
// revision's final approval date, or a pending-approval placeholder string.
type ApprovalDateResolver struct{}

// Key returns the resolver's registry key, "approval_date".
func (ApprovalDateResolver) Key() string { return "approval_date" }

// Version returns the resolver's version.
func (ApprovalDateResolver) Version() int { return 1 }

// Resolve computes the approval_date value for in.
func (ApprovalDateResolver) Resolve(ctx context.Context, in ResolveInput) (ResolvedValue, error) {
	if err := requireTenantID("approval_date", in.TenantID); err != nil {
		return ResolvedValue{}, err
	}
	if err := requireRevisionID("approval_date", in.RevisionID); err != nil {
		return ResolvedValue{}, err
	}
	if in.WorkflowReader == nil {
		return ResolvedValue{}, errors.New("approval_date resolver: workflow reader is nil")
	}
	approvalDate, err := in.WorkflowReader.GetFinalApprovalDate(ctx, in.TenantID, in.RevisionID)
	if err != nil {
		return ResolvedValue{}, err
	}
	value := "[aguardando aprovação]"
	if !approvalDate.IsZero() {
		value = approvalDate.UTC().Format("2006-01-02")
	}

	inputsHash, err := hashInputs(struct {
		TenantID   TenantID   `json:"tenant_id"`
		RevisionID RevisionID `json:"revision_id"`
	}{
		TenantID:   in.TenantID,
		RevisionID: in.RevisionID,
	})
	if err != nil {
		return ResolvedValue{}, err
	}

	return ResolvedValue{
		Value:       value,
		ResolverKey: "approval_date",
		ResolverVer: 1,
		InputsHash:  inputsHash,
		ComputedAt:  time.Now().UTC(),
	}, nil
}
