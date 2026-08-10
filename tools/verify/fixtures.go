package main

// The negative-fixture spine (A1.2).
//
// A guard that has never been observed to fail is not a control, it is a
// decoration. Unit-testing a guard's internal helpers is not enough either:
// the thing CI runs is a *command*, and the property that matters is that the
// command exits non-zero on bad input. A helper can be perfectly tested while
// the command still returns 0 — that exact defect class is why this file
// exists (rulebook V-ARCH-7).
//
// The positive half of the property ("passes on valid input") is already
// proven continuously and at full scale: every one of these checks runs
// against this repo on every PR, and this repo is valid input. Duplicating a
// synthetic good-case fixture would prove strictly less. So this harness
// proves only the half nothing else proves — the failing half.
//
// Shape, deliberately small:
//
//	scripts/testdata/guard-fixtures/<name>/...    a tree of deliberately bad input
//
// The tree is copied into a throwaway directory, made into a git repo, and the
// check's OWN Argv is executed with that directory as the working directory.
// Not a reimplementation of the check, not a library call — the same argv CI
// runs. If it exits 0, the harness fails.
//
// Fixture source files carry a trailing ".txt" (e.g. "bad.go.txt"), stripped
// when copied. This is load-bearing, not cosmetic: tools/cilint's file walker
// does not skip testdata/, `scripts/check-gofmt.sh` scans `git ls-files
// '*.go'`, and check-test-discipline.sh reads tracked sources. A fixture named
// "*.go" would be scanned by the very guards it is meant to break, and this
// repo's own CI would go red on files whose entire purpose is to be wrong.
// The convention matches the one scripts/testdata/test-discipline/ already
// uses.

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
)

// guardFixtureRoot is where fixture trees live, relative to the repo root.
const guardFixtureRoot = "scripts/testdata/guard-fixtures"

// fixtureToken expands to the sandbox's absolute path inside ArgvOverride.
const fixtureToken = "{{fixture}}"

// Fixture declares how to feed a check bad input and prove it says no.
type Fixture struct {
	// Dir is a directory name under scripts/testdata/guard-fixtures.
	//
	// Flat layout: the tree is copied into the sandbox and committed.
	//
	// Layered layout: if the tree contains a "base/" directory, "base/" is
	// copied and committed first, refs/remotes/origin/main is pointed at that
	// commit, then "head/" is copied over it and committed. Diff-shaped
	// guards (governance rules, migration gaplessness) need a real before and
	// after, and inventing that history is the only way to give them one.
	Dir string

	// ArgvOverride replaces the check's Argv, and runs from the REPO ROOT
	// rather than the sandbox. The token {{fixture}} expands to the sandbox's
	// absolute path.
	//
	// Empty is the strong form and the default: run the check's own Argv,
	// byte-identical to CI, with the sandbox as the working directory. Use an
	// override only for guards that take their input path as an argument —
	// there the override is *stronger*, because it needs no sandbox trickery
	// at all, it just hands the real command a bad file.
	ArgvOverride []string

	// CopyFromRepo lists repo-relative files copied verbatim into the sandbox
	// at the same relative path — sibling scripts, jq programs, stub packages
	// a guard reads. The guard script named in Argv is copied automatically;
	// this is for what that script in turn reaches for.
	CopyFromRepo []string

	// Want, when set, must appear in the failing run's combined output. A
	// non-zero exit alone can come from the wrong cause (a syntax error, a
	// missing file); Want pins the failure to the rule under test.
	Want string
}

// fixtureResult is one fixture's outcome.
type fixtureResult struct {
	id     string
	ok     bool
	reason string
	output string
}

