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
			{Order: 1, Name: "QA", RequiredCapability: "document.signoff", Quorum: QuorumAny1Of, OnEligibilityDrift: DriftReduceQuorum, Selectors: []ActorSelector{{Kind: SelectorRoleInFixedArea, Role: "approver", AreaCode: "qa"}}},
			{Order: 2, Name: "Manager", RequiredCapability: "document.signoff", Quorum: QuorumAllOf, OnEligibilityDrift: DriftKeepSnapshot, Selectors: []ActorSelector{{Kind: SelectorRoleInFixedArea, Role: "manager", AreaCode: "mgmt"}}},
			{Order: 3, Name: "Director", RequiredCapability: "document.signoff", Quorum: QuorumMofN, QuorumM: intPtr(2), OnEligibilityDrift: DriftFailStage, Selectors: []ActorSelector{{Kind: SelectorRoleInFixedArea, Role: "director", AreaCode: "exec"}}},
		},
	}
}

func TestRouteValidateHappy(t *testing.T) {
	if err := happyRoute().Validate(""); err != nil {
		t.Fatalf("happy route invalid: %v", err)
	}
}

// TestRouteValidateEmptyStagesStructuralOnly pins ADR 0087: stage-count is a
// POLICY question. A structural-only Validate("") — what submit runs against an
// already-bound active route — must ACCEPT a zero-stage route, because a livre
// profile's configured route is exactly that.
func TestRouteValidateEmptyStagesStructuralOnly(t *testing.T) {
	r := Route{Stages: []Stage{}}
	if err := r.Validate(""); err != nil {
		t.Errorf("structural-only Validate(\"\") on a zero-stage route = %v, want nil", err)
	}
	if err := (Route{}).Validate(""); err != nil {
		t.Errorf("structural-only Validate(\"\") on a nil-stage route = %v, want nil", err)
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

// zeroStageRoute is the ADR 0087 livre shape: a fully configured, first-class
// route object that carries no stages at all.
func zeroStageRoute() Route {
	return Route{ID: "r1", TenantID: "t1", ProfileCode: "RASCUNHO", Version: 1}
}

func TestRouteValidatePolicy(t *testing.T) {
	cases := []struct {
		name    string
		route   Route
		policy  taxonomydomain.RoutePolicy
		wantErr error
	}{
		{"no-approval-stages rejects a staged route", happyRoute(), taxonomydomain.RoutePolicyNoApprovalStages, ErrRouteStagesNotPermittedForProfile},
		{"no-approval-stages rejects even a review-only stage", reviewOnlyRoute(), taxonomydomain.RoutePolicyNoApprovalStages, ErrRouteStagesNotPermittedForProfile},
		{"no-approval-stages accepts the zero-stage route", zeroStageRoute(), taxonomydomain.RoutePolicyNoApprovalStages, nil},
		{"require-approval rejects review-only", reviewOnlyRoute(), taxonomydomain.RoutePolicyRequireApprovalStage, ErrApprovalStageRequired},
		{"require-approval rejects a zero-stage route", zeroStageRoute(), taxonomydomain.RoutePolicyRequireApprovalStage, ErrApprovalStageRequired},
		{"require-approval accepts explicit approval stage", approvalStageRoute(), taxonomydomain.RoutePolicyRequireApprovalStage, nil},
		{"require-approval accepts unset-Kind route (defaults to approval)", happyRoute(), taxonomydomain.RoutePolicyRequireApprovalStage, nil},
		{"optional accepts review-only", reviewOnlyRoute(), taxonomydomain.RoutePolicyApprovalOptional, nil},
		{"optional rejects a zero-stage route", zeroStageRoute(), taxonomydomain.RoutePolicyApprovalOptional, ErrRouteStageRequired},
		{"empty policy imposes no stage-count or signature constraint", reviewOnlyRoute(), taxonomydomain.RoutePolicy(""), nil},
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

// TestRouteValidateStructuralBeatsPolicy confirms a per-stage structural
// failure is reported even when the policy would also reject — the per-stage
// loop runs before the policy switch.
func TestRouteValidateStructuralBeatsPolicy(t *testing.T) {
	malformed := Route{Stages: []Stage{
		{Order: 2, Name: "Misordered", Quorum: QuorumAny1Of, OnEligibilityDrift: DriftReduceQuorum,
			Selectors: []ActorSelector{{Kind: SelectorRoleInDocumentArea, Role: "reviewer"}}},
	}}
	err := malformed.Validate(taxonomydomain.RoutePolicyNoApprovalStages)
	if err == nil {
		t.Fatal("a misordered stage should fail structurally regardless of policy")
	}
	if errors.Is(err, ErrRouteStagesNotPermittedForProfile) {
		t.Fatal("structural error should be reported before the policy error")
	}
}
