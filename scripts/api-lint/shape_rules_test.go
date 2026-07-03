package main

import (
	"path/filepath"
	"testing"
)

// TestShapeNullableNotRequired covers the three acceptance cases from
// validation-contract.md §F1.2 Part A:
//   - a nullable prop absent from `required` -> exactly 1 violation
//   - the same prop added to `required` -> 0 violations
//   - a non-nullable optional prop -> 0 violations
func TestShapeNullableNotRequired(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		spec string
		want []wantViolation
	}{
		{
			name: "nullable_not_required_bites",
			spec: "shape_nullable_not_required.openapi.yaml",
			want: []wantViolation{
				{Rule: "SHAPE-NULLABLE-NOT-REQUIRED", MessageContains: `schema Widget property "published_version" is nullable but not in required (present-and-null drifts to optional)`},
			},
		},
		{
			name: "nullable_and_required_clean",
			spec: "shape_nullable_required_good.openapi.yaml",
			want: nil,
		},
		{
			name: "optional_not_nullable_clean",
			spec: "shape_optional_not_nullable_good.openapi.yaml",
			want: nil,
		},
		{
			name: "request_only_exempt_response_still_checked",
			spec: "shape_request_scope_good.openapi.yaml",
			want: []wantViolation{
				{Rule: "SHAPE-NULLABLE-NOT-REQUIRED", MessageContains: `schema Widget property "published_version" is nullable but not in required (present-and-null drifts to optional)`},
			},
		}}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			specPath := filepath.Join("testdata", c.spec)
			got, err := RunSpecRules(specPath)
			if err != nil {
				t.Fatalf("RunSpecRules(%q): %v", specPath, err)
			}
			shapeOnly := make([]Violation, 0, len(got))
			for _, v := range got {
				if v.Rule == "SHAPE-NULLABLE-NOT-REQUIRED" {
					shapeOnly = append(shapeOnly, v)
				}
			}
			assertViolations(t, shapeOnly, c.want)
		})
	}
}
