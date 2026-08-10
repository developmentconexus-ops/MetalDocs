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
	// go-vet declares {fast, pr, full} — a member of `pr`, so `changed`
	// (which resolves to `pr`) must include it.
	if !got["go-vet"] {
		t.Errorf("expected --profile=changed to resolve to the pr set including go-vet; got %v", jobs[0].OnlyIDs)
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
	// that real registry's "go-vet" entry at a fake job that runs
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
		if c.ID == "go-vet" {
			regs[i].CIJob = "ci.yml:verify"
		}
	}
	jobs := []workflowJob{
		{Workflow: "ci.yml", Job: "verify", OnlyIDs: idsForProfile(ProfileChanged)},
		{Workflow: "ci.yml", Job: "required", Needs: []string{"verify"}},
	}
	found := false
	for _, id := range jobs[0].OnlyIDs {
		if id == "go-vet" {
			found = true
		}
	}
	if !found {
		t.Fatalf("idsForProfile(changed) does not include go-vet; got %v", jobs[0].OnlyIDs)
	}
	for _, f := range auditFindings(regs, jobs, nil) {
		if strings.Contains(f, " go-vet ") {
			t.Errorf("want no finding for go-vet (profile-satisfied CIJob, inside the closure), got %q", f)
		}
	}
}

// ---- A7: waiver kinds are a closed set ----

// A waiver used to be free prose, and prose let three repo-authored guards
// declare themselves uncovered ("TRANSITIONAL, no fixture yet") while A7
// counted them as covered. The kind is what closes that: a check that fits no
// admissible kind needs a fixture, not a new kind.
func TestAuditA7RejectsWaiverKindOutsideTheEnum(t *testing.T) {
	regs := []Check{{
		ID:            "invented",
		Profiles:      []string{ProfilePR},
		CIJob:         "sample.yml:verify",
		FixtureWaiver: &Waiver{Kind: "transitional", Why: "no fixture yet"},
	}}
	jobs := []workflowJob{{Workflow: "sample.yml", Job: "verify", OnlyIDs: []string{"invented"}}}

	got := strings.Join(auditFindings(regs, jobs, nil), "\n")
	if !strings.Contains(got, `A7 check "invented" has FixtureWaiver kind "transitional"`) {
		t.Errorf("A7 did not reject an invented waiver kind; got:\n%s", got)
	}
}

func TestAuditA7RequiresWhyAlongsideKind(t *testing.T) {
	regs := []Check{{
		ID:            "bare",
		Profiles:      []string{ProfilePR},
		CIJob:         "sample.yml:verify",
		FixtureWaiver: &Waiver{Kind: WaiverThirdParty, Why: "   "},
	}}
	jobs := []workflowJob{{Workflow: "sample.yml", Job: "verify", OnlyIDs: []string{"bare"}}}

	got := strings.Join(auditFindings(regs, jobs, nil), "\n")
	if !strings.Contains(got, `A7 check "bare" has a FixtureWaiver with no Why`) {
		t.Errorf("A7 accepted a waiver with no rationale; got:\n%s", got)
	}
}

func TestAuditA7AcceptsAClassifiedWaiver(t *testing.T) {
	regs := []Check{{
		ID:            "vendor",
		Profiles:      []string{ProfilePR},
		CIJob:         "sample.yml:verify",
		FixtureWaiver: &Waiver{Kind: WaiverThirdParty, Why: "pinned upstream linter; a fixture would test its rule engine"},
	}}
	jobs := []workflowJob{{Workflow: "sample.yml", Job: "verify", OnlyIDs: []string{"vendor"}}}

	for _, f := range auditFindings(regs, jobs, nil) {
		if strings.HasPrefix(f, "A7") {
			t.Errorf("A7 fired on a properly classified waiver: %s", f)
		}
	}
}

// ---- A9 containers, A10 closure gates (#87/A1 review B1 + B2) -------------

