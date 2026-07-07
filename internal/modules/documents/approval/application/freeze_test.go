package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/platform/db"
)

// fakeFreezeRepo is a minimal fake covering only the two repository methods
// executeFreeze calls. It embeds no interface — any missing method is a
// compile error, keeping the test honest about executeFreeze's actual surface.
type fakeFreezeRepo struct {
	unresolvedComments    bool
	unresolvedCommentsErr error
	pinWon                bool
	pinErr                error

	// captured call args, asserted by tests
	commentsCalledWithSince time.Time
	pinCalledWithHash       string
	pinCalls                int
	commentsCalls           int
}

func (f *fakeFreezeRepo) HasUnresolvedInstanceComments(_ context.Context, _ db.Tx, _, _ string, since time.Time) (bool, error) {
	f.commentsCalls++
	f.commentsCalledWithSince = since
	return f.unresolvedComments, f.unresolvedCommentsErr
}

func (f *fakeFreezeRepo) PinFrozenHash(_ context.Context, _ db.Tx, _, _, hash string) (bool, error) {
	f.pinCalls++
	f.pinCalledWithHash = hash
	return f.pinWon, f.pinErr
}

func TestExecuteFreeze_IdempotentReentry_NoOps(t *testing.T) {
	hash := "already-frozen-hash"
	inst := &domain.Instance{
		ID:                  "inst-1",
		TenantID:             "tenant-1",
		DocumentID:           "doc-1",
		SubmittedAt:          time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		ContentHashAtSubmit:  "abc123",
		FrozenContentHash:    &hash,
	}
	repo := &fakeFreezeRepo{}

	if err := executeFreeze(context.Background(), nil, repo, inst.TenantID, inst); err != nil {
		t.Fatalf("executeFreeze() error = %v, want nil", err)
	}
	if repo.commentsCalls != 0 {
		t.Fatalf("expected no comment check on already-frozen instance, got %d calls", repo.commentsCalls)
	}
	if repo.pinCalls != 0 {
		t.Fatalf("expected no pin attempt on already-frozen instance, got %d calls", repo.pinCalls)
	}
	if inst.FrozenContentHash == nil || *inst.FrozenContentHash != hash {
		t.Fatalf("hash must remain unchanged, got %v", inst.FrozenContentHash)
	}
}

func TestExecuteFreeze_UnresolvedComments_Blocks(t *testing.T) {
	inst := &domain.Instance{
		ID:                  "inst-2",
		TenantID:            "tenant-1",
		DocumentID:          "doc-2",
		SubmittedAt:         time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		ContentHashAtSubmit: "abc123",
	}
	repo := &fakeFreezeRepo{unresolvedComments: true}

	err := executeFreeze(context.Background(), nil, repo, inst.TenantID, inst)
	if !errors.Is(err, ErrFreezeBlockedByUnresolvedComments) {
		t.Fatalf("expected ErrFreezeBlockedByUnresolvedComments, got %v", err)
	}
	if repo.pinCalls != 0 {
		t.Fatalf("pin must not run when comments block freeze, got %d calls", repo.pinCalls)
	}
	if inst.FrozenContentHash != nil {
		t.Fatalf("FrozenContentHash must stay nil when freeze is blocked, got %v", *inst.FrozenContentHash)
	}
	if !repo.commentsCalledWithSince.Equal(inst.SubmittedAt) {
		t.Fatalf("comment check must scope to instance.SubmittedAt, got %v want %v", repo.commentsCalledWithSince, inst.SubmittedAt)
	}
}

func TestExecuteFreeze_CommentsCheckError_Propagates(t *testing.T) {
	inst := &domain.Instance{
		ID:                  "inst-3",
		TenantID:            "tenant-1",
		DocumentID:          "doc-3",
		SubmittedAt:         time.Now(),
		ContentHashAtSubmit: "abc123",
	}
	wantErr := errors.New("db exploded")
	repo := &fakeFreezeRepo{unresolvedCommentsErr: wantErr}

	err := executeFreeze(context.Background(), nil, repo, inst.TenantID, inst)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected wrapped %v, got %v", wantErr, err)
	}
	if repo.pinCalls != 0 {
		t.Fatalf("pin must not run when comment check errors, got %d calls", repo.pinCalls)
	}
}

func TestExecuteFreeze_CleanPath_PinsSuccessfully(t *testing.T) {
	inst := &domain.Instance{
		ID:                  "inst-4",
		TenantID:            "tenant-1",
		DocumentID:          "doc-4",
		SubmittedAt:         time.Now(),
		ContentHashAtSubmit: "the-canonical-hash",
	}
	repo := &fakeFreezeRepo{pinWon: true}

	if err := executeFreeze(context.Background(), nil, repo, inst.TenantID, inst); err != nil {
		t.Fatalf("executeFreeze() error = %v, want nil", err)
	}
	if repo.pinCalls != 1 {
		t.Fatalf("expected exactly 1 pin call, got %d", repo.pinCalls)
	}
	if repo.pinCalledWithHash != "the-canonical-hash" {
		t.Fatalf("pin must be called with instance.ContentHashAtSubmit, got %q", repo.pinCalledWithHash)
	}
	if inst.FrozenContentHash == nil || *inst.FrozenContentHash != "the-canonical-hash" {
		t.Fatalf("instance.FrozenContentHash must be set in-memory after a won pin, got %v", inst.FrozenContentHash)
	}
}

func TestExecuteFreeze_PinLostRace_StillTreatedAsSuccess(t *testing.T) {
	// Extremely unlikely given the FOR UPDATE row lock already held by the
	// caller, but handled defensively per plan.md task 4: a lost CAS means
	// SOME concurrent freeze already happened, so this is still success, not
	// an error. Since our fake doesn't return the winning hash from the DB,
	// the in-memory instance's FrozenContentHash is set to the same
	// locally-known ContentHashAtSubmit value as a best-effort reflection —
	// callers only rely on the error being nil for the "an active stage may
	// now proceed" decision, not on this in-memory value's exact content in
	// the lost-race case.
	inst := &domain.Instance{
		ID:                  "inst-5",
		TenantID:            "tenant-1",
		DocumentID:          "doc-5",
		SubmittedAt:         time.Now(),
		ContentHashAtSubmit: "abc123",
	}
	repo := &fakeFreezeRepo{pinWon: false}

	if err := executeFreeze(context.Background(), nil, repo, inst.TenantID, inst); err != nil {
		t.Fatalf("executeFreeze() error = %v, want nil (lost race is still success)", err)
	}
	if repo.pinCalls != 1 {
		t.Fatalf("expected exactly 1 pin call, got %d", repo.pinCalls)
	}
}

func TestExecuteFreeze_PinError_Propagates(t *testing.T) {
	inst := &domain.Instance{
		ID:                  "inst-6",
		TenantID:            "tenant-1",
		DocumentID:          "doc-6",
		SubmittedAt:         time.Now(),
		ContentHashAtSubmit: "abc123",
	}
	wantErr := errors.New("pin write failed")
	repo := &fakeFreezeRepo{pinErr: wantErr}

	err := executeFreeze(context.Background(), nil, repo, inst.TenantID, inst)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected wrapped %v, got %v", wantErr, err)
	}
	if inst.FrozenContentHash != nil {
		t.Fatalf("FrozenContentHash must stay nil on pin error, got %v", *inst.FrozenContentHash)
	}
}
