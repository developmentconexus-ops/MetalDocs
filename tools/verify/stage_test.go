package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

// The F2 property, stated once: the subject vuln-scan scans is a pure function
// of the commit. Same commit → same file set, on CI and on any laptop, whatever
// the working directory happens to contain.
//
// These are not tests of a helper. Each one is the property itself: an untracked
// or ignored artifact cannot enter the scan, a tracked file always does, and the
// staged set equals the commit's tree exactly. The previous design could not
// have passed any of them — it scanned whatever was on disk, and the only
// defence was a hand-maintained --exclude list that a new cache directory
// silently defeated.
//
// Most of them run against a purpose-built repository rather than this one.
// That is not a weaker test, it is a sharper one: a fixture repo can hold a
// tracked file, an untracked file, a gitignored file and a stale build artifact
// at known paths, so each assertion names the exact thing it forbids. It is also
// cheap — staging MetalDocs copies 7484 files, and a unit test that does that
// several times inside `go test ./...` is a tax on every run. One test pays it,
// against the real repo, to prove the mechanism holds at real scale.

// stagedFixture builds a small git repository shaped like the problem: tracked
// sources, a .gitignore, and — planted after the commit — exactly the kinds of
// local artifact that used to reach grype.
func stagedFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v: %s", strings.Join(args, " "), err, out)
		}
	}
	write := func(rel, body string) {
		t.Helper()
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
			t.Fatalf("mkdir for %s: %v", rel, err)
		}
		if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}

	run("init", "--quiet")
	run("config", "user.email", "verify@example.test")
	run("config", "user.name", "verify")
	run("config", "commit.gpgsign", "false")

	// Tracked: what the scan is legitimately about.
	write("go.mod", "module fixture\n\ngo 1.24\n")
	write("package.json", `{"name":"fixture","dependencies":{"left-pad":"1.3.0"}}`+"\n")
	write("internal/app/app.go", "package app\n")
	write(".gitignore", "node_modules/\n*.exe\n")
	run("add", "-A")
	run("commit", "--quiet", "-m", "fixture")

	// Planted AFTER the commit — none of this is in HEAD, so none of it may be
	// scanned. Each line is a real instance from the defect this replaces.
	write("node_modules/lodash/package.json", `{"name":"lodash","version":"4.17.20"}`+"\n") // gitignored dependency tree
	write("metaldocs-api.exe", "stale binary with vendored deps\n")                         // gitignored build output
	write("scratch.json", `{"untracked":true}`+"\n")                                        // untracked, not ignored

	return root
}

func stagedFixtureTree(t *testing.T, root string) (dir string, files map[string]bool) {
	t.Helper()
	dir, cleanup, err := stage(context.Background(), stageTrackedTree, root)
	if err != nil {
		t.Fatalf("stage: %v", err)
	}
	t.Cleanup(cleanup)
	return dir, stagedFiles(t, dir)
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
// reports paths relative to the caller's directory.
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

// The defect, named directly: a file the working directory has and the commit
// does not must be invisible to the scan. A stale .exe reporting x/crypto
// v0.31.0 HIGH while go.mod pins v0.53.0 is what this forbids.
func TestStagedTreeExcludesUntrackedAndIgnoredFiles(t *testing.T) {
	_, files := stagedFixtureTree(t, stagedFixture(t))

	for _, artifact := range []string{
		"node_modules/lodash/package.json", // an ignored dependency tree is not this repo's dependency
		"metaldocs-api.exe",                // an ignored build artifact carries the deps it was built with
		"scratch.json",                     // untracked-but-not-ignored counts too: the rule is HEAD, not .gitignore
	} {
		if files[artifact] {
			t.Errorf("%s is not in HEAD but reached the staged tree: the scan subject still depends on the working directory", artifact)
		}
	}
}

// The other direction, and the one that keeps the gate a gate: a stage that
// produced an empty or partial tree would make grype report zero
// vulnerabilities, and a false green looks exactly like a clean repo.
func TestStagedTreeContainsTrackedFilesWithTheirContent(t *testing.T) {
	root := stagedFixture(t)
	dir, files := stagedFixtureTree(t, root)

	for _, want := range []string{"go.mod", "package.json", "internal/app/app.go"} {
		if !files[want] {
			t.Errorf("%s is tracked but missing from the staged tree — grype would not catalogue it", want)
		}
	}

	// Content, not just names: a stage that copied empty files would satisfy
	// the check above and still scan nothing.
	staged, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		t.Fatalf("read staged package.json: %v", err)
	}
	cmd := exec.Command("git", "show", "HEAD:package.json")
	cmd.Dir = root
	head, err := cmd.Output()
	if err != nil {
		t.Fatalf("git show: %v", err)
	}
	if string(staged) != string(head) {
		t.Error("staged package.json differs from HEAD — the staged subject is not the commit")
	}
}

