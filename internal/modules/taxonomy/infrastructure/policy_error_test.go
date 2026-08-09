package infrastructure

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"metaldocs/internal/modules/taxonomy/domain"
)

// TestMapProfilePolicyErr covers the translation of the governance-policy
// trigger exceptions (P0001 with stable message prefixes; assert_route_shape
// as replaced by migration 0316)
// into taxonomy domain sentinels, keeping driver error codes out of the
// application/domain layers.
func TestMapProfilePolicyErr(t *testing.T) {
	tests := []struct {
		name string
		in   error
		want error
	}{
		{"nil passes through", nil, nil},
		{
			"reclassify conflict -> ErrClassChangeRouteConflict",
			&pgconn.PgError{Code: "P0001", Message: "ErrClassChangeRouteConflict: profile is controlado; route x must contain at least one approval-kind stage"},
			domain.ErrClassChangeRouteConflict,
		},
		{
			"route-shape violation -> ErrRouteViolatesProfilePolicy",
			&pgconn.PgError{Code: "P0001", Message: "ErrRouteViolatesProfilePolicy: profile is livre; route 0f1e must carry zero stages (a livre route is configured to require no approval)"},
			domain.ErrRouteViolatesProfilePolicy,
		},
		{
			"other P0001 passes through untouched",
			&pgconn.PgError{Code: "P0001", Message: "ErrRouteInUse: definition column changed"},
			nil, // sentinel checked below via identity
		},
		{
			"non-pg error passes through",
			errors.New("connection reset"),
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapProfilePolicyErr(tt.in)
			switch tt.want {
			case nil:
				// pass-through cases: the returned error must be identical to input.
				if !errors.Is(got, tt.in) {
					t.Fatalf("mapProfilePolicyErr(%v) = %v, want identical passthrough", tt.in, got)
				}
			default:
				if !errors.Is(got, tt.want) {
					t.Fatalf("mapProfilePolicyErr(%v) = %v, want %v", tt.in, got, tt.want)
				}
			}
		})
	}
}

// TestMapProfilePolicyErr_WrappedCommit confirms the mapping also fires when the
// pg error is wrapped (mirrors how a driver may return a wrapped commit error).
func TestMapProfilePolicyErr_WrappedCommit(t *testing.T) {
	wrapped := fmt.Errorf("commit: %w", &pgconn.PgError{Code: "P0001", Message: "ErrClassChangeRouteConflict: conflict"})
	if !errors.Is(mapProfilePolicyErr(wrapped), domain.ErrClassChangeRouteConflict) {
		t.Fatalf("expected wrapped commit error to map to ErrClassChangeRouteConflict")
	}
}
