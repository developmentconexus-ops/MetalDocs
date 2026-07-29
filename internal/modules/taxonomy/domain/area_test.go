package domain

import (
	"errors"
	"testing"
	"time"

	"metaldocs/internal/platform/iamtypes"
)

func TestProcessAreaIsActiveWhenNotArchived(t *testing.T) {
	a := ProcessArea{}

	if !a.IsActive() {
		t.Fatal("expected process area to be active when archived_at is nil")
	}
}

func TestProcessAreaArchiveSetsArchivedAt(t *testing.T) {
	a := ProcessArea{}
	now := time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC)

	if err := a.Archive(now); err != nil {
		t.Fatalf("unexpected error archiving area: %v", err)
	}
	if a.ArchivedAt == nil {
		t.Fatal("expected archived_at to be set")
	}
	if !a.ArchivedAt.Equal(now) {
		t.Fatalf("expected archived_at %s, got %s", now, a.ArchivedAt)
	}
}

func TestProcessAreaArchiveReturnsErrorWhenAlreadyArchived(t *testing.T) {
	archivedAt := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
	a := ProcessArea{ArchivedAt: &archivedAt}

	err := a.Archive(time.Date(2026, 4, 21, 10, 0, 0, 0, time.UTC))
	if !errors.Is(err, ErrAreaArchived) {
		t.Fatalf("expected ErrAreaArchived, got %v", err)
	}
}

func TestNewProcessArea_NormalizesAndValidates(t *testing.T) {
	parent := AreaCode(" ROOT ")
	owner := " owner-1 "
	role := " approver "
	area, err := NewProcessArea(ProcessArea{
		Code:                " AR-01 ",
		TenantID:            " tenant-a ",
		Name:                " Finance ",
		Description:         " Desc ",
		ParentCode:          &parent,
		OwnerUserID:         &owner,
		DefaultApproverRole: &role,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if area.Code != "AR-01" || area.TenantID != "tenant-a" || area.Name != "Finance" || area.Description != "Desc" {
		t.Fatalf("unexpected normalized area: %#v", area)
	}
	if area.ParentCode == nil || *area.ParentCode != "ROOT" {
		t.Fatalf("unexpected parent code: %#v", area.ParentCode)
	}
	if area.OwnerUserID == nil || *area.OwnerUserID != "owner-1" {
		t.Fatalf("unexpected owner user id: %#v", area.OwnerUserID)
	}
	if area.DefaultApproverRole == nil || *area.DefaultApproverRole != "approver" {
		t.Fatalf("unexpected default approver role: %#v", area.DefaultApproverRole)
	}
}

// F-QA4-2: default_approver_role is bound to the AREA-assignable subset. Free
// text and the legacy phantoms fail closed; crucially so does system_admin,
// which is tenant-wide tier-1 and can never be a user_process_areas membership
// — configuring it would resolve to an empty approver pool at runtime.
func TestNewProcessArea_RejectsNonAreaAssignableDefaultApproverRole(t *testing.T) {
	for _, role := range []string{"system_admin", "manager", "admin", "reviewer", "APPROVER"} {
		r := role
		_, err := NewProcessArea(ProcessArea{
			Code:                "AR-01",
			TenantID:            "tenant-a",
			Name:                "Finance",
			DefaultApproverRole: &r,
		})
		if !errors.Is(err, ErrInvalidDefaultApproverRole) {
			t.Fatalf("role %q: expected ErrInvalidDefaultApproverRole, got %v", role, err)
		}
	}
}

// The field is optional: nil and blank-collapsing-to-nil stay valid.
func TestNewProcessArea_AllowsAbsentDefaultApproverRole(t *testing.T) {
	blank := "   "
	for _, in := range []*string{nil, &blank} {
		area, err := NewProcessArea(ProcessArea{
			Code:                "AR-01",
			TenantID:            "tenant-a",
			Name:                "Finance",
			DefaultApproverRole: in,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if area.DefaultApproverRole != nil {
			t.Fatalf("expected nil default approver role, got %#v", area.DefaultApproverRole)
		}
	}
}

// Every area-assignable role is accepted, and the set is exactly 7 (all
// canonical roles minus system_admin) — the same shape the AreaRole schema
// publishes in api/openapi/v1/openapi.yaml.
func TestNewProcessArea_AcceptsEveryAreaAssignableRole(t *testing.T) {
	roles := iamtypes.AreaRoles()
	if len(roles) != 7 {
		t.Fatalf("expected 7 area-assignable roles, got %d: %v", len(roles), roles)
	}
	for _, role := range roles {
		r := string(role)
		area, err := NewProcessArea(ProcessArea{
			Code:                "AR-01",
			TenantID:            "tenant-a",
			Name:                "Finance",
			DefaultApproverRole: &r,
		})
		if err != nil {
			t.Fatalf("role %q: unexpected error: %v", role, err)
		}
		if area.DefaultApproverRole == nil || *area.DefaultApproverRole != r {
			t.Fatalf("role %q: unexpected default approver role %#v", role, area.DefaultApproverRole)
		}
	}
}

func TestNewProcessArea_RequiredFields(t *testing.T) {
	if _, err := NewProcessArea(ProcessArea{Name: "X", TenantID: "t"}); !errors.Is(err, ErrAreaCodeRequired) {
		t.Fatalf("expected ErrAreaCodeRequired, got %v", err)
	}
	if _, err := NewProcessArea(ProcessArea{Code: "A", Name: "X"}); !errors.Is(err, ErrAreaTenantRequired) {
		t.Fatalf("expected ErrAreaTenantRequired, got %v", err)
	}
	if _, err := NewProcessArea(ProcessArea{Code: "A", TenantID: "t"}); !errors.Is(err, ErrAreaNameRequired) {
		t.Fatalf("expected ErrAreaNameRequired, got %v", err)
	}
}
