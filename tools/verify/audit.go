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
			Run string `yaml:"run"`
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

func parseWorkflows(dir string) ([]workflowJob, error) {
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

	var out []workflowJob
	for _, p := range paths {
		b, err := os.ReadFile(p) //nolint:gosec // G304 — path comes from a glob of a fixed directory.
		if err != nil {
			return nil, err
		}
		var wf rawWorkflow
		if err := yaml.Unmarshal(b, &wf); err != nil {
			return nil, fmt.Errorf("%s: %w", filepath.Base(p), err)
		}
		names := make([]string, 0, len(wf.Jobs))
		for name := range wf.Jobs {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			j := wf.Jobs[name]
			var ids []string
			for _, s := range j.Steps {
				if onlyPattern.MatchString(s.Run) {
					for _, m := range onlyPattern.FindAllStringSubmatch(s.Run, -1) {
						for _, id := range strings.Split(m[1], ",") {
							if id = strings.TrimSpace(id); id != "" {
								ids = append(ids, id)
							}
						}
					}
					continue
				}
				for _, m := range profilePattern.FindAllStringSubmatch(s.Run, -1) {
					ids = append(ids, idsForProfile(m[1])...)
				}
			}
			out = append(out, workflowJob{
				Workflow: filepath.Base(p),
				Job:      name,
				OnlyIDs:  ids,
				Needs:    []string(j.Needs),
			})
		}
	}
	return out, nil
}

// auditFindings applies the five rules. Pure, so the rules are testable
// without a repository on disk — requiredGateKeys is read from
// scripts/required-gate.jq by the caller (parseRequiredGateKeys) and handed
// in already parsed, same reason parseWorkflows' I/O stays out of this
// function.
func auditFindings(regs []Check, jobs []workflowJob, requiredGateKeys []string) []string {
	known := map[string]bool{}
	for _, c := range regs {
		known[c.ID] = true
	}
	runsIn := map[string]map[string]bool{} // "file.yml:job" -> set of IDs
	for _, j := range jobs {
		key := j.Workflow + ":" + j.Job
		if runsIn[key] == nil {
			runsIn[key] = map[string]bool{}
		}
		for _, id := range j.OnlyIDs {
			runsIn[key][id] = true
		}
	}

	var findings []string

	// A1 — a workflow runs an ID the registry does not have.
	for _, j := range jobs {
		for _, id := range j.OnlyIDs {
			if !known[id] {
				findings = append(findings, fmt.Sprintf(
					"A1 %s:%s runs --only=%s, which is not a registry check ID", j.Workflow, j.Job, id))
			}
		}
	}

	for _, c := range regs {
		// A4 — no CI job claimed at all.
		if c.CIJob == "" {
			findings = append(findings, fmt.Sprintf(
				"A4 %s has no CIJob: it runs locally and nothing enforces it on a PR", c.ID))
			continue
		}
		set, ok := runsIn[c.CIJob]
		// A2 — the claimed job does not exist.
		if !ok {
			findings = append(findings, fmt.Sprintf(
				"A2 %s claims CIJob %q, which no workflow defines", c.ID, c.CIJob))
			continue
		}
		// A3 — the claimed job exists but does not run the check.
		if !set[c.ID] {
			findings = append(findings, fmt.Sprintf(
				"A3 %s claims CIJob %q, but that job's --only= set does not include it", c.ID, c.CIJob))
		}
	}

	// A5 — ci.yml's `required` job's needs: list and scripts/required-gate.jq's
	// hard-coded key set must name exactly the same jobs. Nothing else ties
	// them together: `required-gate-selftest` proves the .jq expression
	// itself accepts/rejects the right result sets, but it does that against
	// fixtures, not against ci.yml's live needs: list — so a job added to one
	// and not the other passed every existing check while the gate silently
	// stopped requiring it. This is what scripts/required-gate.jq's dead
	// workflowJob.Needs field (parsed since audit.go's first version, never
	// read until now) was always for.
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

	// A6 — every ProfilePR check's CIJob must be a job inside ci.yml:required's
	// needs: transitive closure. I8: 22 of 42 CIJob values used to name
	// legacy, since-collapsed workflows (module-boundaries.yml, lint.yml,
	// governance-check.yml, fe-ci.yml, req-traceability.yml, test-smoke.yml,
	// test-full.yml, invariants.yml) — a check pointed at one of those still
	// passed A2/A3 (the workflow file and its --only=/--profile= still exist
	// on disk today) while proving nothing about whether the NEW topology
	// (ci.yml's four required jobs) actually gates a merge on it. A6 closes
	// that: it does not care whether the claimed job runs and reports
	// (A2/A3's job) — it asks whether reporting it can ever change whether a
	// PR merges. needs: is same-workflow-file only (a job in another file
	// cannot appear in ci.yml:required's needs: list even indirectly), so the
	// closure is computed only over ci.yml's own jobs; a CIJob pointing at any
	// other workflow file — however real that job and however faithfully it
	// runs the check — fails A6, because that workflow's result cannot reach
	// `required` and therefore cannot block a merge.
	closure := requiredClosure(jobs)
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
	b, err := os.ReadFile(path) //nolint:gosec // G304 — path is the fixed scripts/required-gate.jq literal, not user input.
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
