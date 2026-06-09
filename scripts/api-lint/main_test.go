package main

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

type wantViolation struct {
	Rule            string
	MessageContains string
}

func TestRules(t *testing.T) {
	t.Parallel()

	type tc struct {
		name    string
		spec    string
		modules string
		want    []wantViolation
	}

	cases := []tc{
		{name: "good_spec", spec: "good.openapi.yaml"},
		{name: "path_base_prefix", spec: "path_base_prefix.openapi.yaml", want: []wantViolation{{Rule: "PATH-BASE-PREFIX", MessageContains: "must be relative to servers.url"}}},
		{name: "missing_problem", spec: "missing_problem.openapi.yaml", want: []wantViolation{{Rule: "ENVELOPE-DRIFT", MessageContains: "does not reference Problem"}}},
		{name: "envelope_shared_responses", spec: "envelope_shared_responses.openapi.yaml"},
		{name: "envelope_unresolved_ref", spec: "envelope_unresolved_ref.openapi.yaml", want: []wantViolation{{Rule: "ENVELOPE-DRIFT", MessageContains: "unresolved $ref"}}},
		{name: "missing_cursor", spec: "missing_cursor.openapi.yaml", want: []wantViolation{{Rule: "PAGINATION-DRIFT", MessageContains: "missing query param cursor"}}},
		{name: "casing_bad", spec: "casing_bad.openapi.yaml", want: []wantViolation{{Rule: "CASING-DRIFT", MessageContains: "is not snake_case"}}},
		{name: "casing_exempt_good", spec: "casing_exempt_good.openapi.yaml"},
		// Parameter-name casing (Family 5): a camelCase query/path param is drift; a
		// snake param is clean, and header params are skipped (kebab/Pascal allowed).
		{name: "casing_param_bad", spec: "casing_param_bad.openapi.yaml", want: []wantViolation{{Rule: "CASING-DRIFT", MessageContains: `query parameter "includeArchived" is not snake_case`}}},
		{name: "casing_param_good", spec: "casing_param_good.openapi.yaml"},
		{name: "missing_security", spec: "missing_security.openapi.yaml", want: []wantViolation{{Rule: "AUTHZ-DRIFT", MessageContains: "missing security declaration"}}},
		{name: "global_security_good", spec: "global_security_good.openapi.yaml"},
		{name: "state_transition_no_area", spec: "state_transition_no_area.openapi.yaml", want: []wantViolation{{Rule: "AUTHZ-DRIFT", MessageContains: "state-transition op"}}},
		// FD-1 (Phase F): positive authz-area markers replace the deleted negative
		// x-authz-skip-area escape hatch. A tx-derived marker satisfies the
		// state-transition requirement; area-none does too; a malformed marker is
		// drift.
		{name: "state_transition_area_tx", spec: "state_transition_area_tx.openapi.yaml"},
		{name: "state_transition_area_none", spec: "state_transition_area_none.openapi.yaml"},
		{name: "authz_area_tx_no_derived", spec: "authz_area_tx_no_derived.openapi.yaml", want: []wantViolation{{Rule: "AUTHZ-DRIFT", MessageContains: "source:tx requires a non-empty derived_from"}}},
		{name: "authz_area_bad_source", spec: "authz_area_bad_source.openapi.yaml", want: []wantViolation{{Rule: "AUTHZ-DRIFT", MessageContains: "invalid source"}}},
		{name: "pagination_exempt_good", spec: "pagination_exempt_good.openapi.yaml"},
		{name: "pagination_exempt_no_reason", spec: "pagination_exempt_no_reason.openapi.yaml", want: []wantViolation{{Rule: "PAGINATION-DRIFT", MessageContains: "without a non-empty x-pagination-exempt-reason"}}},
		{name: "repo_good", spec: "repo_good/spec.yaml", modules: "repo_good"},
		{name: "repo_missing", spec: "repo_missing/spec.yaml", modules: "repo_missing", want: []wantViolation{{Rule: "tripwire-pairing", MessageContains: "without authz.Require call"}}},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			specPath := filepath.Join("testdata", c.spec)

			got, err := RunSpecRules(specPath)
			if err != nil {
				t.Fatalf("RunSpecRules(%q): %v", specPath, err)
			}
			if c.modules != "" {
				moduleRoot := filepath.Join("testdata", c.modules)
				codeGot, err := RunCodeRules(specPath, moduleRoot, false)
				if err != nil {
					t.Fatalf("RunCodeRules(%q, %q): %v", specPath, moduleRoot, err)
				}
				got = append(got, codeGot...)
			}

			assertViolations(t, got, c.want)
		})
	}
}

func assertViolations(t *testing.T, got []Violation, want []wantViolation) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("violation count mismatch: got=%d want=%d\nfull got=%#v\nfull want=%#v", len(got), len(want), got, want)
	}

	gotPairs := make([]string, 0, len(got))
	wantPairs := make([]string, 0, len(want))
	used := make([]bool, len(got))

	for _, w := range want {
		matched := -1
		for i, g := range got {
			if used[i] {
				continue
			}
			if g.Rule == w.Rule && strings.Contains(g.Message, w.MessageContains) {
				used[i] = true
				matched = i
				gotPairs = append(gotPairs, fmt.Sprintf("%s|%s", g.Rule, w.MessageContains))
				wantPairs = append(wantPairs, fmt.Sprintf("%s|%s", w.Rule, w.MessageContains))
				break
			}
		}
		if matched == -1 {
			t.Fatalf("missing expected violation rule=%q message_contains=%q\nfull got=%#v", w.Rule, w.MessageContains, got)
		}
	}

	sort.Strings(gotPairs)
	sort.Strings(wantPairs)
	for i := range wantPairs {
		if gotPairs[i] != wantPairs[i] {
			t.Fatalf("violation multiset mismatch index=%d got=%q want=%q\nfull got=%#v\nfull want=%#v", i, gotPairs[i], wantPairs[i], got, want)
		}
	}
}
