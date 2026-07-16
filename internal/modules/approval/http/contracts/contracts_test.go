package contracts

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestSubmitRequestReasonField pins the F6.3 structured reason-for-change
// fields on the submit-for-review request contract (validation-contract.md
// §5.1/§5.4): reason_for_change (+ optional reason_category) round-trip
// through JSON as *string, distinguishing "absent" from an explicit "".
// REV>=1-requiredness is NOT enforced here (contracts.SubmitRequest.Validate
// does not know the governed revision number) — that gate lives in
// application.SubmitRevisionForReview (submit_service_test.go).
func TestSubmitRequestReasonField(t *testing.T) {
	reason := "Corrected per audit finding #12"
	category := "corrective"
	body := []byte(`{"route_id":"3fa85f64-5717-4562-b3fc-2c963f66afa6","content_hash":"` + strings.Repeat("a", 64) + `","reason_for_change":"` + reason + `","reason_category":"` + category + `"}`)

	var req SubmitRequest
	if err := json.Unmarshal(body, &req); err != nil {
		t.Fatalf("unmarshal SubmitRequest: %v", err)
	}
	if req.ReasonForChange == nil || *req.ReasonForChange != reason {
		t.Fatalf("ReasonForChange = %v, want %q", req.ReasonForChange, reason)
	}
	if req.ReasonCategory == nil || *req.ReasonCategory != category {
		t.Fatalf("ReasonCategory = %v, want %q", req.ReasonCategory, category)
	}
	if err := req.Validate(); err != nil {
		t.Fatalf("expected valid request with reason fields, got error: %v", err)
	}

	// Absent fields must decode to nil, not a zero-value "" pointer — the
	// handler distinguishes "omitted" from "explicit empty string".
	absentBody := []byte(`{"route_id":"3fa85f64-5717-4562-b3fc-2c963f66afa6","content_hash":"` + strings.Repeat("b", 64) + `"}`)
	var absent SubmitRequest
	if err := json.Unmarshal(absentBody, &absent); err != nil {
		t.Fatalf("unmarshal SubmitRequest (absent reason): %v", err)
	}
	if absent.ReasonForChange != nil {
		t.Fatalf("ReasonForChange = %v, want nil when omitted from the body", absent.ReasonForChange)
	}
	if absent.ReasonCategory != nil {
		t.Fatalf("ReasonCategory = %v, want nil when omitted from the body", absent.ReasonCategory)
	}
	if err := absent.Validate(); err != nil {
		t.Fatalf("expected valid request with reason fields omitted (contract-level Validate does not gate REV>=1), got error: %v", err)
	}
}

func TestSubmitRequestValidate(t *testing.T) {
	valid := SubmitRequest{RouteID: "3fa85f64-5717-4562-b3fc-2c963f66afa6", ContentHash: strings.Repeat("a", 64)}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid request, got error: %v", err)
	}

	// ADR 0073: route_id and content_hash are optional on the wire — the canonical
	// /submit handler resolves them in-tx when the client omits them. Validate is
	// format-only: an omitted route_id (and omitted content_hash) must pass.
	omittedGovernance := SubmitRequest{}
	if err := omittedGovernance.Validate(); err != nil {
		t.Fatalf("expected valid request with route_id/content_hash omitted (ADR 0073 in-tx resolution), got error: %v", err)
	}

	// When present, route_id must still be a well-formed UUID.
	badRoute := SubmitRequest{RouteID: "not-a-uuid", ContentHash: strings.Repeat("a", 64)}
	if err := badRoute.Validate(); err == nil {
		t.Fatalf("expected error for malformed route_id")
	}

	// When present, content_hash must still be 64-hex.
	badHash := SubmitRequest{RouteID: "3fa85f64-5717-4562-b3fc-2c963f66afa6", ContentHash: "abc"}
	if err := badHash.Validate(); err == nil {
		t.Fatalf("expected error for invalid content_hash")
	}
}

func TestSignoffRequestValidate(t *testing.T) {
	valid := SignoffRequest{
		Decision:      "approve",
		PasswordToken: "token",
		ContentHash:   strings.Repeat("b", 64),
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid request, got error: %v", err)
	}

	invalidDecision := SignoffRequest{Decision: "maybe", PasswordToken: "token", ContentHash: strings.Repeat("b", 64)}
	if err := invalidDecision.Validate(); err == nil {
		t.Fatalf("expected error for invalid decision")
	}

	missingRejectReason := SignoffRequest{Decision: "reject", PasswordToken: "token", ContentHash: strings.Repeat("b", 64)}
	if err := missingRejectReason.Validate(); err == nil {
		t.Fatalf("expected error for missing reason on reject")
	}

	missingPassword := SignoffRequest{Decision: "approve", ContentHash: strings.Repeat("b", 64)}
	if err := missingPassword.Validate(); err == nil {
		t.Fatalf("expected error for missing password_token")
	}
}