// closureJobs builds the smallest job set that puts `probe` inside
// ci.yml:required's needs: closure, so the A10 tests below exercise the rule
// rather than the closure computation.
func closureJobs(probe workflowJob) []workflowJob {
	probe.Workflow, probe.Job = "ci.yml", "probe"
	return []workflowJob{
		{Workflow: "ci.yml", Job: "required", Needs: []string{"probe"}},
		probe,
	}
}

func TestAuditA10FiresOnUnclassifiedStepInRequiredClosure(t *testing.T) {
	jobs := closureJobs(workflowJob{Steps: []workflowStep{
		{Name: "prereq: checkout", Uses: "actions/checkout@" + strings.Repeat("a", 40)},
		{Name: "Run gitleaks", Run: `docker run --rm gitleaks@sha256:` + strings.Repeat("b", 64) + ` detect --exit-code 1`},
	}})

	got := strings.Join(auditFindings(nil, jobs, nil), "\n")
	if !strings.Contains(got, "A10 ci.yml:probe step 2 (Run gitleaks)") {
		t.Errorf("A10 did not reject an unregistered blocking gate inside the required closure; got:\n%s", got)
	}
}

func TestAuditA10AcceptsVerifyPrereqAndReportSteps(t *testing.T) {
	jobs := closureJobs(workflowJob{Steps: []workflowStep{
		{Name: "prereq: checkout", Uses: "actions/checkout@" + strings.Repeat("a", 40)},
		{Name: "verify (secret-scan)", Run: "go run ./tools/verify --require-infra --ci-job=ci.yml:probe --profile=pr"},
		{Name: "report: upload SARIF", Uses: "github/codeql-action/upload-sarif@" + strings.Repeat("c", 40)},
	}})

	for _, f := range auditFindings(nil, jobs, nil) {
		if strings.HasPrefix(f, "A10") {
			t.Errorf("A10 fired on a fully classified job: %s", f)
		}
	}
}

func TestAuditA10IgnoresJobsOutsideTheRequiredClosure(t *testing.T) {
	jobs := []workflowJob{
		{Workflow: "ci.yml", Job: "required", Needs: []string{"verify"}},
		{Workflow: "ci.yml", Job: "verify", Steps: []workflowStep{
			{Name: "verify", Run: "go run ./tools/verify --profile=changed"},
		}},
		// nightly runs plenty of things nobody's merge waits on.
		{Workflow: "nightly.yml", Job: "scan", Steps: []workflowStep{{Name: "scan", Run: "trivy fs ."}}},
	}

	for _, f := range auditFindings(nil, jobs, nil) {
		if strings.HasPrefix(f, "A10") {
			t.Errorf("A10 fired outside the required closure: %s", f)
		}
	}
}

func TestAuditA9FiresOnMutableServiceImage(t *testing.T) {
	jobs := []workflowJob{{Workflow: "ci.yml", Job: "test-integration", Images: []string{"postgres:16"}}}

	got := strings.Join(auditFindings(nil, jobs, nil), "\n")
	if !strings.Contains(got, `A9: ci.yml:test-integration runs service container "postgres:16"`) {
		t.Errorf("A9 did not reject a tag-pinned service container; got:\n%s", got)
	}
}

func TestAuditA9AcceptsDigestPinnedServiceImage(t *testing.T) {
	jobs := []workflowJob{{Workflow: "ci.yml", Job: "test-integration",
		Images: []string{"postgres@sha256:" + strings.Repeat("0", 64)}}}

	for _, f := range auditFindings(nil, jobs, nil) {
		if strings.HasPrefix(f, "A9") {
			t.Errorf("A9 fired on a digest-pinned service container: %s", f)
		}
	}
}

func TestAuditA9FiresOnMutableDockerRunInWorkflowStep(t *testing.T) {
	jobs := closureJobs(workflowJob{Steps: []workflowStep{{
		Name: "prereq: scan",
		Run:  `docker run --rm -v "$PWD:/repo" ghcr.io/gitleaks/gitleaks:v8.24.3 detect --source /repo`,
	}}})

	got := strings.Join(auditFindings(nil, jobs, nil), "\n")
	if !strings.Contains(got, `runs container "ghcr.io/gitleaks/gitleaks:v8.24.3"`) {
		t.Errorf("A9 did not reject a tag-pinned docker run; got:\n%s", got)
	}
}

