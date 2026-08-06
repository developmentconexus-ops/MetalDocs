// Command gen-http-surface validates the operation declarations in one or
// more OpenAPI documents against the six rules in validate.go, then emits a
// generated Go source file per document: a map from mux pattern to
// surfaceRule (Task 5, §3/§10).
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// defaultCapabilityRegistrySource is the single source of truth for wire
// capability strings: internal/modules/iam/domain/model.go's Capability
// constants (ADR 0022), addressed relative to the repository root — the
// default cwd when this command is run directly (`go run
// metaldocs/cmd/gen-http-surface ...` from the repo root). loadCapabilityRegistry
// reads whichever path -registry resolves to, to build the wire string → Go
// constant name map. The registry-sourcing mechanism is shared: the
// validator uses it to check rule 3 against the registry, and the emitter
// uses it to emit `iamdomain.CapMetricsView` from the wire strings found in
// the spec.
//
// A flag, not a hardcoded constant, because `go generate` runs this command
// with its cwd set to the directory holding the //go:generate directive
// (apps/api/cmd/metaldocs-api), not the repo root — the directive passes
// -registry with a path relative to there.
const defaultCapabilityRegistrySource = "internal/modules/iam/domain/model.go"

// emitTarget binds one loaded Document to the file/identifier names its
// generated output takes. The public spec always emits
// httpsurface_gen.go/httpSurface/specTags; the optional -e2e spec (Task 12)
// emits httpsurface_e2e_gen.go/httpSurfaceE2E/specTagsE2E — distinct
// identifiers because both files land in the same package
// (apps/api/cmd/metaldocs-api).
type emitTarget struct {
	doc        Document
	fileName   string
	surfaceVar string
	tagsVar    string
}

func main() {
	publicPath := flag.String("public", "", "path to the public OpenAPI document (required)")
	e2ePath := flag.String("e2e", "", "path to the internal-e2e OpenAPI document (optional)")
	outDir := flag.String("out-dir", "", "directory to write generated files into (required)")
	registryPath := flag.String("registry", defaultCapabilityRegistrySource, "path to iam/domain/model.go, the capability registry source")
	flag.Parse()

	if *publicPath == "" || *outDir == "" {
		fmt.Fprintln(os.Stderr, "usage: gen-http-surface -public <spec.yaml> -out-dir <dir> [-e2e <spec.yaml>] [-registry <path>]")
		os.Exit(1)
	}

	registry, err := loadCapabilityRegistry(*registryPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gen-http-surface: loading capability registry: %v\n", err)
		os.Exit(1)
	}

	publicDoc, err := LoadDocument(*publicPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gen-http-surface: %v\n", err)
		os.Exit(1)
	}

	targets := []emitTarget{
		{doc: publicDoc, fileName: "httpsurface_gen.go", surfaceVar: "httpSurface", tagsVar: "specTags"},
	}

	docs := []Document{publicDoc}
	if *e2ePath != "" {
		e2eDoc, err := LoadDocument(*e2ePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "gen-http-surface: %v\n", err)
			os.Exit(1)
		}
		docs = append(docs, e2eDoc)
		targets = append(targets, emitTarget{doc: e2eDoc, fileName: "httpsurface_e2e_gen.go", surfaceVar: "httpSurfaceE2E", tagsVar: "specTagsE2E"})
	}

	// Validate is called once, over every document together: rule 6
	// (cross-document key collision) can only be checked that way.
	errs := Validate(docs, registry)
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, e)
		}
		os.Exit(1)
	}

	for i, t := range targets {
		src, err := Emit(t.doc, EmitOptions{SurfaceVar: t.surfaceVar, TagsVar: t.tagsVar, Registry: registry, SkipTypeDecl: i > 0})
		if err != nil {
			fmt.Fprintf(os.Stderr, "gen-http-surface: emitting %s: %v\n", t.fileName, err)
			os.Exit(1)
		}
		outPath := filepath.Join(*outDir, t.fileName)
		if err := os.WriteFile(outPath, []byte(src), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "gen-http-surface: writing %s: %v\n", outPath, err)
			os.Exit(1)
		}
	}
}
