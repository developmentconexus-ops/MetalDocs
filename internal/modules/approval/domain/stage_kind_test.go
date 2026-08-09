package domain

import (
	"errors"
	"testing"
)

// TestStageKindValues asserts the two spec-fixed enum values exist with the
// exact string representations the DB CHECK constraint (migration 0286) uses.
func TestStageKindValues(t *testing.T) {
	if StageKindReview != "review" {
		t.Fatalf("StageKindReview = %q, want %q", StageKindReview, "review")
	}
	if StageKindApproval != "approval" {
		t.Fatalf("StageKindApproval = %q, want %q", StageKindApproval, "approval")
	}
}

// TestStageKindValidate asserts Validate() accepts exactly the two spec values
// and rejects everything else with the sentinel ErrInvalidStageKind.
func TestStageKindValidate(t *testing.T) {
	valid := []StageKind{StageKindReview, StageKindApproval}
	for _, k := range valid {
		if err := k.Validate(); err != nil {
			t.Errorf("StageKind(%q).Validate() = %v, want nil", k, err)
		}
	}

	invalid := []StageKind{"", "signature", "REVIEW", "Approval", "approve"}
	for _, k := range invalid {
		err := k.Validate()
		if err == nil {
			t.Errorf("StageKind(%q).Validate() = nil, want ErrInvalidStageKind", k)
			continue
		}
		if !errors.Is(err, ErrInvalidStageKind) {
			t.Errorf("StageKind(%q).Validate() = %v, want ErrInvalidStageKind", k, err)
		}
	}
}
