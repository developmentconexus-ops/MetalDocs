package resolvers

import (
	"context"
	"time"
)

// ControlledByAreaResolver resolves the "controlled_by_area" placeholder to
// the revision's snapshotted area name, falling back to its area code.
type ControlledByAreaResolver struct{}

// Key returns the resolver's registry key, "controlled_by_area".
func (ControlledByAreaResolver) Key() string { return "controlled_by_area" }

// Version returns the resolver's version.
func (ControlledByAreaResolver) Version() int { return 2 }

// Resolve computes the controlled_by_area value for in.
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
