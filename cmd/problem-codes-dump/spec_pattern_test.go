package main

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"metaldocs/internal/platform/problem"
)

// The OpenAPI spec constrains Problem.code with a `pattern` that names the ten
// semantic families in an alternation. That alternation is a SECOND copy of a
// list whose first copy lives in problem.familyDefaults — precisely the
// hand-synced-enumeration shape ADR 0089 exists to remove.
//
// It is copied anyway, because the alternative is worse: the spec is the wire
// contract for clients that never link Go code, and a `type: string` with no
// shape tells them nothing. So the copy stays, and this test is what keeps it
// from drifting — add or rename a family and it fails here, at the line that
// names the fix, rather than in a client months later.
//
// This test lives in cmd/problem-codes-dump because that is the only package
// that links EVERY registering module in (enforced by the PROBLEM-DUMP-IMPORT
// api-lint rule), so `regs` below is the whole catalog and not a subset.

const specRelPath = "api/openapi/v1/openapi.yaml"

// codePatternRe pulls the pattern line out of the Problem.code schema. Matched
// textually rather than by parsing YAML: a full parse would need the spec's
// $ref graph, and the failure this guards against is a one-line edit.
var codePatternRe = regexp.MustCompile(`(?m)^\s*pattern: '(\^\(.*?\)\\\..*?\$)'\s*$`)

func repoRoot(t *testing.T) string {
	t.Helper()
	// The test binary runs in cmd/problem-codes-dump.
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func specCodePattern(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(repoRoot(t), filepath.FromSlash(specRelPath)))
	if err != nil {
		t.Fatalf("read %s: %v", specRelPath, err)
	}
	m := codePatternRe.FindSubmatch(raw)
	if m == nil {
		t.Fatalf("no Problem.code `pattern` found in %s — it constrains the wire vocabulary and must not be dropped", specRelPath)
	}
	return string(m[1])
}

// TestSpecPatternMatchesRegistry is the anti-drift half: the spec's family
// alternation must be exactly the registry's family set.
func TestSpecPatternMatchesRegistry(t *testing.T) {
	pattern := specCodePattern(t)

	open := strings.Index(pattern, "(")
	closeIdx := strings.Index(pattern, ")")
	if open < 0 || closeIdx < open {
		t.Fatalf("pattern %q has no family alternation group", pattern)
	}
	inSpec := strings.Split(pattern[open+1:closeIdx], "|")
	sort.Strings(inSpec)

	inRegistry := make([]string, 0, 10)
	for _, f := range problem.Families() {
		inRegistry = append(inRegistry, f.Name)
	}
	sort.Strings(inRegistry)

	if strings.Join(inSpec, "|") != strings.Join(inRegistry, "|") {
		t.Fatalf("family drift between the spec pattern and problem.Families():\n"+
			"  spec:     %v\n"+
			"  registry: %v\n\n"+
			"Update the `pattern` on Problem.code in %s, then re-run `go generate ./...`.",
			inSpec, inRegistry, specRelPath)
	}
}

// TestEveryRegisteredCodeMatchesSpecPattern is the coverage half: a pattern that
// the real codes violate is worse than no pattern, because a strict client would
// reject a valid response.
func TestEveryRegisteredCodeMatchesSpecPattern(t *testing.T) {
	re, err := regexp.Compile(specCodePattern(t))
	if err != nil {
		t.Fatalf("Problem.code pattern does not compile as a Go regexp: %v", err)
	}

	regs := problem.Registrations()
	if len(regs) == 0 {
		t.Fatal("registry is empty; no registering package was linked in")
	}
	for _, r := range regs {
		if !re.MatchString(r.Code.String()) {
			t.Errorf("registered code %q (%s, %s) does not match the spec pattern — a strict client would reject a valid response",
				r.Code.String(), r.Module, r.DeclaredAt)
		}
	}
}
