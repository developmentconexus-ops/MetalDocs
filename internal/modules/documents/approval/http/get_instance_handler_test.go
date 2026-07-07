package approvalhttp

import (
	"strings"
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

// TestMapInstanceResponse_RendersSignatureMeaning asserts the F7 manifest
// wiring: each signoff's SignatureMeaning() is rendered in the mapped
// contracts.SignoffRecord.SignatureMeaning field (spec.md §2.1 — "rendered
// in the signature manifest and audit trail"), for both approve and reject.
func TestMapInstanceResponse_RendersSignatureMeaning(t *testing.T) {
	approved, err := domain.NewSignoff(domain.SignoffParams{
		ID:                       "sig-approve",
		ApprovalInstanceID:       "inst-1",
		StageInstanceID:          "stage-1",
		ActorUserID:              "user-1",
		ActorTenantID:            "tenant-1",
		Decision:                 domain.DecisionApprove,
		SignedAt:                 time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		SignatureMethod:          "password_reauth",
		ContentHash:              strings.Repeat("a", 64),
		ActorDisplayNameSnapshot: "Maria Souza",
		SignatureMeaning:         "approval",
	})
	if err != nil {
		t.Fatalf("new signoff (approve): %v", err)
	}
	rejected, err := domain.NewSignoff(domain.SignoffParams{
		ID:                       "sig-reject",
		ApprovalInstanceID:       "inst-1",
		StageInstanceID:          "stage-1",
		ActorUserID:              "user-2",
		ActorTenantID:            "tenant-1",
		Decision:                 domain.DecisionReject,
		SignedAt:                 time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		SignatureMethod:          "password_reauth",
		ContentHash:              strings.Repeat("a", 64),
		ActorDisplayNameSnapshot: "Joao Lima",
		SignatureMeaning:         "rejection",
	})
	if err != nil {
		t.Fatalf("new signoff (reject): %v", err)
	}

	h := &Handler{}
	inst := &domain.Instance{
		ID:          "inst-1",
		DocumentID:  "doc-1",
		RouteID:     "route-1",
		TenantID:    "tenant-1",
		Status:      domain.InstanceInProgress,
		SubmittedBy: "author-1",
		SubmittedAt: time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC),
		Stages: []domain.StageInstance{
			{
				ID:           "stage-1",
				NameSnapshot: "Qualidade",
				Status:       domain.StageActive,
				Signoffs:     []*domain.Signoff{approved, rejected},
			},
		},
	}

	resp, err := h.mapInstanceResponse(t.Context(), "tenant-1", inst)
	if err != nil {
		t.Fatalf("map response: %v", err)
	}
	if len(resp.Stages) != 1 || len(resp.Stages[0].Signoffs) != 2 {
		t.Fatalf("unexpected shape: %+v", resp.Stages)
	}
	got := map[string]string{}
	for _, rec := range resp.Stages[0].Signoffs {
		got[rec.ID] = rec.SignatureMeaning
	}
	if got["sig-approve"] != "approval" {
		t.Errorf("sig-approve SignatureMeaning = %q; want %q", got["sig-approve"], "approval")
	}
	if got["sig-reject"] != "rejection" {
		t.Errorf("sig-reject SignatureMeaning = %q; want %q", got["sig-reject"], "rejection")
	}
}

func TestMapInstanceResponse_UsesRevisionVersionETag(t *testing.T) {
	h := &Handler{}
	inst := &domain.Instance{
		ID:              "inst-1",
		DocumentID:      "doc-1",
		RouteID:         "route-1",
		TenantID:        "tenant-1",
		Status:          domain.InstanceInProgress,
		SubmittedBy:     "actor-1",
		SubmittedAt:     time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC),
		RevisionVersion: 8,
	}

	resp, err := h.mapInstanceResponse(t.Context(), "tenant-1", inst)
	if err != nil {
		t.Fatalf("map response: %v", err)
	}
	if resp.ETag != "\"v8\"" {
		t.Fatalf("etag = %q, want %q", resp.ETag, "\"v8\"")
	}
}
