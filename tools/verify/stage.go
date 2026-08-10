package main

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// STAGING — why the verifier grew a pre-step (#87/A1 review F2).
//
// vuln-scan used to point grype at the working directory. That makes the gate's
// SUBJECT depend on the shape of the machine it runs on: CI gets a bare
// checkout, a developer's tree also holds node_modules, build outputs, Go and
// pnpm caches, and sibling worktrees under .claude/. Those are not dependencies
// of this repo — a stale metaldocs-api.exe reported x/crypto v0.31.0 as HIGH
// while go.mod pins v0.53.0 — so local and CI could disagree while both were
// "green". The cure was a growing --exclude list, which is the shape of a
// workaround: every new gitignored directory needs another line, and the day one
// is missed the two runs mean different things again.
//
// The property this replaces it with is exact: the scanned subject is a pure
// function of the commit. `git archive HEAD` writes the tracked tree at HEAD and
// nothing else — no untracked file, no ignored file, no cache, whatever the
// working directory happens to contain. CI and a laptop scanning the same commit
// scan the same bytes, and no --exclude line is load-bearing anymore.
//
// It is a staging step, not a framework: one Check field, one placeholder, one
// function. Argv stays argv — no shell, no pipeline — so A9 still reads the
// grype digest straight out of the registry.
//
// The deliberate trade-off: an UNCOMMITTED dependency change is not scanned,
// because it is not in HEAD. That is the same answer CI gives, which is the
// point; `git stash` semantics are not something a verifier should invent.

// stageTrackedTree is the only staging mode. A closed set, like Needs and the
// waiver kinds: a mode nobody defined must not be silently ignored.
const stageTrackedTree = "tracked-tree"

// trackedTreePlaceholder expands to the absolute path of the staged tree. Second
// member of the closed substitution set alongside repoRootPlaceholder — same
// rule: the value comes from the verifier, never from a check's input.
const trackedTreePlaceholder = "{{tracked}}"

// stagePrefix names the staging directories, which live in the SYSTEM TEMP DIR,
// never inside the repository.
//
// The first cut put them under the repo root, on the theory that the repo root
// is a path this repo's container checks already bind-mount successfully while
// %TEMP% might not be a shared drive under Docker Desktop. That theory was
// wrong on both halves. %TEMP% mounts fine here (verified: `docker run -v
// <tempdir>:/src alpine ls /src`), and putting a second copy of the tree inside
// the repository broke other checks in the same run: api-lint walks the tree
// with a hard-coded skip list (.git, .claude, node_modules, vendor), so it
// found every module file twice and reported 40+ tripwire-pairing violations
// against paths under .verify-stage-*. .gitignore does not help — a
// filepath.WalkDir does not consult it.
//
// Keeping it in the repo would mean teaching every current and future
// tree-walker about this prefix, which is the exact exclude-list workaround F2
// exists to delete, one level up. Outside the repo the hazard is
// unrepresentable: no walker rooted at the repo can reach the stage at all, no
// .gitignore entry is load-bearing, and clean-tree assertions have nothing to
// see. TestStagedTreeLivesOutsideTheRepository holds it there.
const stagePrefix = "verify-stage-"

// stage materialises a check's staging mode and returns the directory plus its
// cleanup. The cleanup only ever removes a directory this function created.
func stage(ctx context.Context, mode, root string) (string, func(), error) {
	switch mode {
	case stageTrackedTree:
		return stageTracked(ctx, root)
	default:
		return "", func() {}, fmt.Errorf("unknown staging mode %q", mode)
	}
}

