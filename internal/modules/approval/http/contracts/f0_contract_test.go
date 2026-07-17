package contracts

import "testing"

// TestInstanceStatusWireEnumComplete pins F0: the wire InstanceStatus enum must
// cover every domain status (domain/instance.go). changes_requested was the
// carried M2b gap — a review verdict of request_changes moves the instance to
// changes_requested, and a read of that instance must serialize a value the FE
// can parse.
func TestInstanceStatusWireEnumComplete(t *testing.T) {
	want := []string{"in_progress", "approved", "rejected", "cancelled", "changes_requested"}
	for _, s := range want {
		if !IsValidInstanceStatus(s) {
			t.Fatalf("wire InstanceStatus missing %q", s)
		}
	}
	if IsValidInstanceStatus("bogus") {
		t.Fatalf("IsValidInstanceStatus accepted an unknown status")
	}
}

// TestStageRequestStageKindValidation pins F0: stage_kind is optional (empty
// defaults to approval downstream) but, when supplied, must be one of the two
// canonical kinds. Without this a review-kind route cannot be created; with a
// free-text kind, an unpersistable value slips through to the DB.
func TestStageRequestStageKindValidation(t *testing.T) {
	base := StageRequest{
		Order:              1,
		Name:               "Revisão",
		RequiredCapability: "document.signoff",
		Quorum:             QuorumKindAny1Of, DriftPolicy: DriftPolicyKindReduceQuorum,
		Selectors: []ActorSelector{{Kind: SelectorKindRoleInFixedArea, Role: "approver", AreaCode: "ops"}},
	}
	cases := []struct {
		name    string
		kind    StageKind
		wantErr bool
	}{
		{"empty defaults ok", "", false},
		{"review ok", StageKindReview, false},
		{"approval ok", StageKindApproval, false},
		{"free text rejected", StageKind("reviewer"), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := base
			s.StageKind = tc.kind
			req := CreateRouteRequest{ProfileCode: strp("sop"), Name: "SOP", Stages: []StageRequest{s}}
			err := req.Validate()
			if tc.wantErr && err == nil {
				t.Fatalf("stage_kind %q: want validation error, got nil", tc.kind)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("stage_kind %q: want nil, got %v", tc.kind, err)
			}
		})
	}
}
