package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		fmt.Fprintln(os.Stderr, "usage: api-lint <openapi.yaml> [<go-modules-root>]")
		os.Exit(1)
	}

	specPath := os.Args[1]
	modulesRoot := ""
	if len(os.Args) == 3 {
		modulesRoot = os.Args[2]
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

	for _, v := range violations {
		fmt.Printf("%s:%d: %s: %s\n", v.File, v.Line, v.Rule, v.Message)
	}
	fmt.Printf("%d violation(s)\n", len(violations))

	if len(violations) > 0 {
		os.Exit(1)
	}
}
