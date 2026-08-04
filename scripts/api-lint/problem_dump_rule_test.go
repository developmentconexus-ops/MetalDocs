package main

import (
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTree materializes a minimal fake repo: a dumper with the given imports
// and one delivery package per entry in registering.
func writeTree(t *testing.T, dumperImports []string, registering map[string]string) string {
	t.Helper()
	root := t.TempDir()

	dumperDir := filepath.Join(root, "cmd", "problem-codes-dump")
	if err := os.MkdirAll(dumperDir, 0o755); err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	b.WriteString("package main\n\nimport (\n")
	for _, imp := range dumperImports {
		b.WriteString("\t_ \"" + imp + "\"\n")
	}
	b.WriteString(")\n\nfunc main() {}\n")
	if err := os.WriteFile(filepath.Join(dumperDir, "main.go"), []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}

	for dir, src := range registering {
		full := filepath.Join(root, filepath.FromSlash(dir))
		if err := os.MkdirAll(full, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(full, "errors.go"), []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

const registersOne = `package http

import "metaldocs/internal/platform/problem"

var codeX = problem.Register("mod", "state.thing", 409)
`

func TestProblemDumpImports_MissingPackageIsReported(t *testing.T) {
	root := writeTree(t,
		[]string{"metaldocs/internal/modules/alpha/delivery/http"},
		map[string]string{
			"internal/modules/alpha/delivery/http": registersOne,
			"internal/modules/beta/delivery/http":  registersOne,
		})

	got, err := checkProblemDumpImports(root, token.NewFileSet(), true)
	if err != nil {
		t.Fatalf("checkProblemDumpImports: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d violations, want 1: %+v", len(got), got)
	}
	if got[0].Rule != "PROBLEM-DUMP-IMPORT" {
		t.Fatalf("rule = %q", got[0].Rule)
	}
	// The message must name the import line to add, not merely the offending
	// package: the whole point is that the reader does not have to work out the
	// module path themselves.
	if !strings.Contains(got[0].Message, `"metaldocs/internal/modules/beta/delivery/http"`) {
		t.Fatalf("message does not carry the import to add:\n%s", got[0].Message)
	}
	// Slash-normalized even on Windows — a message quoting a path with backslashes
	// cannot be pasted into an import block.
	if strings.Contains(got[0].Message, `\`) {
		t.Fatalf("message carries OS-native separators:\n%s", got[0].Message)
	}
}

func TestProblemDumpImports_FullyImportedIsClean(t *testing.T) {
	root := writeTree(t,
		[]string{
			"metaldocs/internal/modules/alpha/delivery/http",
			"metaldocs/internal/modules/beta/delivery/http",
		},
		map[string]string{
			"internal/modules/alpha/delivery/http": registersOne,
			"internal/modules/beta/delivery/http":  registersOne,
		})

	got, err := checkProblemDumpImports(root, token.NewFileSet(), true)
	if err != nil {
		t.Fatalf("checkProblemDumpImports: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("got %d violations, want 0: %+v", len(got), got)
	}
}

// A package with no Register call must not be demanded in the import block —
// otherwise the rule would force the dumper to import the whole tree, and an
// import list that includes everything proves nothing.
func TestProblemDumpImports_NonRegisteringPackageIsNotRequired(t *testing.T) {
	root := writeTree(t, nil, map[string]string{
		"internal/modules/gamma/delivery/http": "package http\n\nfunc Handler() {}\n",
	})

	got, err := checkProblemDumpImports(root, token.NewFileSet(), true)
	if err != nil {
		t.Fatalf("checkProblemDumpImports: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("got %d violations, want 0: %+v", len(got), got)
	}
}

// The registry package itself defines Register; requiring the dumper to import
// it explicitly would be circular noise (it is linked in transitively).
func TestProblemDumpImports_PlatformProblemPackageIsSkipped(t *testing.T) {
	root := writeTree(t, nil, map[string]string{
		"internal/platform/problem": "package problem\n\nvar c = problem.Register(\"platform\", \"state.x\", 409)\n",
	})

	got, err := checkProblemDumpImports(root, token.NewFileSet(), true)
	if err != nil {
		t.Fatalf("checkProblemDumpImports: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("got %d violations, want 0: %+v", len(got), got)
	}
}

// -strict exists so a wrong root cannot masquerade as a clean pass. A missing
// dumper is the clearest case of that.
func TestProblemDumpImports_MissingDumperStrictVsLenient(t *testing.T) {
	root := t.TempDir()

	if _, err := checkProblemDumpImports(root, token.NewFileSet(), true); err == nil {
		t.Fatal("strict run with no dumper returned nil error; a wrong root would pass silently")
	}
	got, err := checkProblemDumpImports(root, token.NewFileSet(), false)
	if err != nil {
		t.Fatalf("lenient run: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("lenient run returned %d violations, want 0", len(got))
	}
}
