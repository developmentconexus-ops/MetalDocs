package resolvers

import (
	"context"
	"time"
)

type EffectiveDateResolver struct{}

func (EffectiveDateResolver) Key() string { return "effective_date" }

func (EffectiveDateResolver) Version() int { return 1 }

func (EffectiveDateResolver) Resolve(ctx context.Context, in ResolveInput) (ResolvedValue, error) {
	if err := requireTenantID("effective_date", in.TenantID); err != nil {
		return ResolvedValue{}, err
	}
	if err := requireRevisionID("effective_date", in.RevisionID); err != nil {
		return ResolvedValue{}, err
	}
	effectiveFrom, err := in.RevisionReader.GetEffectiveFrom(ctx, in.TenantID, in.RevisionID)
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

	value := ""
	if !effectiveFrom.IsZero() {
		value = effectiveFrom.UTC().Format("2006-01-02")
	}
	return ResolvedValue{
		Value:       value,
		ResolverKey: "effective_date",
		ResolverVer: 1,
		InputsHash:  inputsHash,
		ComputedAt:  time.Now().UTC(),
	}, nil
}
