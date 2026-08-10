package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// The F2 property, stated once: the subject vuln-scan scans is a pure function
// of the commit. Same commit → same file set, on CI and on any laptop, whatever
// the working directory happens to contain.
//
// These are not tests of a helper. Each one is the property itself: a gitignored
// artifact cannot enter the scan, a tracked file always does, and the staged set
// equals `git ls-tree -r HEAD` exactly. The previous design could not have
// passed any of them — it scanned whatever was on disk, and the only defence
// was a hand-maintained --exclude list that a new cache directory silently
// defeated.

// Staging the whole tracked tree is not free (7k+ files), so the tests that
// only READ a stage share one, built in TestMain. The probe test builds its own,
// because it has to stage after planting an untracked file.
var (
	sharedStageDir  string
	sharedStageRoot string
)

func TestMain(m *testing.M) {
	root, err := gitRoot()
	if err != nil {
		// Not a work tree: every test here skips itself on the same condition.
		os.Exit(m.Run())
	}
	dir, cleanup, err := stage(context.Background(), stageTrackedTree, root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "stage tracked tree: %v\n", err)
		os.Exit(1)
	}
	sharedStageRoot, sharedStageDir = root, dir
	code := m.Run()
	cleanup()
	os.Exit(code)
}

func gitRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// stagedTree returns the shared stage, or skips when there is nothing to stage.
func stagedTree(t *testing.T) (root, dir string) {
	t.Helper()
	if sharedStageDir == "" {
		t.Skip("not in a git work tree")
	}
	return sharedStageRoot, sharedStageDir
}

// stagedFiles lists a staged tree as repo-relative slash paths.
func stagedFiles(t *testing.T, dir string) map[string]bool {
	t.Helper()
	out := map[string]bool{}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		out[filepath.ToSlash(rel)] = true
		return nil
	})
	if err != nil {
		t.Fatalf("walk staged tree: %v", err)
	}
	return out
}

// trackedAtHead is the reference set. --full-name matters: without it git
// reports paths relative to the caller's directory, and this test runs from
// tools/verify.
func trackedAtHead(t *testing.T, root string) map[string]bool {
	t.Helper()
	cmd := exec.Command("git", "ls-tree", "-r", "--full-name", "--name-only", "HEAD")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git ls-tree: %v", err)
	}
	tracked := map[string]bool{}
	for _, line := range strings.Split(strings.ReplaceAll(string(out), "\r\n", "\n"), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			tracked[line] = true
		}
	}
	if len(tracked) == 0 {
		t.Fatal("git ls-tree returned nothing")
	}
	return tracked
}

// A file the working directory has and the commit does not must be invisible to
// the scan. This is the exact failure the --exclude list was chasing: a stale
// .exe, a .gocache object, a sibling worktree under .claude/ — all untracked,
// all previously catalogued as dependencies of this repo.
func TestStagedTreeExcludesUntrackedAndIgnoredFiles(t *testing.T) {
	root, _ := stagedTree(t)

	// The name matches a .gitignore pattern, so the probe itself can never be
	// committed by a careless `git add -A` while this test is being written.
	probe := filepath.Join(root, fmt.Sprintf(".verify-probe-%d.json", os.Getpid()))
	if err := os.WriteFile(probe, []byte(`{"name":"probe","version":"0.0.0"}`), 0o600); err != nil {
		t.Fatalf("write probe: %v", err)
	}
	defer func() { _ = os.Remove(probe) }()

	// A stage taken WITH the probe on disk, so the assertion is about staging
	// and not about ordering.
	dir, cleanup, err := stage(context.Background(), stageTrackedTree, root)
	if err != nil {
		t.Fatalf("stage: %v", err)
	}
	defer cleanup()

	name := filepath.Base(probe)
	if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
		t.Fatalf("%s is untracked but reached the staged tree: the scan subject still depends on the working directory", name)
	}
}

// The other direction, and the one that keeps the gate a gate: the files whose
// contents the scan is actually about are present. If staging ever produced an
// empty or partial tree, grype would report zero vulnerabilities and the check
// would pass — a false green that looks exactly like a clean repo.
func TestStagedTreeContainsTheDependencyManifests(t *testing.T) {
	_, dir := stagedTree(t)
	files := stagedFiles(t, dir)

	for _, want := range []string{"go.mod", "go.sum", "package.json", "pnpm-lock.yaml"} {
		if !files[want] {
			t.Errorf("%s is tracked but missing from the staged tree — grype would not catalogue it", want)
		}
	}
	if len(files) < 100 {
		t.Errorf("staged tree has only %d files; a truncated stage makes a clean scan meaningless", len(files))
	}
}

