package domain

import (
	"testing"
	"time"
)

func TestDocumentFamily_Deactivate(t *testing.T) {
	f := DocumentFamily{Code: "policy", IsActive: true}
	if err := f.Deactivate(time.Now()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.IsActive {
		t.Fatal("expected IsActive=false after Deactivate")
	}
}

func TestDocumentFamily_DeactivateAlreadyInactive(t *testing.T) {
	f := DocumentFamily{Code: "policy", IsActive: false}
	if err := f.Deactivate(time.Now()); err != ErrFamilyAlreadyInactive {
		t.Fatalf("want ErrFamilyAlreadyInactive, got %v", err)
	}
}