// stageTracked extracts `git archive HEAD` of the repo at root into a fresh
// directory in the system temp dir.
func stageTracked(ctx context.Context, root string) (string, func(), error) {
	sweepStages(os.TempDir())

	dir, err := os.MkdirTemp("", stagePrefix)
	if err != nil {
		return "", func() {}, fmt.Errorf("create staging directory: %w", err)
	}
	cleanup := func() { removeStage(dir) }

	// -c core.autocrlf=false -c core.eol=lf is the difference between "same
	// commit, same file set" and "same commit, same bytes". git archive applies
	// working-tree conversion, so on a machine with autocrlf=true (the Windows
	// default, and this box's setting) every staged text file gains CRLF while
	// the committed blob and the Linux runner keep LF. Measured, not assumed.
	// Grype would not read a different dependency out of it, but a scan subject
	// that is byte-identical only on some platforms is the class of divergence
	// this whole change exists to remove. Explicit .gitattributes rules still
	// apply — those are a repo decision and hold on every host.
	cmd := command(ctx, root, []string{
		"git", "-c", "core.autocrlf=false", "-c", "core.eol=lf",
		"archive", "--format=tar", "HEAD",
	})
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cleanup()
		return "", func() {}, fmt.Errorf("git archive: %w", err)
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		cleanup()
		return "", func() {}, fmt.Errorf("git archive: %w", err)
	}
	if err := untar(stdout, dir); err != nil {
		_ = drainAndWait(cmd, stdout)
		cleanup()
		return "", func() {}, err
	}
	if err := drainAndWait(cmd, stdout); err != nil {
		cleanup()
		return "", func() {}, fmt.Errorf("git archive: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return dir, cleanup, nil
}

// drainAndWait reads whatever is left on the pipe before waiting for git.
//
// This is not defensive padding, it is a deadlock fix. tar.Reader stops at the
// end-of-archive marker (two zero blocks), but `git archive` pads its output to
// a 10240-byte blocking factor, so up to ~10KB of zeros are still queued when
// untar returns. A Windows anonymous pipe buffers 4KB by default: git blocks
// writing the tail, exec.Wait blocks on the process, and nobody moves. Whether
// it hangs depends on total archive size mod 10240 — deterministic per commit,
// which is why the fixture repo passed while this repository hung at 7487 files
// with the whole tree already extracted.
func drainAndWait(cmd *exec.Cmd, stdout io.Reader) error {
	_, _ = io.Copy(io.Discard, stdout)
	return cmd.Wait()
}

// staleStage is how old a leftover has to be before a later run will delete it.
// Two verifier processes staging at once is not hypothetical — a test run and a
// profile run overlap easily — so the sweep must never touch a directory that
// could still be live. A stage is written and consumed in minutes; an hour is
// far outside that and far inside "this is junk from a crash".
const staleStage = time.Hour

// removeStage deletes a staging directory, retrying briefly. A single
// os.RemoveAll is not enough on Windows: a virus scanner or the indexer opens
// freshly written files and holds the handle for a moment, and the delete comes
// back with "used by another process" having already removed part of the tree.
// Measured here: the first call failed, the second succeeded.
//
// A cleanup failure is reported, never swallowed. It does not fail the check —
// the check's verdict is about the scan, not about housekeeping — but 7k files
// left in the repo must not be something the tool does quietly.
func removeStage(dir string) {
	var err error
	for i := 0; i < 5; i++ {
		if err = os.RemoveAll(dir); err == nil {
			return
		}
		time.Sleep(time.Duration(i+1) * 200 * time.Millisecond)
	}
	fmt.Fprintf(os.Stderr, "verify: could not remove staging directory %s: %v\n"+
		"verify: it is outside the repository and safe to delete by hand; the next run older than an hour will sweep it\n", dir, err)
}

// sweepStages removes leftovers a crashed or blocked run left behind in tmp, so
// they cannot accumulate a full tracked-tree copy per run.
func sweepStages(tmp string) {
	entries, err := os.ReadDir(tmp)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() || !strings.HasPrefix(e.Name(), stagePrefix) {
			continue
		}
		info, err := e.Info()
		if err != nil || time.Since(info.ModTime()) < staleStage {
			continue
		}
		_ = os.RemoveAll(filepath.Join(tmp, e.Name()))
	}
}

// untar writes a tar stream into dest. Only regular files and directories are
// materialised: git archive can carry symlinks, and a link is a pointer out of
// the staged tree — exactly the escape hatch staging exists to close.
func untar(r io.Reader, dest string) error {
	tr := tar.NewReader(r)
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read staged archive: %w", err)
		}

		target, err := safeJoin(dest, h.Name)
		if err != nil {
			return err
		}

		switch h.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o750); err != nil {
				return fmt.Errorf("stage %s: %w", h.Name, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o750); err != nil {
				return fmt.Errorf("stage %s: %w", h.Name, err)
			}
			if err := writeStagedFile(target, tr); err != nil {
				return err
			}
		default:
			// Symlinks, hardlinks, devices: skipped deliberately. Nothing a
			// dependency scanner reads is one of them.
		}
	}
}

func writeStagedFile(target string, r io.Reader) error {
	f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600) // #nosec G304 -- target is safeJoin'd under the staging dir
	if err != nil {
		return fmt.Errorf("stage %s: %w", target, err)
	}
	// Bounded copy: a tar entry claiming to be larger than this repo could ever
	// be is a malformed stream, not a file worth writing.
	const maxStagedFile = 512 << 20
	if _, err := io.CopyN(f, r, maxStagedFile); err != nil && !errors.Is(err, io.EOF) {
		_ = f.Close()
		return fmt.Errorf("stage %s: %w", target, err)
	}
	return f.Close()
}

// safeJoin rejects any archive entry that would land outside dest. git archive
// does not emit such paths; the check is here because "the producer is
// trustworthy" is not a property a file-writing loop should assume.
func safeJoin(dest, name string) (string, error) {
	// Rooted and volume-qualified names are rejected on their tar spelling, not
	// on the host's. filepath.IsAbs("/etc/passwd") is FALSE on Windows, so a
	// check that trusted it alone would reject the entry on Linux and quietly
	// rewrite it to <dest>\etc\passwd on Windows — the same input, two answers.
	if strings.HasPrefix(name, "/") || strings.HasPrefix(name, `\`) || filepath.VolumeName(name) != "" {
		return "", fmt.Errorf("staged archive entry escapes the staging directory: %q", name)
	}
	clean := filepath.Clean(filepath.FromSlash(name))
	if filepath.IsAbs(clean) || strings.HasPrefix(clean, "..") {
		return "", fmt.Errorf("staged archive entry escapes the staging directory: %q", name)
	}
	target := filepath.Join(dest, clean)
	rel, err := filepath.Rel(dest, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("staged archive entry escapes the staging directory: %q", name)
	}
	return target, nil
}

// withStageDir substitutes trackedTreePlaceholder into a copy of argv.
func withStageDir(argv []string, dir string) []string {
	out := make([]string, len(argv))
	for i, a := range argv {
		out[i] = strings.ReplaceAll(a, trackedTreePlaceholder, dir)
	}
	return out
}
