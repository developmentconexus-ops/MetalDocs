package domain

import (
	"errors"
	"testing"
)

func TestDocumentFamily_Deactivate(t *testing.T) {
	f := DocumentFamily{Code: "policy", IsActive: true}
	if err := f.Deactivate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.IsActive {
		t.Fatal("expected IsActive=false after Deactivate")
	}
}

func TestDocumentFamily_DeactivateAlreadyInactive(t *testing.T) {
	f := DocumentFamily{Code: "policy", IsActive: false}
	if err := f.Deactivate(); err != ErrFamilyAlreadyInactive {
		t.Fatalf("want ErrFamilyAlreadyInactive, got %v", err)
	}
}

func TestNewDocumentFamily_NormalizesAndActivates(t *testing.T) {
	family, err := NewDocumentFamily(DocumentFamily{
		Code:        " policy ",
		Name:        " Policy ",
		Description: " Desc ",
		IsActive:    false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if family.Code != "policy" || family.Name != "Policy" || family.Description != "Desc" {
		t.Fatalf("unexpected normalized family: %#v", family)
	}
	if !family.IsActive {
		t.Fatal("expected IsActive=true")
	}
}

func TestNewDocumentFamily_RequiredFields(t *testing.T) {
	if _, err := NewDocumentFamily(DocumentFamily{Name: "Policy"}); !errors.Is(err, ErrFamilyCodeRequired) {
		t.Fatalf("expected ErrFamilyCodeRequired, got %v", err)
	}
	if _, err := NewDocumentFamily(DocumentFamily{Code: "policy"}); !errors.Is(err, ErrFamilyNameRequired) {
		t.Fatalf("expected ErrFamilyNameRequired, got %v", err)
	}
}
