package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	only := flag.String("only", "", "if set, report only violations for this rule (e.g. PATH-BASE-PREFIX)")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: api-lint [-only RULE] <openapi.yaml> [<go-modules-root>]")
		flag.PrintDefaults()
	}
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 || len(args) > 2 {
		flag.Usage()
		os.Exit(1)
	}

	specPath := args[0]
	modulesRoot := ""
	if len(args) == 2 {
		modulesRoot = args[1]
	}

	violations, err := RunSpecRules(specPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if modulesRoot != "" {
		codeViolations, err := RunCodeRules(specPath, modulesRoot)
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
