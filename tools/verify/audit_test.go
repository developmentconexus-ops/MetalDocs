package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func writeWorkflow(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestParseWorkflowsExtractsOnlyIDsAndNeeds(t *testing.T) {
	dir := t.TempDir()
	writeWorkflow(t, dir, "sample.yml", `
name: sample
jobs:
  lint-go:
    runs-on: ubuntu-latest
    steps:
      - run: go run ./tools/verify --only=gofmt,go-vet --require-infra
  required:
    needs: [lint-go]
    runs-on: ubuntu-latest
    steps:
      - run: echo done
`)

	jobs, err := parseWorkflows(dir)
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]workflowJob{}
	for _, j := range jobs {
		byName[j.Workflow+":"+j.Job] = j
	}

	lint, ok := byName["sample.yml:lint-go"]
	if !ok {
		t.Fatalf("lint-go not parsed; got %v", byName)
	}
	if strings.Join(lint.OnlyIDs, ",") != "gofmt,go-vet" {
		t.Errorf("OnlyIDs = %v, want [gofmt go-vet]", lint.OnlyIDs)
	}
	req := byName["sample.yml:required"]
	if strings.Join(req.Needs, ",") != "lint-go" {
		t.Errorf("Needs = %v, want [lint-go]", req.Needs)
	}
}

func TestAuditRules(t *testing.T) {
	regs := []Check{
		{ID: "gofmt", CIJob: "sample.yml:lint-go"},
		{ID: "orphan", CIJob: "gone.yml:vanished"},
		{ID: "unclaimed", CIJob: ""},
		{ID: "not-run", CIJob: "sample.yml:lint-go"},
	}
	jobs := []workflowJob{
		{Workflow: "sample.yml", Job: "lint-go", OnlyIDs: []string{"gofmt", "typo-id"}},
	}

	got := strings.Join(auditFindings(regs, jobs, nil), "\n")

	for _, want := range []string{
		"typo-id",   // A1: workflow runs an ID the registry does not have
		"gone.yml",  // A2: registry points at a job that does not exist
		"not-run",   // A3: the named job does not run the check
		"unclaimed", // A4: no CI job at all
	} {
		if !strings.Contains(got, want) {
			t.Errorf("findings missing %q; got:\n%s", want, got)
		}
	}
}

func TestAuditCleanRegistryHasNoFindings(t *testing.T) {
	regs := []Check{{ID: "gofmt", CIJob: "sample.yml:lint-go"}}
	jobs := []workflowJob{{Workflow: "sample.yml", Job: "lint-go", OnlyIDs: []string{"gofmt"}}}
	if f := auditFindings(regs, jobs, nil); len(f) != 0 {
		t.Errorf("want no findings, got %v", f)
	}
}

// ---- A5: ci.yml `required`'s needs: vs scripts/required-gate.jq's keys ----

func TestAuditA5FiresOnDesyncedRequiredNeeds(t *testing.T) {
	jobs := []workflowJob{
		{Workflow: "ci.yml", Job: "required", Needs: []string{"verify", "test-integration", "security", "lint-go", "extra-job"}},
	}
	got := strings.Join(auditFindings(nil, jobs, []string{"verify", "test-integration", "security", "lint-go"}), "\n")
	if !strings.Contains(got, "A5") {
		t.Fatalf("want an A5 finding when needs: and required-gate.jq disagree, got:\n%s", got)
	}
	if !strings.Contains(got, "extra-job") {
		t.Errorf("finding should name the job causing the mismatch, got:\n%s", got)
	}
}

func TestAuditA5CleanWhenSynced(t *testing.T) {
	jobs := []workflowJob{
		{Workflow: "ci.yml", Job: "required", Needs: []string{"verify", "test-integration", "security", "lint-go"}},
	}
	got := auditFindings(nil, jobs, []string{"lint-go", "security", "test-integration", "verify"})
	for _, f := range got {
		if strings.Contains(f, "A5") {
			t.Errorf("want no A5 finding when needs: and required-gate.jq agree (order-independent), got %v", got)
		}
	}
}