func TestSchedulePublishRequestValidate(t *testing.T) {
	valid := SchedulePublishRequest{EffectiveFrom: "2026-12-31T18:00:00Z"}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid request, got error: %v", err)
	}

	validWithSupersededID := SchedulePublishRequest{
		EffectiveFrom:        "2026-12-31T18:00:00Z",
		SupersededDocumentID: "3fa85f64-5717-4562-b3fc-2c963f66afa6",
	}
	if err := validWithSupersededID.Validate(); err != nil {
		t.Fatalf("expected valid request with superseded_document_id, got error: %v", err)
	}

	missing := SchedulePublishRequest{}
	if err := missing.Validate(); err == nil {
		t.Fatalf("expected error for missing effective_from")
	}

	invalid := SchedulePublishRequest{EffectiveFrom: "31-12-2026"}
	if err := invalid.Validate(); err == nil {
		t.Fatalf("expected error for invalid effective_from")
	}

	nonUTC := SchedulePublishRequest{EffectiveFrom: "2026-12-31T18:00:00-03:00"}
	if err := nonUTC.Validate(); err == nil {
		t.Fatalf("expected error for non-UTC effective_from")
	}

	invalidSupersededID := SchedulePublishRequest{
		EffectiveFrom:        "2026-12-31T18:00:00Z",
		SupersededDocumentID: "not-a-uuid",
	}
	if err := invalidSupersededID.Validate(); err == nil {
		t.Fatalf("expected error for malformed superseded_document_id")
	}
}

func TestSupersedeRequestValidate(t *testing.T) {
	valid := SupersedeRequest{SupersededDocumentID: "3fa85f64-5717-4562-b3fc-2c963f66afa6"}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid request, got error: %v", err)
	}

	missing := SupersedeRequest{}
	if err := missing.Validate(); err == nil {
		t.Fatalf("expected error for missing superseded_document_id")
	}
}

func TestObsoleteRequestValidate(t *testing.T) {
	valid := ObsoleteRequest{Reason: "Replaced by newer standard"}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid request, got error: %v", err)
	}

	missing := ObsoleteRequest{}
	if err := missing.Validate(); err == nil {
		t.Fatalf("expected error for missing reason")
	}
}

func TestCancelRequestValidate(t *testing.T) {
	valid := CancelRequest{Reason: "Withdrawn"}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid request, got error: %v", err)
	}

	missing := CancelRequest{}
	if err := missing.Validate(); err == nil {
		t.Fatalf("expected error for missing reason")
	}
}

func TestCreateRouteRequestValidate(t *testing.T) {
	m := 2
	valid := CreateRouteRequest{
		ProfileCode: "ops",
		Name:        "Default Route",
		Stages: []StageRequest{
			{
				Order:              1,
				Name:               "Review",
				RequiredCapability: "document.signoff",
				Quorum:             "any_1_of",
				DriftPolicy:        "reduce_quorum",
				Selectors:          []ActorSelector{{Kind: SelectorKindRoleInFixedArea, Role: "approver", AreaCode: "ops"}},
			},
			{
				Order:              2,
				Name:               "Approval",
				RequiredCapability: "document.signoff",
				Quorum:             "m_of_n",
				QuorumM:            &m,
				DriftPolicy:        "keep_snapshot",
				Selectors:          []ActorSelector{{Kind: SelectorKindRoleInFixedArea, Role: "approver", AreaCode: "ops"}},
			},
		},
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid request, got error: %v", err)
	}

	missingStages := CreateRouteRequest{ProfileCode: "ops", Name: "x"}
	if err := missingStages.Validate(); err == nil {
		t.Fatalf("expected error for empty stages")
	}

	badQuorum := valid
	badQuorum.Stages[0].Quorum = "half"
	if err := badQuorum.Validate(); err == nil {
		t.Fatalf("expected error for invalid quorum")
	}

	badOrder := valid
	badOrder.Stages[0].Order = 2
	if err := badOrder.Validate(); err == nil {
		t.Fatalf("expected error for invalid order")
	}

	badCapability := valid
	badCapability.Stages[0].RequiredCapability = "doc_signoff"
	if err := badCapability.Validate(); err == nil {
		t.Fatalf("expected error for non-namespaced required_capability")
	}

	// Registry binding: a well-formed but UNREGISTERED capability must be
	// rejected at config time (no silent phantom), mirroring required_role.
	phantomCapability := valid
	phantomCapability.Stages[0].RequiredCapability = "document.notarealcap"
	if err := phantomCapability.Validate(); err == nil {
		t.Fatalf("expected error for required_capability that is not a registered capability")
	}
}

