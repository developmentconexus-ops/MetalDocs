package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// workflowJob is one job in one workflow file, reduced to the two facts the
// audit needs: which registry checks it runs, and what it waits for.
type workflowJob struct {
	Workflow string
	Job      string
	OnlyIDs  []string
	Needs    []string
	// Uses holds this job's `uses:` values verbatim, for the pinning rule (A9).
	Uses []string
	// Images holds this job's `services.<name>.image` values, for A9's
	// container half — a service container is code CI executes just as much
	// as an action is.
	Images []string
	// Steps holds every step reduced to what A9 and A10 read: the declared
	// name (A10's classification carrier) and the run: body.
	Steps []workflowStep
	// ShardMatrix holds this job's `strategy.matrix.shard` values, for A11.
	// nil means the job declares no shard matrix, which is the normal case.
	ShardMatrix []int
}

// workflowStep is one step, reduced to the fields the audit rules read.
type workflowStep struct {
	Name string
	Run  string
	Uses string
}

// needs: accepts a bare string or a sequence. Both appear in this repo, so
// the audit reads both rather than assuming the shape it prefers.
type stringOrSlice []string

func (s *stringOrSlice) UnmarshalYAML(n *yaml.Node) error {
	switch n.Kind {
	case yaml.ScalarNode:
		var one string
		if err := n.Decode(&one); err != nil {
			return err
		}
		*s = []string{one}
		return nil
	case yaml.SequenceNode:
		var many []string
		if err := n.Decode(&many); err != nil {
			return err
		}
		*s = many
		return nil
	default:
		return fmt.Errorf("needs: unexpected YAML kind %v", n.Kind)
	}
}

type rawWorkflow struct {
	Jobs map[string]struct {
		Needs    stringOrSlice `yaml:"needs"`
		Services map[string]struct {
			Image string `yaml:"image"`
		} `yaml:"services"`
		Strategy struct {
			Matrix struct {
				Shard []int `yaml:"shard"`
			} `yaml:"matrix"`
		} `yaml:"strategy"`
		Steps []struct {
			Name string `yaml:"name"`
			Run  string `yaml:"run"`
			Uses string `yaml:"uses"`
		} `yaml:"steps"`
	} `yaml:"jobs"`
}

// onlyPattern reads the flag out of a run: block. Both --only=a,b and
// -only=a,b are valid Go flag syntax, so both are matched; a job that used
// only one form would otherwise be invisible to the audit.
var onlyPattern = regexp.MustCompile(`--?only=([A-Za-z0-9_,-]+)`)

// profilePattern reads a --profile=X / -profile=X flag out of a run: block.
// A job that selects its checks by profile instead of by explicit --only
// list (ci.yml:verify runs `--profile=changed`) still names a full,
// enumerable set of registry IDs — just indirectly, through profile
// membership rather than a literal list. Without resolving this, every check
// whose CIJob points at that job would read as A3 ("claimed job exists but
// does not run the check"), which is false: the job does run it, by profile.
var profilePattern = regexp.MustCompile(`--?profile=([a-z]+)`)

// ciJobPattern reads a --ci-job=file.yml:job flag out of a run: block.
var ciJobPattern = regexp.MustCompile(`--?ci-job=([A-Za-z0-9_.:-]+)`)

// idsForProfile resolves a --profile=X value to the check IDs it selects, by
// the same rule selectChecks (main.go) uses at run time: "changed" means the
// `pr` set (the profile it filters), everything else means checks that
// declare that literal profile. This mirrors run-time selection for the
// purpose of the audit only — it does not attempt to reproduce --changed's
// diff-scoping, which is a per-run runtime decision, not a fact about which
// checks a job is wired to.
func idsForProfile(profile string) []string {
	resolved := profile
	if resolved == ProfileChanged {
		resolved = ProfilePR
	}
	var ids []string
	for _, c := range checks {
		if hasProfile(c, resolved) {
			ids = append(ids, c.ID)
		}
	}
	return ids
}

// workflowFiles globs *.yml and *.yaml under dir, sorted for a deterministic
// scan order.
func workflowFiles(dir string) ([]string, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.yml"))
	if err != nil {
		return nil, err
	}
	more, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, err
	}
	paths = append(paths, more...)
	sort.Strings(paths)
	return paths, nil
}

