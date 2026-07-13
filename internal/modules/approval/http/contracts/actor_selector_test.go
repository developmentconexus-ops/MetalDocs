package contracts

import "testing"

// TestActorSelectorValidate pins the M4 ActorSelector wire Validate contract
// (unit 3.2, slice 4): it must mirror domain.ActorSelector.Validate exactly —
// one valid shape per kind, plus the malformed variants for each.
func TestActorSelectorValidate(t *testing.T) {
	cases := []struct {
		name    string
		sel     ActorSelector
		wantErr bool
	}{
		{"named_user valid", ActorSelector{Kind: SelectorKindNamedUser, UserID: "user-1"}, false},
		{"named_user missing user_id", ActorSelector{Kind: SelectorKindNamedUser}, true},
		{"named_user with role", ActorSelector{Kind: SelectorKindNamedUser, UserID: "user-1", Role: "approver"}, true},
		{"named_user with area_code", ActorSelector{Kind: SelectorKindNamedUser, UserID: "user-1", AreaCode: "ops"}, true},

		{"role_in_fixed_area valid", ActorSelector{Kind: SelectorKindRoleInFixedArea, Role: "approver", AreaCode: "ops"}, false},
		{"role_in_fixed_area missing area_code", ActorSelector{Kind: SelectorKindRoleInFixedArea, Role: "approver"}, true},
		{"role_in_fixed_area missing role", ActorSelector{Kind: SelectorKindRoleInFixedArea, AreaCode: "ops"}, true},
		{"role_in_fixed_area with user_id", ActorSelector{Kind: SelectorKindRoleInFixedArea, Role: "approver", AreaCode: "ops", UserID: "user-1"}, true},

		{"role_in_document_area valid", ActorSelector{Kind: SelectorKindRoleInDocumentArea, Role: "approver"}, false},
		{"role_in_document_area missing role", ActorSelector{Kind: SelectorKindRoleInDocumentArea}, true},
		{"role_in_document_area with area_code", ActorSelector{Kind: SelectorKindRoleInDocumentArea, Role: "approver", AreaCode: "ops"}, true},
		{"role_in_document_area with user_id", ActorSelector{Kind: SelectorKindRoleInDocumentArea, Role: "approver", UserID: "user-1"}, true},

		{"submit_choice valid", ActorSelector{Kind: SelectorKindSubmitChoice, Role: "approver", AreaCode: "ops"}, false},
		{"submit_choice missing area_code", ActorSelector{Kind: SelectorKindSubmitChoice, Role: "approver"}, true},
		{"submit_choice missing role", ActorSelector{Kind: SelectorKindSubmitChoice, AreaCode: "ops"}, true},
		{"submit_choice with user_id", ActorSelector{Kind: SelectorKindSubmitChoice, Role: "approver", AreaCode: "ops", UserID: "user-1"}, true},

		{"unknown kind", ActorSelector{Kind: SelectorKind("bogus"), UserID: "user-1"}, true},
		{"empty kind", ActorSelector{}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.sel.Validate()
			if tc.wantErr && err == nil {
				t.Fatalf("Validate() = nil, want error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("Validate() = %v, want nil", err)
			}
		})
	}
}

// TestValidateStages_SelectorsField pins validateStages' handling of the
// optional per-stage Selectors slice (M4, unit 3.2, slice 4): a stage with a
// valid named_user selector passes, and a stage with a malformed selector is
// rejected with a stages[i].selectors[j]-scoped error.
func TestValidateStages_SelectorsField(t *testing.T) {
	baseStage := StageRequest{
		Order:              1,
		Name:               "Review",
		RequiredCapability: "document.signoff",
		Quorum:             QuorumKindAny1Of, DriftPolicy: DriftPolicyKindReduceQuorum,
	}

	t.Run("valid named_user selector accepted", func(t *testing.T) {
		s := baseStage
		s.Selectors = []ActorSelector{{Kind: SelectorKindNamedUser, UserID: "user-1"}}
		if err := validateStages([]StageRequest{s}); err != nil {
			t.Fatalf("validateStages() = %v, want nil", err)
		}
	})

	t.Run("malformed selector rejected", func(t *testing.T) {
		s := baseStage
		s.Selectors = []ActorSelector{{Kind: SelectorKindNamedUser}}
		if err := validateStages([]StageRequest{s}); err == nil {
			t.Fatalf("validateStages() = nil, want error")
		}
	})

	// unit 3.2 slice 7a (wire contract extermination): selectors is now the
	// REQUIRED sole source of truth — an empty slice is rejected, not
	// silently bridged from flat required_role/area_code (which no longer
	// exist on the wire).
	t.Run("empty selectors rejected", func(t *testing.T) {
		s := baseStage
		if err := validateStages([]StageRequest{s}); err == nil {
			t.Fatalf("validateStages() = nil, want error for empty selectors")
		}
	})
}
