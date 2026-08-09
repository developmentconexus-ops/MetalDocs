package resolvers

import (
	"context"
	"strings"
	"time"
)

// ApproversResolver resolves the "approvers" placeholder to a comma-separated
// list of approver display names, or a pending-approval placeholder string.
type ApproversResolver struct{}

// Key returns the resolver's registry key, "approvers".
func (ApproversResolver) Key() string { return "approvers" }

// Version returns the resolver's version.
func (ApproversResolver) Version() int { return 1 }

// Resolve computes the approvers value for in.
func (ApproversResolver) Resolve(ctx context.Context, in ResolveInput) (ResolvedValue, error) {
	if err := requireTenantID("approvers", in.TenantID); err != nil {
		return ResolvedValue{}, err
	}
	if err := requireRevisionID("approvers", in.RevisionID); err != nil {
		return ResolvedValue{}, err
	}
	approvers, err := in.WorkflowReader.GetApprovers(ctx, in.TenantID, in.RevisionID, in.ApprovalInstanceID)
	if err != nil {
		return ResolvedValue{}, err
	}

	var value string
	names := make([]string, 0, len(approvers))
	for _, a := range approvers {
		if a.DisplayName != "" {
			names = append(names, a.DisplayName)
		}
	}
	if len(names) == 0 {
		value = "[aguardando aprovação]"
	} else {
		value = strings.Join(names, ", ")
	}

	inputsHash, err := hashInputs(struct {
		TenantID           TenantID           `json:"tenant_id"`
		RevisionID         RevisionID         `json:"revision_id"`
		ApprovalInstanceID ApprovalInstanceID `json:"approval_instance_id"`
	}{in.TenantID, in.RevisionID, in.ApprovalInstanceID})
	if err != nil {
		return ResolvedValue{}, err
	}

	return ResolvedValue{
		Value:       value,
		ResolverKey: "approvers",
		ResolverVer: 1,
		InputsHash:  inputsHash,
		ComputedAt:  time.Now().UTC(),
	}, nil
}
