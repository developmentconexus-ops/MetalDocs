package resolvers

import (
	"context"
	"time"
)

type ControlledByAreaResolver struct{}

func (ControlledByAreaResolver) Key() string { return "controlled_by_area" }

func (ControlledByAreaResolver) Version() int { return 2 }

func (ControlledByAreaResolver) Resolve(ctx context.Context, in ResolveInput) (ResolvedValue, error) {
	if err := requireTenantID("controlled_by_area", in.TenantID); err != nil {
		return ResolvedValue{}, err
	}
	value := in.AreaNameSnapshot
	if value == "" {
		value = in.AreaCodeSnapshot
	}
	inputsHash, err := hashInputs(struct {
		TenantID         TenantID `json:"tenant_id"`
		AreaNameSnapshot string   `json:"area_name_snapshot"`
		AreaCodeFallback string   `json:"area_code_fallback,omitempty"`
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