// TestCreateRouteRequestValidate_ProfileCodeSubjectKindRule pins the S1 (F18)
// contract fix: profile_code presence is conditional on subject_kind, mirroring
// DB truth (migration 0297 approval_routes_template_subject_projection_check,
// ADR 0082). A template route has no profile — carrying a non-empty
// profile_code must fail validation (not be silently accepted and later
// dropped); a document (or absent-kind) route still requires profile_code.
func TestCreateRouteRequestValidate_ProfileCodeSubjectKindRule(t *testing.T) {
	stages := func() []StageRequest {
		return []StageRequest{
			{
				Order:              1,
				Name:               "Review",
				RequiredCapability: "document.signoff",
				Quorum:             "any_1_of",
				DriftPolicy:        "reduce_quorum",
				Selectors:          []ActorSelector{{Kind: SelectorKindRoleInFixedArea, Role: "approver", AreaCode: "ops"}},
			},
		}
	}

	// (b) template + non-empty profile_code must fail.
	templateWithProfileCode := CreateRouteRequest{
		ProfileCode: "ops",
		Name:        "Template Route",
		SubjectKind: "template",
		SubjectKey:  "tmpl-1",
		Stages:      stages(),
	}
	if err := templateWithProfileCode.Validate(); err == nil {
		t.Fatalf("expected error: profile_code must be absent for template routes")
	}

	// (b-corollary) template + empty profile_code must succeed.
	templateNoProfileCode := CreateRouteRequest{
		Name:        "Template Route",
		SubjectKind: "template",
		SubjectKey:  "tmpl-1",
		Stages:      stages(),
	}
	if err := templateNoProfileCode.Validate(); err != nil {
		t.Fatalf("expected template route with no profile_code to be valid, got: %v", err)
	}

	// (c) document (absent subject_kind) + empty profile_code must fail.
	absentKindNoProfileCode := CreateRouteRequest{
		Name:   "Doc Route",
		Stages: stages(),
	}
	if err := absentKindNoProfileCode.Validate(); err == nil {
		t.Fatalf("expected error: profile_code required when subject_kind is absent")
	}

	// (c) document (explicit subject_kind) + empty profile_code must fail.
	documentKindNoProfileCode := CreateRouteRequest{
		Name:        "Doc Route",
		SubjectKind: "document",
		Stages:      stages(),
	}
	if err := documentKindNoProfileCode.Validate(); err == nil {
		t.Fatalf("expected error: profile_code required when subject_kind is document")
	}
}

// TestStageSelectorRoleBoundToRegistry locks the ADR 0022 binding: a stage
// selector's role (role_in_fixed_area, role_in_document_area, submit_choice)
// must be a canonical AREA role. A decommissioned phantom ("reviewer") and a
// valid-but-non-area role ("system_admin", which passes the [a-z0-9_-]+
// format check) must both be rejected at config time, so an admin can never
// create a silently-unsatisfiable stage. Builds fresh requests to avoid the
// shared-slice aliasing in the sibling tests. Unit 3.2 slice 7a re-homes this
// binding from the removed flat required_role field onto sel.Role.
func TestStageSelectorRoleBoundToRegistry(t *testing.T) {
	build := func(role string) CreateRouteRequest {
		return CreateRouteRequest{
			ProfileCode: "ops",
			Name:        "Route",
			Stages: []StageRequest{{
				Order:              1,
				Name:               "Review",
				RequiredCapability: "document.signoff",
				Quorum:             "any_1_of",
				DriftPolicy:        "reduce_quorum",
				Selectors:          []ActorSelector{{Kind: SelectorKindRoleInFixedArea, Role: role, AreaCode: "ops"}},
			}},
		}
	}
	for _, bad := range []string{"reviewer", "system_admin"} {
		req := build(bad)
		if err := req.Validate(); err == nil {
			t.Fatalf("create: selector role %q must be rejected (not a canonical area role)", bad)
		}
		upd := UpdateRouteRequest{Name: "Route", Stages: req.Stages}
		if err := upd.Validate(); err == nil {
			t.Fatalf("update: selector role %q must be rejected", bad)
		}
	}
	for _, good := range []string{"viewer", "editor", "approver", "author", "signer", "area_admin", "qms_admin"} {
		if err := build(good).Validate(); err != nil {
			t.Fatalf("selector role %q must be accepted: %v", good, err)
		}
	}
}

func TestUpdateRouteRequestValidate(t *testing.T) {
	valid := UpdateRouteRequest{
		Name: "Updated Route",
		Stages: []StageRequest{
			{
				Order:              1,
				Name:               "Review",
				RequiredCapability: "document.signoff",
				Quorum:             "all_of",
				DriftPolicy:        "fail_stage",
				Selectors:          []ActorSelector{{Kind: SelectorKindRoleInFixedArea, Role: "approver", AreaCode: "ops"}},
			},
		},
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid request, got error: %v", err)
	}

	missingName := valid
	missingName.Name = ""
	if err := missingName.Validate(); err == nil {
		t.Fatalf("expected error for missing name")
	}

	badDrift := valid
	badDrift.Stages[0].DriftPolicy = "ignore"
	if err := badDrift.Validate(); err == nil {
		t.Fatalf("expected error for invalid drift_policy")
	}
}
