//go:build integration

package application_test

import (
	"testing"

	"metaldocs/internal/modules/documents_v2/approval/application"
)

// Compile-check: ensures the application package exports NewServices and RealClock.
// Full integration coverage requires a live DB — run with -tags integration.
func TestEligibleActors_CompileCheck(t *testing.T) {
	t.Skip("requires live DB — run with -tags integration and real document/route IDs")
	_ = application.NewServices
	_ = application.RealClock{}
}
