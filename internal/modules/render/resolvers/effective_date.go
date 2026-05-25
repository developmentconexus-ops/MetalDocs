package resolvers

import (
	"context"
	"fmt"
	"time"
)

type EffectiveDateResolver struct{}

func (EffectiveDateResolver) Key() string { return "effective_date" }

func (EffectiveDateResolver) Version() int { return 1 }

func (r EffectiveDateResolver) Resolve(ctx context.Context, in ResolveInput) (ResolvedValue, error) {
	if in.TenantID == "" {
		return ResolvedValue{}, fmt.Errorf("%s: TenantID is required", r.Key())
	}
	if in.RevisionID == "" {
		return ResolvedValue{}, fmt.Errorf("%s: RevisionID is required", r.Key())
	}
	effectiveFrom, err := in.RevisionReader.GetEffectiveFrom(ctx, in.TenantID, in.RevisionID)
	if err != nil {
		return ResolvedValue{}, err
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
