package infrastructure

import (
	"context"
	"errors"
	"testing"

	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
)

// fakeProfileGetter is a hand-rolled stub of the narrow taxonomy profile read
// slice the adapter depends on. It records the args it was called with so the
// test can assert the ProfileCode conversion, and returns a canned profile/err.
type fakeProfileGetter struct {
	profile *taxonomydomain.DocumentProfile
	err     error

	gotTenantID string
	gotCode     taxonomydomain.ProfileCode
}

func (f *fakeProfileGetter) GetByCode(_ context.Context, tenantID string, code taxonomydomain.ProfileCode) (*taxonomydomain.DocumentProfile, error) {
	f.gotTenantID = tenantID
	f.gotCode = code
	return f.profile, f.err
}

func TestProfilePolicyReader_RoutePolicy_DerivesFromGovernanceClass(t *testing.T) {
	cases := []struct {
		name  string
		class taxonomydomain.GovernanceClass
		want  taxonomydomain.RoutePolicy
	}{
		{"controlado requires approval stage", taxonomydomain.GovernanceControlado, taxonomydomain.RoutePolicyRequireApprovalStage},
		{"simples optional", taxonomydomain.GovernanceSimples, taxonomydomain.RoutePolicyApprovalOptional},
		{"livre requires a zero-stage route", taxonomydomain.GovernanceLivre, taxonomydomain.RoutePolicyNoApprovalStages},
		{"unknown fail-closes to require", taxonomydomain.GovernanceClass("mystery"), taxonomydomain.RoutePolicyRequireApprovalStage},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			getter := &fakeProfileGetter{
				profile: &taxonomydomain.DocumentProfile{GovernanceClass: tc.class},
			}
			r := NewProfilePolicyReader(getter)

			got, err := r.RoutePolicy(context.Background(), "tenant-1", "POP")
			if err != nil {
				t.Fatalf("RoutePolicy() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("RoutePolicy() = %q, want %q", got, tc.want)
			}
			if getter.gotTenantID != "tenant-1" {
				t.Fatalf("tenantID passed = %q, want %q", getter.gotTenantID, "tenant-1")
			}
			if getter.gotCode != taxonomydomain.ProfileCode("POP") {
				t.Fatalf("code passed = %q, want %q", getter.gotCode, "POP")
			}
		})
	}
}

func TestProfilePolicyReader_RoutePolicy_PropagatesError(t *testing.T) {
	sentinel := errors.New("profile not found")
	r := NewProfilePolicyReader(&fakeProfileGetter{err: sentinel})

	_, err := r.RoutePolicy(context.Background(), "tenant-1", "GONE")
	if !errors.Is(err, sentinel) {
		t.Fatalf("RoutePolicy() error = %v, want %v", err, sentinel)
	}
}
