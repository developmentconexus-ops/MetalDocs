package main

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// The property that matters about sharding is not that it is fast. It is that
// splitting the suite cannot lose a package: a shard scheme that silently
// drops one runs a smaller suite and reports the same green. Every test below
// is aimed at that, not at balance.

func TestEveryPackageLandsInExactlyOneShard(t *testing.T) {
	pkgs := []string{
		"metaldocs/apps/api", "metaldocs/internal/a", "metaldocs/internal/b",
		"metaldocs/internal/c", "metaldocs/tests/d", "metaldocs/tests/e",
		"metaldocs/tests/f",
	}
	weights := map[string]float64{
		"metaldocs/tests/d": 400, "metaldocs/tests/e": 90, "metaldocs/internal/a": 5,
	}
	for total := 1; total <= len(pkgs); total++ {
		seen := map[string]int{}
		for i := 1; i <= total; i++ {
			for _, p := range shardOf(pkgs, weights, i, total) {
				seen[p]++
			}
		}
		for _, p := range pkgs {
			switch seen[p] {
			case 1: // exactly what we want
			case 0:
				t.Errorf("total=%d: %s is in no shard — it would never run, and CI would be green anyway", total, p)
			default:
				t.Errorf("total=%d: %s is in %d shards — it runs %d times", total, p, seen[p], seen[p])
			}
		}
		if len(seen) != len(pkgs) {
			t.Errorf("total=%d: the shards between them cover %d packages, the suite has %d", total, len(seen), len(pkgs))
		}
	}
}

func TestShardAssignmentIsDeterministic(t *testing.T) {
	pkgs := []string{"z/a", "z/b", "z/c", "z/d", "z/e"}
	weights := map[string]float64{"z/a": 10, "z/c": 10}
	for i := 1; i <= 3; i++ {
		first := shardOf(pkgs, weights, i, 3)
		for again := 0; again < 5; again++ {
			if got := shardOf(pkgs, weights, i, 3); !reflect.DeepEqual(first, got) {
				t.Fatalf("shard %d/3 assigned %v then %v: a shard that moves cannot be reproduced locally by number", i, first, got)
			}
		}
	}
}

// Equal weights must not degenerate into "everything in shard 1".
func TestShardsSplitEvenlyWhenNothingIsWeighted(t *testing.T) {
	pkgs := make([]string, 12)
	for i := range pkgs {
		pkgs[i] = fmt.Sprintf("z/p%02d", i)
	}
	for i := 1; i <= 4; i++ {
		if n := len(shardOf(pkgs, map[string]float64{}, i, 4)); n != 3 {
			t.Errorf("shard %d/4 of 12 unweighted packages holds %d, want 3", i, n)
		}
	}
}

// The two heaviest packages must not share a shard while a light shard idles —
// that is the whole reason this is bin-packing rather than round-robin.
func TestTheHeaviestPackagesAreSpreadAcrossShards(t *testing.T) {
	pkgs := []string{"z/heavy1", "z/heavy2", "z/light1", "z/light2"}
	weights := map[string]float64{"z/heavy1": 300, "z/heavy2": 300, "z/light1": 1, "z/light2": 1}
	var withHeavy1, withHeavy2 int
	for i := 1; i <= 2; i++ {
		got := shardOf(pkgs, weights, i, 2)
		for _, p := range got {
			if p == "z/heavy1" {
				withHeavy1 = i
			}
			if p == "z/heavy2" {
				withHeavy2 = i
			}
		}
	}
	if withHeavy1 == withHeavy2 {
		t.Errorf("both 300s landed in shard %d: that shard is the run's wall clock and the other is idle", withHeavy1)
	}
}

// A package nobody has measured is the NEW one, and a new package is the one
// most likely to be slow. Weighting it at zero would pile every new package
// into the same shard.
func TestAnUnmeasuredPackageIsNotAssumedFree(t *testing.T) {
	weights := map[string]float64{"a": 100, "b": 100, "c": 100}
	if got := medianWeight(weights); got != 100 {
		t.Errorf("an unlisted package is weighted %v among three 100s; want the median 100", got)
	}
	if got := medianWeight(map[string]float64{}); got != 1 {
		t.Errorf("with no weights at all the default is %v; want 1, which balances by package count", got)
	}
}

