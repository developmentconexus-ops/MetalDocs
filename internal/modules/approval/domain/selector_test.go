package domain

import (
	"errors"
	"testing"
)

// TestActorSelectorValidateValid asserts the one canonical valid shape for
// each SelectorKind passes.
func TestActorSelectorValidateValid(t *testing.T) {
	cases := []struct {
		name string
		sel  ActorSelector
	}{
		{"named_user", ActorSelector{Kind: SelectorNamedUser, UserID: "u1"}},
		{"role_in_fixed_area", ActorSelector{Kind: SelectorRoleInFixedArea, Role: "approver", AreaCode: "qa"}},
		{"role_in_document_area", ActorSelector{Kind: SelectorRoleInDocumentArea, Role: "approver"}},
		{"submit_choice", ActorSelector{Kind: SelectorSubmitChoice, Role: "approver", AreaCode: "qa"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.sel.Validate(); err != nil {
				t.Fatalf("Validate() = %v, want nil", err)
			}
		})
	}
}

// TestActorSelectorValidateInvalidKind asserts an unrecognized Kind fails
// with ErrInvalidSelectorKind regardless of which fields are set.
func TestActorSelectorValidateInvalidKind(t *testing.T) {
	sel := ActorSelector{Kind: "bogus_kind", UserID: "u1"}
	if err := sel.Validate(); !errors.Is(err, ErrInvalidSelectorKind) {
		t.Fatalf("Validate() = %v, want ErrInvalidSelectorKind", err)
	}
}

// TestActorSelectorValidateFieldPermutations asserts every missing-required /
// forbidden-field-present permutation for each SelectorKind fails with
// ErrSelectorFieldsInvalid.
func TestActorSelectorValidateFieldPermutations(t *testing.T) {
	cases := []struct {
		name string
		sel  ActorSelector
	}{
		// named_user: requires UserID; forbids Role, AreaCode.
		{"named_user missing UserID", ActorSelector{Kind: SelectorNamedUser}},
		{"named_user with Role", ActorSelector{Kind: SelectorNamedUser, UserID: "u1", Role: "approver"}},
		{"named_user with AreaCode", ActorSelector{Kind: SelectorNamedUser, UserID: "u1", AreaCode: "qa"}},

		// role_in_fixed_area: requires Role and AreaCode; forbids UserID.
		{"role_in_fixed_area missing Role", ActorSelector{Kind: SelectorRoleInFixedArea, AreaCode: "qa"}},
		{"role_in_fixed_area missing AreaCode", ActorSelector{Kind: SelectorRoleInFixedArea, Role: "approver"}},
		{"role_in_fixed_area with UserID", ActorSelector{Kind: SelectorRoleInFixedArea, Role: "approver", AreaCode: "qa", UserID: "u1"}},

		// role_in_document_area: requires Role; forbids UserID, AreaCode.
		{"role_in_document_area missing Role", ActorSelector{Kind: SelectorRoleInDocumentArea}},
		{"role_in_document_area with UserID", ActorSelector{Kind: SelectorRoleInDocumentArea, Role: "approver", UserID: "u1"}},
		{"role_in_document_area with AreaCode", ActorSelector{Kind: SelectorRoleInDocumentArea, Role: "approver", AreaCode: "qa"}},

		// submit_choice: requires Role and AreaCode; forbids UserID.
		{"submit_choice missing Role", ActorSelector{Kind: SelectorSubmitChoice, AreaCode: "qa"}},
		{"submit_choice missing AreaCode", ActorSelector{Kind: SelectorSubmitChoice, Role: "approver"}},
		{"submit_choice with UserID", ActorSelector{Kind: SelectorSubmitChoice, Role: "approver", AreaCode: "qa", UserID: "u1"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.sel.Validate()
			if !errors.Is(err, ErrSelectorFieldsInvalid) {
				t.Fatalf("Validate() = %v, want ErrSelectorFieldsInvalid", err)
			}
		})
	}
}

