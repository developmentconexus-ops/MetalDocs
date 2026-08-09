package main

import (
	"os"
	"strconv"
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
//   - mode=append   arg=<file> <id>                       append <id>\n to <file>, exit 0
//   - mode=sleep    arg=<file> <id> <millis>               sleep, then append, exit 0
//   - mode=fail                                            exit 1 immediately
//   - mode=barrier  arg=<arrivedFile> <doneFile> <id> <n>  see below
//
// barrier proves genuine concurrency without measuring duration: it appends
// <id> to arrivedFile, then polls arrivedFile until it holds >= n lines (n
// being the number of peers in the barrier), then appends <id> to doneFile
// and exits 0. A peer can only ever observe n arrivals if all n peers were
// spawned and got far enough to write their own line — which is impossible
// if the runner spawned them one-after-another and waited for each to exit
// before starting the next, because the first spawned would still be
// blocked *waiting* on the others' arrivals (not yet exited) when the next
// would need to start. So reaching doneFile is direct proof of overlap, not
// an inference from how long anything took. A capped poll deadline guards
// against a genuine hang (e.g. the runner regresses to true serialization)
// turning this into an indefinite hang; it is deliberately generous (10s)
// so it is a deadlock backstop, not a race window — crossing it is only
// possible when the checks truly never overlapped.
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
	case "barrier":
		arrivedFile, doneFile, id, nStr := args[1], args[2], args[3], args[4]
		n, err := strconv.Atoi(nStr)
		if err != nil {
			os.Exit(2)
		}
		appendLine(arrivedFile, id)
		deadline := time.Now().Add(10 * time.Second)
		for countLines(arrivedFile) < n {
			if time.Now().After(deadline) {
				// Never observed every peer arrive: they did not overlap in
				// time. Reported as a normal check failure (non-zero exit),
				// not a test-framework panic, so it surfaces through the
				// same PASS/FAIL path every other check result does.
				os.Exit(3)
			}
			time.Sleep(5 * time.Millisecond)
		}
		appendLine(doneFile, id)
		os.Exit(0)
	default:
		os.Exit(2)
	}
}

// countLines reports how many non-empty lines a file has, treating a
// missing file as zero. Used by the barrier helper mode above; kept
// separate from readLines because that one takes a *testing.T and this
// runs inside a re-exec'd subprocess with no testing context.
func countLines(file string) int {
	b, err := os.ReadFile(file)
	if err != nil {
		return 0
	}
	s := strings.TrimSpace(string(b))
	if s == "" {
		return 0
	}
	return len(strings.Split(s, "\n"))
}