// onlyIDsForStep extracts the registry IDs a single step's `run:` block
// selects, whether it names them directly (--only=a,b) or indirectly
// (--profile=x). Split out of parseWorkflows so the two id-sources
// (explicit list vs. profile resolution) are each one small branch instead
// of one loop doing both.
func onlyIDsForStep(run string) []string {
	if onlyPattern.MatchString(run) {
		return idsFromOnlyFlags(run)
	}
	var ids []string
	for _, m := range profilePattern.FindAllStringSubmatch(run, -1) {
		ids = append(ids, idsForProfile(m[1])...)
	}
	// --ci-job=X narrows a profile selection to the checks that job owns
	// (scopeToCIJob, main.go). Without mirroring it here the audit would
	// credit ci.yml:verify with running every `pr` check, including the ones
	// it deliberately leaves to another job — an audit that describes a
	// selection the command does not make is the drift this tool exists to
	// catch.
	for _, m := range ciJobPattern.FindAllStringSubmatch(run, -1) {
		ids = filterIDsByCIJob(ids, m[1])
	}
	return ids
}

// idsFromOnlyFlags reads every --only=a,b list out of one run: block.
func idsFromOnlyFlags(run string) []string {
	var ids []string
	for _, m := range onlyPattern.FindAllStringSubmatch(run, -1) {
		for _, id := range strings.Split(m[1], ",") {
			if id = strings.TrimSpace(id); id != "" {
				ids = append(ids, id)
			}
		}
	}
	return ids
}

// filterIDsByCIJob keeps only the IDs whose registry entry declares ciJob as
// its owner — the audit-side mirror of main.go's scopeToCIJob.
func filterIDsByCIJob(ids []string, ciJob string) []string {
	owned := map[string]bool{}
	for _, c := range checks {
		if c.CIJob == ciJob {
			owned[c.ID] = true
		}
	}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if owned[id] {
			out = append(out, id)
		}
	}
	return out
}

// jobsInWorkflow turns one parsed workflow file's Jobs map into the sorted
// []workflowJob the audit works with. Split out of parseWorkflows so the
// YAML-walking loop (this function) is separate from the per-file I/O
// (parseWorkflows itself).
func jobsInWorkflow(fileBase string, wf rawWorkflow) []workflowJob {
	names := make([]string, 0, len(wf.Jobs))
	for name := range wf.Jobs {
		names = append(names, name)
	}
	sort.Strings(names)

	out := make([]workflowJob, 0, len(names))
	for _, name := range names {
		j := wf.Jobs[name]
		var ids, uses []string
		var steps []workflowStep
		for _, s := range j.Steps {
			ids = append(ids, onlyIDsForStep(s.Run)...)
			if s.Uses != "" {
				uses = append(uses, s.Uses)
			}
			steps = append(steps, workflowStep{Name: s.Name, Run: s.Run, Uses: s.Uses})
		}
		svcNames := make([]string, 0, len(j.Services))
		for svc := range j.Services {
			svcNames = append(svcNames, svc)
		}
		sort.Strings(svcNames)
		images := make([]string, 0, len(svcNames))
		for _, svc := range svcNames {
			if img := j.Services[svc].Image; img != "" {
				images = append(images, img)
			}
		}
		out = append(out, workflowJob{
			Workflow:    fileBase,
			Job:         name,
			OnlyIDs:     ids,
			Needs:       []string(j.Needs),
			Uses:        uses,
			Images:      images,
			Steps:       steps,
			ShardMatrix: j.Strategy.Matrix.Shard,
		})
	}
	return out
}

func parseWorkflows(dir string) ([]workflowJob, error) {
	paths, err := workflowFiles(dir)
	if err != nil {
		return nil, err
	}

	var out []workflowJob
	for _, p := range paths {
		b, err := os.ReadFile(p) // #nosec G304 -- path comes from a glob of a fixed directory.
		if err != nil {
			return nil, err
		}
		var wf rawWorkflow
		if err := yaml.Unmarshal(b, &wf); err != nil {
			return nil, fmt.Errorf("%s: %w", filepath.Base(p), err)
		}
		out = append(out, jobsInWorkflow(filepath.Base(p), wf)...)
	}
	return out, nil
}

// runsInByJob indexes jobs by "file.yml:job" -> the set of registry IDs that
// job's --only=/--profile= flags resolve to. Shared by the A2/A3/A4 and A6
// rules below.
func runsInByJob(jobs []workflowJob) map[string]map[string]bool {
	runsIn := map[string]map[string]bool{}
	for _, j := range jobs {
		key := j.Workflow + ":" + j.Job
		if runsIn[key] == nil {
			runsIn[key] = map[string]bool{}
		}
		for _, id := range j.OnlyIDs {
			runsIn[key][id] = true
		}
	}
	return runsIn
}

