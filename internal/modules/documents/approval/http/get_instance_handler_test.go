package approvalhttp

import (
	"testing"
	"time"

	"metaldocs/internal/modules/documents/approval/domain"
)

func TestBuildStageActors_ExpandsSignedAndPendingActors(t *testing.T) {
	approved, err := domain.NewSignoff(domain.SignoffParams{
		ID:                       "sig-1",
		ApprovalInstanceID:       "inst-1",
		StageInstanceID:          "stage-1",
		ActorUserID:              "user-1",
		ActorTenantID:            "tenant-1",
		Decision:                 domain.DecisionApprove,
		SignedAt:                 time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		SignatureMethod:          "password_reauth",
		ContentHash:              "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		ActorDisplayNameSnapshot: "Maria Souza",
	})
	if err != nil {
		t.Fatalf("new signoff: %v", err)
	}

	stage := domain.StageInstance{
		ID:               "stage-1",
		NameSnapshot:     "Qualidade",
		Status:           domain.StageActive,
		EligibleActorIDs: []string{"user-1", "user-2"},
		Signoffs:         []*domain.Signoff{approved},
	}

	actors := buildStageActors(stage, map[string]string{
		"user-1": "Maria Souza",
		"user-2": "Joao Lima",
	})

	if len(actors) != 2 {
		t.Fatalf("len(actors) = %d, want 2", len(actors))
	}
	if actors[0].DisplayName != "Maria Souza" || actors[0].Status != "approved" {
		t.Fatalf("first actor = %+v", actors[0])
	}
	if actors[1].DisplayName != "Joao Lima" || actors[1].Status != "active" {
		t.Fatalf("second actor = %+v", actors[1])
	}
}
