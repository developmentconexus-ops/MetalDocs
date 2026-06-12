package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPlaceholderType_AllConstants(t *testing.T) {
	types := []PlaceholderType{PHText, PHDate, PHNumber, PHSelect, PHUser, PHPicture, PHComputed}
	wants := []string{"text", "date", "number", "select", "user", "picture", "computed"}
	for i, pt := range types {
		if string(pt) != wants[i] {
			t.Fatalf("PlaceholderType[%d] = %q, want %q", i, pt, wants[i])
		}
	}
}

func TestPlaceholder_JSONRoundTrip_AllFields(t *testing.T) {
	regex := "^[A-Z]{3}-\\d{4}$"
	mn, mx := 0.0, 100.0
	maxLen := 120
	rkey := "doc_code"
	ph := Placeholder{
		ID: "p1", Label: "Doc Code", Type: PHText, Required: true,
		Regex: &regex, MaxLength: &maxLen, MinNumber: &mn, MaxNumber: &mx,
		VisibleIf: &VisibilityCondition{PlaceholderID: "p0", Op: "eq", Value: "x"},
		Computed:  true, ResolverKey: &rkey,
	}
	b, err := json.Marshal(ph)
	if err != nil {
		t.Fatal(err)
	}
	var back Placeholder
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if back.ID != "p1" || !back.Computed || back.ResolverKey == nil || *back.ResolverKey != "doc_code" {
		t.Fatalf("round-trip mismatch: %+v", back)
	}
	if back.VisibleIf == nil || back.VisibleIf.Op != "eq" {
		t.Fatalf("visible_if lost: %+v", back.VisibleIf)
	}
}

func TestPlaceholder_NameField_JSONRoundTrip(t *testing.T) {
	ph := Placeholder{ID: "p1", Name: "doc_code", Label: "Doc Code", Type: PHText}
	b, err := json.Marshal(ph)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"name":"doc_code"`) {
		t.Fatalf("JSON missing name field: %s", b)
	}
	var got Placeholder
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got.Name != "doc_code" {
		t.Fatalf("Name = %q, want %q", got.Name, "doc_code")
	}
}

func TestPlaceholder_NameField_OmitEmpty(t *testing.T) {
	ph := Placeholder{ID: "p1", Label: "Doc Code", Type: PHText}
	b, err := json.Marshal(ph)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), `"name"`) {
		t.Fatalf("JSON contains omitted name field: %s", b)
	}
}
