package domain_test

import (
	"strings"
	"testing"
	"time"

	"metaldocs/internal/modules/approval/domain"
)

func mustSignoff(t *testing.T, stageInstanceID, actorUserID string) *domain.Signoff {
	t.Helper()
	s, err := domain.NewSignoff(domain.SignoffParams{
		ID:                 "signoff-" + stageInstanceID + "-" + actorUserID,
		ApprovalInstanceID: "instance-1",
		StageInstanceID:    stageInstanceID,
		ActorUserID:        actorUserID,
		ActorTenantID:      "tenant-1",
		Decision:           domain.DecisionApprove,
		SignedAt:           time.Now().UTC(),
		ContentHash:        strings.Repeat("0", 64),
	})
	if err != nil {
		t.Fatalf("NewSignoff: %v", err)
	}
	return s
}

func mustDelegation(t *testing.T, delegatorID, delegateID string) domain.Delegation {
	t.Helper()
	now := time.Now().UTC()
	d, err := domain.NewDelegation(domain.DelegationParams{
		ID:          "delegation-" + delegatorID + "-" + delegateID,
		TenantID:    "tenant-1",
		DelegatorID: delegatorID,
		DelegateID:  delegateID,
		StartsAt:    now.Add(-time.Hour),
		EndsAt:      now.Add(time.Hour),
		Reason:      "coverage",
		CreatedBy:   delegatorID,
		CreatedAt:   now.Add(-time.Hour),
	})
	if err != nil {
		t.Fatalf("NewDelegation: %v", err)
	}
	return *d
}

func TestViewerEligibility(t *testing.T) {
	const authorID = "author-1"
	const approverID = "approver-1"
	const delegatorID = "delegator-1"
	const delegateID = "delegate-1"
	const observerID = "observer-1"

	activeStage := func(eligible []string, signoffs ...*domain.Signoff) *domain.StageInstance {
		return &domain.StageInstance{
			ID:               "stage-active",
			Status:           domain.StageActive,
			EligibleActorIDs: eligible,
			Signoffs:         signoffs,
		}
	}

	t.Run("author: is_author true, eligible false by author-rule", func(t *testing.T) {
		active := activeStage([]string{approverID})
		facts := domain.ViewerEligibility(authorID, authorID, active, nil, []domain.StageInstance{*active})

		if !facts.IsAuthor {
			t.Error("expected IsAuthor=true")
		}
		if facts.EligibleForActiveStage {
			t.Error("expected EligibleForActiveStage=false for author")
		}
		if facts.HasSignedActiveStage {
			t.Error("expected HasSignedActiveStage=false")
		}
		if facts.ViaDelegationFromUserID != "" {
			t.Errorf("expected no delegation, got %q", facts.ViaDelegationFromUserID)
		}
	})

	t.Run("snapshot actor on active stage: eligible true, via_delegation none", func(t *testing.T) {
		active := activeStage([]string{approverID})
		facts := domain.ViewerEligibility(approverID, authorID, active, nil, []domain.StageInstance{*active})

		if facts.IsAuthor {
			t.Error("expected IsAuthor=false")
		}
		if !facts.EligibleForActiveStage {
			t.Error("expected EligibleForActiveStage=true")
		}
		if facts.ViaDelegationFromUserID != "" {
			t.Errorf("expected no delegation, got %q", facts.ViaDelegationFromUserID)
		}
		if facts.HasSignedActiveStage {
			t.Error("expected HasSignedActiveStage=false")
		}
	})

	t.Run("delegate: eligible true, via_delegation_from = delegator", func(t *testing.T) {
		active := activeStage([]string{delegatorID})
		delegations := []domain.Delegation{mustDelegation(t, delegatorID, delegateID)}
		facts := domain.ViewerEligibility(delegateID, authorID, active, delegations, []domain.StageInstance{*active})

		if !facts.EligibleForActiveStage {
			t.Error("expected EligibleForActiveStage=true")
		}
		if facts.ViaDelegationFromUserID != delegatorID {
			t.Errorf("expected via_delegation_from=%q, got %q", delegatorID, facts.ViaDelegationFromUserID)
		}
	})

	t.Run("already-signed: eligible true, has_signed true", func(t *testing.T) {
		sig := mustSignoff(t, "stage-active", approverID)
		active := activeStage([]string{approverID}, sig)
		facts := domain.ViewerEligibility(approverID, authorID, active, nil, []domain.StageInstance{*active})

		if !facts.EligibleForActiveStage {
			t.Error("expected EligibleForActiveStage=true")
		}
		if !facts.HasSignedActiveStage {
			t.Error("expected HasSignedActiveStage=true")
		}
	})

	t.Run("observer: not author, not eligible, all false/null", func(t *testing.T) {
		active := activeStage([]string{approverID})
		facts := domain.ViewerEligibility(observerID, authorID, active, nil, []domain.StageInstance{*active})

		if facts.IsAuthor {
			t.Error("expected IsAuthor=false")
		}
		if facts.EligibleForActiveStage {
			t.Error("expected EligibleForActiveStage=false")
		}
		if facts.HasSignedActiveStage {
			t.Error("expected HasSignedActiveStage=false")
		}
		if facts.ViaDelegationFromUserID != "" {
			t.Errorf("expected no delegation, got %q", facts.ViaDelegationFromUserID)
		}
	})

	t.Run("author who is also an approver: excluded by author-rule", func(t *testing.T) {
		active := activeStage([]string{authorID})
		facts := domain.ViewerEligibility(authorID, authorID, active, nil, []domain.StageInstance{*active})

		if !facts.IsAuthor {
			t.Error("expected IsAuthor=true")
		}
		if facts.EligibleForActiveStage {
			t.Error("expected EligibleForActiveStage=false even though author is in the eligible-actor snapshot")
		}
	})

	t.Run("no active stage: all false/null (terminal instance)", func(t *testing.T) {
		facts := domain.ViewerEligibility(approverID, authorID, nil, nil, nil)

		if facts.IsAuthor {
			t.Error("expected IsAuthor=false")
		}
		if facts.EligibleForActiveStage {
			t.Error("expected EligibleForActiveStage=false")
		}
		if facts.HasSignedActiveStage {
			t.Error("expected HasSignedActiveStage=false")
		}
		if facts.ViaDelegationFromUserID != "" {
			t.Errorf("expected no delegation, got %q", facts.ViaDelegationFromUserID)
		}
	})

	t.Run("cross-stage double-sign: eligible in snapshot but already signed earlier stage", func(t *testing.T) {
		priorSig := mustSignoff(t, "stage-prior", approverID)
		priorStage := domain.StageInstance{
			ID:       "stage-prior",
			Status:   domain.StageCompleted,
			Signoffs: []*domain.Signoff{priorSig},
		}
		active := activeStage([]string{approverID})
		facts := domain.ViewerEligibility(approverID, authorID, active, nil, []domain.StageInstance{priorStage, *active})

		if facts.EligibleForActiveStage {
			t.Error("expected EligibleForActiveStage=false due to cross-stage double-sign SoD")
		}
	})
}
