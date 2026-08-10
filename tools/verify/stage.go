package main

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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

// stagePrefix names the staging directories. They live under the repo root
// rather than the system temp dir for one practical reason: the repo root is
// already a path this repo's container checks bind-mount successfully on every
// developer machine, and %TEMP% is not guaranteed to be a shared drive under
// Docker Desktop. .gitignore covers the prefix, so a crashed run leaves a
// directory that is visibly junk and cannot be committed.
const stagePrefix = ".verify-stage-"

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

// stageTracked extracts `git archive HEAD` into a fresh directory under root.
func stageTracked(ctx context.Context, root string) (string, func(), error) {
	dir, err := os.MkdirTemp(root, stagePrefix)
	if err != nil {
		return "", func() {}, fmt.Errorf("create staging directory: %w", err)
	}
	cleanup := func() { _ = os.RemoveAll(dir) }

	cmd := command(ctx, root, []string{"git", "archive", "--format=tar", "HEAD"})
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
		_ = cmd.Wait()
		cleanup()
		return "", func() {}, err
	}
	if err := cmd.Wait(); err != nil {
		cleanup()
		return "", func() {}, fmt.Errorf("git archive: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return dir, cleanup, nil
}

// untar writes a tar stream into dest. Only regular files and directories are
// materialised: git archive can carry symlinks, and a link is a pointer out of
// the staged tree — exactly the escape hatch staging exists to close.
func untar(r io.Reader, dest string) error {
	tr := tar.NewReader(r)
	for {
		h, err := tr.Next()
		if err == io.EOF {
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
	if _, err := io.CopyN(f, r, maxStagedFile); err != nil && err != io.EOF {
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