// auditUnknownIDs is A1 — a workflow runs an ID the registry does not have.
func auditUnknownIDs(jobs []workflowJob, known map[string]bool) []string {
	var findings []string
	for _, j := range jobs {
		for _, id := range j.OnlyIDs {
			if !known[id] {
				findings = append(findings, fmt.Sprintf(
					"A1 %s:%s runs --only=%s, which is not a registry check ID", j.Workflow, j.Job, id))
			}
		}
	}
	return findings
}

// auditCIJobClaims is A2 (claimed job does not exist), A3 (claimed job
// exists but does not run the check), and A4 (no CIJob claimed at all) — one
// pass over the registry, since each check falls into exactly one of the
// three.
func auditCIJobClaims(regs []Check, runsIn map[string]map[string]bool) []string {
	var findings []string
	for _, c := range regs {
		if c.CIJob == "" {
			findings = append(findings, fmt.Sprintf(
				"A4 %s has no CIJob: it runs locally and nothing enforces it on a PR", c.ID))
			continue
		}
		set, ok := runsIn[c.CIJob]
		if !ok {
			findings = append(findings, fmt.Sprintf(
				"A2 %s claims CIJob %q, which no workflow defines", c.ID, c.CIJob))
			continue
		}
		if !set[c.ID] {
			findings = append(findings, fmt.Sprintf(
				"A3 %s claims CIJob %q, but that job's --only= set does not include it", c.ID, c.CIJob))
		}
	}
	return findings
}

// auditRequiredGateParity is A5 — ci.yml's `required` job's needs: list and
// scripts/required-gate.jq's hard-coded key set must name exactly the same
// jobs. Nothing else ties them together: `required-gate-selftest` proves the
// .jq expression itself accepts/rejects the right result sets, but it does
// that against fixtures, not against ci.yml's live needs: list — so a job
// added to one and not the other passed every existing check while the gate
// silently stopped requiring it. This is what scripts/required-gate.jq's
// dead workflowJob.Needs field (parsed since audit.go's first version, never
// read until now) was always for.
func auditRequiredGateParity(jobs []workflowJob, requiredGateKeys []string) []string {
	var findings []string
	for _, j := range jobs {
		if j.Workflow != "ci.yml" || j.Job != "required" {
			continue
		}
		got := append([]string(nil), j.Needs...)
		sort.Strings(got)
		want := append([]string(nil), requiredGateKeys...)
		sort.Strings(want)
		if !sameStrings(got, want) {
			findings = append(findings, fmt.Sprintf(
				"A5 ci.yml:required needs=%v but scripts/required-gate.jq requires=%v — a job in one and not the other can fail while `required` still reports success",
				got, want))
		}
	}
	return findings
}

// auditRequiredClosure is A6 — every ProfilePR check's CIJob must be a job
// inside ci.yml:required's needs: transitive closure. I8: 22 of 42 CIJob
// values used to name legacy, since-collapsed workflows
// (module-boundaries.yml, lint.yml, governance-check.yml, fe-ci.yml,
// req-traceability.yml, test-smoke.yml, test-full.yml, invariants.yml) — a
// check pointed at one of those still passed A2/A3 (the workflow file and
// its --only=/--profile= still exist on disk today) while proving nothing
// about whether the NEW topology (ci.yml's four required jobs) actually
// gates a merge on it. A6 closes that: it does not care whether the claimed
// job runs and reports (A2/A3's job) — it asks whether reporting it can ever
// change whether a PR merges. needs: is same-workflow-file only (a job in
// another file cannot appear in ci.yml:required's needs: list even
// indirectly), so the closure is computed only over ci.yml's own jobs; a
// CIJob pointing at any other workflow file — however real that job and
// however faithfully it runs the check — fails A6, because that workflow's
// result cannot reach `required` and therefore cannot block a merge.
func auditRequiredClosure(regs []Check, jobs []workflowJob) []string {
	closure := requiredClosure(jobs)
	var findings []string
	for _, c := range regs {
		if !hasProfile(c, ProfilePR) {
			continue
		}
		if c.CIJob == "" { // already reported by A4
			continue
		}
		if !closure[c.CIJob] {
			findings = append(findings, fmt.Sprintf(
				"A6 %s claims CIJob %q, which is not inside ci.yml:required's needs: transitive closure — it cannot gate a merge", c.ID, c.CIJob))
		}
	}
	return findings
}