// Same commit, same BYTES — not merely the same file names. git archive applies
// working-tree conversion, so with core.autocrlf=true (Windows' default) a
// staged text file gains CRLF while the commit and every Linux runner hold LF.
// The scan subject would then be platform-dependent again, which is the exact
// class of divergence this change removes.
func TestStagedTreeIsByteIdenticalWhateverAutocrlfSays(t *testing.T) {
	root := stagedFixture(t)

	cmd := exec.Command("git", "config", "core.autocrlf", "true")
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git config: %v: %s", err, out)
	}

	dir, _ := stagedFixtureTree(t, root)
	staged, err := os.ReadFile(filepath.Join(dir, "internal/app/app.go"))
	if err != nil {
		t.Fatalf("read staged file: %v", err)
	}
	if strings.Contains(string(staged), "\r\n") {
		t.Error("the stage rewrote line endings; the scanned bytes now depend on the host's git config")
	}

	show := exec.Command("git", "show", "HEAD:internal/app/app.go")
	show.Dir = root
	blob, err := show.Output()
	if err != nil {
		t.Fatalf("git show: %v", err)
	}
	if string(staged) != string(blob) {
		t.Errorf("staged bytes differ from the committed blob: %q vs %q", staged, blob)
	}
}

// A tracked file CAN change the result: edit it, commit, and the stage carries
// the new bytes. Together with the test above, this is the full property —
// local noise is inert, the commit is what counts.
func TestStagedTreeFollowsTheCommit(t *testing.T) {
	root := stagedFixture(t)

	bump := `{"name":"fixture","dependencies":{"left-pad":"1.3.1"}}` + "\n"
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(bump), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Uncommitted, so not yet in HEAD — the deliberate trade-off, asserted
	// rather than assumed.
	dir, cleanup, err := stage(context.Background(), stageTrackedTree, root)
	if err != nil {
		t.Fatalf("stage: %v", err)
	}
	before, err := os.ReadFile(filepath.Join(dir, "package.json"))
	cleanup()
	if err != nil {
		t.Fatalf("read staged package.json: %v", err)
	}
	if string(before) == bump {
		t.Error("an uncommitted edit reached the stage; the subject is the working directory again")
	}

	for _, args := range [][]string{{"add", "-A"}, {"commit", "--quiet", "-m", "bump"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v: %s", strings.Join(args, " "), err, out)
		}
	}

	dir2, cleanup2, err := stage(context.Background(), stageTrackedTree, root)
	if err != nil {
		t.Fatalf("stage after commit: %v", err)
	}
	defer cleanup2()
	after, err := os.ReadFile(filepath.Join(dir2, "package.json"))
	if err != nil {
		t.Fatalf("read staged package.json: %v", err)
	}
	if string(after) != bump {
		t.Error("a committed dependency bump did not reach the stage — the gate would scan the wrong version")
	}
}

// Cleanup is part of the mechanism, not an afterthought: a stage is a full copy
// of the tracked tree, so a cleanup that half-works leaves thousands of files in
// the repo. This caught a real one — a single os.RemoveAll loses to a scanner
// holding a handle on a just-written file.
func TestStagedTreeDoesNotSurviveCleanup(t *testing.T) {
	root := stagedFixture(t)
	dir, cleanup, err := stage(context.Background(), stageTrackedTree, root)
	if err != nil {
		t.Fatalf("stage: %v", err)
	}
	cleanup()

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Errorf("staging directory %s survived cleanup (stat err: %v)", dir, err)
	}
	// And it never leaves a sibling behind either.
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read root: %v", err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), stagePrefix) {
			t.Errorf("%s was left in the repository after cleanup", e.Name())
		}
	}
}

