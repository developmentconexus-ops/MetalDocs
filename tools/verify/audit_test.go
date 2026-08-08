package main

import (
	"os"
	"path/filepath"
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

	got := strings.Join(auditFindings(regs, jobs), "\n")

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
	if f := auditFindings(regs, jobs); len(f) != 0 {
		t.Errorf("want no findings, got %v", f)
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
	if f := auditFindings(regs, jobs); len(f) != 0 {
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

func TestAuditProfileInvocationSatisfiesCIJob(t *testing.T) {
	// idsForProfile(changed) is computed against the real global registry
	// (it has to be — it is not injectable), so this test points a copy of
	// that real registry's "go-build" entry at a fake job that runs
	// `--profile=changed`, and asserts no A1/A2/A3 finding mentions that job.
	// Using the real registry (rather than a two-entry fake one) avoids a
	// storm of spurious A1 "unknown ID" findings for every other real
	// pr-scoped check idsForProfile also returns.
	regs := make([]Check, len(checks))
	copy(regs, checks)
	for i, c := range regs {
		if c.ID == "go-build" {
			regs[i].CIJob = "sample.yml:verify"
		}
	}
	jobs := []workflowJob{
		{Workflow: "sample.yml", Job: "verify", OnlyIDs: idsForProfile(ProfileChanged)},
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
	for _, f := range auditFindings(regs, jobs) {
		if strings.Contains(f, "sample.yml:verify") {
			t.Errorf("want no finding referencing the profile-satisfied CIJob, got %q", f)
		}
	}
}
