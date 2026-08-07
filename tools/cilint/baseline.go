// Baseline-and-ratchet mechanism for cilint.
//
// The analyzers in internal/analyzers are never weakened by this file. What
// this file does is let CI go green on the *known* findings that exist today
// (recorded in baseline.json) while still failing hard on anything new. The
// ratchet fires in both directions: a finding count going UP is a new
// regression (fail), and a finding count going DOWN without the baseline
// being tightened is also a fail — a red build is what forces someone to run
// --update-baseline and shrink the recorded debt. See baseline.json's "_doc"
// field for the transitional framing (CLAUDE.md: an unlabelled local maximum
// is a defect).
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"metaldocs/tools/cilint/internal/analyzers"
)

// baselineDoc is the mandatory transitional label written into every
// generated baseline.json. It must survive every --update-baseline run
// verbatim (only entries change) so nobody mistakes the file for a
// permission slip.
const baselineDoc = "This file is TRANSITIONAL. It is a record of known lint " +
	"debt as of the last --update-baseline run, not a permission slip to add " +
	"more of it — new violations still fail CI. The global-maximum end state " +
	"is an empty entries array (all analyzers clean with no baseline needed " +
	"at all). The baseline may only ever shrink: every entry must eventually " +
	"be remediated and removed, never grown to accommodate new debt. If you " +
	"are here to add an entry for a NEW finding, stop — fix the finding " +
	"instead."

// BaselineEntry records one known, accepted lint finding.
type BaselineEntry struct {
	Analyzer string `json:"analyzer"`
	File     string `json:"file"`
	Message  string `json:"message"`
	Count    int    `json:"count"`
	Reason   string `json:"reason,omitempty"`
	Owner    string `json:"owner,omitempty"`
}

// Baseline is the on-disk shape of tools/cilint/baseline.json.
type Baseline struct {
	Doc     string          `json:"_doc"`
	Entries []BaselineEntry `json:"entries"`
}

// baselineKey identifies a finding class for baseline purposes. Deliberately
// NOT keyed on line number: line numbers shift on every unrelated edit in
// the same file, which would make the baseline churn constantly and train
// people to blindly regenerate it. File is expected to already be
// slash-normalized by the caller.
func baselineKey(analyzer, file, message string) string {
	return analyzer + "\x00" + file + "\x00" + message
}

// normalizeFindings returns a copy of findings with File normalized to
// forward slashes. cilint prints backslash-separated paths on Windows and
// forward-slash paths in CI (Linux); without normalizing at the single point
// findings are read, the baseline matches locally and fails in CI (or vice
// versa).
func normalizeFindings(findings []analyzers.Finding) []analyzers.Finding {
	out := make([]analyzers.Finding, len(findings))
	for i, f := range findings {
		f.File = filepath.ToSlash(f.File)
		out[i] = f
	}
	return out
}

// countByKey aggregates findings into per-key counts, keeping one sample
// Finding per key for reporting (file/line/message/analyzer).
func countByKey(findings []analyzers.Finding) (counts map[string]int, sample map[string]analyzers.Finding) {
	counts = map[string]int{}
	sample = map[string]analyzers.Finding{}
	for _, f := range findings {
		k := baselineKey(f.Analyzer, f.File, f.Message)
		counts[k]++
		if _, ok := sample[k]; !ok {
			sample[k] = f
		}
	}
	return counts, sample
}

// loadBaseline reads baseline.json from path. A missing file is not an
// error: it returns (nil, nil), which callers treat as "no baseline
// configured" — behave exactly as if the ratchet did not exist, i.e. any
// finding fails.
func loadBaseline(path string) (*Baseline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var bl Baseline
	if err := json.Unmarshal(data, &bl); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	for i := range bl.Entries {
		bl.Entries[i].File = filepath.ToSlash(bl.Entries[i].File)
	}
	return &bl, nil
}

// writeBaseline writes bl to path as indented JSON.
func writeBaseline(path string, bl *Baseline) error {
	data, err := json.MarshalIndent(bl, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

// buildBaseline generates a fresh Baseline from the current findings. When
// existing is non-nil, reason/owner annotations are carried over for keys
// that already had them — --update-baseline must not silently wipe the
// human-authored justification when a count merely shifts.
func buildBaseline(findings []analyzers.Finding, existing *Baseline) *Baseline {
	type annotation struct{ reason, owner string }
	carried := map[string]annotation{}
	if existing != nil {
		for _, e := range existing.Entries {
			carried[baselineKey(e.Analyzer, e.File, e.Message)] = annotation{e.Reason, e.Owner}
		}
	}

	counts, sample := countByKey(findings)
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	entries := make([]BaselineEntry, 0, len(keys))
	for _, k := range keys {
		f := sample[k]
		ann := carried[k]
		entries = append(entries, BaselineEntry{
			Analyzer: f.Analyzer,
			File:     f.File,
			Message:  f.Message,
			Count:    counts[k],
			Reason:   ann.reason,
			Owner:    ann.owner,
		})
	}

	return &Baseline{Doc: baselineDoc, Entries: entries}
}

// compareBaseline checks the current findings against a loaded baseline.
// It returns ok=false with human-readable diagnostic lines whenever:
//   - a finding key is not recorded in the baseline at all (new violation);
//   - a recorded key's actual count exceeds the baseline count (regression);
//   - a recorded key's actual count is below the baseline count, including
//     zero (stale baseline — debt was fixed but not ratcheted down).
//
// Lines are sorted for deterministic output.
func compareBaseline(findings []analyzers.Finding, bl *Baseline) (ok bool, lines []string) {
	ok = true
	actualCounts, actualSample := countByKey(findings)

	baselineByKey := map[string]BaselineEntry{}
	for _, e := range bl.Entries {
		baselineByKey[baselineKey(e.Analyzer, e.File, e.Message)] = e
	}

	for k, cnt := range actualCounts {
		if _, known := baselineByKey[k]; known {
			continue
		}
		ok = false
		f := actualSample[k]
		_ = cnt
		lines = append(lines, fmt.Sprintf(
			"NEW VIOLATION (not in baseline): %s:%d: [%s] %s",
			f.File, f.Line, f.Analyzer, f.Message))
	}

	for k, e := range baselineByKey {
		actual := actualCounts[k]
		switch {
		case actual > e.Count:
			ok = false
			lines = append(lines, fmt.Sprintf(
				"COUNT INCREASED: [%s] %s: %s (baseline=%d actual=%d)",
				e.Analyzer, e.File, e.Message, e.Count, actual))
		case actual < e.Count:
			ok = false
			lines = append(lines, fmt.Sprintf(
				"STALE BASELINE: [%s] %s: %s (baseline=%d actual=%d) — "+
					"baseline is stale: violations were fixed but the baseline was not "+
					"tightened. Run `go run ./tools/cilint --update-baseline` and commit.",
				e.Analyzer, e.File, e.Message, e.Count, actual))
		}
	}

	sort.Strings(lines)
	return ok, lines
}
