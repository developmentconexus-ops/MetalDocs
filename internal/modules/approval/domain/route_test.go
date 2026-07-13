package domain

import (
	"errors"
	"testing"

	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
)

func intPtr(n int) *int { return &n }

func happyRoute() Route {
	return Route{
		ID: "r1", TenantID: "t1", ProfileCode: "SOP", Version: 1,
		Stages: []Stage{
			{Order: 1, Name: "QA", RequiredRole: "approver", RequiredCapability: "document.signoff", AreaCode: "qa", Quorum: QuorumAny1Of, OnEligibilityDrift: DriftReduceQuorum},
			{Order: 2, Name: "Manager", RequiredRole: "manager", RequiredCapability: "document.signoff", AreaCode: "mgmt", Quorum: QuorumAllOf, OnEligibilityDrift: DriftKeepSnapshot},
			{Order: 3, Name: "Director", RequiredRole: "director", RequiredCapability: "document.signoff", AreaCode: "exec", Quorum: QuorumMofN, QuorumM: intPtr(2), OnEligibilityDrift: DriftFailStage},
		},
	}
}

func TestRouteValidateHappy(t *testing.T) {
	if err := happyRoute().Validate(""); err != nil {
		t.Fatalf("happy route invalid: %v", err)
	}
}

func TestRouteValidateEmptyStages(t *testing.T) {
	r := Route{Stages: []Stage{}}
	if err := r.Validate(""); err == nil {
		t.Error("empty stages should fail validation")
	}
}

func TestRouteValidateNonDenseOrder(t *testing.T) {
	r := Route{
		Stages: []Stage{
			{Order: 1, Name: "A", Quorum: QuorumAny1Of, OnEligibilityDrift: DriftReduceQuorum},
			{Order: 3, Name: "B", Quorum: QuorumAny1Of, OnEligibilityDrift: DriftReduceQuorum},
		},
	}
	if err := r.Validate(""); err == nil {
		t.Error("non-dense order [1,3] should fail validation")
	}
}

func TestRouteValidateMofNWithoutM(t *testing.T) {
	r := Route{
		Stages: []Stage{
			{Order: 1, Name: "A", Quorum: QuorumMofN, OnEligibilityDrift: DriftReduceQuorum},
		},
	}
	if err := r.Validate(""); err == nil {
		t.Error("m_of_n without QuorumM should fail")
	}
}

func TestRouteValidateAny1OfWithM(t *testing.T) {
	r := Route{
		Stages: []Stage{
			{Order: 1, Name: "A", Quorum: QuorumAny1Of, QuorumM: intPtr(1), OnEligibilityDrift: DriftReduceQuorum},
		},
	}
	if err := r.Validate(""); err == nil {
		t.Error("any_1_of with QuorumM set should fail")
	}
}

func TestRouteValidateMofNZeroM(t *testing.T) {
	r := Route{
		Stages: []Stage{
			{Order: 1, Name: "A", Quorum: QuorumMofN, QuorumM: intPtr(0), OnEligibilityDrift: DriftReduceQuorum},
		},
	}
	if err := r.Validate(""); err == nil {
		t.Error("QuorumM=0 should fail")
	}
}

func TestRouteValidateDuplicateNames(t *testing.T) {
	r := Route{
		Stages: []Stage{
			{Order: 1, Name: "Dup", Quorum: QuorumAny1Of, OnEligibilityDrift: DriftReduceQuorum},
			{Order: 2, Name: "Dup", Quorum: QuorumAny1Of, OnEligibilityDrift: DriftReduceQuorum},
		},
	}
	if err := r.Validate(""); err == nil {
		t.Error("duplicate stage names should fail")
	}
}

// reviewOnlyRoute is a structurally valid route whose single stage is a review
// (non-signature) stage — the shape controlado must reject and simples permits.
func reviewOnlyRoute() Route {
	return Route{
		ID: "r1", TenantID: "t1", ProfileCode: "REL", Version: 1,
		Stages: []Stage{
			{Order: 1, Name: "Peer review", Quorum: QuorumAny1Of, OnEligibilityDrift: DriftReduceQuorum, Kind: StageKindReview, Selectors: []ActorSelector{{Kind: SelectorRoleInDocumentArea, Role: "reviewer"}}},
		},
	}
}

// approvalStageRoute is a structurally valid route with one explicit approval
// (signature) stage — satisfies controlado.
func approvalStageRoute() Route {
	return Route{
		ID: "r1", TenantID: "t1", ProfileCode: "POP", Version: 1,
		Stages: []Stage{
			{Order: 1, Name: "Review", Quorum: QuorumAny1Of, OnEligibilityDrift: DriftReduceQuorum, Kind: StageKindReview, Selectors: []ActorSelector{{Kind: SelectorRoleInDocumentArea, Role: "reviewer"}}},
			{Order: 2, Name: "QA signoff", Quorum: QuorumAllOf, OnEligibilityDrift: DriftKeepSnapshot, Kind: StageKindApproval, Selectors: []ActorSelector{{Kind: SelectorRoleInFixedArea, Role: "approver", AreaCode: "qa"}}},
		},
	}
}

func TestRouteValidatePolicy(t *testing.T) {
	cases := []struct {
		name    string
		route   Route
		policy  taxonomydomain.RoutePolicy
		wantErr error
	}{
		{"no-route-permitted rejects any route", happyRoute(), taxonomydomain.RoutePolicyNoRoutePermitted, ErrRouteNotPermittedForProfile},
		{"require-approval rejects review-only", reviewOnlyRoute(), taxonomydomain.RoutePolicyRequireApprovalStage, ErrApprovalStageRequired},
		{"require-approval accepts explicit approval stage", approvalStageRoute(), taxonomydomain.RoutePolicyRequireApprovalStage, nil},
		{"require-approval accepts unset-Kind route (defaults to approval)", happyRoute(), taxonomydomain.RoutePolicyRequireApprovalStage, nil},
		{"optional accepts review-only", reviewOnlyRoute(), taxonomydomain.RoutePolicyApprovalOptional, nil},
		{"empty policy imposes no signature constraint", reviewOnlyRoute(), taxonomydomain.RoutePolicy(""), nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.route.Validate(tc.policy)
			if tc.wantErr == nil {
				if err != nil {
					t.Fatalf("Validate(%q) = %v, want nil", tc.policy, err)
				}
				return
			}
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("Validate(%q) = %v, want %v", tc.policy, err, tc.wantErr)
			}
		})
	}
}

// TestRouteValidateStructuralBeatsPolicy confirms a structural failure is
// reported even when a policy would also reject — structural checks run first.
func TestRouteValidateStructuralBeatsPolicy(t *testing.T) {
	empty := Route{Stages: []Stage{}}
	if err := empty.Validate(taxonomydomain.RoutePolicyNoRoutePermitted); err == nil {
		t.Fatal("empty stages should fail structurally regardless of policy")
	} else if errors.Is(err, ErrRouteNotPermittedForProfile) {
		t.Fatal("structural error should be reported before the policy error")
	}
}
