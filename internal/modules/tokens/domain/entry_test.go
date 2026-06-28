package domain_test

import (
	"errors"
	"strings"
	"testing"

	"metaldocs/internal/modules/tokens/domain"
)

func validInput() domain.NewEntryInput {
	return domain.NewEntryInput{
		TenantID: "11111111-1111-1111-1111-111111111111",
		ActorID:  "22222222-2222-2222-2222-222222222222",
		Name:     "company_slogan",
		Value:    "Quality first",
		Label:    "Company slogan",
	}
}

func TestNewEntry_OK(t *testing.T) {
	e, err := domain.NewEntry(validInput())
	if err != nil {
		t.Fatalf("NewEntry = %v, want nil", err)
	}
	if e.Name != "company_slogan" || e.CreatedBy != e.UpdatedBy {
		t.Fatalf("unexpected entry %+v", e)
	}
}

func TestNewEntry_RejectsBadName(t *testing.T) {
	for _, bad := range []string{"", "has space", "dotted.name", "hyphen-name", "{braced}", strings.Repeat("a", 65)} {
		in := validInput()
		in.Name = bad
		if _, err := domain.NewEntry(in); err == nil {
			t.Fatalf("NewEntry(name=%q) = nil, want ValidationError", bad)
		} else {
			var ve *domain.ValidationError
			if !errors.As(err, &ve) || ve.Field != "name" {
				t.Fatalf("NewEntry(name=%q) err = %v, want ValidationError on name", bad, err)
			}
		}
	}
}

func TestNewEntry_RejectsEmptyValueAndLabel(t *testing.T) {
	in := validInput()
	in.Value = ""
	if _, err := domain.NewEntry(in); err == nil {
		t.Fatal("empty value accepted, want ValidationError")
	}
	in = validInput()
	in.Label = ""
	if _, err := domain.NewEntry(in); err == nil {
		t.Fatal("empty label accepted, want ValidationError")
	}
}

func TestNewEntry_RejectsOverlongValue(t *testing.T) {
	in := validInput()
	in.Value = strings.Repeat("x", 4097)
	if _, err := domain.NewEntry(in); err == nil {
		t.Fatal("4097-char value accepted, want ValidationError")
	}
}