func TestAuditA9FiresOnMutableDockerRunInRegistryArgv(t *testing.T) {
	regs := []Check{{
		ID:            "container-check",
		Profiles:      []string{ProfilePR},
		CIJob:         "ci.yml:probe",
		FixtureWaiver: &Waiver{Kind: WaiverThirdParty, Why: "pinned upstream scanner"},
		Argv:          []string{"docker", "run", "--rm", "-v", repoRootPlaceholder + ":/src", "anchore/grype:v0.116.1", "dir:/src"},
	}}
	jobs := closureJobs(workflowJob{Steps: []workflowStep{
		{Name: "verify", Run: "go run ./tools/verify --ci-job=ci.yml:probe --profile=pr"},
	}})

	got := strings.Join(auditFindings(regs, jobs, nil), "\n")
	if !strings.Contains(got, `A9: check "container-check" runs container "anchore/grype:v0.116.1"`) {
		t.Errorf("A9 did not reject a tag-pinned container in a registry Argv; got:\n%s", got)
	}
}

func TestAuditA9AcceptsDigestPinnedRegistryArgv(t *testing.T) {
	regs := []Check{{
		ID:            "container-check",
		Profiles:      []string{ProfilePR},
		CIJob:         "ci.yml:probe",
		FixtureWaiver: &Waiver{Kind: WaiverThirdParty, Why: "pinned upstream scanner"},
		Argv: []string{"docker", "run", "--rm", "-v", repoRootPlaceholder + ":/src",
			"anchore/grype@sha256:" + strings.Repeat("f", 64), "dir:/src"},
	}}
	jobs := closureJobs(workflowJob{Steps: []workflowStep{
		{Name: "verify", Run: "go run ./tools/verify --ci-job=ci.yml:probe --profile=pr"},
	}})

	for _, f := range auditFindings(regs, jobs, nil) {
		if strings.HasPrefix(f, "A9") {
			t.Errorf("A9 fired on a digest-pinned registry Argv: %s", f)
		}
	}
}