// auditFindings applies the rules. Pure, so the rules are testable
// without a repository on disk — requiredGateKeys is read from
// scripts/required-gate.jq by the caller (parseRequiredGateKeys) and handed
// in already parsed, same reason parseWorkflows' I/O stays out of this
// function.
func auditFindings(regs []Check, jobs []workflowJob, requiredGateKeys []string) []string {
	known := map[string]bool{}
	for _, c := range regs {
		known[c.ID] = true
	}
	runsIn := runsInByJob(jobs)

	var findings []string
	findings = append(findings, auditUnknownIDs(jobs, known)...)
	findings = append(findings, auditCIJobClaims(regs, runsIn)...)
	findings = append(findings, auditRequiredGateParity(jobs, requiredGateKeys)...)
	findings = append(findings, auditRequiredClosure(regs, jobs)...)
	findings = append(findings, auditFixtureCoverage(regs, jobs)...)
	findings = append(findings, auditDuplicateIDs(regs)...)
	findings = append(findings, auditToolPinning(regs, jobs)...)
	findings = append(findings, auditRequiredClosureGates(jobs)...)
	findings = append(findings, auditShardCoverage(regs, jobs)...)

	sort.Strings(findings)
	return findings
}

// requiredClosure returns the set of "ci.yml:job" keys reachable from
// ci.yml's `required` job by following needs: edges, plus "ci.yml:required"
// itself. needs: only ever names a job in the same workflow file (GitHub
// Actions has no cross-file needs:), so the closure never leaves ci.yml —
// which is exactly why a CIJob naming any other workflow file can never be in
// it, no matter how real that job is.
func requiredClosure(jobs []workflowJob) map[string]bool {
	needsByJob := map[string][]string{}
	for _, j := range jobs {
		if j.Workflow != "ci.yml" {
			continue
		}
		needsByJob[j.Job] = j.Needs
	}

	closure := map[string]bool{}
	var visit func(job string)
	visit = func(job string) {
		key := "ci.yml:" + job
		if closure[key] {
			return
		}
		closure[key] = true
		for _, dep := range needsByJob[job] {
			visit(dep)
		}
	}
	if _, ok := needsByJob["required"]; ok {
		visit("required")
	}
	return closure
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// requiredGateKeyPattern extracts the jq array literal
// `["a", "b", "c"]` that scripts/required-gate.jq compares `keys | sort`
// against. The audit reads this instead of hard-coding the job list a third
// time (ci.yml's needs:, the .jq file, and this audit would otherwise all
// have to agree by hand).
var requiredGateKeyArrayPattern = regexp.MustCompile(`\[\s*((?:"[^"]*"\s*,?\s*)+)\]`)
var requiredGateKeyPattern = regexp.MustCompile(`"([^"]*)"`)

func parseRequiredGateKeys(path string) ([]string, error) {
	b, err := os.ReadFile(path) // #nosec G304 -- path is the fixed scripts/required-gate.jq literal, not user input.
	if err != nil {
		return nil, err
	}
	m := requiredGateKeyArrayPattern.FindSubmatch(b)
	if m == nil {
		return nil, fmt.Errorf("%s: no jq array literal found (expected [\"job\", ...])", path)
	}
	var keys []string
	for _, km := range requiredGateKeyPattern.FindAllSubmatch(m[1], -1) {
		keys = append(keys, string(km[1]))
	}
	return keys, nil
}

func printAudit(dir string) int {
	jobs, err := parseWorkflows(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "verify --audit: cannot read workflows: %v\n", err)
		return 1
	}
	gateKeys, err := parseRequiredGateKeys(filepath.Join("scripts", "required-gate.jq"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "verify --audit: cannot read scripts/required-gate.jq: %v\n", err)
		return 1
	}
	findings := auditFindings(checks, jobs, gateKeys)
	fmt.Printf("verify --audit: %d checks, %d workflow jobs, %d findings\n",
		len(checks), len(jobs), len(findings))
	if len(findings) == 0 {
		return 0
	}
	fmt.Println()
	for _, f := range findings {
		fmt.Printf("  %s\n", f)
	}
	return 1
}