// runGuardFixtures executes every declared fixture and reports. It is the
// body of `verify --guard-fixtures`, and is itself registered as the
// `guard-fixtures` check so CI runs it like any other.
func runGuardFixtures(ctx context.Context, root string, only []string) error {
	selected := fixtureChecks(only)
	if len(selected) == 0 {
		return fmt.Errorf("no checks with fixtures selected")
	}

	// Each fixture is an independent subprocess against its own temp dir, so
	// they parallelize freely. Serially this check is minutes; that cost lands
	// on every PR, and a control nobody wants to wait for is a control that
	// gets argued out of the gate.
	results := make([]fixtureResult, len(selected))
	sem := make(chan struct{}, defaultParallelism())
	var wg sync.WaitGroup
	for i, c := range selected {
		wg.Add(1)
		go func(i int, c Check) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			results[i] = runOneFixture(ctx, root, c)
		}(i, c)
	}
	wg.Wait()

	failed := 0
	for _, r := range results {
		status := "ok"
		if !r.ok {
			status = "FAIL"
			failed++
		}
		fmt.Printf("%-6s %-26s %s\n", status, r.id, r.reason)
		if !r.ok && r.output != "" {
			fmt.Printf("%s\n", indent(r.output))
		}
	}
	fmt.Printf("\nguard fixtures: %d run, %d failed\n", len(results), failed)
	if failed > 0 {
		return fmt.Errorf("%d guard fixture(s) did not prove their check can fail", failed)
	}
	return nil
}