// TestFlatRoleArea covers the reverse mapping used to repopulate audit
// snapshots and compat wire fields: a single role_in_fixed_area selector
// yields its (role, area_code); zero, multiple, or other-kind-only selectors
// yield ("", "").
func TestFlatRoleArea(t *testing.T) {
	t.Run("single role_in_fixed_area", func(t *testing.T) {
		role, area := FlatRoleArea([]ActorSelector{{Kind: SelectorRoleInFixedArea, Role: "approver", AreaCode: "qa"}})
		if role != "approver" || area != "qa" {
			t.Fatalf("FlatRoleArea() = (%q, %q), want (approver, qa)", role, area)
		}
	})

	t.Run("no role_in_fixed_area", func(t *testing.T) {
		role, area := FlatRoleArea([]ActorSelector{{Kind: SelectorNamedUser, UserID: "u1"}})
		if role != "" || area != "" {
			t.Fatalf("FlatRoleArea() = (%q, %q), want (\"\", \"\")", role, area)
		}
	})

	t.Run("multiple role_in_fixed_area", func(t *testing.T) {
		role, area := FlatRoleArea([]ActorSelector{
			{Kind: SelectorRoleInFixedArea, Role: "approver", AreaCode: "qa"},
			{Kind: SelectorRoleInFixedArea, Role: "director", AreaCode: "exec"},
		})
		if role != "" || area != "" {
			t.Fatalf("FlatRoleArea() = (%q, %q), want (\"\", \"\")", role, area)
		}
	})

	t.Run("empty", func(t *testing.T) {
		role, area := FlatRoleArea(nil)
		if role != "" || area != "" {
			t.Fatalf("FlatRoleArea() = (%q, %q), want (\"\", \"\")", role, area)
		}
	})
}

// TestRouteValidateSelectors covers the Route.Validate integration: a stage
// with zero selectors fails with ErrStageNoSelector, a stage with an invalid
// selector propagates the selector error, a legacy RequiredRole/AreaCode stage
// passes via the synthesized selector, and a happy-path route with mixed
// selector kinds across stages validates cleanly.
func TestRouteValidateSelectors(t *testing.T) {
	t.Run("zero selectors and no legacy fields", func(t *testing.T) {
		r := Route{
			Stages: []Stage{
				{Order: 1, Name: "A", Quorum: QuorumAny1Of, OnEligibilityDrift: DriftReduceQuorum},
			},
		}
		if err := r.Validate(""); !errors.Is(err, ErrStageNoSelector) {
			t.Fatalf("Validate() = %v, want ErrStageNoSelector", err)
		}
	})

	t.Run("selector-populated stage passes", func(t *testing.T) {
		r := Route{
			Stages: []Stage{
				{
					Order: 1, Name: "A", Quorum: QuorumAny1Of, OnEligibilityDrift: DriftReduceQuorum,
					Selectors: []ActorSelector{{Kind: SelectorRoleInFixedArea, Role: "approver", AreaCode: "qa"}},
				},
			},
		}
		if err := r.Validate(""); err != nil {
			t.Fatalf("Validate() = %v, want nil", err)
		}
	})

	t.Run("invalid selector propagates", func(t *testing.T) {
		r := Route{
			Stages: []Stage{
				{
					Order: 1, Name: "A", Quorum: QuorumAny1Of, OnEligibilityDrift: DriftReduceQuorum,
					Selectors: []ActorSelector{{Kind: SelectorNamedUser, Role: "approver"}},
				},
			},
		}
		if err := r.Validate(""); !errors.Is(err, ErrSelectorFieldsInvalid) {
			t.Fatalf("Validate() = %v, want ErrSelectorFieldsInvalid", err)
		}
	})

	t.Run("happy path with mixed selectors", func(t *testing.T) {
		r := Route{
			Stages: []Stage{
				{
					Order: 1, Name: "Review", Quorum: QuorumAny1Of, OnEligibilityDrift: DriftReduceQuorum, Kind: StageKindReview,
					Selectors: []ActorSelector{{Kind: SelectorRoleInDocumentArea, Role: "reviewer"}},
				},
				{
					Order: 2, Name: "QA", Quorum: QuorumAllOf, OnEligibilityDrift: DriftKeepSnapshot, Kind: StageKindApproval,
					Selectors: []ActorSelector{
						{Kind: SelectorRoleInFixedArea, Role: "approver", AreaCode: "qa"},
						{Kind: SelectorNamedUser, UserID: "u1"},
					},
				},
				{
					Order: 3, Name: "Submit", Quorum: QuorumAny1Of, OnEligibilityDrift: DriftReduceQuorum, Kind: StageKindApproval,
					Selectors: []ActorSelector{{Kind: SelectorSubmitChoice, Role: "director", AreaCode: "exec"}},
				},
			},
		}
		if err := r.Validate(""); err != nil {
			t.Fatalf("Validate() = %v, want nil", err)
		}
	})
}
