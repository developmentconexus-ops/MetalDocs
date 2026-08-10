package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

// Sharding a check across CI runners.
//
// One check in this registry dominates PR wall-clock: go-test-integration is
// ~11 minutes of the ~17 a PR takes, and it is a single `go test` invocation
// on a single runner. Splitting it is the one lever that returns minutes
// rather than seconds — but only if the split cannot lie about coverage.
//
// The failure mode a naive split has, and this one does not: a hand-written
// list of packages per CI job. That is a second definition of "the integration
// suite", it drifts the moment a package is added, and its drift is SILENT —
// CI stays green over a package nobody runs. This repo has a name for that
// shape and a register full of instances (docs/engineering/
// mechanical-enforcement-register.md).
//
// So the subject is never enumerated by hand. It is `go list` over the same
// roots the unsharded Argv already names, resolved at run time, on the commit
// under test. A new package joins the suite by existing. What the weights file
// below influences is only WHICH shard a package lands in — never whether it
// runs at all. Deleting the weights file makes the shards lumpy; it cannot
// make a package disappear. TestEveryPackageLandsInExactlyOneShard pins that.

// Partition declares that a check's subject is a set of Go test packages which
// may be split across CI runners. A check without one runs whole, everywhere.
type Partition struct {
	// Roots are `go list` patterns naming the subject. These are the same
	// patterns the check's own Argv would carry unsharded — stated once, here,
	// so the listed set and the run set cannot diverge.
	Roots []string

	// Tags is the build-tag set the packages are listed under. It must match
	// the -tags the Argv passes, or `go list` reports a different universe
	// than `go test` runs: integration tests live behind `//go:build
	// integration`, and listing without the tag hides every one of them.
	Tags string
}

// packagesPlaceholder is where a partitioned check's Argv receives its package
// list. It is not part of expandArgv's closed token set because it expands to
// MANY argv elements rather than one string, which is exactly why it is
// handled by its own pass instead of being bolted onto that one.
const packagesPlaceholder = "{{packages}}"

// testWeights maps import path -> measured seconds, and is advisory only: it
// balances the shards and nothing else. It is embedded rather than read from
// disk so a check running against a staged tree sees the same weights as one
// running against the repo.
//
// Regenerate with tools/verify/testweights.md's one-liner after a real run.
// A stale entry costs a lumpy shard. A missing entry costs nothing but
// precision — an unlisted package is weighted at the median of the listed
// ones, so a brand-new heavy package is not assumed free.
//
// The current values come from CI run 31412988724 and are worth reading once:
// apps/api/cmd/metaldocs-api is 439.9s of 864.8s. No package-level splitter
// can produce a shard faster than its slowest package, so that one entry is
// the floor of this whole scheme, and adding shards past four buys nothing.
//
//go:embed testweights.json
var testWeightsJSON []byte

func testWeights() map[string]float64 {
	var w map[string]float64
	if err := json.Unmarshal(testWeightsJSON, &w); err != nil {
		// Advisory data: a malformed file must not stop the suite from
		// running. Equal weights are a worse split, not a wrong one.
		return map[string]float64{}
	}
	return w
}

// shardSpec is a parsed --shard=i/n.
type shardSpec struct {
	index int // 1-based
	total int
}

// parseShard reads "i/n". Empty means "no sharding", which is the local
// default and what every unsharded caller gets.
func parseShard(s string) (*shardSpec, error) {
	if s == "" {
		return nil, nil
	}
	i, n, ok := strings.Cut(s, "/")
	if !ok {
		return nil, fmt.Errorf("--shard=%q: want i/n, e.g. 2/4", s)
	}
	index, err := strconv.Atoi(strings.TrimSpace(i))
	if err != nil {
		return nil, fmt.Errorf("--shard=%q: %q is not a number", s, i)
	}
	total, err := strconv.Atoi(strings.TrimSpace(n))
	if err != nil {
		return nil, fmt.Errorf("--shard=%q: %q is not a number", s, n)
	}
	if total < 1 {
		return nil, fmt.Errorf("--shard=%q: shard count must be at least 1", s)
	}
	if index < 1 || index > total {
		return nil, fmt.Errorf("--shard=%q: shard index must be within 1..%d", s, total)
	}
	return &shardSpec{index: index, total: total}, nil
}

