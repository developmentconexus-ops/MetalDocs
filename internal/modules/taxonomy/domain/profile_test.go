package domain

import (
	"errors"
	"testing"
	"time"

	"metaldocs/internal/platform/iamtypes"
)

func TestDocumentProfileIsActiveWhenNotArchived(t *testing.T) {
	p := DocumentProfile{}

	if !p.IsActive() {
		t.Fatal("expected profile to be active when archived_at is nil")
	}
}

func TestDocumentProfileArchiveSetsArchivedAt(t *testing.T) {
	p := DocumentProfile{}
	now := time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC)

	if err := p.Archive(now); err != nil {
		t.Fatalf("unexpected error archiving profile: %v", err)
	}
	if p.ArchivedAt == nil {
		t.Fatal("expected archived_at to be set")
	}
	if !p.ArchivedAt.Equal(now) {
		t.Fatalf("expected archived_at %s, got %s", now, p.ArchivedAt)
	}
}

func TestDocumentProfileArchiveReturnsErrorWhenAlreadyArchived(t *testing.T) {
	archivedAt := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
	p := DocumentProfile{ArchivedAt: &archivedAt}

	err := p.Archive(time.Date(2026, 4, 21, 10, 0, 0, 0, time.UTC))
	if !errors.Is(err, ErrProfileArchived) {
		t.Fatalf("expected ErrProfileArchived, got %v", err)
	}
}

func TestNewDocumentProfile_NormalizesAndDefaultsAlias(t *testing.T) {
	templateID := " tpl-1 "
	ownerID := " owner-1 "
	profile, err := NewDocumentProfile(DocumentProfile{
		Code:                     " PROC-AUDIT-VERY-LONG-CODE ",
		TenantID:                 " tenant-a ",
		FamilyCode:               " compliance ",
		Name:                     " Audit ",
		Description:              " Desc ",
		Alias:                    " ",
		DefaultTemplateVersionID: &templateID,
		OwnerUserID:              &ownerID,
		EditableByRole:           " editor ",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.Code != "PROC-AUDIT-VERY-LONG-CODE" || profile.TenantID != "tenant-a" || profile.FamilyCode != "compliance" || profile.Name != "Audit" || profile.Description != "Desc" {
		t.Fatalf("unexpected normalized profile: %#v", profile)
	}
	if profile.Alias != "PROC-AUDIT-VERY-LONG-COD" {
		t.Fatalf("unexpected alias: %q", profile.Alias)
	}
	if profile.DefaultTemplateVersionID == nil || *profile.DefaultTemplateVersionID != "tpl-1" {
		t.Fatalf("unexpected default template version id: %#v", profile.DefaultTemplateVersionID)
	}
	if profile.OwnerUserID == nil || *profile.OwnerUserID != "owner-1" {
		t.Fatalf("unexpected owner user id: %#v", profile.OwnerUserID)
	}
	if profile.EditableByRole != "editor" {
		t.Fatalf("unexpected editableByRole: %q", profile.EditableByRole)
	}
}

// F-QA4-2: editable_by_role is a bound vocabulary (iamtypes canonical 8), not
// free text. The legacy 'admin'/'manager' phantoms and an empty value must all
// fail closed at construction rather than reach the DB.
func TestNewDocumentProfile_RejectsNonCanonicalEditableByRole(t *testing.T) {
	for _, role := range []string{"manager", "admin", "reviewer", "", "   ", "SYSTEM_ADMIN"} {
		_, err := NewDocumentProfile(DocumentProfile{
			Code:           "PO-01",
			TenantID:       "tenant-a",
			FamilyCode:     "policy",
			Name:           "Procedure",
			EditableByRole: role,
		})
		if !errors.Is(err, ErrInvalidEditableByRole) {
			t.Fatalf("role %q: expected ErrInvalidEditableByRole, got %v", role, err)
		}
	}
}

// system_admin IS valid for editable_by_role (document-edit authority is not
// area-scoped) — the opposite of default_approver_role. Guards the two
// vocabularies from being collapsed into one.
func TestNewDocumentProfile_AcceptsAllCanonicalRolesIncludingSystemAdmin(t *testing.T) {
	for _, role := range iamtypes.Roles() {
		profile, err := NewDocumentProfile(DocumentProfile{
			Code:           "PO-01",
			TenantID:       "tenant-a",
			FamilyCode:     "policy",
			Name:           "Procedure",
			EditableByRole: string(role),
		})
		if err != nil {
			t.Fatalf("role %q: unexpected error: %v", role, err)
		}
		if profile.EditableByRole != string(role) {
			t.Fatalf("role %q: unexpected editableByRole %q", role, profile.EditableByRole)
		}
	}
}

func TestNewDocumentProfile_RequiredFields(t *testing.T) {
	if _, err := NewDocumentProfile(DocumentProfile{Name: "x", TenantID: "t", FamilyCode: "f"}); !errors.Is(err, ErrProfileCodeRequired) {
		t.Fatalf("expected ErrProfileCodeRequired, got %v", err)
	}
	if _, err := NewDocumentProfile(DocumentProfile{Code: "p", Name: "x", FamilyCode: "f"}); !errors.Is(err, ErrProfileTenantRequired) {
		t.Fatalf("expected ErrProfileTenantRequired, got %v", err)
	}
	if _, err := NewDocumentProfile(DocumentProfile{Code: "p", TenantID: "t", Name: "x"}); !errors.Is(err, ErrProfileFamilyRequired) {
		t.Fatalf("expected ErrProfileFamilyRequired, got %v", err)
	}
	if _, err := NewDocumentProfile(DocumentProfile{Code: "p", TenantID: "t", FamilyCode: "f"}); !errors.Is(err, ErrProfileNameRequired) {
		t.Fatalf("expected ErrProfileNameRequired, got %v", err)
	}
}