// splitIDs parses a comma-separated --only value.
func splitIDs(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// fixtureChecks returns the checks that declare a fixture, filtered by --only
// when given.
func fixtureChecks(only []string) []Check {
	want := map[string]bool{}
	for _, id := range only {
		want[id] = true
	}
	var out []Check
	for _, c := range checks {
		if c.Fixture == nil {
			continue
		}
		if len(want) > 0 && !want[c.ID] {
			continue
		}
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// runOneFixture materializes one fixture and asserts the check rejects it.
func runOneFixture(ctx context.Context, root string, c Check) fixtureResult {
	fail := func(format string, a ...any) fixtureResult {
		return fixtureResult{id: c.ID, reason: fmt.Sprintf(format, a...)}
	}

	sandbox, err := os.MkdirTemp("", "verify-fixture-")
	if err != nil {
		return fail("mkdtemp: %v", err)
	}
	defer func() { _ = os.RemoveAll(sandbox) }()

	// Symlinked temp roots (macOS /var -> /private/var) make a guard's own
	// path arithmetic disagree with ours; resolve once, up front.
	if resolved, err := filepath.EvalSymlinks(sandbox); err == nil {
		sandbox = resolved
	}

	if err := materializeFixture(ctx, root, sandbox, c); err != nil {
		return fail("materialize: %v", err)
	}

	argv, dir, err := fixtureArgv(ctx, root, sandbox, c)
	if err != nil {
		return fail("build argv: %v", err)
	}

	cmd := command(ctx, dir, argv)
	cmd.Env = fixtureEnv()
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	runErr := cmd.Run()
	out := buf.String()

	if runErr == nil {
		return fixtureResult{
			id:     c.ID,
			reason: fmt.Sprintf("exited 0 on bad input — the guard does not guard (%s)", strings.Join(argv, " ")),
			output: out,
		}
	}
	var exitErr *exec.ExitError
	if !errors.As(runErr, &exitErr) {
		return fixtureResult{id: c.ID, reason: fmt.Sprintf("could not run %s: %v", strings.Join(argv, " "), runErr), output: out}
	}
	if c.Fixture.Want != "" && !strings.Contains(out, c.Fixture.Want) {
		return fixtureResult{
			id:     c.ID,
			reason: fmt.Sprintf("failed, but not for the reason under test: output does not contain %q", c.Fixture.Want),
			output: out,
		}
	}
	return fixtureResult{id: c.ID, ok: true, reason: fmt.Sprintf("rejected bad input (exit %d)", exitErr.ExitCode())}
}

// materializeFixture builds the sandbox: guard scripts, fixture tree, git.
func materializeFixture(ctx context.Context, root, sandbox string, c Check) error {
	for _, rel := range append(scriptsInArgv(c.Argv), c.Fixture.CopyFromRepo...) {
		if err := copyFile(filepath.Join(root, rel), filepath.Join(sandbox, filepath.FromSlash(rel))); err != nil {
			return fmt.Errorf("copy %s: %w", rel, err)
		}
	}

	src := filepath.Join(root, filepath.FromSlash(guardFixtureRoot), c.Fixture.Dir)
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("fixture tree: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("fixture tree %s is not a directory", src)
	}

	layered := false
	if fi, err := os.Stat(filepath.Join(src, "base")); err == nil && fi.IsDir() {
		layered = true
	}

	if err := gitRun(ctx, sandbox, "init", "-q", "-b", "main"); err != nil {
		return err
	}

	if !layered {
		if err := copyTree(src, sandbox); err != nil {
			return err
		}
		return gitCommit(ctx, sandbox, "fixture")
	}

	if err := copyTree(filepath.Join(src, "base"), sandbox); err != nil {
		return err
	}
	if err := gitCommit(ctx, sandbox, "base"); err != nil {
		return err
	}
	// The diff-shaped guards resolve their base as origin/<branch>; give the
	// sandbox that ref rather than teaching every guard a second code path.
	if err := gitRun(ctx, sandbox, "update-ref", "refs/remotes/origin/main", "HEAD"); err != nil {
		return err
	}
	if err := copyTree(filepath.Join(src, "head"), sandbox); err != nil {
		return err
	}
	return gitCommit(ctx, sandbox, "head")
}

// fixtureArgv resolves what to run and from where.
func fixtureArgv(ctx context.Context, root, sandbox string, c Check) ([]string, string, error) {
	if len(c.Fixture.ArgvOverride) > 0 {
		argv := make([]string, len(c.Fixture.ArgvOverride))
		for i, a := range c.Fixture.ArgvOverride {
			argv[i] = strings.ReplaceAll(a, fixtureToken, sandbox)
		}
		return argv, root, nil
	}

	dir := sandbox
	if c.Dir != "" {
		dir = filepath.Join(sandbox, filepath.FromSlash(c.Dir))
	}

	// `go run ./pkg` cannot work from the sandbox: there is no module there,
	// and there must not be — a module file would drag the whole dependency
	// graph into a throwaway directory. Build the same package in the real
	// module and run the resulting binary instead. Same code, same flags.
	if len(c.Argv) >= 3 && c.Argv[0] == "go" && c.Argv[1] == "run" {
		pkg := c.Argv[2]
		name := filepath.Base(pkg)
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		bin := filepath.Join(sandbox, ".guardbin", name)
		build := command(ctx, root, []string{"go", "build", "-o", bin, pkg})
		build.Env = fixtureEnv()
		if out, err := build.CombinedOutput(); err != nil {
			return nil, "", fmt.Errorf("go build %s: %w\n%s", pkg, err, out)
		}
		return append([]string{bin}, c.Argv[3:]...), dir, nil
	}

	return c.Argv, dir, nil
}

// fixtureEnv is the sandbox environment. GITHUB_BASE_REF is dropped so a
// diff-shaped guard resolves the base the sandbox actually has, and so the
// harness behaves identically on a laptop and inside Actions.
func fixtureEnv() []string {
	var out []string
	for _, kv := range os.Environ() {
		if strings.HasPrefix(kv, "GITHUB_BASE_REF=") {
			continue
		}
		out = append(out, kv)
	}
	return out
}

// scriptsInArgv returns the repo-relative script paths an argv names, so the
// sandbox gets the real guard rather than a copy that can drift from it.
func scriptsInArgv(argv []string) []string {
	var out []string
	for _, a := range argv {
		if !strings.HasSuffix(a, ".sh") && !strings.HasSuffix(a, ".ps1") {
			continue
		}
		out = append(out, strings.TrimPrefix(filepath.ToSlash(a), "./"))
	}
	return out
}

func gitRun(ctx context.Context, dir string, args ...string) error {
	cmd := command(ctx, dir, append([]string{"git"}, args...))
	cmd.Env = fixtureEnv()
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, out)
	}
	return nil
}

// gitCommit stages everything and commits with an identity supplied on the
// command line, so the harness never depends on the machine's git config.
func gitCommit(ctx context.Context, dir, msg string) error {
	if err := gitRun(ctx, dir, "add", "-A"); err != nil {
		return err
	}
	return gitRun(ctx, dir,
		"-c", "user.name=verify-fixture",
		"-c", "user.email=verify-fixture@invalid",
		"commit", "-q", "--allow-empty", "-m", msg,
	)
}

// copyTree copies src into dst, stripping one trailing ".txt" per file name.
func copyTree(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		target := filepath.Join(dst, strings.TrimSuffix(rel, ".txt"))
		if info.IsDir() {
			return os.MkdirAll(target, 0o750)
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o750); err != nil {
		return err
	}
	in, err := os.Open(src) // #nosec G304 -- src is a repo-relative fixture path from the registry.
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600) // #nosec G304 -- dst is inside the harness's own temp dir.
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}
