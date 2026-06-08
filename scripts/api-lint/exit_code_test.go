package main

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// This is the Family 3 · C1 regression guard for the api-lint exit-code contract
// (CWE-693). The reported-only/deferred exit tier and the -enforce three-way
// split were retired: EVERY rule now gates, so the binary must exit non-zero on
// ANY violation and zero only on a fully clean spec. The per-rule unit tests in
// main_test.go assert RunSpecRules' VIOLATION SET; they cannot assert main()'s
// EXIT CODE because main() calls os.Exit. This test drives the real built binary
// as a subprocess so a regression that re-introduces a silent-bypass exit class
// (e.g. a continue-on-error path or a non-blocking rule tier) turns the build
// red here.

// buildLinter compiles the api-lint binary once into the test's temp dir and
// returns its path. Building (rather than `go run`) keeps each case a single
// fast exec and isolates the exit code from the go toolchain's own.
func buildLinter(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "api-lint")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	build := exec.Command("go", "build", "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build api-lint: %v\n%s", err, out)
	}
	return bin
}

// repoRoot resolves the repository root from this package (scripts/api-lint),
// matching the existing convention in registry_rules_test.go.
func repoRoot(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	return abs
}

func TestExitCode_FailsOnAnyViolationZeroOnCleanSpec(t *testing.T) {
	bin := buildLinter(t)
	root := repoRoot(t)
	cleanSpec := filepath.Join(root, "api", "openapi", "v1", "openapi.yaml")

	cases := []struct {
		name       string
		args       []string
		wantNonNil bool // true => expect a non-zero exit (violation gated)
	}{
		{
			// The live spec from the repo root must pass: every blocking rule
			// (spec + code, incl. pagination-codec) is 0 on a clean tree.
			name: "clean_spec_repo_root",
			args: []string{"-strict", cleanSpec, root},
		},
		{
			// PATH-BASE-PREFIX is a binding/structural guard.
			name:       "bad_path_base_prefix",
			args:       []string{filepath.Join("testdata", "path_base_prefix.openapi.yaml")},
			wantNonNil: true,
		},
		{
			// ENVELOPE-DRIFT is a formerly "reported-only" spec-drift family; it
			// must now gate the exit code with no -enforce escape hatch.
			name:       "bad_envelope_drift_gates",
			args:       []string{filepath.Join("testdata", "missing_problem.openapi.yaml")},
			wantNonNil: true,
		},
		{
			// AUTHZ-DRIFT — the other formerly-deferred family — must gate too.
			name:       "bad_authz_drift_gates",
			args:       []string{filepath.Join("testdata", "missing_security.openapi.yaml")},
			wantNonNil: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			cmd := exec.Command(bin, c.args...)
			out, err := cmd.CombinedOutput()
			exitErr, isExit := err.(*exec.ExitError)
			switch {
			case c.wantNonNil && err == nil:
				t.Fatalf("want non-zero exit on a known-bad spec, got exit 0\nargs=%v\noutput:\n%s", c.args, out)
			case c.wantNonNil && !isExit:
				t.Fatalf("want a clean non-zero process exit, got non-exit error %v\nargs=%v\noutput:\n%s", err, c.args, out)
			case c.wantNonNil && exitErr.ExitCode() != 1:
				t.Fatalf("want exit code 1 on a known-bad spec, got %d\nargs=%v\noutput:\n%s", exitErr.ExitCode(), c.args, out)
			case !c.wantNonNil && err != nil:
				t.Fatalf("want exit 0 on the clean spec, got %v\nargs=%v\noutput:\n%s", err, c.args, out)
			}
		})
	}
}
