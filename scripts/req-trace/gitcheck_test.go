package main

import "testing"

// TestCommitExists_RealRepo checks a known-good commit hash (this repo's
// initial-ish history) resolves, and a bogus hash does not. Uses the real
// repo root and real git binary — this test is skipped, not failed, when
// .git is absent (matches plan.md Task 2: "warn-skip otherwise").
func TestCommitExists_RealRepo(t *testing.T) {
	root := repoRootForTest(t)

	// HEAD always resolves in a real git checkout.
	ok, err := CommitExists(root, "HEAD")
	if err != nil {
		t.Fatalf("CommitExists(HEAD): %v", err)
	}
	if !ok {
		t.Fatalf("CommitExists(HEAD) = false, want true in a real git checkout")
	}

	ok, err = CommitExists(root, "0000000000000000000000000000000000000000")
	if err != nil {
		t.Fatalf("CommitExists(bogus): %v", err)
	}
	if ok {
		t.Fatalf("CommitExists(bogus all-zero hash) = true, want false")
	}
}
