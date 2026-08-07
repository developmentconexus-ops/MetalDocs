// cilint — MetalDocs custom Go/TS linters.
// Runs all analyzers and emits SARIF to stdout.
//
// Usage:
//
//	go run ./tools/cilint [--sarif] [--baseline path] [--update-baseline] [./...]
//
// Baseline-and-ratchet: by default cilint compares findings against
// tools/cilint/baseline.json (see that file's "_doc" field, and baseline.go
// for the mechanism). Findings recorded there at their recorded count do not
// fail the build; anything new, or any recorded count that goes UP, does.
// A recorded count that goes DOWN also fails — with --update-baseline
// pointed out explicitly — so the baseline is forced to ratchet down over
// time instead of silently going stale. If baseline.json is absent, cilint
// falls back to its original behavior: any finding at all fails.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"metaldocs/tools/cilint/internal/analyzers"
)

func main() {
	sarif := flag.Bool("sarif", false, "emit SARIF output for GitHub code-scanning")
	baselinePath := flag.String("baseline", "tools/cilint/baseline.json", "path to the baseline file governing the exit code (absent file = fail on any finding)")
	updateBaseline := flag.Bool("update-baseline", false, "rewrite the baseline from current findings and exit 0")
	flag.Parse()
	targets := flag.Args()
	if len(targets) == 0 {
		targets = []string{"./..."}
	}

	results := normalizeFindings(analyzers.RunAll(targets))

	if *sarif {
		out := buildSARIF(results)
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(out)
	}

	if *updateBaseline {
		existing, err := loadBaseline(*baselinePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cilint: cannot read existing baseline %s: %v\n", *baselinePath, err)
			os.Exit(1)
		}
		bl := buildBaseline(results, existing)
		if err := writeBaseline(*baselinePath, bl); err != nil {
			fmt.Fprintf(os.Stderr, "cilint: cannot write baseline %s: %v\n", *baselinePath, err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "cilint: baseline updated — %d findings, %d keys recorded in %s\n",
			len(results), len(bl.Entries), *baselinePath)
		os.Exit(0)
	}

	baseline, err := loadBaseline(*baselinePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cilint: cannot read baseline %s: %v\n", *baselinePath, err)
		os.Exit(1)
	}

	if !*sarif {
		for _, r := range results {
			fmt.Fprintf(os.Stderr, "%s:%d: [%s] %s\n", r.File, r.Line, r.Analyzer, r.Message)
		}
	}

	if baseline == nil {
		// No baseline file configured: behave exactly as before it existed.
		if len(results) > 0 {
			os.Exit(1)
		}
		return
	}

	ok, diagLines := compareBaseline(results, baseline)
	for _, line := range diagLines {
		fmt.Fprintln(os.Stderr, line)
	}
	if !ok {
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "cilint: %d findings, all within baseline (0 new)\n", len(results))
}

// sarifResult is a minimal SARIF 2.1.0 structure for GitHub code-scanning.
func buildSARIF(results []analyzers.Finding) map[string]any {
	rules := map[string]bool{}
	sarfRules := []map[string]any{}
	sarfResults := []map[string]any{}

	for _, r := range results {
		if !rules[r.Analyzer] {
			rules[r.Analyzer] = true
			sarfRules = append(sarfRules, map[string]any{
				"id": r.Analyzer,
				"shortDescription": map[string]any{"text": r.Analyzer},
			})
		}
		sarfResults = append(sarfResults, map[string]any{
			"ruleId":  r.Analyzer,
			"message": map[string]any{"text": r.Message},
			"locations": []map[string]any{
				{
					"physicalLocation": map[string]any{
						"artifactLocation": map[string]any{"uri": r.File},
						"region":           map[string]any{"startLine": r.Line},
					},
				},
			},
		})
	}

	return map[string]any{
		"$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Documents/CommitteeSpecifications/2.1.0/sarif-schema-2.1.0.json",
		"version": "2.1.0",
		"runs": []map[string]any{
			{
				"tool": map[string]any{
					"driver": map[string]any{
						"name":  "cilint",
						"rules": sarfRules,
					},
				},
				"results": sarfResults,
			},
		},
	}
}
