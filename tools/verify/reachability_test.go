package main

import (
	"os/exec"
	"sort"
	"strings"
	"testing"
)

// The `changed` profile's reachability proof (#87/A1 review B6a).
//
// `changed` is `pr` intersected with the diff, and the intersection is a
// PREFIX match of Check.Paths against changed file names (matchesPaths). That
// makes Paths a silent narrowing device: a prefix with a typo in it
// ("tools/verifyy/"), or one left behind by a rename, matches no file that
// can ever change, so the check it guards is selected by `changed` on exactly
// zero pull requests. Nothing else notices. `--audit` sees a check wired to a
// job. The job runs and reports success. `--profile=pr` runs the check
// locally and it passes. The gate is simply never reached on a PR, and every
// signal says it is fine.
//
// So the property proven here is reachability, not matching: for every
// path-scoped blocking check, a file that EXISTS IN THIS REPOSITORY selects
// it through the same code path a PR takes — selectByProfile, then
// filterByChanged. Not a unit test of matchesPaths with invented inputs;
// invented inputs are exactly what a dead prefix would still satisfy.
//
// Deliberately not an --audit rule: --audit is pure over the registry and the
// workflows, and this needs the repository's actual file list. It runs under
// go-test-unit, which is blocking on every PR.

// trackedFiles lists every file git knows about, which is the population a
// diff can draw from. A prefix that matches nothing here can never be matched
// by any pull request.
// Run from the repo root, not the package directory: `go test` sets the
// working directory to the package, where `git ls-files` would list only
// tools/verify's own files and every prefix outside it would look dead.
func trackedFiles(t *testing.T) []string {
	t.Helper()
	top, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		t.Fatalf("git rev-parse --show-toplevel: %v", err)
	}
	ls := exec.Command("git", "ls-files")
	ls.Dir = strings.TrimSpace(string(top))
	out, err := ls.Output()
	if err != nil {
		t.Fatalf("git ls-files: %v", err)
	}
	var files []string
	for _, line := range strings.Split(string(out), "\n") {
		if f := strings.TrimSpace(line); f != "" {
			files = append(files, f)
		}
	}
	if len(files) == 0 {
		t.Fatal("git ls-files returned nothing; this test needs the real repository")
	}
	return files
}

// TestEveryDeclaredPathPrefixMatchesATrackedFile is the dead-prefix half. A
// prefix nobody can touch is a check nobody runs.
func TestEveryDeclaredPathPrefixMatchesATrackedFile(t *testing.T) {
	files := trackedFiles(t)
	for _, c := range checks {
		if !hasProfile(c, ProfilePR) {
			continue
		}
		for _, p := range c.Paths {
			if !anyFileHasPrefix(files, p) {
				t.Errorf("check %q declares Paths prefix %q, which no tracked file matches: "+
					"under --profile=changed this prefix can never select the check, so the guard is unreachable on a PR",
					c.ID, p)
			}
		}
	}
}

// TestChangedProfileSelectsEveryPathScopedCheck is the reachability half, and
// the one that is a proof rather than a smell test: it drives the real
// selection path. If a future edit narrows Paths past the point of usefulness
// — or reorders selection so a path-scoped check drops out — this fails with
// the check named.
func TestChangedProfileSelectsEveryPathScopedCheck(t *testing.T) {
	files := trackedFiles(t)
	for _, c := range checks {
		if !hasProfile(c, ProfilePR) || len(c.Paths) == 0 {
			continue
		}
		witness := firstFileUnderAnyPrefix(files, c.Paths)
		if witness == "" {
			continue // already reported, precisely, by the test above
		}
		// The same two steps a PR takes: `changed` resolves to the `pr` set,
		// then the diff filters it.
		selected := filterByChanged(selectByProfile(ProfilePR), []string{witness})
		if !containsCheck(selected, c.ID) {
			t.Errorf("check %q is not selected by --profile=changed even when %q changes; "+
				"declared Paths: %v", c.ID, witness, c.Paths)
		}
	}
}

// TestPathlessChecksAlwaysRunUnderChanged pins the fail-closed direction. A
// check with no Paths is repo-scoped: it must survive ANY diff, including one
// that touches a single unrelated file. If this ever inverts, the security
// scanners and every other deliberately unscoped guard become dodgeable by
// touching nothing they declare.
func TestPathlessChecksAlwaysRunUnderChanged(t *testing.T) {
	selected := filterByChanged(selectByProfile(ProfilePR), []string{"README.md"})
	var missing []string
	for _, c := range checks {
		if !hasProfile(c, ProfilePR) || len(c.Paths) > 0 {
			continue
		}
		if !containsCheck(selected, c.ID) {
			missing = append(missing, c.ID)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("path-less checks dropped by a one-file diff: %v", missing)
	}
}

func anyFileHasPrefix(files []string, prefix string) bool {
	for _, f := range files {
		if strings.HasPrefix(f, prefix) {
			return true
		}
	}
	return false
}

func firstFileUnderAnyPrefix(files, prefixes []string) string {
	for _, f := range files {
		for _, p := range prefixes {
			if strings.HasPrefix(f, p) {
				return f
			}
		}
	}
	return ""
}

func containsCheck(cs []Check, id string) bool {
	for _, c := range cs {
		if c.ID == id {
			return true
		}
	}
	return false
}
