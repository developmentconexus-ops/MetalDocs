package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// EVERY rule this linter emits is BLOCKING: any violation fails CI. There is no
// reported-only/deferred tier — the spec-contract-drift families (ENVELOPE-,
// AUTHZ-, PAGINATION-DRIFT) all graduated to blocking across api-contract-
// hardening Phases D/E2 and the deferral backlog is empty, so a separate
// non-blocking exit class would only let a real regression be silently bypassed
// (CWE-693). A NEW rule is blocking by construction. This is Principle 5 ("the
// model is bound by CI, not by discipline") made real.

func main() {
	strict := flag.Bool("strict", false,
		"hard-error when an EXPECTED core registry file (model.go, seed sql, tripwire allow-list, wiki authz doc) is missing instead of treating it as an empty set. Use for production/CI runs so a wrong modulesRoot can never masquerade as a clean pass (ADR 0022 Phase 13).")
	only := flag.String("only", "", "if set, report only violations for this rule (e.g. PATH-BASE-PREFIX)")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: api-lint [-strict] [-only RULE] <openapi.yaml> [<repo-root>]")
		flag.PrintDefaults()
	}
	flag.Parse()

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

	for _, v := range violations {
		fmt.Printf("%s:%d: %s: %s\n", v.File, v.Line, v.Rule, v.Message)
	}
	fmt.Printf("%d violation(s)\n", len(violations))

	if len(violations) > 0 {
		os.Exit(1)
	}
}
