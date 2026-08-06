// Command gen-http-surface validates the operation declarations in one or
// more OpenAPI documents against the six rules in validate.go.
//
// This is the minimal entry point for Task 4: load every document named on
// the command line, validate them together (rule 6 is cross-document), and
// fail loudly. Task 5 extends this with the actual table emission; nothing
// downstream consumes this command's output yet.
package main

import (
	"fmt"
	"os"
)

// capabilityRegistrySource is the single source of truth for wire capability
// strings: internal/modules/iam/domain/model.go's Capability constants
// (ADR 0022). This is a stopgap reader for Task 4's validation-only command —
// it exists because Validate requires a registry to check rule 3 against, not
// because Task 4 owns registry sourcing. Task 5 may replace it wholesale when
// it wires the real emitter.
const capabilityRegistrySource = "internal/modules/iam/domain/model.go"

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: gen-http-surface <spec.yaml> [spec2.yaml ...]")
		os.Exit(1)
	}

	registry, err := loadCapabilityRegistry(capabilityRegistrySource)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gen-http-surface: loading capability registry: %v\n", err)
		os.Exit(1)
	}

	docs := make([]Document, 0, len(os.Args)-1)
	for _, path := range os.Args[1:] {
		doc, err := LoadDocument(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "gen-http-surface: %v\n", err)
			os.Exit(1)
		}
		docs = append(docs, doc)
	}

	errs := Validate(docs, registry)
	if len(errs) == 0 {
		return
	}
	for _, e := range errs {
		fmt.Fprintln(os.Stderr, e)
	}
	os.Exit(1)
}