// A5 must only look at ci.yml's `required` job — a job that happens to be
// named `required` in some other workflow, or a `required` job during a
// synthetic test that has nothing to do with the real gate, must not fire.
func TestAuditA5IgnoresUnrelatedRequiredJobs(t *testing.T) {
	jobs := []workflowJob{
		{Workflow: "sample.yml", Job: "required", Needs: []string{"anything"}},
	}
	got := auditFindings(nil, jobs, []string{"verify"})
	for _, f := range got {
		if strings.Contains(f, "A5") {
			t.Errorf("A5 must be scoped to ci.yml:required, got %v", got)
		}
	}
}

func TestParseRequiredGateKeysReadsTheRealGateFile(t *testing.T) {
	// scripts/required-gate.jq relative to the repo root; go test runs with
	// the package directory as cwd, so walk up.
	keys, err := parseRequiredGateKeys(filepath.Join("..", "..", "scripts", "required-gate.jq"))
	if err != nil {
		t.Fatalf("parseRequiredGateKeys: %v", err)
	}
	want := []string{"lint-go", "security", "test-integration", "verify"}
	got := append([]string(nil), keys...)
	sort.Strings(got)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// A job may run a check that names a different CIJob. Phase 2 runs every
// check twice on purpose, and the audit must not call that a defect.
func TestAuditAllowsDuplicateInvocation(t *testing.T) {
	regs := []Check{{ID: "gofmt", CIJob: "old.yml:legacy"}}
	jobs := []workflowJob{
		{Workflow: "old.yml", Job: "legacy", OnlyIDs: []string{"gofmt"}},
		{Workflow: "ci.yml", Job: "lint-go", OnlyIDs: []string{"gofmt"}},
	}
	if f := auditFindings(regs, jobs, nil); len(f) != 0 {
		t.Errorf("duplicate invocation must be allowed during Phase 2, got %v", f)
	}
}

// A job selecting its checks with --profile= instead of --only= (ci.yml's
// verify job runs `--profile=changed`) still satisfies every check whose
// CIJob points at it, for every check that profile's membership includes.
func TestParseWorkflowsResolvesProfileFlag(t *testing.T) {
	dir := t.TempDir()
	writeWorkflow(t, dir, "sample.yml", `
name: sample
jobs:
  verify:
    runs-on: ubuntu-latest
    steps:
      - run: go run ./tools/verify --require-infra --profile=changed
`)
	jobs, err := parseWorkflows(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 {
		t.Fatalf("want 1 job, got %d", len(jobs))
	}
	got := map[string]bool{}
	for _, id := range jobs[0].OnlyIDs {
		got[id] = true
	}
	// go-build declares {fast, pr, full} — a member of `pr`, so `changed`
	// (which resolves to `pr`) must include it.
	if !got["go-build"] {
		t.Errorf("expected --profile=changed to resolve to the pr set including go-build; got %v", jobs[0].OnlyIDs)
	}
	// go-test-integration is `full`-only, never `pr`, so `changed` must not
	// claim it — that would let a check with no PR-time coverage read as
	// audited when it is not.
	if got["go-test-integration"] {
		t.Errorf("--profile=changed must not resolve to full-only checks; got %v", jobs[0].OnlyIDs)
	}
}

// ---- A6: ProfilePR check's CIJob must be inside ci.yml:required's needs: closure ----

func TestAuditA6FiresOnCIJobOutsideRequiredClosure(t *testing.T) {
	regs := []Check{
		{ID: "orphan-pr-check", Profiles: []string{ProfilePR}, CIJob: "side.yml:node"},
	}
	jobs := []workflowJob{
		{Workflow: "side.yml", Job: "node", OnlyIDs: []string{"orphan-pr-check"}},
		{Workflow: "ci.yml", Job: "verify", OnlyIDs: nil, Needs: nil},
		{Workflow: "ci.yml", Job: "required", Needs: []string{"verify"}},
	}
	got := strings.Join(auditFindings(regs, jobs, nil), "\n")
	if !strings.Contains(got, "A6") || !strings.Contains(got, "orphan-pr-check") {
		t.Fatalf("want an A6 finding naming orphan-pr-check (its CIJob side.yml:node is outside ci.yml:required's closure), got:\n%s", got)
	}
}

func TestAuditA6CleanWhenCIJobInsideClosure(t *testing.T) {
	regs := []Check{
		{ID: "in-closure-check", Profiles: []string{ProfilePR}, CIJob: "ci.yml:verify"},
	}
	jobs := []workflowJob{
		{Workflow: "ci.yml", Job: "verify", OnlyIDs: []string{"in-closure-check"}},
		{Workflow: "ci.yml", Job: "required", Needs: []string{"verify"}},
	}
	got := auditFindings(regs, jobs, nil)
	for _, f := range got {
		if strings.Contains(f, "A6") {
			t.Errorf("want no A6 finding when CIJob is inside the closure, got %v", got)
		}
	}
}

// A6 must not fire for a check that is not in ProfilePR at all (e.g. a
// full-only integration suite) — its CIJob is allowed to be a job that never
// reports to `required`, because it was never claimed to gate a merge.
func TestAuditA6IgnoresNonPRProfileChecks(t *testing.T) {
	regs := []Check{
		{ID: "full-only-check", Profiles: []string{ProfileFull}, CIJob: "side.yml:node"},
	}
	jobs := []workflowJob{
		{Workflow: "side.yml", Job: "node", OnlyIDs: []string{"full-only-check"}},
		{Workflow: "ci.yml", Job: "required", Needs: nil},
	}
	got := auditFindings(regs, jobs, nil)
	for _, f := range got {
		if strings.Contains(f, "A6") {
			t.Errorf("A6 must not fire for a non-ProfilePR check, got %v", got)
		}
	}
}

func TestAuditProfileInvocationSatisfiesCIJob(t *testing.T) {
	// idsForProfile(changed) is computed against the real global registry
	// (it has to be — it is not injectable), so this test points a copy of
	// that real registry's "go-build" entry at a fake job that runs
	// `--profile=changed`, and asserts no A1/A2/A3 finding mentions that job.
	// Using the real registry (rather than a two-entry fake one) avoids a
	// storm of spurious A1 "unknown ID" findings for every other real
	// pr-scoped check idsForProfile also returns.
	// Workflow/job named "ci.yml:verify" (not a "sample.yml" stand-in): A6
	// requires every ProfilePR check's CIJob to be inside ci.yml:required's
	// needs: closure, so the fake job here has to actually live in that
	// closure for this test to isolate what it means to test (A1/A2/A3
	// satisfaction via --profile=), rather than tripping A6 as a side effect.
	regs := make([]Check, len(checks))
	copy(regs, checks)
	for i, c := range regs {
		if c.ID == "go-build" {
			regs[i].CIJob = "ci.yml:verify"
		}
	}
	jobs := []workflowJob{
		{Workflow: "ci.yml", Job: "verify", OnlyIDs: idsForProfile(ProfileChanged)},
		{Workflow: "ci.yml", Job: "required", Needs: []string{"verify"}},
	}
	found := false
	for _, id := range jobs[0].OnlyIDs {
		if id == "go-build" {
			found = true
		}
	}
	if !found {
		t.Fatalf("idsForProfile(changed) does not include go-build; got %v", jobs[0].OnlyIDs)
	}
	for _, f := range auditFindings(regs, jobs, nil) {
		if strings.Contains(f, " go-build ") {
			t.Errorf("want no finding for go-build (profile-satisfied CIJob, inside the closure), got %q", f)
		}
	}
}
