package resolvers

import (
	"context"
	"fmt"
	"time"
)

type ControlledByAreaResolver struct{}

func (ControlledByAreaResolver) Key() string { return "controlled_by_area" }

func (ControlledByAreaResolver) Version() int { return 2 }

func (r ControlledByAreaResolver) Resolve(ctx context.Context, in ResolveInput) (ResolvedValue, error) {
	if in.TenantID == "" {
		return ResolvedValue{}, fmt.Errorf("%s: TenantID is required", r.Key())
	}
	value := in.AreaNameSnapshot
	if value == "" {
		value = in.AreaCodeSnapshot
	}
	inputsHash, err := hashInputs(struct {
		TenantID         string `json:"tenant_id"`
		AreaNameSnapshot string `json:"area_name_snapshot"`
		AreaCodeFallback string `json:"area_code_fallback,omitempty"`
	}{
		TenantID:         in.TenantID,
		AreaNameSnapshot: in.AreaNameSnapshot,
		AreaCodeFallback: in.AreaCodeSnapshot,
	})
	if err != nil {
		return ResolvedValue{}, err
	}

	return ResolvedValue{
		Value:       value,
		ResolverKey: "controlled_by_area",
		ResolverVer: 2,
		InputsHash:  inputsHash,
		ComputedAt:  time.Now().UTC(),
	}, nil
}
