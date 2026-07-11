package application

import (
	"context"
	"testing"

	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
)

// stubPolicyReader is an inert ProfilePolicyReader used only to assert wiring.
type stubPolicyReader struct{}

func (stubPolicyReader) RoutePolicy(context.Context, string, string) (taxonomydomain.RoutePolicy, error) {
	return taxonomydomain.RoutePolicyApprovalOptional, nil
}

// TestWithProfilePolicyReader_ThreadsIntoBothSites verifies the option installs
// the reader on both route-shape enforcement services (Submit + RouteAdmin) and
// leaves them nil-configured otherwise. Nil-safe on empty services.
func TestWithProfilePolicyReader_ThreadsIntoBothSites(t *testing.T) {
	s := &Services{Submit: &SubmitService{}, RouteAdmin: &RouteAdminService{}}

	if s.Submit.policyReader != nil || s.RouteAdmin.policyReader != nil {
		t.Fatalf("precondition: expected unset policyReader before wiring")
	}

	reader := stubPolicyReader{}
	got := s.WithProfilePolicyReader(reader)

	if got != s {
		t.Fatalf("WithProfilePolicyReader should return the same *Services for chaining")
	}
	if s.Submit.policyReader == nil {
		t.Fatalf("Submit.policyReader not wired")
	}
	if s.RouteAdmin.policyReader == nil {
		t.Fatalf("RouteAdmin.policyReader not wired")
	}
}

func TestWithProfilePolicyReader_NilServicesSafe(t *testing.T) {
	var s *Services
	if got := s.WithProfilePolicyReader(stubPolicyReader{}); got != nil {
		t.Fatalf("nil Services should stay nil")
	}
}
