package domain_test

import (
	"errors"
	"testing"
	"time"

	"metaldocs/internal/modules/templates/domain"
)

func TestCanTransition(t *testing.T) {
	tests := []struct {
		name        string
		from        domain.VersionStatus
		next        domain.VersionStatus
		hasReviewer bool
		wantErr     error
	}{
		{
			name:        "draft to under_review",
			from:        domain.VersionStatusDraft,
			next:        domain.VersionStatusUnderReview,
			hasReviewer: true,
		},
		{
			name:        "under_review to approved when reviewer required",
			from:        domain.VersionStatusUnderReview,
			next:        domain.VersionStatusApproved,
			hasReviewer: true,
		},
		{
			name:        "under_review to published when reviewer not required",
			from:        domain.VersionStatusUnderReview,
			next:        domain.VersionStatusPublished,
			hasReviewer: false,
		},
		{
			name:        "approved to published",
			from:        domain.VersionStatusApproved,
			next:        domain.VersionStatusPublished,
			hasReviewer: true,
		},
		{
			name:        "published to obsolete",
			from:        domain.VersionStatusPublished,
			next:        domain.VersionStatusObsolete,
			hasReviewer: true,
		},
		{
			name:        "under_review to draft reject",
			from:        domain.VersionStatusUnderReview,
			next:        domain.VersionStatusDraft,
			hasReviewer: true,
		},
		{
			name:        "approved to draft reject",
			from:        domain.VersionStatusApproved,
			next:        domain.VersionStatusDraft,
			hasReviewer: true,
		},
		{
			name:        "under_review to approved denied when reviewer not required",
			from:        domain.VersionStatusUnderReview,
			next:        domain.VersionStatusApproved,
			hasReviewer: false,
			wantErr:     domain.ErrInvalidStateTransition,
		},
		{
			name:        "under_review to published denied when reviewer required",
			from:        domain.VersionStatusUnderReview,
			next:        domain.VersionStatusPublished,
			hasReviewer: true,
			wantErr:     domain.ErrInvalidStateTransition,
		},
		{
			name:        "draft to published invalid",
			from:        domain.VersionStatusDraft,
			next:        domain.VersionStatusPublished,
			hasReviewer: true,
			wantErr:     domain.ErrInvalidStateTransition,
		},
		{
			name:        "draft to approved invalid",
			from:        domain.VersionStatusDraft,
			next:        domain.VersionStatusApproved,
			hasReviewer: true,
			wantErr:     domain.ErrInvalidStateTransition,
		},
		{
			name:        "draft to obsolete invalid",
			from:        domain.VersionStatusDraft,
			next:        domain.VersionStatusObsolete,
			hasReviewer: true,
			wantErr:     domain.ErrInvalidStateTransition,
		},
		{
			name:        "approved to obsolete invalid",
			from:        domain.VersionStatusApproved,
			next:        domain.VersionStatusObsolete,
			hasReviewer: true,
			wantErr:     domain.ErrInvalidStateTransition,
		},
		{
			name:        "published to draft invalid",
			from:        domain.VersionStatusPublished,
			next:        domain.VersionStatusDraft,
			hasReviewer: true,
			wantErr:     domain.ErrInvalidStateTransition,
		},
		{
			name:        "obsolete to draft invalid",
			from:        domain.VersionStatusObsolete,
			next:        domain.VersionStatusDraft,
			hasReviewer: true,
			wantErr:     domain.ErrInvalidStateTransition,
		},
		{
			name:        "obsolete to published invalid",
			from:        domain.VersionStatusObsolete,
			next:        domain.VersionStatusPublished,
			hasReviewer: true,
			wantErr:     domain.ErrInvalidStateTransition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := domain.TemplateVersion{Status: tt.from}

			err := v.CanTransition(tt.next, tt.hasReviewer)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CanTransition(%q -> %q, hasReviewer=%v) error = %v, want %v", tt.from, tt.next, tt.hasReviewer, err, tt.wantErr)
			}
		})
	}
}

func TestNewTemplateVersionDraft(t *testing.T) {
	createdAt := time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)
	metadata := domain.MetadataSchema{DocCodePattern: "ABC-###"}
	placeholders := []domain.Placeholder{{ID: "p1", Label: "Name", Type: domain.PHText}}
	const hash = "5cdae1bb25103bbc121cdc696ed11eb09aa22041940f199164ebc302f6923d2e"
	v := domain.NewTemplateVersionDraft("ver-1", "tenant-1", "tpl-1", "user-1", "templates/tpl-1/versions/1.docx", hash, 1, metadata, placeholders, createdAt)

	if v.Status != domain.VersionStatusDraft {
		t.Fatalf("expected draft status, got %q", v.Status)
	}
	// ADR 0088 (inverted assertion): a draft is born WITH the verified hash of
	// the object it points at. The constructor no longer hardcodes "" — the old
	// expectation encoded the very overload the ADR removed.
	if v.ContentHash != hash {
		t.Fatalf("expected content hash %q, got %q", hash, v.ContentHash)
	}
	if v.AuthorID != "user-1" || v.TemplateID != "tpl-1" || v.ID != "ver-1" || v.TenantID != "tenant-1" {
		t.Fatalf("unexpected identity fields: %+v", v)
	}
}