// A tracked file's CONTENT reaches the scan, not just its name. Content is what
// a scanner reads: a stage that copied empty files would satisfy the test above
// and still scan nothing.
func TestStagedTreeCarriesTrackedContent(t *testing.T) {
	root, dir := stagedTree(t)

	for _, name := range []string{"go.mod", "pnpm-lock.yaml"} {
		staged, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("read staged %s: %v", name, err)
		}
		cmd := exec.Command("git", "show", "HEAD:"+name)
		cmd.Dir = root
		head, err := cmd.Output()
		if err != nil {
			t.Fatalf("git show HEAD:%s: %v", name, err)
		}
		if string(staged) != string(head) {
			t.Errorf("staged %s differs from HEAD:%s — the staged subject is not the commit", name, name)
		}
	}
}

// The whole claim in one assertion: the staged set IS the tracked set at HEAD.
// Not a superset (working-directory leakage), not a subset (a truncated stage).
// This is what makes "CI and local scan the same subject" checkable rather than
// argued: both run `git archive HEAD`, and `git ls-tree -r HEAD` is what that
// contains.
func TestStagedTreeEqualsTrackedTreeAtHead(t *testing.T) {
	root, dir := stagedTree(t)
	tracked := trackedAtHead(t, root)
	staged := stagedFiles(t, dir)

	var extra, missing []string
	for f := range staged {
		if !tracked[f] {
			extra = append(extra, f)
		}
	}
	for f := range tracked {
		if !staged[f] {
			missing = append(missing, f)
		}
	}
	sort.Strings(extra)
	sort.Strings(missing)

	// Symlinks are the one deliberate difference: untar skips them, because a
	// link is a pointer out of the staged tree. If the repo ever gains one this
	// fails, and that is the right moment to decide what a scanner should do
	// with it — not silently.
	if len(missing) > 0 {
		t.Errorf("%d tracked file(s) missing from the stage (first few: %s)", len(missing), strings.Join(firstFew(missing), ", "))
	}
	if len(extra) > 0 {
		t.Errorf("%d file(s) in the stage that HEAD does not track (first few: %s)", len(extra), strings.Join(firstFew(extra), ", "))
	}
}

// Staging is scoped: it never writes outside the directory it created, whatever
// the archive claims. git archive does not emit escaping paths — the point is
// that a file-writing loop must not depend on the producer being well-behaved,
// and that it must give the same answer on every host. "/etc/passwd" is not
// absolute to Windows' filepath.IsAbs; rejecting it is not platform-conditional.
func TestStagedArchiveCannotEscape(t *testing.T) {
	dest := t.TempDir()
	for _, name := range []string{"../escape.txt", "/etc/passwd", `\escape.txt`, "C:/escape.txt", "a/../../escape.txt"} {
		if _, err := safeJoin(dest, name); err == nil {
			t.Errorf("%q was accepted; it escapes the staging directory", name)
		}
	}
	for _, name := range []string{"go.mod", "tools/verify/stage.go"} {
		if _, err := safeJoin(dest, name); err != nil {
			t.Errorf("normal entry %q was rejected: %v", name, err)
		}
	}
}

// A check that declares a staging mode must actually consume it, and a check
// that consumes the placeholder must declare the mode — otherwise the token
// reaches docker verbatim and the mount silently points at a directory named
// "{{tracked}}".
func TestStagedChecksAndPlaceholderAgree(t *testing.T) {
	for _, c := range checks {
		uses := false
		for _, a := range c.Argv {
			if strings.Contains(a, trackedTreePlaceholder) {
				uses = true
			}
		}
		switch {
		case c.Stage != "" && c.Stage != stageTrackedTree:
			t.Errorf("%s declares unknown staging mode %q", c.ID, c.Stage)
		case c.Stage != "" && !uses:
			t.Errorf("%s declares Stage=%q but never uses %s", c.ID, c.Stage, trackedTreePlaceholder)
		case c.Stage == "" && uses:
			t.Errorf("%s uses %s but declares no Stage — the token would reach the command unexpanded", c.ID, trackedTreePlaceholder)
		}
	}
}

func firstFew(s []string) []string {
	if len(s) > 5 {
		return s[:5]
	}
	return s
}
