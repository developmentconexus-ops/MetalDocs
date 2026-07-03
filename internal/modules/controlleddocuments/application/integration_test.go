//go:build integration

package application

import (
	"errors"
	"testing"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
)

// TestCrossProfileOverride_Rejected pins a real domain invariant: an override
// template belonging to a different profile than the resolution input must be
// rejected with ErrTemplateProfileMismatch. Does not require a live DB; the
// integration build tag is kept only to match this file's existing scope.
func TestCrossProfileOverride_Rejected(t *testing.T) {
	input := controlleddocumentsdomain.TemplateResolutionInput{
		ProfileCode: "po",
		OverrideTemplate: &controlleddocumentsdomain.TemplateVersionCandidate{
			ID:          "some-uuid",
			ProfileCode: "it",
			Status:      func() *string { s := "published"; return &s }(),
		},
	}
	_, err := controlleddocumentsdomain.Resolve(input)
	if !errors.Is(err, controlleddocumentsdomain.ErrTemplateProfileMismatch) {
		t.Fatalf("expected ErrTemplateProfileMismatch, got %v", err)
	}
}
