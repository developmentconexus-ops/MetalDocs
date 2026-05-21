package domain_test

import (
	"testing"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
)

func TestAutoCode_Format(t *testing.T) {
	got := controlleddocumentsdomain.AutoCode("po", "rh", 5)
	want := "PO-RH-005"
	if got != want {
		t.Errorf("AutoCode: got %q, want %q", got, want)
	}
}

func TestAutoCode_ThreeSegmentThreeDigit(t *testing.T) {
	got := controlleddocumentsdomain.AutoCode("dc", "rh", 3)
	if got != "DC-RH-003" {
		t.Fatalf("got %q want DC-RH-003", got)
	}
}

