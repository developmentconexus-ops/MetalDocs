package infrastructure

import (
	"reflect"
	"testing"
)

// ---------------------------------------------------------------------------
// TestUnionDedupSort — the pure union+dedup+sort core of
// ResolveEligibleActorsForSelectors (M4, unit 3.2, slice 3). Each case models
// one selector kind's contributed pool (or lack thereof) at the point it
// reaches the union step, so this exercises the resolver's documented
// contract without a DB:
//   - role_in_fixed_area / role_in_document_area selectors contribute their
//     resolved role×area pool (possibly overlapping another selector's pool).
//   - named_user contributes a singleton pool (the one checked-active user).
//   - submit_choice contributes NO pool at all (it is simply never appended
//     to the pools slice by ResolveEligibleActorsForSelectors) — modeled here
//     by omitting it, not by passing an empty slice for it.
//   - no contributing selectors (or all-empty pools) returns []string{}, never
//     nil.
//
// The DB-backed per-kind query correctness (v_active_user_areas membership,
// role_in_document_area's subjectArea substitution, parity with the legacy
// flat ResolveEligibleActors) is covered by the real-DB integration test in
// tests/integration/approval (this slice's parity + union test).
// ---------------------------------------------------------------------------

func TestUnionDedupSort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		pools [][]string
		want  []string
	}{
		{
			name:  "no selectors contributed a pool returns empty not nil",
			pools: nil,
			want:  []string{},
		},
		{
			name: "two role selectors with overlapping ids dedup to the union",
			pools: [][]string{
				{"user-a", "user-b"}, // role_in_fixed_area(role=qa, area=area-1)
				{"user-b", "user-c"}, // role_in_document_area(role=manager)
			},
			want: []string{"user-a", "user-b", "user-c"},
		},
		{
			name: "named_user contributes exactly its single checked-active id",
			pools: [][]string{
				{"user-a", "user-b"}, // role_in_fixed_area pool
				{"user-z"},           // named_user(user-z), confirmed active
			},
			want: []string{"user-a", "user-b", "user-z"},
		},
		{
			name: "submit_choice contributes no pool (omitted entirely, not an empty slice)",
			pools: [][]string{
				{"user-a"}, // role_in_fixed_area pool
				// submit_choice selector: no entry appended by the resolver.
			},
			want: []string{"user-a"},
		},
		{
			name: "an inactive named_user and an empty role pool both contribute nothing",
			pools: [][]string{
				{},
				nil,
			},
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := unionDedupSort(tt.pools...)
			if got == nil {
				t.Fatalf("unionDedupSort(%v) = nil, want non-nil", tt.pools)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("unionDedupSort(%v) = %v, want %v", tt.pools, got, tt.want)
			}
		})
	}
}
