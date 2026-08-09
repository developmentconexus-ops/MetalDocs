// Command gen-tripwire writes the machine-generated tripwire SQL
// (internal/platform/tripwire.RenderMigration()) to the committed golden render
// at internal/platform/tripwire/golden/ (0301 as of unit 3.1a S5, ADR 0082
// phase c: templates_template_version arm drops 'template.review' — the legacy
// reviewer stage was deleted in S4 and the capability is retired from the
// IAM registry in the same change-set).
//
// The default target is a GOLDEN, not a migration: the 2026-07-29 fold squashed
// migrations 0257-0315 into db/baseline and archived the files under
// archive/migrations/post-baseline-2026-07-fold/ (immutable historical record),
// so the regenerable artifact now lives next to the renderer. Both consumers —
// scripts/api-lint's TRIPWIRE-ARM-PARITY rule and
// internal/platform/tripwire.TestRenderMigration_MatchesCommittedFile — read
// that same path.
//
// Any future tripwire vocabulary change ships as a NEW forward migration in
// db/migrations/ regenerated to a new version:
//
//	go run ./cmd/gen-tripwire db/migrations/<NNNN>_<slug>.sql
//
// and then advances defaultRelPath + both consumers in lockstep. See
// internal/platform/tripwire/golden/README.md.
//
// Usage: go run ./cmd/gen-tripwire [output-path]
// With no argument, writes to the canonical golden path relative to the repo
// root (found by walking up from the working directory for go.mod).
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"metaldocs/internal/platform/tripwire"
)

const defaultRelPath = "internal/platform/tripwire/golden/0301_tripwire_template_review_retired.sql"

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
	if err := os.WriteFile(out, []byte(content), 0o644); err != nil { // #nosec G306 G703 -- out is either the fixed defaultRelPath golden file or an explicit operator-supplied CLI arg (os.Args[1]) to this local codegen command, never external/request input; the output is a committed SQL migration/golden file that must stay git readable at 0644.
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