// shardPattern reads a --shard=i/n flag out of a run: block, capturing only
// the denominator.
//
// The index half is a `${{ matrix.shard }}` expression, not a literal — that
// is the entire shape A11 exists to check — and a GitHub expression contains
// SPACES. The first spelling of this pattern used \S*? for the index and
// therefore matched nothing, which made A11 report the real job as "declares
// a matrix but passes no --shard". Matching to end-of-line instead of
// end-of-token is what makes the expression form visible.
var shardPattern = regexp.MustCompile(`--?shard=[^\n]*?/([0-9]+)`)

// auditShardCoverage is A11 — a sharded job's matrix must cover exactly the
// shards its own --shard denominator promises.
//
// The hazard is specific and silent. `--shard=${{ matrix.shard }}/4` with a
// matrix of [1, 2, 3] runs three quarters of the integration suite and
// reports green: no shard fails, no step errors, and nothing anywhere else in
// this repo would notice that a quarter of the packages were never executed.
// The same is true of a matrix that skips an index, repeats one, or drifts
// after someone edits one line of the pair.
//
// The two numbers are two spellings of the same fact in the same YAML block,
// which is precisely the hand-synced-enumeration shape that keeps producing
// defects here. This rule does not remove the duplication — a generated CI
// manifest (ROADMAP 4.7) would — but it makes the drift a finding instead of
// a green build over a suite that did not fully run.
//
// It runs in both directions: a job that shards must declare a matrix, and a
// check that declares a Partition must be run by a job that shards it or
// deliberately runs it whole (a single `--shard` absent is fine — that is the
// unsharded path, and it runs everything).
func auditShardCoverage(regs []Check, jobs []workflowJob) []string {
	var out []string
	for _, j := range jobs {
		key := j.Workflow + ":" + j.Job
		denom, findings := shardDenominator(key, j.Steps)
		out = append(out, findings...)
		out = append(out, shardMatrixFindings(key, denom, j.ShardMatrix)...)
	}

	// The reverse direction: a partitioned check whose CI job never shards it
	// is not a defect (it runs whole), but a partitioned check whose CI job
	// does not exist or does not run it is already A2/A3's finding, so there
	// is nothing to add here. What IS worth stating is a Partition that no
	// workflow can ever shard because the check has no CI job at all — A4
	// reports the missing job, and this rule would only repeat it.
	_ = regs
	return out
}

// shardDenominator reads the n of --shard=i/n out of a job's steps. It returns
// 0 when no step shards, which is the ordinary unsharded job.
func shardDenominator(key string, steps []workflowStep) (int, []string) {
	var out []string
	denom := 0
	for _, s := range steps {
		m := shardPattern.FindStringSubmatch(s.Run)
		if m == nil {
			continue
		}
		n, err := strconv.Atoi(m[1])
		if err != nil || n < 1 {
			out = append(out, fmt.Sprintf(
				"A11 %s passes --shard with denominator %q, which is not a positive number", key, m[1]))
			continue
		}
		if denom != 0 && denom != n {
			out = append(out, fmt.Sprintf(
				"A11 %s passes two different --shard denominators (%d and %d): one job cannot partition its subject two ways", key, denom, n))
		}
		denom = n
	}
	return denom, out
}

