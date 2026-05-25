package domain

import (
	"errors"
	"testing"
	"time"
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
	parent := " ROOT "
	owner := " owner-1 "
	role := " manager "
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
	if area.DefaultApproverRole == nil || *area.DefaultApproverRole != "manager" {
		t.Fatalf("unexpected default approver role: %#v", area.DefaultApproverRole)
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
