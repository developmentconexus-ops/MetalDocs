// Command gen-tripwire writes the machine-generated tripwire migration SQL
// (internal/platform/tripwire.RenderMigration()) to the latest canonical
// migration path (0275 as of M6 F6.2, documents/UPDATE gains document.review).
//
// Usage: go run ./cmd/gen-tripwire [output-path]
// With no argument, writes to the canonical path relative to the repo root
// (found by walking up from the working directory for go.mod).
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"metaldocs/internal/platform/tripwire"
)

const defaultRelPath = "db/migrations/0275_documents_update_tripwire_review_cap.sql"

func main() {
	out := defaultRelPath
	if len(os.Args) > 1 {
		out = os.Args[1]
	} else {
		root, err := findRepoRoot()
		if err != nil {
			fmt.Fprintf(os.Stderr, "gen-tripwire: locate repo root: %v\n", err)
			os.Exit(1)
		}
		out = filepath.Join(root, defaultRelPath)
	}

	content := tripwire.RenderMigration()
	if err := os.WriteFile(out, []byte(content), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "gen-tripwire: write %s: %v\n", out, err)
		os.Exit(1)
	}
	fmt.Printf("gen-tripwire: wrote %s (%d bytes)\n", out, len(content))
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}
