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
