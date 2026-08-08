package main

import (
	"os"
	"strings"
	"testing"
	"time"
)

// ---- TestHelperProcess: a re-exec'd instance of this same test binary,
// used as a fast, dependency-free stand-in for a real subprocess whose
// timing and side effects a test can control. Standard Go idiom (see
// os/exec_test.go). A no-op unless GO_WANT_HELPER_PROCESS=1 is set, so it is
// silent and harmless under a normal `go test ./tools/verify/...`.
//
// Argv shape: {os.Args[0], "-test.run=^TestHelperProcess$", "--", mode, arg...}
//   - mode=append  arg=<file> <id>            append <id>\n to <file>, exit 0
//   - mode=sleep   arg=<file> <id> <millis>    sleep, then append, exit 0
//   - mode=fail                                exit 1 immediately
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	args := os.Args
	for i, a := range args {
		if a == "--" {
			args = args[i+1:]
			break
		}
	}
	if len(args) == 0 {
		os.Exit(2)
	}
	switch args[0] {
	case "fail":
		os.Exit(1)
	case "sleep":
		file, id, millis := args[1], args[2], args[3]
		d, err := time.ParseDuration(millis + "ms")
		if err != nil {
			os.Exit(2)
		}
		time.Sleep(d)
		appendLine(file, id)
		os.Exit(0)
	case "append":
		file, id := args[1], args[2]
		appendLine(file, id)
		os.Exit(0)
	default:
		os.Exit(2)
	}
}

func appendLine(file, id string) {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		os.Exit(1)
	}
	defer f.Close()
	if _, err := f.WriteString(id + "\n"); err != nil {
		os.Exit(1)
	}
}

func helperArgv(mode string, rest ...string) []string {
	argv := []string{os.Args[0], "-test.run=^TestHelperProcess$", "--", mode}
	return append(argv, rest...)
}

func readLines(t *testing.T, file string) []string {
	t.Helper()
	b, err := os.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatal(err)
	}
	s := strings.TrimSpace(string(b))
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

// ---- ordering: After honoured, independent checks still concurrent -------

func TestRunHonoursAfterOrdering(t *testing.T) {
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	file := t.TempDir() + "/order.log"

	build := Check{ID: "fixture-build", Argv: helperArgv("sleep", file, "fixture-build", "50")}
	test := Check{ID: "fixture-test", Argv: helperArgv("append", file, "fixture-test"), After: []string{"fixture-build"}}

	// Dependent listed first in the slice on purpose — ordering must come
	// from the After edge, not from slice position.
	got := run([]Check{test, build}, 4, false, false)
	for _, r := range got {
		if r.status != statusPass {
			t.Fatalf("check %s: want PASS, got %s (%s)", r.check.ID, r.status, r.reason)
		}
	}

	lines := readLines(t, file)
	if strings.Join(lines, ",") != "fixture-build,fixture-test" {
		t.Fatalf("want fixture-build before fixture-test, got %v", lines)
	}
}

func TestRunKeepsIndependentChecksConcurrent(t *testing.T) {
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	file := t.TempDir() + "/concurrent.log"

	a := Check{ID: "fixture-a", Argv: helperArgv("sleep", file, "fixture-a", "150")}
	b := Check{ID: "fixture-b", Argv: helperArgv("sleep", file, "fixture-b", "150")}

	start := time.Now()
	got := run([]Check{a, b}, 2, false, false)
	elapsed := time.Since(start)

	for _, r := range got {
		if r.status != statusPass {
			t.Fatalf("check %s: want PASS, got %s", r.check.ID, r.status)
		}
	}
	// Serialized would take >= 300ms; concurrent should finish well under
	// that. Generous margin for CI jitter.
	if elapsed >= 280*time.Millisecond {
		t.Fatalf("independent checks did not run concurrently: took %s", elapsed)
	}
}

func TestRunDoesNotRunDependentWhenPredecessorFails(t *testing.T) {
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	file := t.TempDir() + "/never-written.log"

	build := Check{ID: "fixture-build-fails", Argv: helperArgv("fail")}
	test := Check{ID: "fixture-test-blocked", Argv: helperArgv("append", file, "fixture-test-blocked"), After: []string{"fixture-build-fails"}}

	got := run([]Check{build, test}, 2, false, false)

	byID := map[string]result{}
	for _, r := range got {
		byID[r.check.ID] = r
	}
	if byID["fixture-build-fails"].status != statusFail {
		t.Fatalf("predecessor: want FAIL, got %s", byID["fixture-build-fails"].status)
	}
	dep := byID["fixture-test-blocked"]
	if dep.status == statusPass {
		t.Fatalf("dependent must never be reported as PASS when its predecessor failed; got %+v", dep)
	}
	if dep.status != statusSkip {
		t.Fatalf("dependent: want SKIP (not-run), got %s", dep.status)
	}
	if dep.reason == "" {
		t.Error("a skipped dependent must carry a reason naming the failed predecessor")
	}
	if lines := readLines(t, file); len(lines) != 0 {
		t.Fatalf("dependent's Argv must never have executed; file has %v", lines)
	}
}

