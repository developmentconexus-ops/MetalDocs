//go:build integration

package application

import (
	"errors"
	"testing"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
)

func TestFullFlow_CreateProfile_SetTemplate_CreateCD_CreateDocument(t *testing.T) {
	t.Skip("requires live DB")
}

func TestBackfill_SeedLegacyDocs_RunBackfill_AssertAllLinked(t *testing.T) {
	t.Skip("requires live DB")
}

func TestBackfill_ReRunIsNoop(t *testing.T) {
	t.Skip("requires live DB")
}

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

func TestRename_Flow(t *testing.T) {
	t.Skip("requires live DB")
}