// The embedded weights file must parse. It is advisory, so a broken one must
// not stop the suite — but it should not be broken silently either.
func TestEmbeddedWeightsParse(t *testing.T) {
	if len(testWeightsJSON) == 0 {
		t.Fatal("testweights.json is empty; go:embed needs the file to exist even when there are no weights yet")
	}
	if testWeights() == nil {
		t.Error("testWeights() returned nil")
	}
}

func TestParseShard(t *testing.T) {
	for _, tc := range []struct {
		in      string
		wantErr bool
		index   int
		total   int
	}{
		{in: "", index: 0, total: 0},
		{in: "1/4", index: 1, total: 4},
		{in: "4/4", index: 4, total: 4},
		{in: "1/1", index: 1, total: 1},
		{in: "0/4", wantErr: true},
		{in: "5/4", wantErr: true},
		{in: "1/0", wantErr: true},
		{in: "-1/4", wantErr: true},
		{in: "4", wantErr: true},
		{in: "a/4", wantErr: true},
		{in: "1/b", wantErr: true},
	} {
		got, err := parseShard(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("--shard=%q was accepted; a malformed shard must fail loudly, not run a wrong subset", tc.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("--shard=%q: %v", tc.in, err)
			continue
		}
		if tc.in == "" {
			if got != nil {
				t.Errorf("--shard=%q gave %v; empty means no sharding", tc.in, got)
			}
			continue
		}
		if got.index != tc.index || got.total != tc.total {
			t.Errorf("--shard=%q gave %d/%d, want %d/%d", tc.in, got.index, got.total, tc.index, tc.total)
		}
	}
}

// Registry-level: the placeholder and the Partition are two halves of one
// declaration, and either half alone is a misconfiguration that would
// otherwise surface as a confusing `go test` error at run time.
func TestPartitionAndPlaceholderComeInPairs(t *testing.T) {
	for _, c := range checks {
		n := 0
		for _, a := range c.Argv {
			if a == packagesPlaceholder {
				n++
			}
		}
		switch {
		case c.Partition == nil && n > 0:
			t.Errorf("%s spends %s but declares no Partition", c.ID, packagesPlaceholder)
		case c.Partition != nil && n == 0:
			t.Errorf("%s declares a Partition but never spends %s, so the package list has nowhere to go", c.ID, packagesPlaceholder)
		case c.Partition != nil && n > 1:
			t.Errorf("%s spends %s %d times; the package list would be inserted twice", c.ID, packagesPlaceholder, n)
		}
		if c.Partition == nil {
			continue
		}
		if len(c.Partition.Roots) == 0 {
			t.Errorf("%s declares a Partition with no Roots: `go list` over nothing finds nothing", c.ID)
		}
		// The tag set is what decides whether integration tests are visible at
		// all. If it ever stops matching the -tags in Argv, `go list` reports
		// a different universe than `go test` runs — and the smaller one wins.
		if tag := c.Partition.Tags; tag != "" {
			var argvTag string
			for i, a := range c.Argv {
				if a == "-tags" && i+1 < len(c.Argv) {
					argvTag = c.Argv[i+1]
				}
			}
			if argvTag != tag {
				t.Errorf("%s lists packages under -tags %q but runs them under -tags %q", c.ID, tag, argvTag)
			}
		}
	}
}

// The real thing: resolve go-test-integration's Partition against this
// repository and prove the roots and tags name a universe that actually
// contains the integration suite.
func TestGoTestIntegrationResolvesToRealPackages(t *testing.T) {
	c := checkByID(t, "go-test-integration")
	pkgs, err := listTestPackages(*c.Partition)
	if err != nil {
		t.Fatalf("go list: %v", err)
	}
	if len(pkgs) < 20 {
		t.Fatalf("go list over %v found %d test packages; this suite is much larger than that, so the roots or tags are wrong", c.Partition.Roots, len(pkgs))
	}
	var integration int
	for _, p := range pkgs {
		if strings.Contains(p, "/tests/integration/") {
			integration++
		}
	}
	if integration == 0 {
		t.Errorf("not one tests/integration package was listed: the -tags are not reaching the //go:build integration files")
	}
}

// Falsification of the tag half: without -tags integration the same roots
// resolve to a STRICTLY SMALLER set. If this ever stops being true the tag is
// no longer doing anything and TestGoTestIntegrationResolvesToRealPackages
// would keep passing while the suite silently shrank.
func TestDroppingTheBuildTagShrinksTheUniverse(t *testing.T) {
	c := checkByID(t, "go-test-integration")
	tagged, err := listTestPackages(*c.Partition)
	if err != nil {
		t.Fatalf("go list (tagged): %v", err)
	}
	bare, err := listTestPackages(Partition{Roots: c.Partition.Roots})
	if err != nil {
		t.Fatalf("go list (untagged): %v", err)
	}
	if len(bare) >= len(tagged) {
		t.Errorf("untagged listing found %d packages and tagged found %d: -tags %q is not selecting anything, so it is not protecting anything either",
			len(bare), len(tagged), c.Partition.Tags)
	}
}

// applyPartition with no shard must produce the same package set the four
// shards produce between them — one definition of the suite, not two.
func TestUnshardedRunEqualsTheUnionOfItsShards(t *testing.T) {
	c := checkByID(t, "go-test-integration")

	whole, err := applyPartition([]Check{c}, nil)
	if err != nil {
		t.Fatalf("applyPartition(nil): %v", err)
	}
	wholeSet := argvPackages(whole[0].Argv)

	union := map[string]bool{}
	for i := 1; i <= 4; i++ {
		spec, err := parseShard(fmt.Sprintf("%d/4", i))
		if err != nil {
			t.Fatal(err)
		}
		part, err := applyPartition([]Check{c}, spec)
		if err != nil {
			t.Fatalf("applyPartition(%d/4): %v", i, err)
		}
		for _, p := range argvPackages(part[0].Argv) {
			if union[p] {
				t.Errorf("%s is in more than one shard", p)
			}
			union[p] = true
		}
	}

	if len(union) != len(wholeSet) {
		t.Fatalf("the unsharded run covers %d packages, the four shards cover %d between them", len(wholeSet), len(union))
	}
	for _, p := range wholeSet {
		if !union[p] {
			t.Errorf("%s runs unsharded but no shard claims it: CI would never execute it", p)
		}
	}
}

// A Check that carries neither half is untouched by --shard. Sharding is a
// property a check declares, not a flag that reshapes arbitrary commands.
func TestShardingLeavesUnpartitionedChecksAlone(t *testing.T) {
	c := checkByID(t, "gofmt")
	spec, err := parseShard("2/4")
	if err != nil {
		t.Fatal(err)
	}
	got, err := applyPartition([]Check{c}, spec)
	if err != nil {
		t.Fatalf("applyPartition: %v", err)
	}
	if !reflect.DeepEqual(got[0].Argv, c.Argv) {
		t.Errorf("--shard rewrote an unpartitioned check's argv: %v -> %v", c.Argv, got[0].Argv)
	}
}

func checkByID(t *testing.T, id string) Check {
	t.Helper()
	for _, c := range checks {
		if c.ID == id {
			return c
		}
	}
	t.Fatalf("no check %q in the registry", id)
	return Check{}
}

// argvPackages pulls the import paths back out of an expanded argv: everything
// that looks like a package path rather than a flag.
func argvPackages(argv []string) []string {
	var out []string
	for _, a := range argv {
		if strings.HasPrefix(a, "metaldocs/") {
			out = append(out, a)
		}
	}
	sort.Strings(out)
	return out
}
