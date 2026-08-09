package main

import (
	"strings"
	"testing"
)

// ---- filterByChanged: the pure intersection --------------------------------

// A check named by --only whose declared Paths do not intersect the diff
// must be dropped. This is the whole point of composing --changed with
// --only: "run this specific check, but only if the diff actually touches
// it".
func TestFilterByChangedExcludesNonMatchingCheck(t *testing.T) {
	selected := []Check{{ID: "scoped-check", Paths: []string{"frontend/"}}}
	changed := []string{"internal/foo/bar.go"}

	got := filterByChanged(selected, changed)
	if len(got) != 0 {
		t.Fatalf("want the check excluded — its Paths don't intersect the diff — got %+v", got)
	}
}

// A check whose declared Paths do intersect the diff survives.
func TestFilterByChangedKeepsMatchingCheck(t *testing.T) {
	selected := []Check{{ID: "scoped-check", Paths: []string{"frontend/"}}}
	changed := []string{"frontend/apps/web/x.tsx"}

	got := filterByChanged(selected, changed)
	if len(got) != 1 {
		t.Fatalf("want the check included — diff touches its declared path — got %+v", got)
	}
}

// A check with no declared Paths is repo-scoped and always runs — the
// fail-closed default matchesPaths already guarantees must survive
// composition with --only too.
func TestFilterByChangedKeepsPathlessCheck(t *testing.T) {
	selected := []Check{{ID: "repo-scoped-check"}}
	changed := []string{"some/unrelated/file.txt"}

	got := filterByChanged(selected, changed)
	if len(got) != 1 {
		t.Fatalf("a check with no declared Paths must survive diff-scoping, got %+v", got)
	}
}

// ---- runningAsNonPRCI: the fallback trigger --------------------------------

func TestRunningAsNonPRCI(t *testing.T) {
	cases := []struct {
		event string
		want  bool
	}{
		{"", false}, // unset: not GitHub Actions at all (a laptop) — must NOT trigger the fallback
		{"push", true},
		{"schedule", true},
		{"workflow_dispatch", true},
		{"pull_request", false},
		{"pull_request_target", false},
	}
	for _, c := range cases {
		t.Setenv("GITHUB_EVENT_NAME", c.event)
		if got := runningAsNonPRCI(); got != c.want {
			t.Errorf("GITHUB_EVENT_NAME=%q: runningAsNonPRCI() = %v, want %v", c.event, got, c.want)
		}
	}
}

// ---- scopeToChanged: the fallback itself -----------------------------------

// Outside a pull request, scoping must return the input set unfiltered and
// must never even ask git for a diff — proven here by handing it a --base
// that would fail refPattern validation if changedFiles were called.
func TestScopeToChangedFallsBackOutsidePR(t *testing.T) {
	t.Setenv("GITHUB_EVENT_NAME", "push")
	selected := []Check{{ID: "a"}, {ID: "b", Paths: []string{"frontend/"}}}

	got, err := scopeToChanged(selected, "not a valid ref!!")
	if err != nil {
		t.Fatalf("fallback must not need a valid --base, got error: %v", err)
	}
	if len(got) != len(selected) {
		t.Fatalf("want the full input set on non-PR fallback, got %+v", got)
	}
}

// Inside a pull request, scoping is real: an invalid --base must surface as
// an error rather than being silently absorbed by the fallback.
func TestScopeToChangedRunsDiffOnPR(t *testing.T) {
	t.Setenv("GITHUB_EVENT_NAME", "pull_request")
	_, err := scopeToChanged([]Check{{ID: "a"}}, "not a valid ref!!")
	if err == nil {
		t.Fatal("want an error: an invalid --base must be validated when scoping actually runs")
	}
}

// ---- emptySelectionMessage: the loud zero ----------------------------------

func TestEmptySelectionMessageDistinguishesScoped(t *testing.T) {
	scoped := emptySelectionMessage(true, "origin/main")
	unscoped := emptySelectionMessage(false, "origin/main")

	if scoped == unscoped {
		t.Fatal("a diff-scoped empty selection must read differently from an ordinary empty selection")
	}
	if !strings.Contains(scoped, "origin/main") {
		t.Errorf("scoped message should name the base ref, got %q", scoped)
	}
}

// ---- selectChecks: the composed flag surface -------------------------------

