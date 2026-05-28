package resolvers

import (
	"context"
	"errors"
	"time"
)

type ApprovalDateResolver struct{}

func (ApprovalDateResolver) Key() string { return "approval_date" }

func (ApprovalDateResolver) Version() int { return 1 }

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
	if approvalDate.IsZero() {
		return ResolvedValue{}, errors.New("approval_date resolver: final approval date is zero")
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
		Value:       approvalDate.UTC().Format("2006-01-02"),
		ResolverKey: "approval_date",
		ResolverVer: 1,
		InputsHash:  inputsHash,
		ComputedAt:  time.Now().UTC(),
	}, nil
}
