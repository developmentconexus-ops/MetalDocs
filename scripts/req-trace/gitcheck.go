package main

import (
	"os"
	"os/exec"
	"path/filepath"
)

// CommitExists reports whether ref resolves to a real object in the repo at
// root, via `git cat-file -e <ref>`. When root has no .git directory (e.g. a
// shallow/exported tree with history stripped), it returns (false, nil) —
// callers should treat that as "skip verification with a warning", not a
// hard failure (plan.md Task 2 / mission box constraint: CI checkouts may be
// shallow).
func CommitExists(root, ref string) (bool, error) {
	gitDir := filepath.Join(root, ".git")
	if !hasGitDir(gitDir) {
		return false, nil
	}
	cmd := exec.Command("git", "cat-file", "-e", ref) // #nosec G204 -- argv, not a shell string, so ref cannot inject shell metacharacters; it only selects which read-only git object to probe, and the only caller in this tree is the test suite (CommitExists is not wired to any external/request input).
	cmd.Dir = root
	if err := cmd.Run(); err != nil {
		if _, isExit := err.(*exec.ExitError); isExit {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// HasGitRepo reports whether root looks like a full (non-shallow-in-a-way-
// that-hides-.git) git checkout, i.e. whether commit-hash verification is
// even possible here.
func HasGitRepo(root string) bool {
	return hasGitDir(filepath.Join(root, ".git"))
}

func hasGitDir(gitDir string) bool {
	_, err := os.Stat(gitDir)
	// .git can be a directory (normal clone) or a file (worktree gitlink) —
	// either way its presence means git commands are usable here.
	return err == nil
}
