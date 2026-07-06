package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Class is the RFC 2119 classification carried in a REQ line's trailing
// parenthetical, per wiki/architecture/backend-target-architecture.md's
// documented convention: `- **REQ-<AREA>-<n>** text (MUST[ — annotation])`.
type Class string

const (
	ClassMust   Class = "MUST"
	ClassShould Class = "SHOULD"
	ClassMay    Class = "MAY"
)

// Req is one extracted requirement line.
type Req struct {
	ID string
	// Class is the RFC 2119 classification (MUST/SHOULD/MAY).
	Class Class
	// Annotation is the free text following the classification keyword and an
	// em-dash inside the same parenthetical, e.g. "satisfied Wave 1 (F-01): ...".
	// Empty when the REQ line carries a bare classification with no annotation.
	Annotation string
	// Line is the 1-based line number of the line that introduced this ID
	// (the first normative mention — duplicates are collapsed onto it).
	Line int
}

// reqBulletRe matches the start of a REQ bullet: `- **REQ-AREA-n** rest...`.
var reqBulletRe = regexp.MustCompile(`^\s*-\s+\*\*(REQ-[A-Z]+-[0-9]+)\*\*`)

// classStartRe finds the opening paren of a classification parenthetical.
var classStartRe = regexp.MustCompile(`\((MUST|SHOULD|MAY)\b`)

// ExtractReqs parses a governing architecture doc and returns one Req per
// unique REQ ID, in first-normative-mention order.
//
// Extraction works in two passes:
//  1. Split the file into "blocks" — a block starts at a line matching
//     reqBulletRe or a plain non-bullet line, and a REQ bullet's block
//     continues accumulating subsequent lines only while it is still a
//     "continuation" (does not itself start a new bullet and is not blank),
//     which lets a classification annotation wrap onto following physical
//     lines (e.g. the doc's REQ-OBS-3 line) without swallowing the next
//     REQ bullet or an unrelated non-REQ bullet.
//  2. For each REQ block, extract the ID and the classification parenthetical
//     (balanced-paren scan, so a nested "(F-01)" inside the annotation does
//     not truncate it), keeping only the first mention of each ID.
func ExtractReqs(path string) ([]Req, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan %s: %w", path, err)
	}

	var reqs []Req
	seen := map[string]bool{}

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		m := reqBulletRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		id := m[1]
		startLine := i + 1

		// Accumulate continuation lines: a bare-text line (not a new bullet,
		// not blank) that follows, up until the classification paren closes
		// or a boundary (new bullet / blank line / EOF) is hit.
		acc := line
		j := i + 1
		for !hasBalancedClassParen(acc) && j < len(lines) {
			next := lines[j]
			if reqBulletRe.MatchString(next) || strings.TrimSpace(next) == "" {
				break
			}
			acc += "\n" + next
			j++
		}
		i = j - 1 // resume outer loop after the consumed continuation lines

		if seen[id] {
			continue
		}
		seen[id] = true

		class, annotation := parseClassification(acc)
		reqs = append(reqs, Req{
			ID:         id,
			Class:      class,
			Annotation: strings.TrimSpace(annotation),
			Line:       startLine,
		})
	}
	return reqs, nil
}

// hasBalancedClassParen reports whether acc contains a classification
// parenthetical whose parens are balanced (i.e. any nested annotation parens
// have fully closed rather than being cut off mid-wrap).
func hasBalancedClassParen(acc string) bool {
	start := classStartRe.FindStringIndex(acc)
	if start == nil {
		return false
	}
	depth := 0
	for _, r := range acc[start[0]:] {
		switch r {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return true
			}
		}
	}
	return false
}

// parseClassification extracts the RFC 2119 keyword and optional annotation
// text from the accumulated bullet text, honoring balanced parens so a nested
// `(F-01)` reference inside an annotation does not truncate it.
func parseClassification(acc string) (Class, string) {
	loc := classStartRe.FindStringIndex(acc)
	if loc == nil {
		return "", ""
	}
	start := loc[0]
	depth := 0
	end := -1
	for i, r := range acc[start:] {
		switch r {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				end = start + i
			}
		}
		if end >= 0 {
			break
		}
	}
	if end < 0 {
		return "", ""
	}
	inner := acc[start+1 : end] // strip outer parens
	kw := classStartRe.FindStringSubmatch(acc[start:])
	var class Class
	if len(kw) > 1 {
		class = Class(kw[1])
	}
	rest := strings.TrimPrefix(inner, string(class))
	rest = strings.TrimSpace(rest)
	rest = strings.TrimPrefix(rest, "—")
	return class, strings.TrimSpace(rest)
}
