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

// auditFindings applies the four rules. Pure, so the rules are testable
// without a repository on disk.
func auditFindings(regs []Check, jobs []workflowJob) []string {
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

	sort.Strings(findings)
	return findings
}

func printAudit(dir string) int {
	jobs, err := parseWorkflows(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "verify --audit: cannot read workflows: %v\n", err)
		return 1
	}
	findings := auditFindings(checks, jobs)
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