// applyPartition resolves packagesPlaceholder in every selected check that
// declares a Partition, and returns the selection with those Argvs rewritten.
//
// spec == nil expands to the whole package set, so a local `verify --only=
// go-test-integration` runs exactly what the four CI shards run between them.
// There is no separate unsharded code path to keep in sync.
//
// A check WITHOUT a Partition is untouched, including by --shard: sharding is
// a property a check declares, not a flag that reshapes arbitrary commands.
func applyPartition(selected []Check, spec *shardSpec) ([]Check, error) {
	out := make([]Check, len(selected))
	copy(out, selected)
	for i, c := range out {
		if c.Partition == nil {
			if idx := indexOf(c.Argv, packagesPlaceholder); idx >= 0 {
				return nil, fmt.Errorf("check %q carries %s but declares no Partition", c.ID, packagesPlaceholder)
			}
			continue
		}
		idx := indexOf(c.Argv, packagesPlaceholder)
		if idx < 0 {
			return nil, fmt.Errorf("check %q declares a Partition but its Argv never spends %s", c.ID, packagesPlaceholder)
		}
		pkgs, err := listTestPackages(*c.Partition)
		if err != nil {
			return nil, fmt.Errorf("check %q: %w", c.ID, err)
		}
		if len(pkgs) == 0 {
			return nil, fmt.Errorf("check %q: `go list` over %v found no test packages; the suite would report green over nothing", c.ID, c.Partition.Roots)
		}
		mine := pkgs
		if spec != nil {
			mine = shardOf(pkgs, testWeights(), spec.index, spec.total)
			if len(mine) == 0 {
				// Not an error: with more shards than packages some shard is
				// empty by arithmetic. Say so loudly rather than exiting 0
				// over nothing with no explanation in the log.
				return nil, fmt.Errorf("check %q: shard %d/%d is empty — %d packages cannot fill %d shards; lower the shard count", c.ID, spec.index, spec.total, len(pkgs), spec.total)
			}
		}
		argv := make([]string, 0, len(c.Argv)+len(mine))
		argv = append(argv, c.Argv[:idx]...)
		argv = append(argv, mine...)
		argv = append(argv, c.Argv[idx+1:]...)
		out[i].Argv = argv
	}
	return out, nil
}

func indexOf(ss []string, want string) int {
	for i, s := range ss {
		if s == want {
			return i
		}
	}
	return -1
}

// listTestPackages asks the toolchain which packages under Roots actually
// carry tests. Packages with no test files are dropped here rather than
// handed to `go test` — they contribute "no test files" noise to every shard
// and would skew an otherwise even split by count.
func listTestPackages(p Partition) ([]string, error) {
	// Roots are repo-relative (./tests/...), so this must run at the repo
	// root and not merely wherever the process happens to be. main() chdirs
	// there before anything else; a test binary does not, and a `go list`
	// resolved against tools/verify silently reports a DIFFERENT universe
	// rather than failing — which would be a shard scheme that runs the wrong
	// set and says nothing.
	root, err := repoRoot()
	if err != nil {
		return nil, fmt.Errorf("cannot locate repo root to list %v: %w", p.Roots, err)
	}
	argv := []string{"go", "list"}
	if p.Tags != "" {
		argv = append(argv, "-tags", p.Tags)
	}
	argv = append(argv, "-f", "{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}")
	argv = append(argv, p.Roots...)
	out, err := command(context.Background(), root, argv).Output()
	if err != nil {
		// go list writes the reason (a broken import, a bad tag) to stderr and
		// nothing to stdout; without this the caller sees only "exit status 1".
		var ee *exec.ExitError
		if errors.As(err, &ee) && len(ee.Stderr) > 0 {
			return nil, fmt.Errorf("go list over %v: %w: %s", p.Roots, err, strings.TrimSpace(string(ee.Stderr)))
		}
		return nil, fmt.Errorf("go list over %v: %w", p.Roots, err)
	}
	var pkgs []string
	for _, line := range strings.Split(string(out), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			pkgs = append(pkgs, line)
		}
	}
	sort.Strings(pkgs) // `go list` is already ordered; sorting makes that a guarantee rather than an observation.
	return pkgs, nil
}

// shardOf assigns packages to shards by longest-processing-time-first: sort
// heaviest to lightest, drop each into whichever shard is currently lightest.
// It is the standard greedy makespan heuristic, and it is deterministic —
// same package set plus same weights gives the same assignment on every host,
// which is what lets a failing shard be reproduced locally by number.
//
// Round-robin would be simpler and wrong: this suite's cost is dominated by a
// handful of packages, and round-robin over a sorted list happily puts the two
// heaviest in the same shard.
func shardOf(pkgs []string, weights map[string]float64, index, total int) []string {
	type bin struct {
		load float64
		pkgs []string
	}
	bins := make([]bin, total)

	def := medianWeight(weights)
	type wp struct {
		pkg string
		w   float64
	}
	items := make([]wp, len(pkgs))
	for i, p := range pkgs {
		w, ok := weights[p]
		if !ok {
			w = def
		}
		items[i] = wp{pkg: p, w: w}
	}
	// Heaviest first; ties broken by import path so the order is total.
	sort.Slice(items, func(a, b int) bool {
		if items[a].w != items[b].w {
			return items[a].w > items[b].w
		}
		return items[a].pkg < items[b].pkg
	})

	for _, it := range items {
		lightest := 0
		for j := 1; j < total; j++ {
			if bins[j].load < bins[lightest].load {
				lightest = j
			}
		}
		bins[lightest].load += it.w
		bins[lightest].pkgs = append(bins[lightest].pkgs, it.pkg)
	}

	got := bins[index-1].pkgs
	sort.Strings(got) // `go test` gets them in a stable, readable order.
	return got
}

// medianWeight is what an unlisted package is assumed to cost. Median rather
// than mean because this distribution has a long tail: the mean is dragged up
// by the few heavy packages and would over-weight every new small one. With
// no weights at all it is 1, which degenerates to balancing by package count.
func medianWeight(weights map[string]float64) float64 {
	if len(weights) == 0 {
		return 1
	}
	ws := make([]float64, 0, len(weights))
	for _, w := range weights {
		ws = append(ws, w)
	}
	sort.Float64s(ws)
	return ws[len(ws)/2]
}
