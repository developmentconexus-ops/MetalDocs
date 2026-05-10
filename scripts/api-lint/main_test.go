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
		{name: "missing_problem", spec: "missing_problem.openapi.yaml", want: []wantViolation{{Rule: "ENVELOPE-DRIFT", MessageContains: "does not reference Problem"}}},
		{name: "missing_cursor", spec: "missing_cursor.openapi.yaml", want: []wantViolation{{Rule: "PAGINATION-DRIFT", MessageContains: "missing query param cursor"}}},
		{name: "missing_security", spec: "missing_security.openapi.yaml", want: []wantViolation{{Rule: "AUTHZ-DRIFT", MessageContains: "missing security declaration"}}},
		{name: "state_transition_no_area", spec: "state_transition_no_area.openapi.yaml", want: []wantViolation{{Rule: "AUTHZ-DRIFT", MessageContains: "state-transition op"}}},
		{name: "state_transition_skip_area", spec: "state_transition_skip_area.openapi.yaml"},
		{name: "handlers_good", spec: "handlers_good/spec.yaml", modules: "handlers_good"},
		{name: "handlers_missing_call", spec: "handlers_missing_call/spec.yaml", modules: "handlers_missing_call", want: []wantViolation{{Rule: "authz-call-present", MessageContains: "does not call authz.Require"}}},
		{name: "handlers_wrong_field", spec: "handlers_wrong_field/spec.yaml", modules: "handlers_wrong_field", want: []wantViolation{{Rule: "authz-call-present", MessageContains: "expected req.Body.AreaCode"}}},
		{name: "handlers_custom", spec: "handlers_custom/spec.yaml", modules: "handlers_custom"},
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
				codeGot, err := RunCodeRules(specPath, moduleRoot)
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
