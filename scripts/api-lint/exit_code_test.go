package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
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
	dir, err := os.MkdirTemp("", "api-lint-test")
	if err != nil {
		t.Fatalf("mkdir temp: %v", err)
	}
	t.Cleanup(func() {
		// Windows may briefly hold a scan/exec lock on the just-built .exe, so a
		// plain RemoveAll (what t.TempDir does) can fail with "Access is denied".
		// Retry, then swallow a persistent failure rather than failing the test on
		// a cleanup race (a stray temp dir is harmless).
		for i := 0; i < 10; i++ {
			if rmErr := os.RemoveAll(dir); rmErr == nil {
				return
			}
			time.Sleep(200 * time.Millisecond)
		}
	})
	bin := filepath.Join(dir, "api-lint")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	build := exec.Command("go", "build", "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build api-lint: %v\n%s", err, out)
	}
	return bin
}

// runLinter execs the built binary, retrying transient Windows launch failures.
// Windows Defender holds an exclusive scan lock on a freshly written .exe, which
// surfaces as "Access is denied" on fork/exec; that is a launch failure, not a
// program result, so retry it. A real *exec.ExitError (the process ran and
// exited) is returned immediately.
func runLinter(bin string, args ...string) ([]byte, error) {
	var out []byte
	var err error
	for attempt := 0; attempt < 8; attempt++ {
		out, err = exec.Command(bin, args...).CombinedOutput()
		if err == nil {
			return out, nil
		}
		if _, isExit := err.(*exec.ExitError); isExit {
			return out, err
		}
		time.Sleep(150 * time.Millisecond)
	}
	return out, err
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
	t.Parallel()
	bin := buildLinter(t)
	root := repoRoot(t)
	cleanSpec := filepath.Join(root, "api", "openapi", "v1", "openapi.yaml")

	cases := []struct {
		name            string
		args            []string
		wantNonZeroExit bool // true => expect a non-zero exit (violation gated)
	}{
		{
			// The live spec from the repo root must pass: every blocking rule
			// (spec + code, incl. pagination-codec) is 0 on a clean tree.
			name: "clean_spec_repo_root",
			args: []string{"-strict", cleanSpec, root},
		},
		{
			// PATH-BASE-PREFIX is a binding/structural guard.
			name:            "bad_path_base_prefix",
			args:            []string{filepath.Join("testdata", "path_base_prefix.openapi.yaml")},
			wantNonZeroExit: true,
		},
		{
			// ENVELOPE-DRIFT is a formerly "reported-only" spec-drift family; it
			// must now gate the exit code with no -enforce escape hatch.
			name:            "bad_envelope_drift_gates",
			args:            []string{filepath.Join("testdata", "missing_problem.openapi.yaml")},
			wantNonZeroExit: true,
		},
		{
			// AUTHZ-DRIFT — the other formerly-deferred family — must gate too.
			name:            "bad_authz_drift_gates",
			args:            []string{filepath.Join("testdata", "missing_security.openapi.yaml")},
			wantNonZeroExit: true,
		},
		{
			// PAGINATION-DRIFT — the third formerly-deferred family — must gate too,
			// closing the CWE-693 guard coverage gap.
			name:            "bad_pagination_drift_gates",
			args:            []string{filepath.Join("testdata", "missing_cursor.openapi.yaml")},
			wantNonZeroExit: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			out, err := runLinter(bin, c.args...)
			exitErr, isExit := err.(*exec.ExitError)
			switch {
			case c.wantNonZeroExit && err == nil:
				t.Fatalf("want non-zero exit on a known-bad spec, got exit 0\nargs=%v\noutput:\n%s", c.args, out)
			case c.wantNonZeroExit && !isExit:
				t.Fatalf("want a clean non-zero process exit, got non-exit error %v\nargs=%v\noutput:\n%s", err, c.args, out)
			case c.wantNonZeroExit && isExit && exitErr.ExitCode() != 1:
				t.Fatalf("want exit code 1 on a known-bad spec, got %d\nargs=%v\noutput:\n%s", exitErr.ExitCode(), c.args, out)
			case !c.wantNonZeroExit && err != nil:
				t.Fatalf("want exit 0 on the clean spec, got %v\nargs=%v\noutput:\n%s", err, c.args, out)
			}
		})
	}
}
