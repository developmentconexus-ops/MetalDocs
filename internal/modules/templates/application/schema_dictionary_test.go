package application

import (
	"errors"
	"testing"

	"metaldocs/internal/modules/templates/domain"
)

func nativeSet() map[string]int {
	return map[string]int{
		"doc_code": 1, "doc_title": 1, "revision_number": 1, "effective_date": 1,
		"controlled_by_area": 1, "author": 1, "approvers": 1, "approval_date": 1,
	}
}

func TestValidatePlaceholders_Dictionary(t *testing.T) {
	ck := "author"
	cases := []struct {
		name    string
		ph      domain.Placeholder
		wantErr error
	}{
		{"valid dictionary ref", domain.Placeholder{ID: "p1", Type: domain.PHDictionary, Name: "company_name", Label: "Co"}, nil},
		{"native name rejected", domain.Placeholder{ID: "p2", Type: domain.PHDictionary, Name: "author", Label: "A"}, domain.ErrPlaceholderReservedName},
		{"registry-only native (approval_date) rejected", domain.Placeholder{ID: "p3", Type: domain.PHDictionary, Name: "approval_date", Label: "AD"}, domain.ErrPlaceholderReservedName},
		{"bad format rejected", domain.Placeholder{ID: "p4", Type: domain.PHDictionary, Name: "Company Name", Label: "C"}, domain.ErrPlaceholderNameInvalid},
		{"dictionary with resolver_key rejected", domain.Placeholder{ID: "p5", Type: domain.PHDictionary, Name: "company_name", Label: "C", ResolverKey: &ck}, domain.ErrPlaceholderDictionaryInvalid},
		{"dictionary computed=true rejected", domain.Placeholder{ID: "p6", Type: domain.PHDictionary, Name: "company_name", Label: "C", Computed: true}, domain.ErrPlaceholderDictionaryInvalid},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePlaceholders([]domain.Placeholder{tc.ph}, nativeSet())
			if tc.wantErr == nil {
				if err != nil {
					t.Fatalf("want nil, got %v", err)
				}
				return
			}
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("want %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestValidatePlaceholders_ComputedStillValid(t *testing.T) {
	rk := "author"
	ph := domain.Placeholder{ID: "c1", Type: domain.PHComputed, Name: "author", Label: "A", Computed: true, ResolverKey: &rk}
	if err := ValidatePlaceholders([]domain.Placeholder{ph}, nativeSet()); err != nil {
		t.Fatalf("computed placeholder must still validate, got %v", err)
	}
}