// The sweep is the backstop for a run killed before its cleanup. It must clear
// junk without ever touching a stage another verifier process is still using.
func TestSweepRemovesOnlyStaleStages(t *testing.T) {
	root := t.TempDir()

	fresh := filepath.Join(root, stagePrefix+"fresh")
	stale := filepath.Join(root, stagePrefix+"stale")
	other := filepath.Join(root, "node_modules")
	for _, d := range []string{fresh, stale, other} {
		if err := os.MkdirAll(d, 0o750); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}
	old := time.Now().Add(-2 * staleStage)
	if err := os.Chtimes(stale, old, old); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	sweepStages(root)

	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Error("a stage older than an hour was left behind; leftovers accumulate a tracked-tree copy per run")
	}
	if _, err := os.Stat(fresh); err != nil {
		t.Error("a fresh stage was swept — that would delete a concurrent verifier run's scan subject out from under it")
	}
	if _, err := os.Stat(other); err != nil {
		t.Error("the sweep removed a directory that is not a stage")
	}
}

// A stage is a full second copy of the tree, so WHERE it lands is part of the
// mechanism. The first cut put it under the repo root and relied on .gitignore.
// That failed in the first full `--profile=pr` run, and not subtly: api-lint
// walks the tree with filepath.WalkDir and a hard-coded skip list (.git,
// .claude, node_modules, vendor), which does not consult .gitignore at all. It
// read every module file twice and reported 40+ tripwire-pairing violations
// against paths under the stage — vuln-scan failing OTHER checks in the same
// run, on input that does not exist in the commit.
//
// Excluding the prefix in each walker is the exclude-list workaround F2 deletes,
// one level up: it needs every current AND future tree-walker to know about it.
// Outside the repository the hazard is unrepresentable instead — nothing rooted
// at the repo can reach the stage, and no .gitignore entry is load-bearing.
func TestStagedTreeLivesOutsideTheRepository(t *testing.T) {
	root := stagedFixture(t)

	dir, cleanup, err := stage(context.Background(), stageTrackedTree, root)
	if err != nil {
		t.Fatalf("stage: %v", err)
	}
	defer cleanup()

	rel, err := filepath.Rel(root, dir)
	if err != nil {
		return // different volume: further outside than "outside" already requires
	}
	if !strings.HasPrefix(rel, "..") {
		t.Errorf("the stage landed at %s, inside the repository at %s: every check that walks the tree would read the repo twice", dir, root)
	}
}

// Staging is scoped: it never writes outside the directory it created, whatever
// the archive claims. git archive does not emit escaping paths — the point is
// that a file-writing loop must not depend on the producer being well-behaved,
// and that it must give the same answer on every host.
//
// Both directions of that were real. "/etc/passwd" is not absolute to Windows'
// filepath.IsAbs, and "C:/escape.txt" has no volume name on Linux — this test
// passed on Windows and failed on ubuntu-latest until safeJoin stopped asking
// the host and started matching the tar spelling.
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

// One test at real scale. The fixture repo proves the rules; this proves they
// still hold over 7000+ files, nested vendor trees and every path this
// repository actually contains — the input vuln-scan gets in production.
//
// It is the expensive one (~80s on Windows, seconds on a Linux runner), so it is
// the only test here that stages MetalDocs, and it makes the whole claim in a
// single assertion: the staged set IS the tracked set at HEAD. Not a superset
// (working-directory leakage), not a subset (a truncated stage).
//
// It is also the deadlock regression. Scale is the trigger: whether the unread
// tar padding exceeds the pipe buffer depends on total archive size, so a
// fixture repo cannot reproduce it and this repository did — reliably, with the
// whole tree already extracted and git still blocked on its last write.
func TestStagedTreeEqualsThisRepositoryAtHead(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Skipf("not in a git work tree: %v", err)
	}

	dir, cleanup, err := stage(context.Background(), stageTrackedTree, root)
	if err != nil {
		t.Fatalf("stage: %v", err)
	}
	defer cleanup()

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
	// link is a pointer out of the staged tree. If this repository ever gains
	// one, this fails — which is the right moment to decide what a scanner
	// should do with it, rather than finding out silently.
	if len(missing) > 0 {
		t.Errorf("%d tracked file(s) missing from the stage (first few: %s)", len(missing), strings.Join(firstFew(missing), ", "))
	}
	if len(extra) > 0 {
		t.Errorf("%d file(s) in the stage that HEAD does not track (first few: %s)", len(extra), strings.Join(firstFew(extra), ", "))
	}
	t.Logf("staged tree == HEAD: %d files", len(tracked))
}

func firstFew(s []string) []string {
	if len(s) > 5 {
		return s[:5]
	}
	return s
}
