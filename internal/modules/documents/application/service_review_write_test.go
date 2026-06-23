package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"metaldocs/internal/modules/documents/application"
	"metaldocs/internal/modules/documents/domain"
)

// fakeEligibilityReader is a test double for domain.ApproverEligibilityReader.
type fakeEligibilityReader struct {
	eligible bool
	err      error
}

func (f *fakeEligibilityReader) IsEligibleApprover(_ context.Context, _, _, _ string) (bool, error) {
	return f.eligible, f.err
}

// newSvcWithEligibility builds a Service wired with a fakeEligibilityReader
// and a minimal presigner/audit so the autosave paths can run.
func newSvcWithEligibility(repo *fakeRepo, eligibility *fakeEligibilityReader) *application.Service {
	return application.New(repo, &fakePresigner{
		hashReturn: "h_ok",
		sizeReturn: 512,
	}, nil, nil, &noopAudit{}).WithEligibility(eligibility)
}

// basePendingMeta is a valid PendingCommitMeta for CommitAutosave tests.
var basePendingMetaForReview = &application.PendingCommitMeta{
	SessionID:           "sess_r",
	DocumentID:          "doc_r",
	BaseRevisionID:      "rev_base",
	ExpectedContentHash: "h_ok",
	StorageKey:          "documents/doc_r/pending/p_r.docx",
	ExpiresAt:           time.Now().Add(10 * time.Minute),
}

// --- PresignAutosave matrix ---

func TestPresignAutosave_ReviewWrite(t *testing.T) {
	tests := []struct {
		name       string
		status     domain.DocumentStatus
		isOwner    bool
		isEligible bool
		wantErr    error
	}{
		{
			name:    "draft_owner_ok",
			status:  domain.DocStatusDraft,
			isOwner: true,
			wantErr: nil,
		},
		{
			name:    "draft_not_owner_rejected",
			status:  domain.DocStatusDraft,
			isOwner: false,
			wantErr: domain.ErrInvalidStateTransition,
		},
		{
			name:       "under_review_eligible_approver_ok",
			status:     domain.DocStatusUnderReview,
			isEligible: true,
			wantErr:    nil,
		},
		{
			name:       "under_review_not_eligible_rejected",
			status:     domain.DocStatusUnderReview,
			isEligible: false,
			wantErr:    domain.ErrInvalidStateTransition,
		},
		{
			name:    "approved_rejected",
			status:  domain.DocStatusApproved,
			isOwner: true, // ownership doesn't matter for non-draft/non-review
			wantErr: domain.ErrInvalidStateTransition,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeRepo{
				docReturn:   &domain.Document{ID: "doc_r", TenantID: "tenant_r", Status: tc.status},
				ownerReturn: tc.isOwner,
			}
			eligibility := &fakeEligibilityReader{eligible: tc.isEligible}
			svc := newSvcWithEligibility(repo, eligibility)

			_, err := svc.PresignAutosave(context.Background(), application.PresignAutosaveCmd{
				TenantID:       "tenant_r",
				ActorUserID:    "user_r",
				DocumentID:     "doc_r",
				SessionID:      "sess_r",
				BaseRevisionID: "rev_base",
				ContentHash:    "h_ok",
			})

			if tc.wantErr == nil {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
			} else if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected %v, got %v", tc.wantErr, err)
			}
		})
	}
}

// --- CommitAutosave matrix ---

func TestCommitAutosave_ReviewWrite(t *testing.T) {
	tests := []struct {
		name       string
		status     domain.DocumentStatus
		isOwner    bool
		isEligible bool
		wantErr    error
	}{
		{
			name:    "draft_owner_ok",
			status:  domain.DocStatusDraft,
			isOwner: true,
			wantErr: nil,
		},
		{
			name:    "draft_not_owner_rejected",
			status:  domain.DocStatusDraft,
			isOwner: false,
			wantErr: domain.ErrInvalidStateTransition,
		},
		{
			name:       "under_review_eligible_approver_ok",
			status:     domain.DocStatusUnderReview,
			isEligible: true,
			wantErr:    nil,
		},
		{
			name:       "under_review_not_eligible_rejected",
			status:     domain.DocStatusUnderReview,
			isEligible: false,
			wantErr:    domain.ErrInvalidStateTransition,
		},
		{
			name:    "approved_rejected",
			status:  domain.DocStatusApproved,
			isOwner: true,
			wantErr: domain.ErrInvalidStateTransition,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeRepo{
				docReturn:   &domain.Document{ID: "doc_r", TenantID: "tenant_r", Status: tc.status},
				ownerReturn: tc.isOwner,
				pendingMeta: &application.PendingCommitMeta{
					SessionID:           basePendingMetaForReview.SessionID,
					DocumentID:          basePendingMetaForReview.DocumentID,
					BaseRevisionID:      basePendingMetaForReview.BaseRevisionID,
					ExpectedContentHash: basePendingMetaForReview.ExpectedContentHash,
					StorageKey:          basePendingMetaForReview.StorageKey,
					ExpiresAt:           basePendingMetaForReview.ExpiresAt,
				},
				commitResult: &application.CommitResult{RevisionID: "rev_r", RevisionNum: 1},
			}
			eligibility := &fakeEligibilityReader{eligible: tc.isEligible}
			svc := newSvcWithEligibility(repo, eligibility)

			_, err := svc.CommitAutosave(context.Background(), application.CommitAutosaveCmd{
				TenantID:         "tenant_r",
				ActorUserID:      "user_r",
				DocumentID:       "doc_r",
				SessionID:        "sess_r",
				PendingUploadID:  "pending_r",
				FormDataSnapshot: []byte(`{}`),
			})

			if tc.wantErr == nil {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
			} else if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected %v, got %v", tc.wantErr, err)
			}
		})
	}
}

// --- AcquireSession matrix ---

func TestAcquireSession_ReviewWrite(t *testing.T) {
	tests := []struct {
		name       string
		status     domain.DocumentStatus
		isOwner    bool
		isEligible bool
		wantErr    error
	}{
		{
			name:    "draft_owner_ok",
			status:  domain.DocStatusDraft,
			isOwner: true,
			wantErr: nil,
		},
		{
			name:    "draft_not_owner_rejected",
			status:  domain.DocStatusDraft,
			isOwner: false,
			wantErr: domain.ErrInvalidStateTransition,
		},
		{
			name:       "under_review_eligible_approver_ok",
			status:     domain.DocStatusUnderReview,
			isEligible: true,
			wantErr:    nil,
		},
		{
			name:       "under_review_not_eligible_rejected",
			status:     domain.DocStatusUnderReview,
			isEligible: false,
			wantErr:    domain.ErrInvalidStateTransition,
		},
		{
			name:    "approved_rejected",
			status:  domain.DocStatusApproved,
			isOwner: true,
			wantErr: domain.ErrInvalidStateTransition,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeRepo{
				docReturn:   &domain.Document{ID: "doc_r", TenantID: "tenant_r", Status: tc.status},
				ownerReturn: tc.isOwner,
				acquireSess: &domain.Session{ID: "sess_r", DocumentID: "doc_r", UserID: "user_r"},
			}
			eligibility := &fakeEligibilityReader{eligible: tc.isEligible}
			svc := newSvcWithEligibility(repo, eligibility)

			_, _, err := svc.AcquireSession(context.Background(), "tenant_r", "doc_r", "user_r")

			if tc.wantErr == nil {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
			} else if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected %v, got %v", tc.wantErr, err)
			}
		})
	}
}