// shardMatrixFindings compares the denominator a job's steps promise against
// the matrix that is supposed to supply every index of it.
func shardMatrixFindings(key string, denom int, matrix []int) []string {
	if denom == 0 {
		if len(matrix) == 0 {
			return nil
		}
		return []string{fmt.Sprintf(
			"A11 %s declares a shard matrix %v but no step passes --shard: every matrix leg runs the whole subject, so the suite runs %d times over", key, matrix, len(matrix))}
	}
	got := append([]int(nil), matrix...)
	sort.Ints(got)
	want := make([]int, denom)
	for i := range want {
		want[i] = i + 1
	}
	if equalInts(got, want) {
		return nil
	}
	return []string{fmt.Sprintf(
		"A11 %s runs --shard=.../%d but its matrix is %v, not %v: the shards it does not run are silently skipped and the job still reports success",
		key, denom, got, want)}
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// auditFixtureCoverage is rule A7: every blocking check declares either a
// negative fixture or a waiver saying why it does not, and never both.
//
// A guard with no evidence that it can fail is indistinguishable from a guard
// that cannot. Before A1.2 that state was invisible — the audit could prove a
// check ran in CI and could not prove it did anything when it ran. This rule
// makes the hole enumerable: a check may lack a fixture, but not silently.
func auditFixtureCoverage(regs []Check, jobs []workflowJob) []string {
	closure := requiredClosure(jobs)
	var out []string
	for _, c := range regs {
		// "Blocking" is not the same as "in the pr profile". go-test-integration
		// is full-only, yet ci.yml:test-integration runs it with
		// --only=go-test-integration and that job is inside ci.yml:required's
		// closure — so it blocks a merge while sitting outside the rule's
		// original scope, and it did: it was the one check in the registry with
		// neither a fixture nor a waiver. Membership in the required closure is
		// the honest definition of blocking, and it is the same closure A10 uses.
		blocksMerge := hasProfile(c, ProfilePR) || (c.CIJob != "" && closure[c.CIJob])
		if !blocksMerge {
			continue
		}
		switch {
		case c.Fixture == nil && c.FixtureWaiver == nil:
			out = append(out, fmt.Sprintf(
				"A7 check %q blocks a merge but declares neither Fixture nor FixtureWaiver: "+
					"nothing proves this guard can fail", c.ID))
		case c.Fixture != nil && c.FixtureWaiver != nil:
			out = append(out, fmt.Sprintf(
				"A7 check %q declares both a Fixture and a FixtureWaiver: a waiver explains an "+
					"absent fixture, so keeping both leaves a stale claim in the registry", c.ID))
		case c.FixtureWaiver != nil:
			// The kind is the load-bearing part. Prose could always describe a
			// fourth reason ("not yet", "the sandbox is hard"), and prose is
			// what let three repo-authored guards count as covered while
			// declaring themselves uncovered in the same sentence.
			if !waiverKinds[c.FixtureWaiver.Kind] {
				out = append(out, fmt.Sprintf(
					"A7 check %q has FixtureWaiver kind %q, which is not one of %s: a guard that fits "+
						"no admissible kind needs a fixture, not a new kind",
					c.ID, c.FixtureWaiver.Kind, knownWaiverKinds()))
			}
			if strings.TrimSpace(c.FixtureWaiver.Why) == "" {
				out = append(out, fmt.Sprintf(
					"A7 check %q has a FixtureWaiver with no Why: the kind says which rule admits the "+
						"waiver, Why must say why this check is an instance of it", c.ID))
			}
		}
	}
	return out
}

// knownWaiverKinds renders the closed set for a finding message.
func knownWaiverKinds() string {
	var kinds []string
	for k := range waiverKinds {
		kinds = append(kinds, string(k))
	}
	sort.Strings(kinds)
	return strings.Join(kinds, ", ")
}

// Step-name prefixes that declare a step is NOT a blocking gate (A10).
// `prereq:` sets the stage (checkout, toolchain, install, fetch a ref);
// `report:` publishes a result somewhere. Neither asserts a property about
// the code under review — that is the verifier's job, and the verifier's
// alone.
const (
	stepPrereqPrefix = "prereq:"
	stepReportPrefix = "report:"
)

// verifyInvocation matches a run: block that calls the verifier. Path form
// rather than binary name: `go run ./tools/verify`, `go run ./tools/verify/`
// and a prebuilt `./tools/verify/verify` all contain it, and nothing else in
// these workflows does.
var verifyInvocation = "tools/verify"

// auditRequiredClosureGates is A10 — inside ci.yml:required's transitive
// closure, the ONLY thing allowed to decide whether a PR merges is the
// verifier.
//
// This rule exists because of a concrete regression, not a hypothetical one:
// ci.yml:security was inside the closure and carried two blocking gates
// (gitleaks via `docker run`, grype via `anchore/scan-action` with
// fail-build: true) that no registry check described. Every other audit rule
// passed — A2/A3 because the job existed and ran what it claimed, A6 because
// the job was in the closure, A7 because the registry had nothing to be
// missing a fixture FOR. The gate was real, blocking, and invisible to the
// product that is supposed to define what "verified" means. Moving those two
// into the registry (review B1) fixed the instance; this rule fixes the class.
//
// Mechanism: every step in a closure job must be either a verifier
// invocation or self-declared as not-a-gate by its `name:` prefix. That is a
// declaration, not a proof — an author can still name a scanner
// "prereq: scan". What it cannot be is silent: the declaration is a diff, in
// the file the reviewer is already reading, on the line that introduces the
// gate.
//
// ci.yml:required itself is excluded: it is the closure's root, and its one
// step's job is precisely to aggregate the closure's verdicts.
func auditRequiredClosureGates(jobs []workflowJob) []string {
	closure := requiredClosure(jobs)
	var findings []string
	for _, j := range jobs {
		key := j.Workflow + ":" + j.Job
		if !closure[key] || key == "ci.yml:required" {
			continue
		}
		for i, s := range j.Steps {
			if strings.Contains(s.Run, verifyInvocation) {
				continue
			}
			name := strings.ToLower(strings.TrimSpace(s.Name))
			if strings.HasPrefix(name, stepPrereqPrefix) || strings.HasPrefix(name, stepReportPrefix) {
				continue
			}
			findings = append(findings, fmt.Sprintf(
				"A10 %s step %d (%s) is inside ci.yml:required's closure but neither invokes %s nor declares itself %s/%s: "+
					"a blocking gate defined here is a second definition of \"verified\" that --audit cannot see and a local run does not execute",
				key, i+1, describeStep(s), verifyInvocation, stepPrereqPrefix, stepReportPrefix))
		}
	}
	return findings
}

// describeStep names a step for a finding message, preferring what the author
// wrote over what the audit inferred.
func describeStep(s workflowStep) string {
	switch {
	case strings.TrimSpace(s.Name) != "":
		return strings.TrimSpace(s.Name)
	case s.Uses != "":
		return "uses " + s.Uses
	default:
		return "run " + strings.TrimSpace(strings.SplitN(strings.TrimSpace(s.Run), "\n", 2)[0])
	}
}

// auditDuplicateIDs rejects two checks sharing an ID. --only, --audit's own
// CIJob mapping and the fixture harness all address a check by ID, so a
// duplicate silently shadows one of the two; this was introduced and caught by
// accident while wiring A1.2, which is reason enough for it to be mechanical.
func auditDuplicateIDs(regs []Check) []string {
	seen := map[string]int{}
	for _, c := range regs {
		seen[c.ID]++
	}
	var out []string
	for id, n := range seen {
		if n > 1 {
			out = append(out, fmt.Sprintf("A8 check ID %q is declared %d times; IDs address checks and must be unique", id, n))
		}
	}
	return out
}

// shaPin matches a 40-hex commit SHA, the only `uses:` form GitHub resolves to
// an immutable object. A tag (@v4, @v4.2.2) is a movable pointer: the tag can
// be re-pointed at different code with no diff in this repository to review.
var shaPin = regexp.MustCompile(`@[0-9a-f]{40}$`)

// localUses matches a `uses:` that names something inside this repository
// (./path) or a reusable workflow in this repository — there is no upstream to
// pin, the referenced code is the code under review.
var localUses = regexp.MustCompile(`^\.{1,2}/`)

// versionSuffix matches the @version of a Go module path in an Argv, e.g. the
// "@latest" in "github.com/x/y/cmd/z@latest".
var versionSuffix = regexp.MustCompile(`@([A-Za-z0-9._+-]+)$`)

// digestPin matches an image reference pinned to a content digest —
// `name@sha256:<64 hex>`. A tag (postgres:16, grype:v0.116.1) is a movable
// pointer exactly like an action tag: the registry can republish it over
// different bytes with no diff in this repository to review.
var digestPin = regexp.MustCompile(`@sha256:[0-9a-f]{64}$`)

// dockerRunImage pulls the image reference out of a `docker run ...` command
// line. Everything before it is flags (and their values), which is why the
// pattern skips them rather than taking the first token after "run".
var dockerRunLine = regexp.MustCompile(`docker\s+run\s+([^\n]*)`)

// dockerValueFlags are the `docker run` flags that consume the NEXT token, so
// the image reference is not the token that follows them.
var dockerValueFlags = map[string]bool{
	"-v": true, "--volume": true, "-e": true, "--env": true,
	"-w": true, "--workdir": true, "-u": true, "--user": true,
	"--name": true, "--network": true, "--platform": true,
	"--entrypoint": true, "-p": true, "--publish": true,
	"--mount": true, "--label": true, "-l": true,
}

// dockerRunImageOf returns the image reference in a `docker run` argument
// list, or "" if the list names none. Shared by A9's two container shapes
// (workflow run: blocks and registry Argvs) so both read the command the same
// way.
func dockerRunImageOf(args []string) string {
	for i := 0; i < len(args); i++ {
		a := args[i]
		if strings.HasPrefix(a, "-") {
			if dockerValueFlags[a] && !strings.Contains(a, "=") {
				i++ // skip this flag's value
			}
			continue
		}
		return a
	}
	return ""
}

// auditToolPinning is A9 — everything CI executes must name an immutable
// version.
//
// Two shapes, one property. A workflow `uses:` must be SHA-pinned, and a check
// whose Argv fetches a tool by module path must not say "@latest". Both are
// the same failure: the bytes CI runs can change with no diff here to review,
// so a check can start or stop failing for reasons nobody chose. That is not
// hypothetical drift — it is the reason a green run stops being evidence.
//
// Four shapes now, still one property (#87/A1 review B2 added the last
// three): a workflow `uses:`, a `services:` container, a `docker run` inside a
// workflow step, and a `docker run` inside a registry Argv. A container is
// code CI executes; leaving images on mutable tags while pinning actions to
// SHAs pinned the half that was easy to see.
//
// Scope is deliberately what CI executes: the workflows and the registry.
// Container images under deploy/ are the same class of gap but belong to the
// deployment axis, not the verifier spine; they are recorded there, not
// silently fixed here.
func auditToolPinning(regs []Check, jobs []workflowJob) []string {
	var findings []string
	seen := map[string]bool{}
	for _, j := range jobs {
		for _, u := range j.Uses {
			if shaPin.MatchString(u) || localUses.MatchString(u) {
				continue
			}
			key := j.Workflow + ":" + u
			if seen[key] {
				continue
			}
			seen[key] = true
			findings = append(findings, fmt.Sprintf(
				"A9: %s uses %q, which is not SHA-pinned — a tag can be re-pointed at different code with no diff here to review",
				j.Workflow, u))
		}
		findings = append(findings, auditJobContainers(j)...)
	}
	findings = append(findings, auditArgvContainers(regs)...)
	for _, c := range regs {
		for _, arg := range c.Argv {
			m := versionSuffix.FindStringSubmatch(arg)
			if m == nil || !strings.Contains(arg, "/") {
				continue
			}
			switch m[1] {
			case "latest", "upgrade", "patch":
				findings = append(findings, fmt.Sprintf(
					"A9: check %q fetches %q — an unpinned tool version changes what this gate accepts without a diff here to review",
					c.ID, arg))
			}
		}
	}
	return findings
}

// auditJobContainers is A9's workflow-container half: `services:` images and
// any `docker run` a step issues must both name a digest.
func auditJobContainers(j workflowJob) []string {
	var findings []string
	for _, img := range j.Images {
		if digestPin.MatchString(img) {
			continue
		}
		findings = append(findings, fmt.Sprintf(
			"A9: %s:%s runs service container %q, which is not digest-pinned — a tag can be republished over different bytes with no diff here to review",
			j.Workflow, j.Job, img))
	}
	for _, s := range j.Steps {
		for _, m := range dockerRunLine.FindAllStringSubmatch(s.Run, -1) {
			img := dockerRunImageOf(dockerFields(m[1]))
			if img == "" || digestPin.MatchString(img) {
				continue
			}
			findings = append(findings, fmt.Sprintf(
				"A9: %s:%s runs container %q, which is not digest-pinned — a tag can be republished over different bytes with no diff here to review",
				j.Workflow, j.Job, img))
		}
	}
	return findings
}

// auditArgvContainers is A9's registry-container half. A check that shells out
// to `docker run` is executing third-party bytes exactly like a workflow step
// is, and moving a gate from YAML into the registry (review B1) must not be a
// way to leave the pin behind.
func auditArgvContainers(regs []Check) []string {
	var findings []string
	for _, c := range regs {
		if len(c.Argv) < 2 || filepath.Base(c.Argv[0]) != "docker" || c.Argv[1] != "run" {
			continue
		}
		img := dockerRunImageOf(c.Argv[2:])
		if img == "" {
			findings = append(findings, fmt.Sprintf(
				"A9: check %q runs docker but names no image", c.ID))
			continue
		}
		if !digestPin.MatchString(img) {
			findings = append(findings, fmt.Sprintf(
				"A9: check %q runs container %q, which is not digest-pinned — a tag can be republished over different bytes with no diff here to review",
				c.ID, img))
		}
	}
	return findings
}

// dockerFields splits a shell-ish command tail into tokens, dropping the
// quotes YAML steps wrap paths in. It is not a shell parser and does not need
// to be: the only thing read out of the result is which token is the image.
func dockerFields(tail string) []string {
	var out []string
	for _, f := range strings.Fields(tail) {
		out = append(out, strings.Trim(f, `"'`))
	}
	return out
}
