package approvalhttp

import (
	"context"
	"testing"
	"time"

	"metaldocs/internal/modules/documents/approval/domain"
	iamdomain "metaldocs/internal/modules/iam/domain"
)

// fakeDisplayNameReader is a test double for iamdomain.UserDisplayNameReader.
type fakeDisplayNameReader struct {
	names map[string]string // userID -> displayName; absent/empty means omitted
}

func (f *fakeDisplayNameReader) DisplayName(_ context.Context, _, userID string) (string, error) {
	return f.names[userID], nil
}

func (f *fakeDisplayNameReader) DisplayNames(_ context.Context, _ string, userIDs []string) (map[string]string, error) {
	out := make(map[string]string)
	for _, uid := range userIDs {
		if name, ok := f.names[uid]; ok && name != "" {
			out[uid] = name
		}
	}
	return out, nil
}

var _ iamdomain.UserDisplayNameReader = (*fakeDisplayNameReader)(nil)

// TestResolveEligibleActorNames_UsesFakePortAndMissingFallback verifies that:
// (a) actors whose display_name the port returns are used as-is,
// (b) actors absent from the port result fall back to their own userID
//
// This is the spec's "byte-identical rendered StageActor names" requirement.
func TestResolveEligibleActorNames_UsesFakePortAndMissingFallback(t *testing.T) {
	knownUser := "user-with-name"
	unknownUser := "user-without-name"

	reader := &fakeDisplayNameReader{
		names: map[string]string{
			knownUser: "Known Actor",
			// unknownUser intentionally absent → port omits it → fallback to userID
		},
	}

	h := &Handler{displayNameReader: reader}

	sig, err := domain.NewSignoff(domain.SignoffParams{
		ID:                       "sig-1",
		ApprovalInstanceID:       "inst-1",
		StageInstanceID:          "stage-1",
		ActorUserID:              knownUser,
		ActorTenantID:            "tenant-1",
		Decision:                 domain.DecisionApprove,
		SignedAt:                 time.Now(),
		SignatureMethod:          "password_reauth",
		ContentHash:              "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		ActorDisplayNameSnapshot: "Known Actor",
	})
	if err != nil {
		t.Fatalf("new signoff: %v", err)
	}

	inst := &domain.Instance{
		ID:       "inst-1",
		TenantID: "tenant-1",
		Stages: []domain.StageInstance{
			{
				ID:               "stage-1",
				EligibleActorIDs: []string{knownUser, unknownUser},
				Signoffs:         []*domain.Signoff{sig},
			},
		},
	}

	names, err := h.resolveEligibleActorNames(context.Background(), "tenant-1", inst)
	if err != nil {
		t.Fatalf("resolveEligibleActorNames: %v", err)
	}

	if names[knownUser] != "Known Actor" {
		t.Fatalf("names[%s] = %q, want %q", knownUser, names[knownUser], "Known Actor")
	}
	// Missing-from-port → fallback to userID (the existing post-loop fallback).
	if names[unknownUser] != unknownUser {
		t.Fatalf("names[%s] = %q, want userID fallback %q", unknownUser, names[unknownUser], unknownUser)
	}
}

// TestResolveEligibleActorNames_EmptyInstanceReturnsEmptyMap ensures no nil-map
// panic when the instance has no stages.
func TestResolveEligibleActorNames_EmptyInstanceReturnsEmptyMap(t *testing.T) {
	h := &Handler{displayNameReader: &fakeDisplayNameReader{names: map[string]string{}}}
	inst := &domain.Instance{TenantID: "tenant-1"}
	names, err := h.resolveEligibleActorNames(context.Background(), "tenant-1", inst)
	if err != nil {
		t.Fatalf("resolveEligibleActorNames(empty): %v", err)
	}
	if len(names) != 0 {
		t.Fatalf("expected empty map, got %v", names)
	}
}
