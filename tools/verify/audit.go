package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
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
		Needs stringOrSlice `yaml:"needs"`
		Steps []struct {
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
	var ids []string
	for _, m := range profilePattern.FindAllStringSubmatch(run, -1) {
		ids = append(ids, idsForProfile(m[1])...)
	}
	return ids
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
		for _, s := range j.Steps {
			ids = append(ids, onlyIDsForStep(s.Run)...)
			if s.Uses != "" {
				uses = append(uses, s.Uses)
			}
		}
		out = append(out, workflowJob{
			Workflow: fileBase,
			Job:      name,
			OnlyIDs:  ids,
			Needs:    []string(j.Needs),
			Uses:     uses,
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
	findings = append(findings, auditFixtureCoverage(regs)...)
	findings = append(findings, auditDuplicateIDs(regs)...)
	findings = append(findings, auditToolPinning(regs, jobs)...)

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

// auditFixtureCoverage is rule A7: every blocking check declares either a
// negative fixture or a waiver saying why it does not, and never both.
//
// A guard with no evidence that it can fail is indistinguishable from a guard
// that cannot. Before A1.2 that state was invisible — the audit could prove a
// check ran in CI and could not prove it did anything when it ran. This rule
// makes the hole enumerable: a check may lack a fixture, but not silently.
func auditFixtureCoverage(regs []Check) []string {
	var out []string
	for _, c := range regs {
		if !hasProfile(c, ProfilePR) {
			continue // only blocking checks are in scope
		}
		switch {
		case c.Fixture == nil && c.FixtureWaiver == "":
			out = append(out, fmt.Sprintf(
				"A7 check %q is in profile %s but declares neither Fixture nor FixtureWaiver: "+
					"nothing proves this guard can fail", c.ID, ProfilePR))
		case c.Fixture != nil && c.FixtureWaiver != "":
			out = append(out, fmt.Sprintf(
				"A7 check %q declares both a Fixture and a FixtureWaiver: a waiver explains an "+
					"absent fixture, so keeping both leaves a stale claim in the registry", c.ID))
		}
	}
	return out
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

// auditToolPinning is A9 — everything CI executes must name an immutable
// version.
//
// Two shapes, one property. A workflow `uses:` must be SHA-pinned, and a check
// whose Argv fetches a tool by module path must not say "@latest". Both are
// the same failure: the bytes CI runs can change with no diff here to review,
// so a check can start or stop failing for reasons nobody chose. That is not
// hypothetical drift — it is the reason a green run stops being evidence.
//
// Scope is deliberately what CI executes. Container images in deploy/ are the
// same class of gap but belong to the deployment axis, not the verifier spine;
// they are recorded there, not silently fixed here.
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
	}
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