// The image is the first NON-flag token, not the first token after `run` — a
// parser that took the latter would read a -v mount as the image and then
// "prove" every container unpinned or every container fine, depending on the
// mount string. Both directions are silent failures, so pin the parse.
func TestDockerRunImageOfSkipsFlagValues(t *testing.T) {
	cases := []struct{ in, want string }{
		{`--rm -v C:/repo:/src anchore/grype@sha256:abc dir:/src`, "anchore/grype@sha256:abc"},
		{`--rm --network host -e FOO=bar postgres:16 psql`, "postgres:16"},
		{`--rm`, ""},
	}
	for _, c := range cases {
		if got := dockerRunImageOf(dockerFields(c.in)); got != c.want {
			t.Errorf("dockerRunImageOf(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// A7's scope is "blocks a merge", not "is in the pr profile". go-test-integration
// is the live instance: full-only, but run by ci.yml:test-integration, which is
// inside ci.yml:required's closure.
func TestAuditA7CoversRequiredClosureChecksOutsideThePRProfile(t *testing.T) {
	jobs := closureJobs(workflowJob{})
	regs := []Check{{
		ID:       "closure-only",
		Profiles: []string{ProfileFull},
		Argv:     []string{"true"},
		CIJob:    "ci.yml:probe",
	}}

	got := strings.Join(auditFindings(regs, jobs, nil), "\n")
	if !strings.Contains(got, `A7 check "closure-only" blocks a merge`) {
		t.Errorf("A7 ignored a merge-blocking check that is not in the pr profile; got:\n%s", got)
	}
}

func TestAuditA7IgnoresChecksThatBlockNothing(t *testing.T) {
	jobs := closureJobs(workflowJob{})
	regs := []Check{{
		ID:       "nightly-only",
		Profiles: []string{ProfileFull},
		Argv:     []string{"true"},
		CIJob:    "nightly.yml:security-scan",
	}}

	for _, f := range auditFindings(regs, jobs, nil) {
		if strings.Contains(f, "A7 ") {
			t.Errorf("A7 fired on a check outside the required closure: %s", f)
		}
	}
}

// ------------------------------------------------------------------ A11

// The hazard A11 exists for: a matrix that does not cover the denominator the
// job's own --shard flag promises. Every case below is a green CI run over a
// suite that did not fully execute.
func TestAuditA11CatchesAShardMatrixThatDoesNotCoverTheSuite(t *testing.T) {
	shardStep := func(denom string) workflowStep {
		return workflowStep{
			Name: "verify (go-test-integration)",
			Run:  "go run ./tools/verify --require-infra --only=go-test-integration --shard=${{ matrix.shard }}/" + denom,
		}
	}

	for _, tc := range []struct {
		name   string
		job    workflowJob
		want   string
		silent bool
	}{{
		name: "matrix short of the denominator",
		job:  workflowJob{Steps: []workflowStep{shardStep("4")}, ShardMatrix: []int{1, 2, 3}},
		want: "not [1 2 3 4]",
	}, {
		name: "matrix repeats an index and skips another",
		job:  workflowJob{Steps: []workflowStep{shardStep("3")}, ShardMatrix: []int{1, 2, 2}},
		want: "not [1 2 3]",
	}, {
		name: "matrix is zero-based",
		job:  workflowJob{Steps: []workflowStep{shardStep("3")}, ShardMatrix: []int{0, 1, 2}},
		want: "not [1 2 3]",
	}, {
		name: "sharded step with no matrix at all",
		job:  workflowJob{Steps: []workflowStep{shardStep("4")}},
		want: "not [1 2 3 4]",
	}, {
		name: "matrix with no sharded step runs the whole suite four times",
		job:  workflowJob{Steps: []workflowStep{{Name: "verify (x)", Run: "go run ./tools/verify --only=x"}}, ShardMatrix: []int{1, 2, 3, 4}},
		want: "no step passes --shard",
	}, {
		name:   "matrix and denominator agree",
		job:    workflowJob{Steps: []workflowStep{shardStep("4")}, ShardMatrix: []int{1, 2, 3, 4}},
		silent: true,
	}, {
		name:   "matrix out of order still covers the suite",
		job:    workflowJob{Steps: []workflowStep{shardStep("4")}, ShardMatrix: []int{4, 2, 1, 3}},
		silent: true,
	}, {
		name:   "an ordinary unsharded job is none of A11's business",
		job:    workflowJob{Steps: []workflowStep{{Name: "verify (x)", Run: "go run ./tools/verify --only=x"}}},
		silent: true,
	}} {
		t.Run(tc.name, func(t *testing.T) {
			job := tc.job
			job.Workflow, job.Job = "ci.yml", "test-integration"
			got := auditShardCoverage(nil, []workflowJob{job})
			if tc.silent {
				if len(got) != 0 {
					t.Errorf("A11 fired on a correct configuration: %v", got)
				}
				return
			}
			if len(got) == 0 {
				t.Fatalf("A11 stayed silent; this configuration runs a partial suite and reports success")
			}
			if !strings.Contains(got[0], tc.want) {
				t.Errorf("A11 said %q, want it to mention %q", got[0], tc.want)
			}
		})
	}
}

// The regex half, pinned separately: the index is a GitHub expression with
// spaces in it, and a pattern that stops at the first space reads the real
// workflow as unsharded. That exact bug shipped for one commit.
func TestAuditA11ReadsTheDenominatorThroughAMatrixExpression(t *testing.T) {
	m := shardPattern.FindStringSubmatch("go run ./tools/verify --only=go-test-integration --changed --shard=${{ matrix.shard }}/4")
	if m == nil {
		t.Fatal("shardPattern did not match a --shard whose index is a ${{ matrix.shard }} expression")
	}
	if m[1] != "4" {
		t.Errorf("read denominator %q, want 4", m[1])
	}
}
