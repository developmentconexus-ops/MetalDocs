package main

import (
	"fmt"
	"sort"
	"strings"

	"metaldocs/internal/platform/httprouter"
)

// assertSurface is the boot-time proof that the mounted HTTP surface is
// exactly the declared one, one publisher per tag, and no publisher mounted
// a route outside its own tag.
//
// mounted is keyed by publisher name, recording the patterns that
// publisher's own Recorder observed — NOT one global pattern set. That
// per-publisher shape is what makes check 4 non-redundant with check 1:
// recording a single global set would let every publisher claim a distinct
// tag while silently mounting each other's operations, and checks 1-3 would
// all still pass. Only per-publisher recording lets check 4 catch
// "documents mounted a templates route" as its own, distinct failure.
//
// It returns one aggregated error naming every failing check across all four
// checks, not just the first, so a boot failure is fixable in one restart.
func assertSurface(
	mounted map[string][]string,
	surface map[string]surfaceRule,
	expectedTags []string,
	publishers []httprouter.SurfacePublisher,
) error {
	var problems []string
	problems = append(problems, checkTagCoverage(expectedTags, publishers)...)

	allMounted := mountedPatternSet(mounted)
	problems = append(problems, checkMountedDeclared(allMounted, surface)...)
	problems = append(problems, checkDeclaredMounted(surface, allMounted)...)
	problems = append(problems, checkPublisherOwnership(mounted, surface, publishers)...)

	if len(problems) == 0 {
		return nil
	}
	sort.Strings(problems)
	return fmt.Errorf("http surface assertion failed:\n  - %s", strings.Join(problems, "\n  - "))
}

// checkTagCoverage is check 1: tag coverage — exactly one publisher per
// expected tag.
func checkTagCoverage(expectedTags []string, publishers []httprouter.SurfacePublisher) []string {
	var problems []string
	tagCount := make(map[string]int)
	tagOwners := make(map[string][]string)
	for _, p := range publishers {
		tagCount[p.Tag()]++
		tagOwners[p.Tag()] = append(tagOwners[p.Tag()], p.Name())
	}
	for _, tag := range expectedTags {
		switch tagCount[tag] {
		case 0:
			problems = append(problems, fmt.Sprintf("tag %q has no publisher", tag))
		case 1:
			// fine
		default:
			owners := append([]string(nil), tagOwners[tag]...)
			sort.Strings(owners)
			problems = append(problems, fmt.Sprintf(
				"tag %q is claimed by more than one publisher: %s",
				tag, strings.Join(owners, ", "),
			))
		}
	}
	return problems
}

// mountedPatternSet builds the union of every pattern actually mounted,
// across publishers.
func mountedPatternSet(mounted map[string][]string) map[string]bool {
	allMounted := make(map[string]bool)
	for _, patterns := range mounted {
		for _, pattern := range patterns {
			allMounted[pattern] = true
		}
	}
	return allMounted
}

// checkMountedDeclared is check 2: mounted ⊆ declared — every mounted
// pattern has a surface rule.
func checkMountedDeclared(allMounted map[string]bool, surface map[string]surfaceRule) []string {
	var problems []string
	for pattern := range allMounted {
		if _, ok := surface[pattern]; !ok {
			problems = append(problems, fmt.Sprintf(
				"pattern %q is mounted but not declared in the surface", pattern,
			))
		}
	}
	return problems
}

// checkDeclaredMounted is check 3: declared ⊆ mounted — every surface rule
// has a mounted handler.
func checkDeclaredMounted(surface map[string]surfaceRule, allMounted map[string]bool) []string {
	var problems []string
	for pattern := range surface {
		if !allMounted[pattern] {
			problems = append(problems, fmt.Sprintf(
				"pattern %q is declared in the surface but was never mounted", pattern,
			))
		}
	}
	return problems
}

// checkPublisherOwnership is check 4: ownership — a publisher may only mount
// patterns whose surface tag matches its own claimed tag. Evaluated
// per-publisher, against that publisher's own recorded patterns, so it is a
// genuinely distinct check from check 1 (tag claims) and cannot be satisfied
// by accident when checks 1-3 pass.
func checkPublisherOwnership(mounted map[string][]string, surface map[string]surfaceRule, publishers []httprouter.SurfacePublisher) []string {
	var problems []string
	publisherTag := make(map[string]string)
	for _, p := range publishers {
		publisherTag[p.Name()] = p.Tag()
	}
	for publisherName, patterns := range mounted {
		claimedTag, known := publisherTag[publisherName]
		if !known {
			// Not a registered publisher at all; check 1 already reports
			// missing/duplicate tag coverage, so skip here to avoid noise.
			continue
		}
		for _, pattern := range patterns {
			rule, ok := surface[pattern]
			if !ok {
				// Already reported by check 2.
				continue
			}
			if rule.tag != claimedTag {
				problems = append(problems, fmt.Sprintf(
					"publisher %q (tag %q) mounted pattern %q, which belongs to tag %q",
					publisherName, claimedTag, pattern, rule.tag,
				))
			}
		}
	}
	return problems
}

// mergedSurface merges b into a, returning a new map. It panics on a
// colliding key: generator rule 6 catches collisions at build time, so
// reaching this panic means the generator was bypassed.
func mergedSurface(a, b map[string]surfaceRule) map[string]surfaceRule {
	out := make(map[string]surfaceRule, len(a)+len(b))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if _, exists := out[k]; exists {
			panic(fmt.Sprintf("mergedSurface: colliding pattern %q; the generator should have caught this at build time", k))
		}
		out[k] = v
	}
	return out
}

// union returns the sorted union of a and b, without duplicates.
func union(a, b []string) []string {
	set := make(map[string]struct{}, len(a)+len(b))
	for _, s := range a {
		set[s] = struct{}{}
	}
	for _, s := range b {
		set[s] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for s := range set {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}
