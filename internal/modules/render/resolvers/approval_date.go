package resolvers

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type ApprovalDateResolver struct{}

func (ApprovalDateResolver) Key() string { return "approval_date" }

func (ApprovalDateResolver) Version() int { return 1 }

func (r ApprovalDateResolver) Resolve(ctx context.Context, in ResolveInput) (ResolvedValue, error) {
	if in.TenantID == "" {
		return ResolvedValue{}, fmt.Errorf("%s: TenantID is required", r.Key())
	}
	if in.RevisionID == "" {
		return ResolvedValue{}, fmt.Errorf("%s: RevisionID is required", r.Key())
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
		TenantID   string `json:"tenant_id"`
		RevisionID string `json:"revision_id"`
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