// The scenario the operator ruling calls out by name: --only narrowed
// further by --changed.
func TestSelectChecksComposesOnlyWithChanged(t *testing.T) {
	t.Setenv("GITHUB_EVENT_NAME", "pull_request")
	// HEAD vs the working tree: this checkout is clean, so the diff is empty
	// and a path-scoped check must be dropped.
	selected, scoped, err := selectChecks(ProfileFast, "css-token-discipline", "HEAD", true)
	if err != nil {
		t.Fatalf("selectChecks: %v", err)
	}
	if !scoped {
		t.Fatal("want scoped=true when --changed is passed")
	}
	if len(selected) != 0 {
		t.Fatalf("want 0 checks (empty diff, path-scoped check) — got %+v", selected)
	}
}

// A pathless check named by --only survives --changed even over an empty
// diff — same fail-closed default as the unscoped case.
func TestSelectChecksKeepsPathlessCheckUnderChanged(t *testing.T) {
	t.Setenv("GITHUB_EVENT_NAME", "pull_request")
	selected, scoped, err := selectChecks(ProfileFast, "go-vet", "HEAD", true)
	if err != nil {
		t.Fatalf("selectChecks: %v", err)
	}
	if !scoped {
		t.Fatal("want scoped=true when --changed is passed")
	}
	if len(selected) != 1 {
		t.Fatalf("want the pathless check to survive an empty diff, got %+v", selected)
	}
}

// --profile=changed must keep working exactly as documented, with no
// --changed flag needed: it resolves to the `pr` check set and implies
// scoping (scoped=true) on its own.
//
// The prior version of this test called selectChecks(ProfileChanged, "",
// "HEAD", false) on a `pull_request` event, which resolves to
// `git diff --name-only HEAD` — a diff against the WORKING TREE, not a
// commit — then asserted no selected check had declared Paths (i.e. the
// diff was empty). That asserts the developer's tree is clean: it fails for
// any uncommitted change to a Paths-scoped file, which is exactly the state
// a developer is in while running `go run ./tools/verify`. It was asserting
// mutable external state, not the behaviour --profile=changed actually
// promises.
//
// filterByChanged (main.go:268) is the pure half of that behaviour and is
// already covered directly, above (TestFilterByChanged*). What was missing
// is proof that --profile=changed reaches that pure function via the same
// composition as `pr` + --changed, with no flag needed. This test proves
// that half deterministically, by taking the non-PR fallback branch of
// scopeToChanged (TestScopeToChangedFallsBackOutsidePR below exercises the
// same branch): outside a pull request the diff step is skipped entirely
// and the input selection passes through unfiltered, so the only things
// left to assert are exactly the two properties --profile=changed must
// have regardless of git state — it resolves to the `pr` set, and it
// implies scoped=true — without ever reading the working tree.
func TestProfileChangedResolvesToPRSetAndImpliesScoping(t *testing.T) {
	t.Setenv("GITHUB_EVENT_NAME", "push")
	selected, scoped, err := selectChecks(ProfileChanged, "", "irrelevant-on-non-pr-fallback", false)
	if err != nil {
		t.Fatalf("selectChecks: %v", err)
	}
	if !scoped {
		t.Fatal("profile=changed must imply scoping even without --changed")
	}
	want := selectByProfile(ProfilePR)
	if len(selected) != len(want) {
		t.Fatalf("profile=changed must resolve to the `pr` check set before diff-scoping is applied, got %d checks, want %d (the `pr` set)", len(selected), len(want))
	}
	for i, c := range selected {
		if c.ID != want[i].ID {
			t.Fatalf("profile=changed's resolved set diverges from `pr` at index %d: got %s, want %s", i, c.ID, want[i].ID)
		}
	}
}

// Requirement 2: a push straight to main (no PR) must fall back to the full
// unscoped set, never to an empty one — that is the exact failure this unit
// exists to close.
func TestSelectChecksNonPRFallsBackToFullSet(t *testing.T) {
	t.Setenv("GITHUB_EVENT_NAME", "push")
	selected, scoped, err := selectChecks(ProfileFast, "", "not-a-real-ref!!", true)
	if err != nil {
		t.Fatalf("fallback must not need a valid --base: %v", err)
	}
	if !scoped {
		t.Fatal("want scoped=true (a --changed request was made)")
	}
	full := selectByProfile(ProfileFast)
	if len(selected) != len(full) {
		t.Fatalf("want the full fast set on non-PR fallback: got %d checks, want %d", len(selected), len(full))
	}
}
