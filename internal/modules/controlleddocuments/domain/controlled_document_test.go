package domain

import (
	"errors"
	"testing"
)

func TestNewControlledDocument_NormalizesAndValidates(t *testing.T) {
	department := "  ops "
	override := " tpl-v1 "
	doc, err := NewControlledDocument(ControlledDocument{
		TenantID:                  " tenant-a ",
		ProfileCode:               " po ",
		ProcessAreaCode:           " qa ",
		DepartmentCode:            &department,
		Code:                      " po-qa-001 ",
		Title:                     " Welding Procedure ",
		OwnerUserID:               " owner-1 ",
		OverrideTemplateVersionID: &override,
		Status:                    CDStatusActive,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.TenantID != "tenant-a" || doc.ProfileCode != "po" || doc.ProcessAreaCode != "qa" || doc.Code != "po-qa-001" || doc.Title != "Welding Procedure" || doc.OwnerUserID != "owner-1" {
		t.Fatalf("unexpected normalized doc: %#v", doc)
	}
	if doc.DepartmentCode == nil || *doc.DepartmentCode != "ops" {
		t.Fatalf("expected trimmed department code, got %#v", doc.DepartmentCode)
	}
	if doc.OverrideTemplateVersionID == nil || *doc.OverrideTemplateVersionID != "tpl-v1" {
		t.Fatalf("expected trimmed override template version id, got %#v", doc.OverrideTemplateVersionID)
	}
}

func TestNewControlledDocument_RequiredFields(t *testing.T) {
	_, err := NewControlledDocument(ControlledDocument{
		ProfileCode:     "po",
		ProcessAreaCode: "qa",
		Code:            "PO-QA-001",
		Title:           "Title",
		OwnerUserID:     "owner-1",
	})
	if !errors.Is(err, ErrCDTenantRequired) {
		t.Fatalf("expected ErrCDTenantRequired, got %v", err)
	}
}
