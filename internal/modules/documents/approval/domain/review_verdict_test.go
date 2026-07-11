package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewVerdict_Ready(t *testing.T) {
	v, err := NewVerdict(VerdictParams{
		ID:                 "v1",
		ApprovalInstanceID: "inst-1",
		StageInstanceID:    "s1",
		StageKind:          StageKindReview,
		ActorUserID:        "u1",
		ActorTenantID:      "t1",
		Verdict:            VerdictReady,
		VerdictAt:          time.Now(),
	})
	if err != nil {
		t.Fatalf("NewVerdict: %v", err)
	}
	if v.Verdict() != VerdictReady {
		t.Errorf("want VerdictReady; got %s", v.Verdict())
	}
}

func TestNewVerdict_RequestChangesRequiresComment(t *testing.T) {
	_, err := NewVerdict(VerdictParams{
		ID:                 "v1",
		ApprovalInstanceID: "inst-1",
		StageInstanceID:    "s1",
		StageKind:          StageKindReview,
		ActorUserID:        "u1",
		ActorTenantID:      "t1",
		Verdict:            VerdictRequestChanges,
		Comment:            "",
		VerdictAt:          time.Now(),
	})
	if !errors.Is(err, ErrVerdictCommentRequired) {
		t.Errorf("want ErrVerdictCommentRequired; got %v", err)
	}
}

func TestNewVerdict_RequestChangesWithComment(t *testing.T) {
	v, err := NewVerdict(VerdictParams{
		ID:                 "v1",
		ApprovalInstanceID: "inst-1",
		StageInstanceID:    "s1",
		StageKind:          StageKindReview,
		ActorUserID:        "u1",
		ActorTenantID:      "t1",
		Verdict:            VerdictRequestChanges,
		Comment:            "please fix section 3",
		VerdictAt:          time.Now(),
	})
	if err != nil {
		t.Fatalf("NewVerdict: %v", err)
	}
	if v.Comment() != "please fix section 3" {
		t.Errorf("want comment preserved; got %q", v.Comment())
	}
}

func TestNewVerdict_ReadyOnApprovalStageRejected(t *testing.T) {
	_, err := NewVerdict(VerdictParams{
		ID:                 "v1",
		ApprovalInstanceID: "inst-1",
		StageInstanceID:    "s1",
		StageKind:          StageKindApproval,
		ActorUserID:        "u1",
		ActorTenantID:      "t1",
		Verdict:            VerdictReady,
		VerdictAt:          time.Now(),
	})
	if !errors.Is(err, ErrVerdictReadyOnApprovalStage) {
		t.Errorf("want ErrVerdictReadyOnApprovalStage; got %v", err)
	}
}

func TestNewVerdict_RequestChangesOnApprovalStageAllowed(t *testing.T) {
	v, err := NewVerdict(VerdictParams{
		ID:                 "v1",
		ApprovalInstanceID: "inst-1",
		StageInstanceID:    "s1",
		StageKind:          StageKindApproval,
		ActorUserID:        "u1",
		ActorTenantID:      "t1",
		Verdict:            VerdictRequestChanges,
		Comment:            "please fix section 3",
		VerdictAt:          time.Now(),
	})
	if err != nil {
		t.Fatalf("NewVerdict: %v", err)
	}
	if v.Verdict() != VerdictRequestChanges {
		t.Errorf("want VerdictRequestChanges; got %s", v.Verdict())
	}
}

func TestNewVerdict_RequestChangesOnApprovalStageWithoutCommentRejected(t *testing.T) {
	_, err := NewVerdict(VerdictParams{
		ID:                 "v1",
		ApprovalInstanceID: "inst-1",
		StageInstanceID:    "s1",
		StageKind:          StageKindApproval,
		ActorUserID:        "u1",
		ActorTenantID:      "t1",
		Verdict:            VerdictRequestChanges,
		Comment:            "",
		VerdictAt:          time.Now(),
	})
	if !errors.Is(err, ErrVerdictCommentRequired) {
		t.Errorf("want ErrVerdictCommentRequired; got %v", err)
	}
}

func TestNewVerdict_UnknownStageKindRejected(t *testing.T) {
	_, err := NewVerdict(VerdictParams{
		ID:                 "v1",
		ApprovalInstanceID: "inst-1",
		StageInstanceID:    "s1",
		StageKind:          StageKind("bogus"),
		ActorUserID:        "u1",
		ActorTenantID:      "t1",
		Verdict:            VerdictReady,
		VerdictAt:          time.Now(),
	})
	if !errors.Is(err, ErrVerdictWrongStageKind) {
		t.Errorf("want ErrVerdictWrongStageKind; got %v", err)
	}
}

func TestNewVerdict_InvalidVerdictValue(t *testing.T) {
	_, err := NewVerdict(VerdictParams{
		ID:                 "v1",
		ApprovalInstanceID: "inst-1",
		StageInstanceID:    "s1",
		StageKind:          StageKindReview,
		ActorUserID:        "u1",
		ActorTenantID:      "t1",
		Verdict:            Verdict("bogus"),
		VerdictAt:          time.Now(),
	})
	if err == nil {
		t.Error("want error for invalid verdict value")
	}
}

func TestNewVerdict_RequiredFields(t *testing.T) {
	base := VerdictParams{
		ID:                 "v1",
		ApprovalInstanceID: "inst-1",
		StageInstanceID:    "s1",
		StageKind:          StageKindReview,
		ActorUserID:        "u1",
		ActorTenantID:      "t1",
		Verdict:            VerdictReady,
		VerdictAt:          time.Now(),
	}

	tests := []struct {
		name    string
		mutate  func(p *VerdictParams)
	}{
		{"missing ID", func(p *VerdictParams) { p.ID = "" }},
		{"missing ApprovalInstanceID", func(p *VerdictParams) { p.ApprovalInstanceID = "" }},
		{"missing StageInstanceID", func(p *VerdictParams) { p.StageInstanceID = "" }},
		{"missing ActorUserID", func(p *VerdictParams) { p.ActorUserID = "" }},
		{"missing ActorTenantID", func(p *VerdictParams) { p.ActorTenantID = "" }},
		{"zero VerdictAt", func(p *VerdictParams) { p.VerdictAt = time.Time{} }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := base
			tt.mutate(&p)
			if _, err := NewVerdict(p); err == nil {
				t.Errorf("%s: want error", tt.name)
			}
		})
	}
}
