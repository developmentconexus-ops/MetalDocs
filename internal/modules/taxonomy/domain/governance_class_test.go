package domain

import (
	"errors"
	"testing"
)

func TestGovernanceClassValidate(t *testing.T) {
	tests := []struct {
		name    string
		class   GovernanceClass
		wantErr error
	}{
		{"controlado ok", GovernanceControlado, nil},
		{"simples ok", GovernanceSimples, nil},
		{"livre ok", GovernanceLivre, nil},
		{"empty invalid", GovernanceClass(""), ErrInvalidGovernanceClass},
		{"unknown invalid", GovernanceClass("mystery"), ErrInvalidGovernanceClass},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.class.Validate()
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Validate() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestDocumentProfileRoutePolicy(t *testing.T) {
	tests := []struct {
		name  string
		class GovernanceClass
		want  RoutePolicy
	}{
		{"controlado requires approval", GovernanceControlado, RoutePolicyRequireApprovalStage},
		{"simples optional", GovernanceSimples, RoutePolicyApprovalOptional},
		{"livre no route", GovernanceLivre, RoutePolicyNoRoutePermitted},
		// Fail-closed: an unset/unknown class never silently drops the signature
		// requirement — it derives the strictest policy.
		{"empty fail-closes to require", GovernanceClass(""), RoutePolicyRequireApprovalStage},
		{"unknown fail-closes to require", GovernanceClass("mystery"), RoutePolicyRequireApprovalStage},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &DocumentProfile{GovernanceClass: tt.class}
			if got := p.RoutePolicy(); got != tt.want {
				t.Fatalf("RoutePolicy() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewDocumentProfileDefaultsGovernanceClass(t *testing.T) {
	// Empty class defaults to controlado (behavior-preserving: the pre-G1 world
	// required a signature on every profile).
	p, err := NewDocumentProfile(DocumentProfile{
		Code:               "pop-001",
		TenantID:           "t1",
		FamilyCode:         "fam",
		Name:               "POP",
		ReviewIntervalDays: 365,
		EditableByRole:     "editor",
	})
	if err != nil {
		t.Fatalf("NewDocumentProfile() error = %v", err)
	}
	if p.GovernanceClass != GovernanceControlado {
		t.Fatalf("GovernanceClass = %q, want %q", p.GovernanceClass, GovernanceControlado)
	}
}

func TestNewDocumentProfileRejectsInvalidGovernanceClass(t *testing.T) {
	_, err := NewDocumentProfile(DocumentProfile{
		Code:               "pop-001",
		TenantID:           "t1",
		FamilyCode:         "fam",
		Name:               "POP",
		ReviewIntervalDays: 365,
		EditableByRole:     "editor",
		GovernanceClass:    GovernanceClass("mystery"),
	})
	if !errors.Is(err, ErrInvalidGovernanceClass) {
		t.Fatalf("NewDocumentProfile() error = %v, want %v", err, ErrInvalidGovernanceClass)
	}
}
