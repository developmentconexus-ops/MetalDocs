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
			name:        "draft to in_review",
			from:        domain.VersionStatusDraft,
			next:        domain.VersionStatusInReview,
			hasReviewer: true,
		},
		{
			name:        "in_review to approved when reviewer required",
			from:        domain.VersionStatusInReview,
			next:        domain.VersionStatusApproved,
			hasReviewer: true,
		},
		{
			name:        "in_review to published when reviewer not required",
			from:        domain.VersionStatusInReview,
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
			name:        "in_review to draft reject",
			from:        domain.VersionStatusInReview,
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
			name:        "in_review to approved denied when reviewer not required",
			from:        domain.VersionStatusInReview,
			next:        domain.VersionStatusApproved,
			hasReviewer: false,
			wantErr:     domain.ErrInvalidStateTransition,
		},
		{
			name:        "in_review to published denied when reviewer required",
			from:        domain.VersionStatusInReview,
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
	v := domain.NewTemplateVersionDraft("ver-1", "tpl-1", "user-1", "templates/tpl-1/versions/1.docx", 1, metadata, placeholders, createdAt)

	if v.Status != domain.VersionStatusDraft {
		t.Fatalf("expected draft status, got %q", v.Status)
	}
	if v.ContentHash != "" {
		t.Fatalf("expected empty content hash, got %q", v.ContentHash)
	}
	if v.AuthorID != "user-1" || v.TemplateID != "tpl-1" || v.ID != "ver-1" {
		t.Fatalf("unexpected identity fields: %+v", v)
	}
}

func TestRoleBindingFor(t *testing.T) {
	reviewer := "reviewer-role"
	cases := []struct {
		name       string
		version    domain.TemplateVersion
		transition domain.VersionStatus
		want       string
	}{
		{"published_returns_approver_when_set", domain.TemplateVersion{PendingApproverRole: "approver-role"}, domain.VersionStatusPublished, "approver-role"},
		{"published_empty_when_unset", domain.TemplateVersion{}, domain.VersionStatusPublished, ""},
		{"in_review_returns_reviewer_when_set", domain.TemplateVersion{PendingReviewerRole: &reviewer}, domain.VersionStatusInReview, reviewer},
		{"in_review_empty_when_nil", domain.TemplateVersion{}, domain.VersionStatusInReview, ""},
		{"approved_returns_reviewer_when_set", domain.TemplateVersion{PendingReviewerRole: &reviewer}, domain.VersionStatusApproved, reviewer},
		{"draft_has_no_binding", domain.TemplateVersion{PendingApproverRole: "x", PendingReviewerRole: &reviewer}, domain.VersionStatusDraft, ""},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := tc.version.RoleBindingFor(tc.transition)
			if got != tc.want {
				t.Fatalf("RoleBindingFor(%q) = %q, want %q", tc.transition, got, tc.want)
			}
		})
	}
}
