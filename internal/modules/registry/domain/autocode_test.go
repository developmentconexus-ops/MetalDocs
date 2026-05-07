package domain_test

import (
	"testing"

	registrydomain "metaldocs/internal/modules/registry/domain"
)

func TestAutoCode_Format(t *testing.T) {
	got := registrydomain.AutoCode("po", "rh", 5)
	want := "PO-RH-005"
	if got != want {
		t.Errorf("AutoCode: got %q, want %q", got, want)
	}
}

func TestAutoCode_ThreeSegmentThreeDigit(t *testing.T) {
	got := registrydomain.AutoCode("dc", "rh", 3)
	if got != "DC-RH-003" {
		t.Fatalf("got %q want DC-RH-003", got)
	}
}
