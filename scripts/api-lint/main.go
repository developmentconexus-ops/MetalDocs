package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// reportedOnlyRules are the spec-contract-drift families (ADR 0022) intentionally
// deferred to a dedicated spec-hygiene phase. They are surfaced and counted but
// MUST NOT gate CI yet. EVERY other rule — the registry-binding guards, the
// dialect bans, and tripwire-pairing — is BLOCKING: a regression turns the build
// red. This is Principle 5 ("the model is bound by CI, not by discipline") made
// real. A NEW rule defaults to blocking by omission, which is the safe default.
var reportedOnlyRules = map[string]struct{}{
	"ENVELOPE-DRIFT":   {},
	"AUTHZ-DRIFT":      {},
	"PAGINATION-DRIFT": {},
}

func isBlocking(rule string) bool {
	_, reported := reportedOnlyRules[rule]
	return !reported
}

func main() {
	strict := flag.Bool("strict", false,
		"hard-error when an EXPECTED core registry file (model.go, seed sql, tripwire allow-list, wiki authz doc) is missing instead of treating it as an empty set. Use for production/CI runs so a wrong modulesRoot can never masquerade as a clean pass (ADR 0022 Phase 13).")
	enforce := flag.String("enforce", "all",
		"which rule-set determines the exit code: all|blocking|reported. blocking = only the binding/dialect/tripwire guards gate (spec-drift printed but non-fatal); reported = only the deferred spec-drift families gate; all = any violation (back-compat default).")
	only := flag.String("only", "", "if set, report only violations for this rule (e.g. PATH-BASE-PREFIX)")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: api-lint [-strict] [-enforce=all|blocking|reported] [-only RULE] <openapi.yaml> [<repo-root>]")
		flag.PrintDefaults()
	}
	flag.Parse()

	switch *enforce {
	case "all", "blocking", "reported":
	default:
		fmt.Fprintf(os.Stderr, "invalid -enforce=%q (want all|blocking|reported)\n", *enforce)
		os.Exit(2)
	}

	args := flag.Args()
	if len(args) < 1 || len(args) > 2 {
		flag.Usage()
		os.Exit(2)
	}

	specPath := args[0]
	modulesRoot := ""
	if len(args) == 2 {
		// The root arg is the REPO ROOT — every binding helper joins its core
		// files off it (internal/modules/..., db/reference-data/..., wiki/...).
		// Normalize so a relative `.`, a trailing slash, or `./internal/modules`
		// (the historical CI bug) all resolve unambiguously.
		abs, err := filepath.Abs(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "resolve repo-root %q: %v\n", args[1], err)
			os.Exit(1)
		}
		modulesRoot = filepath.Clean(abs)
	}

	violations, err := RunSpecRules(specPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if modulesRoot != "" {
		codeViolations, err := RunCodeRules(specPath, modulesRoot, *strict)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		violations = append(violations, codeViolations...)
	}

	if *only != "" {
		kept := violations[:0]
		for _, v := range violations {
			if v.Rule == *only {
				kept = append(kept, v)
			}
		}
		violations = kept
	}

	var blockingCount, reportedCount int
	for _, v := range violations {
		if isBlocking(v.Rule) {
			blockingCount++
		} else {
			reportedCount++
		}
	}

	for _, v := range violations {
		show := *enforce == "all" ||
			(*enforce == "blocking" && isBlocking(v.Rule)) ||
			(*enforce == "reported" && !isBlocking(v.Rule))
		if show {
			fmt.Printf("%s:%d: %s: %s\n", v.File, v.Line, v.Rule, v.Message)
		}
	}
	fmt.Printf("%d blocking violation(s), %d reported-only/deferred violation(s)\n", blockingCount, reportedCount)

	var fail bool
	switch *enforce {
	case "blocking":
		fail = blockingCount > 0
	case "reported":
		fail = reportedCount > 0
	default:
		fail = len(violations) > 0
	}
	if fail {
		os.Exit(1)
	}
}