// ---- validateOrdering: startup failures, not deadlocks or silent no-ops --

func TestValidateOrderingRejectsUnknownEdge(t *testing.T) {
	err := validateOrdering([]Check{
		{ID: "a", After: []string{"does-not-exist"}},
	})
	if err == nil {
		t.Fatal("want an error for an After edge naming an unregistered check ID")
	}
	if !strings.Contains(err.Error(), "does-not-exist") {
		t.Errorf("error should name the unknown ID, got: %v", err)
	}
}

func TestValidateOrderingRejectsCycle(t *testing.T) {
	err := validateOrdering([]Check{
		{ID: "a", After: []string{"b"}},
		{ID: "b", After: []string{"a"}},
	})
	if err == nil {
		t.Fatal("want an error for a cycle in After edges")
	}
	if !strings.Contains(err.Error(), "cycle") {
		t.Errorf("error should say this is a cycle, got: %v", err)
	}
}

func TestValidateOrderingAcceptsRealRegistry(t *testing.T) {
	if err := validateOrdering(checks); err != nil {
		t.Fatalf("the real registry must have a valid ordering graph: %v", err)
	}
}

func TestValidateOrderingAcceptsDAG(t *testing.T) {
	err := validateOrdering([]Check{
		{ID: "build"},
		{ID: "test", After: []string{"build"}},
		{ID: "lint"},
	})
	if err != nil {
		t.Fatalf("a valid DAG must not be rejected: %v", err)
	}
}

// ---- excluded-predecessor ruling: run anyway, but never silently ---------

func TestExcludedPredecessorWarnsAndDoesNotBlock(t *testing.T) {
	selected := []Check{
		{ID: "fixture-test-only", After: []string{"fixture-build-excluded"}},
	}
	warns := excludedPredecessorWarnings(selected)
	if len(warns) != 1 {
		t.Fatalf("want one warning about the excluded predecessor, got %v", warns)
	}
	if !strings.Contains(warns[0], "fixture-build-excluded") || !strings.Contains(warns[0], "fixture-test-only") {
		t.Errorf("warning should name both checks, got %q", warns[0])
	}
}

func TestExcludedPredecessorNoWarningWhenBothSelected(t *testing.T) {
	selected := []Check{
		{ID: "fixture-build"},
		{ID: "fixture-test", After: []string{"fixture-build"}},
	}
	if warns := excludedPredecessorWarnings(selected); len(warns) != 0 {
		t.Errorf("want no warning when the predecessor is in the same selection, got %v", warns)
	}
}

// The excluded-predecessor case must not deadlock run(): a dependent whose
// predecessor is missing from the selection has no done-channel to wait on
// and must run immediately.
func TestRunExcludedPredecessorRunsAnyway(t *testing.T) {
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	file := t.TempDir() + "/ran-anyway.log"
	test := Check{ID: "fixture-test-solo", Argv: helperArgv("append", file, "fixture-test-solo"), After: []string{"fixture-build-not-selected"}}

	got := run([]Check{test}, 1, false, false)
	if len(got) != 1 || got[0].status != statusPass {
		t.Fatalf("want the dependent to run when its predecessor is excluded, got %+v", got)
	}
}

// A check that cannot run must not be reported as a hole in the claim when
// --require-infra is set. In CI, "not verified" and "verified" must not both
// exit 0, because the gate cannot tell them apart.
func TestRequireInfraTurnsSkipIntoFail(t *testing.T) {
	t.Setenv("METALDOCS_DATABASE_URL", "")

	needsDB := Check{
		ID:    "fixture-needs-postgres",
		Desc:  "fixture",
		Argv:  []string{"go", "version"},
		Needs: []string{needsPostgres},
	}

	got := run([]Check{needsDB}, 1, false, false)
	if len(got) != 1 || got[0].status != statusSkip {
		t.Fatalf("without --require-infra: want one SKIP, got %+v", got)
	}

	got = run([]Check{needsDB}, 1, false, true)
	if len(got) != 1 || got[0].status != statusFail {
		t.Fatalf("with --require-infra: want one FAIL, got %+v", got)
	}
	if got[0].reason == "" {
		t.Error("the FAIL must carry the same reason the SKIP carried; an unexplained FAIL is worse than a SKIP")
	}
	if report(got, ProfileFull) != 1 {
		t.Error("report() must exit non-zero when infra is required and missing")
	}
}

// The absence of the flag must not change existing behaviour.
func TestRunnableCheckPassesUnderRequireInfra(t *testing.T) {
	ok := Check{ID: "fixture-runnable", Desc: "fixture", Argv: []string{"go", "version"}}
	got := run([]Check{ok}, 1, false, true)
	if len(got) != 1 || got[0].status != statusPass {
		t.Fatalf("want one PASS, got %+v", got)
	}
}