func appendLine(file, id string) {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		os.Exit(1)
	}
	// No defer here: this function's every exit path is os.Exit, which
	// terminates the process before a deferred call would ever run — so the
	// close (and its error) has to be handled explicitly on each path
	// instead.
	if _, err := f.WriteString(id + "\n"); err != nil {
		_ = f.Close()
		os.Exit(1)
	}
	if err := f.Close(); err != nil {
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

// TestRunKeepsIndependentChecksConcurrent guards a real property of the
// check runner: two checks with no After edge between them must be
// scheduled without one waiting on the other. That matters because this
// registry check runs inside a blocking required CI gate — if independent
// checks silently serialized, a profile sized (and time-budgeted) for
// concurrent execution would blow past CI's timeout, and a flaky assertion
// here would teach people to re-run the gate until green, which defeats the
// point of a single required check.
//
// The previous version inferred concurrency from wall-clock duration
// (two 150ms sleeps, asserting elapsed < 280ms). That inference's premise —
// that re-exec'ing this test binary as a subprocess costs much less than
// 150ms — does not hold everywhere: reproduced locally, both helpers
// consistently finished together (proving they DID run concurrently: same
// individual duration, same finish time) while each individual invocation
// cost 500-900ms of pure process-spawn overhead, pushing total elapsed past
// the fixed threshold even though scheduling was correct. That is a
// too-slow environmental flake against an assumed-cheap constant, not a
// scheduling defect — but a wall-clock threshold cannot tell the two apart,
// which is exactly the failure mode a required gate cannot carry.
//
// This version proves overlap directly instead of inferring it from
// duration: see the "barrier" mode on TestHelperProcess. Neither helper can
// reach doneFile without having observed both arrivals, and neither could
// arrive-then-observe-both unless the other had also already started — a
// spawn-one-wait-then-spawn-the-other execution would leave the first
// helper blocked at the barrier (eventually timing out and FAILing) rather
// than silently reporting a false PASS. So there is no duration threshold
// left to cross, in either direction, under load or on a fast machine.
func TestRunKeepsIndependentChecksConcurrent(t *testing.T) {
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	dir := t.TempDir()
	arrived := dir + "/arrived.log"
	done := dir + "/done.log"

	a := Check{ID: "fixture-a", Argv: helperArgv("barrier", arrived, done, "fixture-a", "2")}
	b := Check{ID: "fixture-b", Argv: helperArgv("barrier", arrived, done, "fixture-b", "2")}

	got := run([]Check{a, b}, 2, false, false)
	for _, r := range got {
		if r.status != statusPass {
			t.Fatalf("check %s: want PASS, got %s (%s)\n%s", r.check.ID, r.status, r.reason, r.output)
		}
	}

	doneIDs := map[string]bool{}
	for _, l := range readLines(t, done) {
		doneIDs[l] = true
	}
	if !doneIDs["fixture-a"] || !doneIDs["fixture-b"] {
		t.Fatalf("want both fixture-a and fixture-b past the barrier, got %v", doneIDs)
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

// ---- excluded-predecessor ruling: refuse rather than run over stale state -
//
// The prior ruling here was "run the dependent anyway, warn loudly, do not
// refuse." A reproduction proved that fail-open: `dist/meta.json` already
// existed on disk from an earlier, unrelated build, and
// `--only=docx-v2-test` ran the bundle-guard test anyway — it read the
// stale file and reported PASS without ever re-validating current source.
// The ruling is now: a selection that includes a check without its After
// predecessor is refused, not run. See validateSelectionOrdering (selection
// boundary, tools/verify/main.go) and run()'s excluded-predecessor branch
// (execution kernel, same file) — both enforce it, independently, so a
// future caller that assembles a selection some other way than
// selectChecks still cannot slip an incomplete one past run().

func TestValidateSelectionOrderingRejectsExcludedPredecessor(t *testing.T) {
	selected := []Check{
		{ID: "fixture-test-only", After: []string{"fixture-build-excluded"}},
	}
	err := validateSelectionOrdering(selected)
	if err == nil {
		t.Fatal("want an error: fixture-test-only's After predecessor is not in the selection")
	}
	if !strings.Contains(err.Error(), "fixture-build-excluded") || !strings.Contains(err.Error(), "fixture-test-only") {
		t.Errorf("error should name both checks, got %q", err.Error())
	}
}

func TestValidateSelectionOrderingAcceptsCompleteSelection(t *testing.T) {
	selected := []Check{
		{ID: "fixture-build"},
		{ID: "fixture-test", After: []string{"fixture-build"}},
	}
	if err := validateSelectionOrdering(selected); err != nil {
		t.Errorf("want no error when the predecessor is in the same selection, got %v", err)
	}
}

// This is the reviewer's exact reproduction, at the unit level: selecting
// docx-v2-test alone, via the real registry, through the same selectChecks
// path `--only=docx-v2-test` goes through. Before the fix this returned the
// check with no error, and main() ran it anyway over whatever
// dist/meta.json happened to be on disk.
func TestSelectChecksRefusesOnlyDocxV2TestAlone(t *testing.T) {
	_, _, err := selectChecks(ProfileFast, "docx-v2-test", "origin/main", false)
	if err == nil {
		t.Fatal("want an error: --only=docx-v2-test excludes its After predecessor docx-v2-build")
	}
	if !strings.Contains(err.Error(), "docx-v2-build") {
		t.Errorf("error should name the missing predecessor docx-v2-build, got %v", err)
	}
}

// selectChecks must still resolve a complete pair without error — the fix
// must not make an ordinary, order-complete selection refuse.
func TestSelectChecksAcceptsDocxV2BuildAndTestTogether(t *testing.T) {
	selected, _, err := selectChecks(ProfileFast, "docx-v2-build,docx-v2-test", "origin/main", false)
	if err != nil {
		t.Fatalf("want no error when both sides of the edge are selected, got %v", err)
	}
	if len(selected) != 2 {
		t.Fatalf("want both checks selected, got %d", len(selected))
	}
}

// The excluded-predecessor case must not deadlock run(): a dependent whose
// predecessor is missing from the selection has no done-channel to wait on,
// and must be refused (FAIL, with a reason) rather than block forever or
// run unguarded.
func TestRunRefusesWhenPredecessorExcluded(t *testing.T) {
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	file := t.TempDir() + "/never-written.log"
	test := Check{ID: "fixture-test-solo", Argv: helperArgv("append", file, "fixture-test-solo"), After: []string{"fixture-build-not-selected"}}

	got := run([]Check{test}, 1, false, false)
	if len(got) != 1 {
		t.Fatalf("want one result, got %+v", got)
	}
	if got[0].status == statusPass {
		t.Fatalf("must never report PASS when the After predecessor is excluded from the selection; got %+v", got[0])
	}
	if got[0].status != statusFail {
		t.Fatalf("want FAIL (a SKIP would exit 0 and hide the refusal), got %s", got[0].status)
	}
	if got[0].reason == "" {
		t.Error("the FAIL must carry a reason naming the missing predecessor")
	}
	if !strings.Contains(got[0].reason, "fixture-build-not-selected") {
		t.Errorf("reason should name the missing predecessor, got %q", got[0].reason)
	}
	if lines := readLines(t, file); len(lines) != 0 {
		t.Fatalf("the check's Argv must never have executed; file has %v", lines)
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
