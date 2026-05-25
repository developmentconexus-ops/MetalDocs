package domain

import (
	"errors"
	"testing"
)

func TestNewVisibility_DefaultsToRestrictedArea(t *testing.T) {
	v, err := NewVisibility("", nil, nil, "RH")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Scope != VisibilityScopeRestricted {
		t.Fatalf("scope = %q, want %q", v.Scope, VisibilityScopeRestricted)
	}
	if len(v.AreaCodes) != 1 || v.AreaCodes[0] != "RH" {
		t.Fatalf("areaCodes = %+v, want [RH]", v.AreaCodes)
	}
}

func TestNewVisibility_CompanyClearsGrants(t *testing.T) {
	v, err := NewVisibility("company", []string{"A1"}, []string{"u1"}, "RH")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Scope != VisibilityScopePublic {
		t.Fatalf("scope = %q, want %q", v.Scope, VisibilityScopePublic)
	}
	if len(v.AreaCodes) != 0 || len(v.UserIDs) != 0 {
		t.Fatalf("grants should be empty for company scope: %+v", v)
	}
}

func TestNewVisibility_InvalidScope(t *testing.T) {
	_, err := NewVisibility("team", nil, nil, "RH")
	if !errors.Is(err, ErrVisibilityScopeInvalid) {
		t.Fatalf("expected ErrVisibilityScopeInvalid, got %v", err)
	}
}
