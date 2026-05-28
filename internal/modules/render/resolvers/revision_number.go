package resolvers

import (
	"context"
	"time"
)

type RevisionNumberResolver struct{}

func (RevisionNumberResolver) Key() string { return "revision_number" }

func (RevisionNumberResolver) Version() int { return 1 }

func (RevisionNumberResolver) Resolve(ctx context.Context, in ResolveInput) (ResolvedValue, error) {
	if err := requireTenantID("revision_number", in.TenantID); err != nil {
		return ResolvedValue{}, err
	}
	if err := requireRevisionID("revision_number", in.RevisionID); err != nil {
		return ResolvedValue{}, err
	}
	revisionNumber, err := in.RevisionReader.GetRevisionNumber(ctx, in.TenantID, in.RevisionID)
	if err != nil {
		return ResolvedValue{}, err
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
		Value:       revisionNumber,
		ResolverKey: "revision_number",
		ResolverVer: 1,
		InputsHash:  inputsHash,
		ComputedAt:  time.Now().UTC(),
	}, nil
}
